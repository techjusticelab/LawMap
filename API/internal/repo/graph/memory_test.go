package graphrepo

import (
    "path/filepath"
    "testing"
)

func exFile() string {
    // tests run from package dir; locate API/docs/EXAMPLES.graph.jsonl relative to repo root
    // this file is at API/internal/repo/graph → up to API, then docs
    return filepath.Clean("../../../docs/EXAMPLES.graph.jsonl")
}

func TestLoadAndGet(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil {
        t.Fatalf("load: %v", err)
    }
    ids := []string{
        "CA",
        "CA:CIV:T02:CH02:§3342",
        "CA:CRC:rule_1.1",
        "CA:CCR:T15:§3044",
    }
    for _, id := range ids {
        if _, ok := m.GetNode(id); !ok {
            t.Fatalf("expected node %s to exist", id)
        }
    }
}

func TestChildren(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil { t.Fatal(err) }
    nodes, edges := m.GetChildren("CA:CIV:T02:CH02")
    if len(nodes) == 0 || len(edges) == 0 {
        t.Fatalf("expected children for chapter; got nodes=%d edges=%d", len(nodes), len(edges))
    }
}

func TestParentsPath(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil { t.Fatal(err) }
    nodes, edges := m.GetParentsPath("CA:CIV:T02:CH02:§3342")
    if len(nodes) < 2 || len(edges) < 1 {
        t.Fatalf("expected ancestry path; got nodes=%v edges=%v", nodes, edges)
    }
}

func TestSliceFromRoot(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil { t.Fatal(err) }
    ns, es, err := m.SliceFromRoot("CA:CIV:T02:CH02", 1, nil)
    if err != nil { t.Fatal(err) }
    if len(ns) < 2 || len(es) < 1 {
        t.Fatalf("expected slice with children; ns=%d es=%d", len(ns), len(es))
    }
}

func TestSearch(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil { t.Fatal(err) }
    got := m.Search("dog bite", "CA", "CIV", 10)
    if len(got) == 0 {
        t.Fatalf("expected search results for 'dog bite'")
    }
}

func TestChildrenOrder(t *testing.T) {
    m := NewMemoryStore()
    if err := m.LoadJSONL(exFile()); err != nil { t.Fatal(err) }
    nodes, _ := m.GetChildren("CA:CIV:T02:CH02")
    if len(nodes) < 2 { t.Fatalf("need at least two children to test order") }
    // Expect §3343 (order 5) to come before §3342 (order 10)
    if nodes[0].ID != "CA:CIV:T02:CH02:§3343" {
        t.Fatalf("expected first child to be §3343, got %s", nodes[0].ID)
    }
}
