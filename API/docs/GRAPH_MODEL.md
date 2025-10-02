# Graph Model

Canonical ID
- Format: `jurisdiction:code:title:chapter:section`
- Examples:
  - State root: `CA`
  - Code: `CA:CIV`
  - Chapter: `CA:CIV:T02:CH02`
  - Section: `CA:CIV:T02:CH02:§3342`

Labels & Edges
- Labels enumerate node types (see `API/docs/model/labels.md`).
- Edges capture relationships (see `API/docs/model/edge_types.md`).

Properties
- Standard keys and types in `API/docs/model/properties.md`.

Hierarchy Normalization
- Map jurisdiction-specific layers to a common ladder: `JURISDICTION → CODE → TITLE → CHAPTER → SECTION`.
- Preserve native layer names in `props` (e.g., `division_num`, `article_num`).

Versioning & Sources
- Nodes carry `version` block and `sources[]` for provenance and change detection.

Minimal Slice Example
```json
{
  "nodes": [
    {"id": "CA:CIV:T02:CH02", "labels": ["CHAPTER"], "title": "Chapter 2"},
    {"id": "CA:CIV:T02:CH02:§3342", "labels": ["SECTION"], "citation": "CIV § 3342"}
  ],
  "edges": [
    {"type": "PARENT_OF", "from_id": "CA:CIV:T02:CH02", "to_id": "CA:CIV:T02:CH02:§3342"}
  ]
}
```
