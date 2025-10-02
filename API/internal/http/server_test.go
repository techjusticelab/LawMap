package httpapi

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    graphrepo "lawmap/internal/repo/graph"
    conf "lawmap/internal/config"
)

func newTestMux(t *testing.T) *http.ServeMux {
    t.Helper()
    store := graphrepo.NewMemoryStore()
    if err := store.LoadJSONL("../../docs/EXAMPLES.graph.jsonl"); err != nil { t.Fatalf("load: %v", err) }
    s := NewServer(store, []conf.SourceDescriptor{})
    mux := http.NewServeMux()
    s.Routes(mux)
    return mux
}

func TestHealth(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/health", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestGetNode(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:§3342", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var body map[string]any
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    if body["id"] != "CA:CIV:T02:CH02:§3342" { t.Fatalf("unexpected id: %v", body["id"]) }
}

func TestChildrenEndpoint(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02/children", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestGraphSliceEndpoint(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/graph?root=CA:CIV:T02:CH02&depth=1", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestSearchEndpoint(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/search?q=dog+bite&jurisdiction=CA&code=CIV&sort=title&limit=1", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestNodeFieldsSelection(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:%C2%A73342?fields=id,title", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var body map[string]any
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    if _, ok := body["id"]; !ok { t.Fatalf("id missing") }
    if _, ok := body["title"]; !ok { t.Fatalf("title missing") }
    if _, ok := body["labels"]; ok { t.Fatalf("labels should be omitted by fields selection") }
}

func TestCitationsHeader(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?limit=1", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    if rr.Header().Get("X-Total-Count") == "" { t.Fatalf("expected X-Total-Count header") }
}

func TestChildrenFilterPaginate(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02/children?labels=SECTION&limit=1&offset=0", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestOutgoingCitesEndpoint(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/nodes/CA:OPN:People_v_Smith_2020_1/cites", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
}

func TestGraphSliceFilterLabels(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/graph?root=CA:CIV:T02:CH02&depth=1&labels=SECTION", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var body struct{
        Nodes []struct{ ID string `json:"id"`; Labels []string `json:"labels"` } `json:"nodes"`
    }
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    if len(body.Nodes) == 0 { t.Fatalf("expected nodes in filtered graph") }
    hasSection := false
    for _, n := range body.Nodes {
        for _, l := range n.Labels { if l == "SECTION" { hasSection = true } }
    }
    if !hasSection { t.Fatalf("expected at least one SECTION node in filtered graph") }
}

func TestVersionsAndDiff(t *testing.T) {
    mux := newTestMux(t)
    // versions
    req := httptest.NewRequest("GET", "/versions/CA:CIV:T02:CH02:§3342", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("versions status=%d", rr.Code) }
    var versions []map[string]any
    _ = json.Unmarshal(rr.Body.Bytes(), &versions)
    if len(versions) == 0 { t.Fatalf("expected at least one version") }

    // diff
    req2 := httptest.NewRequest("GET", "/diff/CA:CIV:T02:CH02:§3342", nil)
    rr2 := httptest.NewRecorder()
    mux.ServeHTTP(rr2, req2)
    if rr2.Code != 200 { t.Fatalf("diff status=%d", rr2.Code) }
    var diff map[string]any
    _ = json.Unmarshal(rr2.Body.Bytes(), &diff)
    if diff["id"] != "CA:CIV:T02:CH02:§3342" { t.Fatalf("unexpected id in diff") }
}

func TestSearchTableDriven(t *testing.T) {
    mux := newTestMux(t)
    cases := []struct{ path string; wantStatus int }{
        {"/search?q=dog", 200},
        {"/search?q=rule&jurisdiction=CA&code=CRC", 200},
        {"/search?q=Regulatory&jurisdiction=CA&code=CCR", 200},
        {"/search?q=", 200},
    }
    for _, c := range cases {
        req := httptest.NewRequest("GET", c.path, nil)
        rr := httptest.NewRecorder()
        mux.ServeHTTP(rr, req)
        if rr.Code != c.wantStatus { t.Fatalf("%s status=%d", c.path, rr.Code) }
    }
}

func TestSourcesEndpoint(t *testing.T) {
    mux := newTestMux(t)
    req := httptest.NewRequest("GET", "/sources", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var body struct{ Sources []map[string]any `json:"sources"` }
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    if len(body.Sources) == 0 { t.Fatalf("expected at least one source") }
}

func TestTopicsEndpoints(t *testing.T) {
    mux := newTestMux(t)
    // list
    req := httptest.NewRequest("GET", "/topics", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("list status=%d", rr.Code) }
    var list struct{ Topics []map[string]any `json:"topics"` }
    _ = json.Unmarshal(rr.Body.Bytes(), &list)
    if len(list.Topics) == 0 { t.Fatalf("expected topics") }

    // by id
    req2 := httptest.NewRequest("GET", "/topics/TOPIC:Dogs", nil)
    rr2 := httptest.NewRecorder()
    mux.ServeHTTP(rr2, req2)
    if rr2.Code != 200 { t.Fatalf("by id status=%d", rr2.Code) }
    var slice struct{ Nodes []map[string]any `json:"nodes"`; Edges []map[string]any `json:"edges"` }
    _ = json.Unmarshal(rr2.Body.Bytes(), &slice)
    if len(slice.Nodes) < 2 || len(slice.Edges) == 0 { t.Fatalf("expected topic slice") }
}

func TestCitationsEndpoint(t *testing.T) {
    mux := newTestMux(t)
    // Known: CA CIV §3342 is cited by at least two opinions in fixtures
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:%C2%A73342/citations", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var slice struct{ Nodes []map[string]any `json:"nodes"`; Edges []map[string]any `json:"edges"` }
    _ = json.Unmarshal(rr.Body.Bytes(), &slice)
    if len(slice.Nodes) == 0 { t.Fatalf("expected at least one reverse citation") }
}

func TestCitationsFilterAndPaginate(t *testing.T) {
    mux := newTestMux(t)
    // filter to opinions only
    req := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?labels=OPINION&limit=1&offset=0", nil)
    rr := httptest.NewRecorder()
    mux.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d", rr.Code) }
    var page1 struct{ Nodes []struct{ Labels []string `json:"labels"` } `json:"nodes"` }
    _ = json.Unmarshal(rr.Body.Bytes(), &page1)
    if len(page1.Nodes) != 1 { t.Fatalf("expected 1 item in page1, got %d", len(page1.Nodes)) }
    // verify label filter applied
    hasOpinion := false
    for _, l := range page1.Nodes[0].Labels { if l == "OPINION" { hasOpinion = true } }
    if !hasOpinion { t.Fatalf("expected OPINION label in filtered results") }

    // second page
    req2 := httptest.NewRequest("GET", "/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?labels=OPINION&limit=1&offset=1", nil)
    rr2 := httptest.NewRecorder()
    mux.ServeHTTP(rr2, req2)
    if rr2.Code != 200 { t.Fatalf("status=%d", rr2.Code) }
}
