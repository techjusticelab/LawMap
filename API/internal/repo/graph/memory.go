package graphrepo

import (
    "bufio"
    "encoding/json"
    "errors"
    "os"
    "sort"
    "strings"

    dgraph "lawmap/internal/domain/graph"
)

// MemoryStore is a simple in-memory graph storage for development and tests.
type MemoryStore struct {
    nodes       map[string]*dgraph.Node
    edges       []*dgraph.Edge
    edgesByFrom map[string][]*dgraph.Edge
    edgesByTo   map[string][]*dgraph.Edge
    parentOf    map[string][]string // parent -> children IDs (PARENT_OF)
    parentID    map[string]string   // child -> parent ID
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        nodes:       make(map[string]*dgraph.Node),
        edges:       make([]*dgraph.Edge, 0, 1024),
        edgesByFrom: make(map[string][]*dgraph.Edge),
        edgesByTo:   make(map[string][]*dgraph.Edge),
        parentOf:    make(map[string][]string),
        parentID:    make(map[string]string),
    }
}

// LoadJSONL loads nodes and edges from a JSONL file following EXAMPLES.graph.jsonl format.
func (m *MemoryStore) LoadJSONL(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        var raw map[string]any
        if err := json.Unmarshal([]byte(line), &raw); err != nil { return err }
        t, _ := raw["type"].(string)
        switch t {
        case "node":
            var n dgraph.Node
            if err := json.Unmarshal([]byte(line), &n); err != nil { return err }
            m.nodes[n.ID] = &n
        case "edge":
            var e dgraph.Edge
            if err := json.Unmarshal([]byte(line), &e); err != nil { return err }
            m.edges = append(m.edges, &e)
            m.edgesByFrom[e.FromID] = append(m.edgesByFrom[e.FromID], &e)
            m.edgesByTo[e.ToID] = append(m.edgesByTo[e.ToID], &e)
            if e.EdgeType == "PARENT_OF" {
                m.parentOf[e.FromID] = append(m.parentOf[e.FromID], e.ToID)
                m.parentID[e.ToID] = e.FromID
            }
        default:
            // ignore unknown lines
        }
    }
    if err := sc.Err(); err != nil { return err }

    // stable child order if "order" property exists
    for p, kids := range m.parentOf {
        sort.SliceStable(kids, func(i, j int) bool {
            a, b := kids[i], kids[j]
            // try to sort by edge props.order
            ai, bi := 0, 0
            for _, e := range m.edgesByFrom[p] {
                if e.ToID == a { if v, ok := e.Props["order"].(float64); ok { ai = int(v) } }
                if e.ToID == b { if v, ok := e.Props["order"].(float64); ok { bi = int(v) } }
            }
            return ai < bi
        })
        m.parentOf[p] = kids
    }
    return nil
}

func (m *MemoryStore) GetNode(id string) (*dgraph.Node, bool) {
    n, ok := m.nodes[id]
    return n, ok
}

func (m *MemoryStore) GetChildren(id string) ([]*dgraph.Node, []*dgraph.Edge) {
    children := m.parentOf[id]
    nodes := make([]*dgraph.Node, 0, len(children))
    edges := make([]*dgraph.Edge, 0, len(children))
    for _, cid := range children {
        if n, ok := m.nodes[cid]; ok { nodes = append(nodes, n) }
        for _, e := range m.edgesByFrom[id] {
            if e.ToID == cid && e.EdgeType == "PARENT_OF" { edges = append(edges, e) }
        }
    }
    return nodes, edges
}

func (m *MemoryStore) GetParentsPath(id string) ([]string, []string) {
    var nodes []string
    var edges []string
    cur := id
    for cur != "" {
        nodes = append([]string{cur}, nodes...)
        p := m.parentID[cur]
        if p == "" { break }
        edges = append([]string{"PARENT_OF"}, edges...)
        cur = p
    }
    return nodes, edges
}

func (m *MemoryStore) SliceFromRoot(root string, depth int, labelFilter map[string]struct{}) ([]*dgraph.Node, []*dgraph.Edge, error) {
    if _, ok := m.nodes[root]; !ok { return nil, nil, errors.New("root not found") }
    visited := make(map[string]struct{})
    q := []struct{ id string; d int }{{root, 0}}
    var nodes []*dgraph.Node
    var edges []*dgraph.Edge
    for len(q) > 0 {
        cur := q[0]; q = q[1:]
        if _, ok := visited[cur.id]; ok { continue }
        visited[cur.id] = struct{}{}
        n := m.nodes[cur.id]
        if n != nil {
            if len(labelFilter) > 0 {
                keep := false
                for _, l := range n.Labels { if _, ok := labelFilter[l]; ok { keep = true; break } }
                if keep || cur.d == 0 { nodes = append(nodes, n) }
            } else {
                nodes = append(nodes, n)
            }
        }
        if cur.d >= depth { continue }
        for _, e := range m.edgesByFrom[cur.id] {
            if e.EdgeType != "PARENT_OF" { continue }
            edges = append(edges, e)
            q = append(q, struct{ id string; d int }{e.ToID, cur.d + 1})
        }
    }
    // Deduplicate edges
    dedup := make(map[string]struct{})
    outEdges := make([]*dgraph.Edge, 0, len(edges))
    for _, e := range edges {
        key := e.FromID + "->" + e.ToID + ":" + e.EdgeType
        if _, ok := dedup[key]; ok { continue }
        dedup[key] = struct{}{}
        outEdges = append(outEdges, e)
    }
    return nodes, outEdges, nil
}

func (m *MemoryStore) Search(q string, jurisdiction, code string, limit int) []dgraph.Node {
    ql := strings.ToLower(q)
    out := make([]dgraph.Node, 0, limit)
    for _, n := range m.nodes {
        if jurisdiction != "" {
            if j, _ := n.Props["jurisdiction"].(string); strings.ToUpper(j) != strings.ToUpper(jurisdiction) { continue }
        }
        if code != "" {
            if c, _ := n.Props["code"].(string); strings.ToUpper(c) != strings.ToUpper(code) { continue }
        }
        if q == "" || strings.Contains(strings.ToLower(n.Title), ql) || strings.Contains(strings.ToLower(n.Text), ql) || strings.Contains(strings.ToLower(n.Citation), ql) {
            out = append(out, *n)
            if len(out) >= limit { break }
        }
    }
    return out
}

// GetCitations returns nodes that cite the given target via CITES edges and those edges.
func (m *MemoryStore) GetCitations(targetID string) ([]*dgraph.Node, []*dgraph.Edge) {
    var nodes []*dgraph.Node
    var edges []*dgraph.Edge
    for _, e := range m.edgesByTo[targetID] {
        if e.EdgeType != "CITES" { continue }
        if n, ok := m.nodes[e.FromID]; ok {
            nodes = append(nodes, n)
            edges = append(edges, e)
        }
    }
    return nodes, edges
}

// GetOutgoingCitations returns nodes that the given source cites via CITES edges and those edges.
func (m *MemoryStore) GetOutgoingCitations(sourceID string) ([]*dgraph.Node, []*dgraph.Edge) {
    var nodes []*dgraph.Node
    var edges []*dgraph.Edge
    for _, e := range m.edgesByFrom[sourceID] {
        if e.EdgeType != "CITES" { continue }
        if n, ok := m.nodes[e.ToID]; ok {
            nodes = append(nodes, n)
            edges = append(edges, e)
        }
    }
    return nodes, edges
}

// GetTopics returns all nodes labeled TOPIC.
func (m *MemoryStore) GetTopics() []*dgraph.Node {
    out := make([]*dgraph.Node, 0)
    for _, n := range m.nodes {
        for _, l := range n.Labels {
            if l == "TOPIC" {
                out = append(out, n)
                break
            }
        }
    }
    return out
}

// GetTopicAssociations returns nodes linked to the given topic via HAS_TOPIC edges and the edges themselves.
func (m *MemoryStore) GetTopicAssociations(topicID string) ([]*dgraph.Node, []*dgraph.Edge) {
    var nodes []*dgraph.Node
    var edges []*dgraph.Edge
    for _, e := range m.edgesByTo[topicID] {
        if e.EdgeType != "HAS_TOPIC" { continue }
        if n, ok := m.nodes[e.FromID]; ok {
            nodes = append(nodes, n)
            edges = append(edges, e)
        }
    }
    return nodes, edges
}
