package graph

// Version metadata for a node's content.
type Version struct {
    FetchedAt     string `json:"fetched_at,omitempty"`
    EffectiveDate string `json:"effective_date,omitempty"`
    Hash          string `json:"hash,omitempty"`
}

// SourceMeta captures provenance for a node's content.
type SourceMeta struct {
    Name        string `json:"name,omitempty"`
    URL         string `json:"url,omitempty"`
    RetrievedAt string `json:"retrieved_at,omitempty"`
}

// Node represents a graph node as stored in the repository.
type Node struct {
    ID       string                 `json:"id"`
    Labels   []string               `json:"labels"`
    Title    string                 `json:"title,omitempty"`
    Citation string                 `json:"citation,omitempty"`
    Text     string                 `json:"text,omitempty"`
    Props    map[string]any         `json:"props,omitempty"`
    Version  *Version               `json:"version,omitempty"`
    Sources  []SourceMeta           `json:"sources,omitempty"`
}

// Edge is a directed relationship in the repository format.
// The field name EdgeType is used for storage to avoid clashing with the JSONL item-level "type" key.
type Edge struct {
    ID       string            `json:"id,omitempty"`
    EdgeType string            `json:"edge_type"`
    FromID   string            `json:"from_id"`
    ToID     string            `json:"to_id"`
    Props    map[string]any    `json:"props,omitempty"`
}

// DTOs used by the HTTP layer.
type NodeDTO struct {
    ID       string                 `json:"id"`
    Labels   []string               `json:"labels"`
    Title    string                 `json:"title,omitempty"`
    Citation string                 `json:"citation,omitempty"`
    Text     string                 `json:"text,omitempty"`
    Props    map[string]any         `json:"props,omitempty"`
    Version  *Version               `json:"version,omitempty"`
    Sources  []SourceMeta           `json:"sources,omitempty"`
}

type EdgeDTO struct {
    ID     string            `json:"id,omitempty"`
    Type   string            `json:"type"`
    FromID string            `json:"from_id"`
    ToID   string            `json:"to_id"`
    Props  map[string]any    `json:"props,omitempty"`
}

type GraphSliceDTO struct {
    Nodes []NodeDTO `json:"nodes"`
    Edges []EdgeDTO `json:"edges"`
}

type PathDTO struct {
    Nodes []string `json:"nodes"`
    Edges []string `json:"edges,omitempty"`
}

type SearchItem struct {
    Type    string `json:"type"`
    ID      string `json:"id"`
    Title   string `json:"title,omitempty"`
    Snippet string `json:"snippet,omitempty"`
}

type SearchResultDTO struct {
    Query      string       `json:"query,omitempty"`
    Items      []SearchItem `json:"items"`
    NextCursor string       `json:"next_cursor,omitempty"`
}

