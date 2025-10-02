# Endpoints (v1)

Health
- `GET /health` → `200 {"ok": true}` when ready.

Sources
- `GET /sources` → `{ "sources": [SourceDescriptorDTO, ...] }`

Topics
- `GET /topics` → `{ "topics": [NodeDTO, ...] }`
- `GET /topics/:id` → GraphSliceDTO (topic node + associated nodes via `HAS_TOPIC` edges)

Nodes
- `GET /nodes/:id` → NodeDTO
  - Path: `:id` canonical ID
  - Query: `expand=parents|children` (optional), `fields=id,title,citation,...` (optional)
- `GET /nodes/:id/children` → GraphSliceDTO
  - Returns direct children nodes and `PARENT_OF` edges
  - Query: `labels=SECTION,CHAPTER` (optional), `fields=...` (optional), `sort=order|title|-title` (default `order`), `limit` (default 1000), `offset` (default 0) or `cursor`
  - Headers: `X-Total-Count` mirrors `total`
  - Query: `sort=order|title|-title` (default `order`)
- `GET /nodes/:id/parents` → PathDTO
  - Returns ancestry path to root
- `GET /nodes/:id/citations` → GraphSliceDTO
  - Returns nodes that cite `:id` and `CITES` edges
  - Query: `labels=OPINION,RULE` (optional), `fields=...` (optional), `pin_cite_contains=...`, `context_contains=...`
  - Query: `sort=title|-title|id|-id` (default `id`), `limit` (default 20), `offset` (default 0) or `cursor`, `count_only=true|false`
  - Headers: `X-Total-Count` mirrors `total`
- `GET /nodes/:id/cites` → GraphSliceDTO
  - Returns nodes cited by `:id` and `CITES` edges
  - Query: `labels=SECTION,OPINION,RULE` (optional), `fields=...` (optional), `pin_cite_contains=...`, `context_contains=...`
  - Query: `sort=title|-title|id|-id` (default `id`), `limit` (default 20), `offset` (default 0) or `cursor`, `count_only=true|false`
  - Headers: `X-Total-Count` mirrors `total`

Graph
- `GET /graph` → GraphSliceDTO
  - Query: `root=:id` (required), `depth=1..5` (default 1), `labels=SECTION,CHAPTER` (optional)

Search
- `GET /search` → SearchResultDTO
  - Query: `q=...` (required), `jurisdiction=CA|US` (optional), `code=CIV|PEN|...` (optional), `sort=title|-title|id|-id` (optional), `limit` (default 20), `offset` (optional), `cursor`

Diffs & Versions
- `GET /diff/:id` → `{ "id": string, "versions": [{"effective_date": string, "hash": string}], "diff": "..." }`
- `GET /versions/:id` → `[{"fetched_at": string, "effective_date": string, "hash": string}]`

Examples
```http
GET /nodes/CA:CIV:T02:CH02:§3342 HTTP/1.1
Host: localhost:8080

200 OK
{
  "id": "CA:CIV:T02:CH02:§3342",
  "labels": ["SECTION"],
  "title": "Section 3342. Dog bite liability",
  "citation": "CIV § 3342"
}
```

```http
GET /graph?root=CA:CIV:T02:CH02&depth=1 HTTP/1.1

200 OK
{
  "nodes": [ {"id": "CA:CIV:T02:CH02"}, {"id": "CA:CIV:T02:CH02:§3342"} ],
  "edges": [ {"type": "PARENT_OF", "from_id": "CA:CIV:T02:CH02", "to_id": "CA:CIV:T02:CH02:§3342"} ]
}
```
Responses include `total`, `next_offset`, and `next_cursor` for pagination when applicable.
