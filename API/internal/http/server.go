package httpapi

import (
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sort"
    "encoding/base64"

    dgraph "lawmap/internal/domain/graph"
    graphrepo "lawmap/internal/repo/graph"
    conf "lawmap/internal/config"
)

type Server struct {
    store   *graphrepo.MemoryStore
    sources []sourceDesc
}

func NewServer(store *graphrepo.MemoryStore, sourcesCfg []conf.SourceDescriptor) *Server {
    // convert config to internal representation
    var sdescs []sourceDesc
    for _, s := range sourcesCfg {
        sdescs = append(sdescs, sourceDesc{
            Name: s.Name, Jurisdictions: s.Jurisdictions, Codes: s.Codes, Kind: s.Kind, URLs: s.URLs,
        })
    }
    return &Server{store: store, sources: sdescs}
}

func (s *Server) Routes(mux *http.ServeMux) {
    mux.HandleFunc("/health", s.handleHealth)
    mux.HandleFunc("/sources", s.handleSources)
    mux.HandleFunc("/topics", s.handleTopics)
    mux.HandleFunc("/topics/", s.handleTopics)
    mux.HandleFunc("/nodes/", s.handleNodes)
    mux.HandleFunc("/graph", s.handleGraph)
    mux.HandleFunc("/search", s.handleSearch)
    mux.HandleFunc("/diff/", s.handleDiff)
    mux.HandleFunc("/versions/", s.handleVersions)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type sourceDesc struct {
    Name          string   `json:"name"`
    Jurisdictions []string `json:"jurisdictions"`
    Codes         []string `json:"codes"`
    Kind          string   `json:"kind"`   // bulk|api|web|mixed
    URLs          []string `json:"urls"`
}

func (s *Server) handleSources(w http.ResponseWriter, r *http.Request) {
    // Static capability list for now; in future, read from config.
    sources := s.sources
    if len(sources) == 0 {
        sources = []sourceDesc{
        {Name: "US Code (OLRC)", Jurisdictions: []string{"US"}, Codes: []string{"USC"}, Kind: "bulk", URLs: []string{"https://uscode.house.gov/download/download.shtml"}},
        {Name: "GovInfo CFR", Jurisdictions: []string{"US"}, Codes: []string{"CFR"}, Kind: "bulk", URLs: []string{"https://www.govinfo.gov/app/collection/CFR"}},
        {Name: "eCFR", Jurisdictions: []string{"US"}, Codes: []string{"CFR"}, Kind: "web", URLs: []string{"https://www.ecfr.gov/"}},
        {Name: "Federal Register", Jurisdictions: []string{"US"}, Codes: []string{"FR"}, Kind: "web", URLs: []string{"https://www.federalregister.gov/"}},
        {Name: "US Sentencing Commission", Jurisdictions: []string{"US"}, Codes: []string{"USSG"}, Kind: "web", URLs: []string{"https://www.ussc.gov/guidelines"}},
        {Name: "CourtListener (opinions/RECAP)", Jurisdictions: []string{"US","CA"}, Codes: []string{"OPN"}, Kind: "web", URLs: []string{"https://www.courtlistener.com/"}},
        {Name: "National Archives (Constitution)", Jurisdictions: []string{"US"}, Codes: []string{"CONST"}, Kind: "web", URLs: []string{"https://www.archives.gov/founding-docs/constitution"}},
        {Name: "CA LegInfo", Jurisdictions: []string{"CA"}, Codes: []string{"CIV","PEN","CONS","BPC", "CRC"}, Kind: "mixed", URLs: []string{"https://leginfo.legislature.ca.gov/"}},
        {Name: "CA OAL / CCR", Jurisdictions: []string{"CA"}, Codes: []string{"CCR"}, Kind: "web", URLs: []string{"https://oal.ca.gov/publications/"}},
        {Name: "California Attorney General Opinions", Jurisdictions: []string{"CA"}, Codes: []string{"OPN"}, Kind: "web", URLs: []string{"https://oag.ca.gov/opinions"}},
        {Name: "California Courts (opinions)", Jurisdictions: []string{"CA"}, Codes: []string{"OPN"}, Kind: "web", URLs: []string{"https://www.courts.ca.gov/opinions.htm"}},
        }
    }
    writeJSON(w, http.StatusOK, map[string]any{"sources": sources})
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/nodes/")
    if path == "" || path == r.URL.Path {
        writeError(w, http.StatusBadRequest, "bad_request", "missing id", nil)
        return
    }
    if strings.HasSuffix(path, "/children") {
        id := strings.TrimSuffix(path, "/children")
        s.handleNodeChildren(w, r, id)
        return
    }
    if strings.HasSuffix(path, "/parents") {
        id := strings.TrimSuffix(path, "/parents")
        s.handleNodeParents(w, r, id)
        return
    }
    if strings.HasSuffix(path, "/citations") || strings.HasSuffix(path, "/citers") {
        id := strings.TrimSuffix(path, "/citations")
        id = strings.TrimSuffix(id, "/citers")
        s.handleNodeCitations(w, r, id)
        return
    }
    if strings.HasSuffix(path, "/cites") {
        id := strings.TrimSuffix(path, "/cites")
        s.handleNodeCites(w, r, id)
        return
    }
    id := path
    n, ok := s.store.GetNode(id)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "Node not found", nil)
        return
    }
    dto := nodeToDTO(n)
    // Optional expansions and field selection
    q := r.URL.Query()
    expand := q.Get("expand")
    fieldsParam := q.Get("fields")
    // If no expand and no fields selection, return DTO directly
    if expand == "" && fieldsParam == "" {
        writeJSON(w, http.StatusOK, dto)
        return
    }
    // Build response map (enables field selection)
    resp := map[string]any{}
    // If fields specified, include only those keys; else include full NodeDTO fields
    include := func(k string) bool { if fieldsParam == "" { return true }; for _, f := range strings.Split(fieldsParam, ",") { if strings.TrimSpace(f) == k { return true } }; return false }
    if include("id") { resp["id"] = dto.ID }
    if include("labels") { resp["labels"] = dto.Labels }
    if include("title") { resp["title"] = dto.Title }
    if include("citation") { resp["citation"] = dto.Citation }
    if include("text") { resp["text"] = dto.Text }
    if include("props") { resp["props"] = dto.Props }
    if include("version") { resp["version"] = dto.Version }
    if include("sources") { resp["sources"] = dto.Sources }
    switch expand {
    case "parents":
        nodes, edges := s.store.GetParentsPath(id)
        resp["parents"] = dgraph.PathDTO{Nodes: nodes, Edges: edges}
    case "children":
        ns, es := s.store.GetChildren(id)
        cn := make([]dgraph.NodeDTO, 0, len(ns))
        for _, n2 := range ns { cn = append(cn, nodeToDTO(n2)) }
        ce := make([]dgraph.EdgeDTO, 0, len(es))
        for _, e := range es { ce = append(ce, edgeToDTO(e)) }
        resp["children"] = dgraph.GraphSliceDTO{Nodes: cn, Edges: ce}
    default:
        // ignore unknown expand
    }
    writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleNodeChildren(w http.ResponseWriter, r *http.Request, id string) {
    ns, es := s.store.GetChildren(id)
    // Optional label filter and pagination
    q := r.URL.Query()
    labelsParam := q.Get("labels")
    sortParam := q.Get("sort") // order|title|-title
    haveFilter := labelsParam != ""
    labelSet := make(map[string]struct{})
    if haveFilter { for _, l := range strings.Split(labelsParam, ",") { labelSet[strings.TrimSpace(l)] = struct{}{} } }
    type pair struct{ n *dgraph.Node; e *dgraph.Edge }
    pairs := make([]pair, 0, len(ns))
    for i := range ns {
        keep := true
        if haveFilter {
            keep = false
            for _, l := range ns[i].Labels { if _, ok := labelSet[l]; ok { keep = true; break } }
        }
        if keep { pairs = append(pairs, pair{ns[i], es[i]}) }
    }
    // Sorting
    switch sortParam {
    case "title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) < strings.ToLower(pairs[j].n.Title) })
    case "-title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) > strings.ToLower(pairs[j].n.Title) })
    default: // "order" (edge props)
        sort.SliceStable(pairs, func(i, j int) bool {
            oi, oj := 0, 0
            if v, ok := pairs[i].e.Props["order"].(float64); ok { oi = int(v) }
            if v, ok := pairs[j].e.Props["order"].(float64); ok { oj = int(v) }
            return oi < oj
        })
    }
    limit := 1000 // children usually small; cap to 1000
    if lv := q.Get("limit"); lv != "" { if n, err := strconv.Atoi(lv); err == nil && n > 0 && n <= 1000 { limit = n } }
    offset := 0
    if cur := q.Get("cursor"); cur != "" {
        if b, err := base64.URLEncoding.DecodeString(cur); err == nil {
            s := string(b)
            if strings.HasPrefix(s, "o:") {
                if n, err := strconv.Atoi(strings.TrimPrefix(s, "o:")); err == nil && n >= 0 { offset = n }
            }
        }
    } else if ov := q.Get("offset"); ov != "" { if n, err := strconv.Atoi(ov); err == nil && n >= 0 { offset = n } }
    start := offset
    if start > len(pairs) { start = len(pairs) }
    end := start + limit
    if end > len(pairs) { end = len(pairs) }
    slice := pairs[start:end]
    nodes := make([]dgraph.NodeDTO, 0, len(slice))
    edges := make([]dgraph.EdgeDTO, 0, len(slice))
    fieldsParam := q.Get("fields")
    var fs map[string]struct{}
    if fieldsParam != "" { fs = make(map[string]struct{}); for _, f := range strings.Split(fieldsParam, ",") { fs[strings.TrimSpace(f)] = struct{}{} } }
    for _, pr := range slice {
        ndto := nodeToDTO(pr.n)
        if fs != nil { ndto = filterNodeFields(ndto, fs) }
        nodes = append(nodes, ndto)
        edges = append(edges, edgeToDTO(pr.e))
    }
    resp := map[string]any{
        "nodes": nodes,
        "edges": edges,
        "total": len(pairs),
        "next_offset": end,
    }
    if end >= len(pairs) {
        resp["next_offset"] = nil
        resp["next_cursor"] = nil
    } else {
        resp["next_cursor"] = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("o:%d", end)))
    }
    w.Header().Set("X-Total-Count", strconv.Itoa(len(pairs)))
    writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleNodeParents(w http.ResponseWriter, r *http.Request, id string) {
    nodes, edges := s.store.GetParentsPath(id)
    writeJSON(w, http.StatusOK, dgraph.PathDTO{Nodes: nodes, Edges: edges})
}

func (s *Server) handleNodeCitations(w http.ResponseWriter, r *http.Request, id string) {
    ns, es := s.store.GetCitations(id)
    // Optional label filter
    q := r.URL.Query()
    labelsParam := q.Get("labels")
    pinFilter := strings.ToLower(q.Get("pin_cite_contains"))
    ctxFilter := strings.ToLower(q.Get("context_contains"))
    sortParam := q.Get("sort") // title|-title|id|-id
    haveFilter := labelsParam != ""
    labelSet := make(map[string]struct{})
    if haveFilter {
        for _, l := range strings.Split(labelsParam, ",") { labelSet[strings.TrimSpace(l)] = struct{}{} }
    }
    type pair struct{ n *dgraph.Node; e *dgraph.Edge }
    pairs := make([]pair, 0, len(ns))
    for i := range ns {
        keep := true
        if haveFilter {
            keep = false
            for _, l := range ns[i].Labels { if _, ok := labelSet[l]; ok { keep = true; break } }
        }
        if keep && pinFilter != "" {
            sVal := ""
            if v, ok := es[i].Props["pin_cite"].(string); ok { sVal = strings.ToLower(v) }
            keep = strings.Contains(sVal, pinFilter)
        }
        if keep && ctxFilter != "" {
            sVal := ""
            if v, ok := es[i].Props["context"].(string); ok { sVal = strings.ToLower(v) }
            keep = strings.Contains(sVal, ctxFilter)
        }
        if keep { pairs = append(pairs, pair{ns[i], es[i]}) }
    }
    // Sorting
    switch sortParam {
    case "title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) < strings.ToLower(pairs[j].n.Title) })
    case "-title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) > strings.ToLower(pairs[j].n.Title) })
    case "-id":
        sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].n.ID > pairs[j].n.ID })
    case "id":
        fallthrough
    default:
        sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].n.ID < pairs[j].n.ID })
    }
    // Pagination
    limit := 20
    if lv := q.Get("limit"); lv != "" { if n, err := strconv.Atoi(lv); err == nil && n > 0 && n <= 100 { limit = n } }
    offset := 0
    if cur := q.Get("cursor"); cur != "" {
        if b, err := base64.URLEncoding.DecodeString(cur); err == nil {
            s := string(b)
            if strings.HasPrefix(s, "o:") {
                if n, err := strconv.Atoi(strings.TrimPrefix(s, "o:")); err == nil && n >= 0 { offset = n }
            }
        }
    } else if ov := q.Get("offset"); ov != "" { if n, err := strconv.Atoi(ov); err == nil && n >= 0 { offset = n } }
    start := offset
    if start > len(pairs) { start = len(pairs) }
    end := start + limit
    if end > len(pairs) { end = len(pairs) }
    slice := pairs[start:end]
    nodes := make([]dgraph.NodeDTO, 0, len(slice))
    edges := make([]dgraph.EdgeDTO, 0, len(slice))
    fieldsParam := q.Get("fields")
    var fs map[string]struct{}
    if fieldsParam != "" { fs = make(map[string]struct{}); for _, f := range strings.Split(fieldsParam, ",") { fs[strings.TrimSpace(f)] = struct{}{} } }
    for _, pr := range slice {
        ndto := nodeToDTO(pr.n)
        if fs != nil { ndto = filterNodeFields(ndto, fs) }
        nodes = append(nodes, ndto)
        edges = append(edges, edgeToDTO(pr.e))
    }
    // Count-only fast path
    if cv := strings.ToLower(q.Get("count_only")); cv == "true" || cv == "1" {
        writeJSON(w, http.StatusOK, map[string]any{"total": len(pairs)})
        return
    }
    resp := map[string]any{
        "nodes": nodes,
        "edges": edges,
        "total": len(pairs),
        "next_offset": end,
    }
    if end >= len(pairs) {
        resp["next_offset"] = nil
        resp["next_cursor"] = nil
    } else {
        resp["next_cursor"] = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("o:%d", end)))
    }
    w.Header().Set("X-Total-Count", strconv.Itoa(len(pairs)))
    writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleNodeCites(w http.ResponseWriter, r *http.Request, id string) {
    ns, es := s.store.GetOutgoingCitations(id)
    q := r.URL.Query()
    labelsParam := q.Get("labels")
    pinFilter := strings.ToLower(q.Get("pin_cite_contains"))
    ctxFilter := strings.ToLower(q.Get("context_contains"))
    sortParam := q.Get("sort") // title|-title|id|-id
    haveFilter := labelsParam != ""
    labelSet := make(map[string]struct{})
    if haveFilter { for _, l := range strings.Split(labelsParam, ",") { labelSet[strings.TrimSpace(l)] = struct{}{} } }
    type pair struct{ n *dgraph.Node; e *dgraph.Edge }
    pairs := make([]pair, 0, len(ns))
    for i := range ns {
        keep := true
        if haveFilter {
            keep = false
            for _, l := range ns[i].Labels { if _, ok := labelSet[l]; ok { keep = true; break } }
        }
        if keep && pinFilter != "" {
            sVal := ""
            if v, ok := es[i].Props["pin_cite"].(string); ok { sVal = strings.ToLower(v) }
            keep = strings.Contains(sVal, pinFilter)
        }
        if keep && ctxFilter != "" {
            sVal := ""
            if v, ok := es[i].Props["context"].(string); ok { sVal = strings.ToLower(v) }
            keep = strings.Contains(sVal, ctxFilter)
        }
        if keep { pairs = append(pairs, pair{ns[i], es[i]}) }
    }
    switch sortParam {
    case "title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) < strings.ToLower(pairs[j].n.Title) })
    case "-title":
        sort.SliceStable(pairs, func(i, j int) bool { return strings.ToLower(pairs[i].n.Title) > strings.ToLower(pairs[j].n.Title) })
    case "-id":
        sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].n.ID > pairs[j].n.ID })
    case "id":
        fallthrough
    default:
        sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].n.ID < pairs[j].n.ID })
    }
    limit := 20
    if lv := q.Get("limit"); lv != "" { if n, err := strconv.Atoi(lv); err == nil && n > 0 && n <= 100 { limit = n } }
    offset := 0
    if cur := q.Get("cursor"); cur != "" {
        if b, err := base64.URLEncoding.DecodeString(cur); err == nil {
            s := string(b)
            if strings.HasPrefix(s, "o:") {
                if n, err := strconv.Atoi(strings.TrimPrefix(s, "o:")); err == nil && n >= 0 { offset = n }
            }
        }
    } else if ov := q.Get("offset"); ov != "" { if n, err := strconv.Atoi(ov); err == nil && n >= 0 { offset = n } }
    start := offset
    if start > len(pairs) { start = len(pairs) }
    end := start + limit
    if end > len(pairs) { end = len(pairs) }
    slice := pairs[start:end]
    nodes := make([]dgraph.NodeDTO, 0, len(slice))
    edges := make([]dgraph.EdgeDTO, 0, len(slice))
    fieldsParam := q.Get("fields")
    var fs map[string]struct{}
    if fieldsParam != "" { fs = make(map[string]struct{}); for _, f := range strings.Split(fieldsParam, ",") { fs[strings.TrimSpace(f)] = struct{}{} } }
    for _, pr := range slice {
        ndto := nodeToDTO(pr.n)
        if fs != nil { ndto = filterNodeFields(ndto, fs) }
        nodes = append(nodes, ndto)
        edges = append(edges, edgeToDTO(pr.e))
    }
    if cv := strings.ToLower(q.Get("count_only")); cv == "true" || cv == "1" {
        writeJSON(w, http.StatusOK, map[string]any{"total": len(pairs)})
        return
    }
    resp := map[string]any{
        "nodes": nodes,
        "edges": edges,
        "total": len(pairs),
        "next_offset": end,
    }
    if end >= len(pairs) {
        resp["next_offset"] = nil
        resp["next_cursor"] = nil
    } else {
        resp["next_cursor"] = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("o:%d", end)))
    }
    w.Header().Set("X-Total-Count", strconv.Itoa(len(pairs)))
    writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    root := q.Get("root")
    if root == "" {
        writeError(w, http.StatusBadRequest, "bad_request", "root is required", nil)
        return
    }
    depth := 1
    if d := q.Get("depth"); d != "" {
        if n, err := strconv.Atoi(d); err == nil && n >= 0 { depth = n }
    }
    labelsParam := q.Get("labels")
    lf := make(map[string]struct{})
    if labelsParam != "" {
        for _, l := range strings.Split(labelsParam, ",") { lf[strings.TrimSpace(l)] = struct{}{} }
    }
    ns, es, err := s.store.SliceFromRoot(root, depth, lf)
    if err != nil {
        writeError(w, http.StatusNotFound, "not_found", err.Error(), nil)
        return
    }
    nodes := make([]dgraph.NodeDTO, 0, len(ns))
    for _, n := range ns { nodes = append(nodes, nodeToDTO(n)) }
    edges := make([]dgraph.EdgeDTO, 0, len(es))
    for _, e := range es { edges = append(edges, edgeToDTO(e)) }
    writeJSON(w, http.StatusOK, dgraph.GraphSliceDTO{Nodes: nodes, Edges: edges})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    query := strings.TrimSpace(q.Get("q"))
    jur := q.Get("jurisdiction")
    code := q.Get("code")
    sortParam := q.Get("sort") // title|-title|id|-id
    limit := 20
    if lv := q.Get("limit"); lv != "" { if n, err := strconv.Atoi(lv); err == nil && n > 0 && n <= 100 { limit = n } }
    offset := 0
    if cur := q.Get("cursor"); cur != "" {
        if b, err := base64.URLEncoding.DecodeString(cur); err == nil {
            s := string(b)
            if strings.HasPrefix(s, "o:") {
                if n, err := strconv.Atoi(strings.TrimPrefix(s, "o:")); err == nil && n >= 0 { offset = n }
            }
        }
    } else if ov := q.Get("offset"); ov != "" { if n, err := strconv.Atoi(ov); err == nil && n >= 0 { offset = n } }
    // get more than we need to compute next_cursor
    cap := offset + limit
    if cap < limit { cap = limit }
    results := s.store.Search(query, jur, code, cap)
    // sort
    switch sortParam {
    case "title":
        sort.SliceStable(results, func(i, j int) bool { return strings.ToLower(results[i].Title) < strings.ToLower(results[j].Title) })
    case "-title":
        sort.SliceStable(results, func(i, j int) bool { return strings.ToLower(results[i].Title) > strings.ToLower(results[j].Title) })
    case "-id":
        sort.SliceStable(results, func(i, j int) bool { return results[i].ID > results[j].ID })
    case "id":
        fallthrough
    default:
        sort.SliceStable(results, func(i, j int) bool { return results[i].ID < results[j].ID })
    }
    if offset > len(results) { offset = len(results) }
    end := offset + limit
    if end > len(results) { end = len(results) }
    page := results[offset:end]
    items := make([]dgraph.SearchItem, 0, len(page))
    for _, n := range page { items = append(items, dgraph.SearchItem{Type: "node", ID: n.ID, Title: n.Title}) }
    resp := dgraph.SearchResultDTO{Query: query, Items: items}
    if end < len(results) {
        resp.NextCursor = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("o:%d", end)))
    }
    writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/diff/")
    n, ok := s.store.GetNode(id)
    if !ok { writeError(w, http.StatusNotFound, "not_found", "Node not found", nil); return }
    var versions []dgraph.Version
    if n.Version != nil { versions = append(versions, *n.Version) }
    writeJSON(w, http.StatusOK, map[string]any{"id": n.ID, "versions": versions, "diff": ""})
}

func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/versions/")
    n, ok := s.store.GetNode(id)
    if !ok { writeError(w, http.StatusNotFound, "not_found", "Node not found", nil); return }
    var versions []dgraph.Version
    if n.Version != nil { versions = append(versions, *n.Version) }
    writeJSON(w, http.StatusOK, versions)
}

// Topics
func (s *Server) handleTopics(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/topics" {
        ts := s.store.GetTopics()
        out := make([]dgraph.NodeDTO, 0, len(ts))
        for _, n := range ts { out = append(out, nodeToDTO(n)) }
        writeJSON(w, http.StatusOK, map[string]any{"topics": out})
        return
    }
    // /topics/{id}
    id := strings.TrimPrefix(r.URL.Path, "/topics/")
    if id == r.URL.Path || id == "" {
        writeError(w, http.StatusBadRequest, "bad_request", "missing id", nil)
        return
    }
    topic, ok := s.store.GetNode(id)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "Topic not found", nil)
        return
    }
    ns, es := s.store.GetTopicAssociations(id)
    // include topic node in results
    nodes := make([]dgraph.NodeDTO, 0, len(ns)+1)
    nodes = append(nodes, nodeToDTO(topic))
    for _, n := range ns { nodes = append(nodes, nodeToDTO(n)) }
    edges := make([]dgraph.EdgeDTO, 0, len(es))
    for _, e := range es { edges = append(edges, edgeToDTO(e)) }
    writeJSON(w, http.StatusOK, dgraph.GraphSliceDTO{Nodes: nodes, Edges: edges})
}

func nodeToDTO(n *dgraph.Node) dgraph.NodeDTO {
    return dgraph.NodeDTO{
        ID: n.ID, Labels: n.Labels, Title: n.Title, Citation: n.Citation, Text: n.Text,
        Props: n.Props, Version: n.Version, Sources: n.Sources,
    }
}

func edgeToDTO(e *dgraph.Edge) dgraph.EdgeDTO {
    return dgraph.EdgeDTO{ID: e.ID, Type: e.EdgeType, FromID: e.FromID, ToID: e.ToID, Props: e.Props}
}

// filterNodeFields returns a copy of n where only keys in fields are preserved.
func filterNodeFields(n dgraph.NodeDTO, fields map[string]struct{}) dgraph.NodeDTO {
    var out dgraph.NodeDTO
    if _, ok := fields["id"]; ok { out.ID = n.ID }
    if _, ok := fields["labels"]; ok { out.Labels = n.Labels }
    if _, ok := fields["title"]; ok { out.Title = n.Title }
    if _, ok := fields["citation"]; ok { out.Citation = n.Citation }
    if _, ok := fields["text"]; ok { out.Text = n.Text }
    if _, ok := fields["props"]; ok { out.Props = n.Props }
    if _, ok := fields["version"]; ok { out.Version = n.Version }
    if _, ok := fields["sources"]; ok { out.Sources = n.Sources }
    return out
}

// Start starts the HTTP server on the given port after loading the example data.
func (s *Server) Start() error {
    mux := http.NewServeMux()
    s.Routes(mux)
    port := os.Getenv("PORT")
    if port == "" { port = "8080" }
    addr := ":" + port
    fmt.Printf("Listening on %s\n", addr)
    return http.ListenAndServe(addr, mux)
}
