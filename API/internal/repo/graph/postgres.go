package graphrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dgraph "lawmap/internal/domain/graph"
	"lawmap/internal/pkg/log"
	"lawmap/internal/repo/db"
)

// PostgresStore implements graph storage using PostgreSQL.
type PostgresStore struct {
	db     *db.DB
	logger *log.Logger
}

// NewPostgresStore creates a new PostgreSQL-backed graph store.
func NewPostgresStore(database *db.DB) *PostgresStore {
	return &PostgresStore{
		db:     database,
		logger: log.Default().WithField("component", "postgres-store"),
	}
}

// UpsertNode creates or updates a node in the database.
func (p *PostgresStore) UpsertNode(ctx context.Context, node *dgraph.Node) error {
	propsJSON, err := json.Marshal(node.Props)
	if err != nil {
		return fmt.Errorf("marshal props: %w", err)
	}

	sourcesJSON, err := json.Marshal(node.Sources)
	if err != nil {
		return fmt.Errorf("marshal sources: %w", err)
	}

	var versionFetchedAt sql.NullString
	var versionEffectiveDate sql.NullString
	var versionHash sql.NullString

	if node.Version != nil {
		if node.Version.FetchedAt != "" {
			versionFetchedAt.String = node.Version.FetchedAt
			versionFetchedAt.Valid = true
		}
		if node.Version.EffectiveDate != "" {
			versionEffectiveDate.String = node.Version.EffectiveDate
			versionEffectiveDate.Valid = true
		}
		if node.Version.Hash != "" {
			versionHash.String = node.Version.Hash
			versionHash.Valid = true
		}
	}

	query := `
		INSERT INTO nodes (id, labels, title, citation, text, props, version_fetched_at, version_effective_date, version_hash, sources)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			labels = EXCLUDED.labels,
			title = EXCLUDED.title,
			citation = EXCLUDED.citation,
			text = EXCLUDED.text,
			props = EXCLUDED.props,
			version_fetched_at = EXCLUDED.version_fetched_at,
			version_effective_date = EXCLUDED.version_effective_date,
			version_hash = EXCLUDED.version_hash,
			sources = EXCLUDED.sources,
			updated_at = NOW()
	`

	_, err = p.db.ExecContext(ctx, query,
		node.ID,
		pgArray(node.Labels),
		node.Title,
		node.Citation,
		node.Text,
		propsJSON,
		versionFetchedAt,
		versionEffectiveDate,
		versionHash,
		sourcesJSON,
	)

	if err != nil {
		return fmt.Errorf("upsert node: %w", err)
	}

	p.logger.Debug("Upserted node", map[string]any{"id": node.ID})
	return nil
}

// GetNode retrieves a node by ID.
func (p *PostgresStore) GetNode(id string) (*dgraph.Node, bool) {
	ctx := context.Background()

	query := `
		SELECT id, labels, title, citation, text, props, version_fetched_at, version_effective_date, version_hash, sources
		FROM nodes
		WHERE id = $1
	`

	var node dgraph.Node
	var labelsArray pgStringArray
	var propsJSON []byte
	var sourcesJSON []byte
	var versionFetchedAt, versionEffectiveDate, versionHash sql.NullString

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID,
		&labelsArray,
		&node.Title,
		&node.Citation,
		&node.Text,
		&propsJSON,
		&versionFetchedAt,
		&versionEffectiveDate,
		&versionHash,
		&sourcesJSON,
	)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		p.logger.Error("Query node failed", map[string]any{"error": err.Error(), "id": id})
		return nil, false
	}

	node.Labels = []string(labelsArray)

	if err := json.Unmarshal(propsJSON, &node.Props); err != nil {
		p.logger.Error("Unmarshal props failed", map[string]any{"error": err.Error()})
		node.Props = make(map[string]any)
	}

	if err := json.Unmarshal(sourcesJSON, &node.Sources); err != nil {
		p.logger.Error("Unmarshal sources failed", map[string]any{"error": err.Error()})
		node.Sources = []dgraph.SourceMeta{}
	}

	if versionFetchedAt.Valid || versionEffectiveDate.Valid || versionHash.Valid {
		node.Version = &dgraph.Version{
			FetchedAt:     versionFetchedAt.String,
			EffectiveDate: versionEffectiveDate.String,
			Hash:          versionHash.String,
		}
	}

	return &node, true
}

// GetChildren retrieves child nodes via PARENT_OF edges.
func (p *PostgresStore) GetChildren(id string) ([]*dgraph.Node, []*dgraph.Edge) {
	ctx := context.Background()

	edgeQuery := `
		SELECT id, edge_type, from_id, to_id, props
		FROM edges
		WHERE from_id = $1 AND edge_type = 'PARENT_OF'
		ORDER BY (props->>'order')::INTEGER NULLS LAST
	`

	rows, err := p.db.QueryContext(ctx, edgeQuery, id)
	if err != nil {
		p.logger.Error("Query children edges failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}, []*dgraph.Edge{}
	}
	defer rows.Close()

	var edges []*dgraph.Edge
	var childIDs []string

	for rows.Next() {
		var edge dgraph.Edge
		var propsJSON []byte

		if err := rows.Scan(&edge.ID, &edge.EdgeType, &edge.FromID, &edge.ToID, &propsJSON); err != nil {
			p.logger.Error("Scan edge failed", map[string]any{"error": err.Error()})
			continue
		}

		if err := json.Unmarshal(propsJSON, &edge.Props); err != nil {
			edge.Props = make(map[string]any)
		}

		edges = append(edges, &edge)
		childIDs = append(childIDs, edge.ToID)
	}

	if len(childIDs) == 0 {
		return []*dgraph.Node{}, []*dgraph.Edge{}
	}

	// Fetch child nodes
	nodes := p.getNodesByIDs(childIDs)
	return nodes, edges
}

// GetParentsPath retrieves the parent chain for a node.
func (p *PostgresStore) GetParentsPath(id string) ([]string, []string) {
	var nodeIDs []string
	var edgeTypes []string

	current := id
	for current != "" {
		nodeIDs = append([]string{current}, nodeIDs...)

		// Find parent
		ctx := context.Background()
		var parentID sql.NullString
		err := p.db.QueryRowContext(ctx,
			"SELECT from_id FROM edges WHERE to_id = $1 AND edge_type = 'PARENT_OF' LIMIT 1",
			current,
		).Scan(&parentID)

		if err == sql.ErrNoRows || !parentID.Valid {
			break
		}

		edgeTypes = append([]string{"PARENT_OF"}, edgeTypes...)
		current = parentID.String
	}

	return nodeIDs, edgeTypes
}

// Search performs a full-text search on nodes.
func (p *PostgresStore) Search(query, jurisdiction, code string, limit int) []dgraph.Node {
	ctx := context.Background()

	var conditions []string
	var args []interface{}
	argNum := 1

	if query != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR text ILIKE $%d OR citation ILIKE $%d)", argNum, argNum, argNum))
		args = append(args, "%"+query+"%")
		argNum++
	}

	if jurisdiction != "" {
		conditions = append(conditions, fmt.Sprintf("props->>'jurisdiction' = $%d", argNum))
		args = append(args, jurisdiction)
		argNum++
	}

	if code != "" {
		conditions = append(conditions, fmt.Sprintf("props->>'code' = $%d", argNum))
		args = append(args, code)
		argNum++
	}

	whereClause := "TRUE"
	if len(conditions) > 0 {
		whereClause = strings.Join(conditions, " AND ")
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, labels, title, citation, text, props, version_fetched_at, version_effective_date, version_hash, sources
		FROM nodes
		WHERE %s
		LIMIT $%d
	`, whereClause, argNum)

	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		p.logger.Error("Search query failed", map[string]any{"error": err.Error()})
		return []dgraph.Node{}
	}
	defer rows.Close()

	var results []dgraph.Node
	for rows.Next() {
		var node dgraph.Node
		var labelsArray pgStringArray
		var propsJSON, sourcesJSON []byte
		var versionFetchedAt, versionEffectiveDate, versionHash sql.NullString

		if err := rows.Scan(&node.ID, &labelsArray, &node.Title, &node.Citation, &node.Text,
			&propsJSON, &versionFetchedAt, &versionEffectiveDate, &versionHash, &sourcesJSON); err != nil {
			continue
		}

		node.Labels = []string(labelsArray)
		json.Unmarshal(propsJSON, &node.Props)
		json.Unmarshal(sourcesJSON, &node.Sources)

		if versionFetchedAt.Valid || versionEffectiveDate.Valid || versionHash.Valid {
			node.Version = &dgraph.Version{
				FetchedAt:     versionFetchedAt.String,
				EffectiveDate: versionEffectiveDate.String,
				Hash:          versionHash.String,
			}
		}

		results = append(results, node)
	}

	return results
}

// UpsertEdge creates or updates an edge.
func (p *PostgresStore) UpsertEdge(ctx context.Context, edge *dgraph.Edge) error {
	propsJSON, err := json.Marshal(edge.Props)
	if err != nil {
		return fmt.Errorf("marshal props: %w", err)
	}

	query := `
		INSERT INTO edges (id, edge_type, from_id, to_id, props)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (from_id, to_id, edge_type) DO UPDATE SET
			props = EXCLUDED.props
	`

	if edge.ID == "" {
		edge.ID = fmt.Sprintf("%s-%s-%s", edge.FromID, edge.EdgeType, edge.ToID)
	}

	_, err = p.db.ExecContext(ctx, query, edge.ID, edge.EdgeType, edge.FromID, edge.ToID, propsJSON)
	if err != nil {
		return fmt.Errorf("upsert edge: %w", err)
	}

	return nil
}

// Helper functions

func (p *PostgresStore) getNodesByIDs(ids []string) []*dgraph.Node {
	if len(ids) == 0 {
		return []*dgraph.Node{}
	}

	ctx := context.Background()
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, labels, title, citation, text, props, version_fetched_at, version_effective_date, version_hash, sources
		FROM nodes
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		p.logger.Error("Query nodes by IDs failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}
	}
	defer rows.Close()

	var nodes []*dgraph.Node
	for rows.Next() {
		var node dgraph.Node
		var labelsArray pgStringArray
		var propsJSON, sourcesJSON []byte
		var versionFetchedAt, versionEffectiveDate, versionHash sql.NullString

		if err := rows.Scan(&node.ID, &labelsArray, &node.Title, &node.Citation, &node.Text,
			&propsJSON, &versionFetchedAt, &versionEffectiveDate, &versionHash, &sourcesJSON); err != nil {
			continue
		}

		node.Labels = []string(labelsArray)
		json.Unmarshal(propsJSON, &node.Props)
		json.Unmarshal(sourcesJSON, &node.Sources)

		if versionFetchedAt.Valid {
			if node.Version == nil {
				node.Version = &dgraph.Version{}
			}
			node.Version.FetchedAt = versionFetchedAt.String
		}

		nodes = append(nodes, &node)
	}

	return nodes
}

// PostgreSQL array helper types
type pgStringArray []string

func (p *pgStringArray) Scan(src interface{}) error {
	if src == nil {
		*p = []string{}
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return p.scanBytes(v)
	case string:
		return p.scanBytes([]byte(v))
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

func (p *pgStringArray) scanBytes(src []byte) error {
	s := string(src)
	s = strings.Trim(s, "{}")
	if s == "" {
		*p = []string{}
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.Trim(part, `"`)
	}
	*p = result
	return nil
}

func pgArray(items []string) interface{} {
	if len(items) == 0 {
		return "{}"
	}
	return "{" + strings.Join(items, ",") + "}"
}

// GetCitations returns nodes that cite the target node.
func (p *PostgresStore) GetCitations(targetID string) ([]*dgraph.Node, []*dgraph.Edge) {
	ctx := context.Background()

	query := `
		SELECT id, edge_type, from_id, to_id, props
		FROM edges
		WHERE to_id = $1 AND edge_type = 'CITES'
	`

	rows, err := p.db.QueryContext(ctx, query, targetID)
	if err != nil {
		p.logger.Error("Query citations failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}, []*dgraph.Edge{}
	}
	defer rows.Close()

	var edges []*dgraph.Edge
	var fromIDs []string

	for rows.Next() {
		var edge dgraph.Edge
		var propsJSON []byte

		if err := rows.Scan(&edge.ID, &edge.EdgeType, &edge.FromID, &edge.ToID, &propsJSON); err != nil {
			continue
		}

		json.Unmarshal(propsJSON, &edge.Props)
		edges = append(edges, &edge)
		fromIDs = append(fromIDs, edge.FromID)
	}

	nodes := p.getNodesByIDs(fromIDs)
	return nodes, edges
}

// GetOutgoingCitations returns nodes that the source node cites.
func (p *PostgresStore) GetOutgoingCitations(sourceID string) ([]*dgraph.Node, []*dgraph.Edge) {
	ctx := context.Background()

	query := `
		SELECT id, edge_type, from_id, to_id, props
		FROM edges
		WHERE from_id = $1 AND edge_type = 'CITES'
	`

	rows, err := p.db.QueryContext(ctx, query, sourceID)
	if err != nil {
		p.logger.Error("Query outgoing citations failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}, []*dgraph.Edge{}
	}
	defer rows.Close()

	var edges []*dgraph.Edge
	var toIDs []string

	for rows.Next() {
		var edge dgraph.Edge
		var propsJSON []byte

		if err := rows.Scan(&edge.ID, &edge.EdgeType, &edge.FromID, &edge.ToID, &propsJSON); err != nil {
			continue
		}

		json.Unmarshal(propsJSON, &edge.Props)
		edges = append(edges, &edge)
		toIDs = append(toIDs, edge.ToID)
	}

	nodes := p.getNodesByIDs(toIDs)
	return nodes, edges
}

// GetTopics returns all TOPIC nodes.
func (p *PostgresStore) GetTopics() []*dgraph.Node {
	ctx := context.Background()

	query := `
		SELECT id, labels, title, citation, text, props, version_fetched_at, version_effective_date, version_hash, sources
		FROM nodes
		WHERE 'TOPIC' = ANY(labels)
		ORDER BY title
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		p.logger.Error("Query topics failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}
	}
	defer rows.Close()

	var nodes []*dgraph.Node
	for rows.Next() {
		var node dgraph.Node
		var labelsArray pgStringArray
		var propsJSON, sourcesJSON []byte
		var versionFetchedAt, versionEffectiveDate, versionHash sql.NullString

		if err := rows.Scan(&node.ID, &labelsArray, &node.Title, &node.Citation, &node.Text,
			&propsJSON, &versionFetchedAt, &versionEffectiveDate, &versionHash, &sourcesJSON); err != nil {
			continue
		}

		node.Labels = []string(labelsArray)
		json.Unmarshal(propsJSON, &node.Props)
		json.Unmarshal(sourcesJSON, &node.Sources)

		nodes = append(nodes, &node)
	}

	return nodes
}

// GetTopicAssociations returns nodes linked to a topic via HAS_TOPIC edges.
func (p *PostgresStore) GetTopicAssociations(topicID string) ([]*dgraph.Node, []*dgraph.Edge) {
	ctx := context.Background()

	query := `
		SELECT id, edge_type, from_id, to_id, props
		FROM edges
		WHERE to_id = $1 AND edge_type = 'HAS_TOPIC'
	`

	rows, err := p.db.QueryContext(ctx, query, topicID)
	if err != nil {
		p.logger.Error("Query topic associations failed", map[string]any{"error": err.Error()})
		return []*dgraph.Node{}, []*dgraph.Edge{}
	}
	defer rows.Close()

	var edges []*dgraph.Edge
	var fromIDs []string

	for rows.Next() {
		var edge dgraph.Edge
		var propsJSON []byte

		if err := rows.Scan(&edge.ID, &edge.EdgeType, &edge.FromID, &edge.ToID, &propsJSON); err != nil {
			continue
		}

		json.Unmarshal(propsJSON, &edge.Props)
		edges = append(edges, &edge)
		fromIDs = append(fromIDs, edge.FromID)
	}

	nodes := p.getNodesByIDs(fromIDs)
	return nodes, edges
}

// SliceFromRoot returns a subgraph starting from a root node.
func (p *PostgresStore) SliceFromRoot(root string, depth int, labelFilter map[string]struct{}) ([]*dgraph.Node, []*dgraph.Edge, error) {
	// For simplicity, just return the node and its immediate children for now
	// A full implementation would recursively traverse to the specified depth
	node, ok := p.GetNode(root)
	if !ok {
		return nil, nil, fmt.Errorf("root node not found: %s", root)
	}

	nodes := []*dgraph.Node{node}
	edges := []*dgraph.Edge{}

	if depth > 0 {
		children, childEdges := p.GetChildren(root)
		nodes = append(nodes, children...)
		edges = append(edges, childEdges...)
	}

	return nodes, edges, nil
}
