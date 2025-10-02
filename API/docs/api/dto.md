# DTO Models

Common
- `id` is a canonical string ID (see `GRAPH_MODEL.md`).
- Timestamps are ISO-8601 (`YYYY-MM-DDTHH:MM:SSZ`).

NodeDTO
```json
{
  "id": "CA:CIV:T02:CH02:§3342",
  "labels": ["SECTION"],
  "title": "Section 3342. Dog bite liability",
  "citation": "CIV § 3342",
  "text": "...normalized section text...",
  "props": {"jurisdiction": "CA", "code": "CIV", "title_num": 2, "chapter_num": 2},
  "version": {"fetched_at": "2025-01-01T00:00:00Z", "effective_date": "2024-01-01", "hash": "sha256:..."},
  "sources": [{"name": "LegInfo", "url": "https://leginfo.legislature.ca.gov/...", "retrieved_at": "2025-01-01T00:00:00Z"}]
}
```

EdgeDTO
```json
{
  "id": "edge-1",
  "type": "PARENT_OF",
  "from_id": "CA:CIV:T02:CH02",
  "to_id": "CA:CIV:T02:CH02:§3342",
  "props": {"order": 10}
}
```

GraphSliceDTO
```json
{
  "nodes": [NodeDTO, ...],
  "edges": [EdgeDTO, ...],
  "total": 123,
  "next_offset": 40,
  "next_cursor": "bzo0MA=="
}
```

PathDTO
```json
{
  "nodes": ["CA", "CA:CIV", "CA:CIV:T02", "CA:CIV:T02:CH02", "CA:CIV:T02:CH02:§3342"],
  "edges": ["PARENT_OF", "PARENT_OF", "PARENT_OF", "PARENT_OF"]
}
```

SearchResultDTO
```json
{
  "query": "dog bite",
  "items": [
    {"type": "node", "id": "CA:CIV:T02:CH02:§3342", "title": "Section 3342. Dog bite liability", "snippet": "..."}
  ],
  "next_cursor": "opaque-token"
}
```

ErrorResponse
```json
{
  "error": {
    "code": "not_found",
    "message": "Node not found",
    "details": null
  }
}
```

SourceDescriptorDTO
```json
{
  "name": "US Code (OLRC)",
  "jurisdictions": ["US"],
  "codes": ["USC"],
  "kind": "bulk",
  "urls": ["https://uscode.house.gov/download/download.shtml"]
}
```
