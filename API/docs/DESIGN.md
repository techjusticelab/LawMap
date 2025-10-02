# API Design Overview

## Goals & Scope
- Build an open, version-aware graph of US law. Start with California; extend to other states and federal sources.
- Ingest from official bulk downloads or public sites listed in `API/source.md`.
- Provide a stable HTTP API (Fiber) for graph traversal, search, and document lookups.

## Architecture
- Fetch: source-specific clients (API or scraper) under `internal/etl/<source>` with rate limits and polite crawling.
- Parse/Normalize: convert HTML/XML/PDF into structured units (code → title → part → chapter → section). Emit canonical citations and metadata.
- Graph Build: create nodes and edges with labels: `STATE`, `CODE`, `TITLE`, `CHAPTER`, `SECTION`, `OPINION`, etc. Relations: `PARENT_OF`, `CITES`, `AMENDS`, `REPEALS`, `INTERPRETS`.
- Store: repository interfaces in `internal/repo/graph` with a simple default (e.g., SQLite/Badger) and room for swapping implementations.
- API: handlers in `internal/http/fiber` expose read endpoints for nodes, edges, traversals, search, and diffs.
- Schedule: `internal/etl/scheduler` runs periodic refreshes (default monthly; per-source override).

## Identity & Deduplication
- Canonical ID: `jurisdiction:code:title:chapter:section` when applicable, e.g., `CA:CIV:T02:CH02:§3342`.
- Normalization: upper-case code keys (BPC, CIV, PEN), trim punctuation, normalize section markers (e.g., `§`, ranges, subdivisions).
- Merge Policy: prefer most recent effective/last-modified content; attach all `sources[]` that refer to the same canonical citation.

## Versioning & Change Detection
- Track `version{ fetched_at, effective_date, hash }` per node.
- Detect changes via ETag/Last-Modified when available; else hash normalized text.
- Maintain lineage edges `AMENDS`, `REPEALS`, and keep prior text for diffs.

## Classification & Reverse Citations
- Classification uses `TOPIC` nodes with `HAS_TOPIC` edges from statutes, regulations, and opinions to topical tags (e.g., TOPIC:Dogs).
- Reverse citation traversal collects inbound `CITES` edges to a node (e.g., opinions that cite a statute) and exposes them via `GET /nodes/:id/citations`.
- Topic browsing endpoints: `GET /topics` lists topics; `GET /topics/:id` returns the topic node and associated items.
- Store association edges sparsely with optional context (e.g., `pin_cite`).

## Error Handling & Observability
- Structured logging with request/source context. Metrics for fetch latency, items processed, dedup rate, failures.
- Idempotent ETL jobs: safe re-runs by keying writes on canonical IDs.

## Non-Goals (Pilot)
- Full-text search at scale and full federal coverage. Pilot focuses on CA codes + a small set of opinions.

## Compliance
- Respect robots.txt and site terms. Prefer official bulk endpoints (LegInfo downloads, US Code bulk XML, GovInfo).
