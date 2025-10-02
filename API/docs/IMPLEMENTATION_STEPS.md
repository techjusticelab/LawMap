# Implementation Steps (API)

1) Bootstrap & Config
- Ensure Go toolchain; `cd API && go mod tidy`.
- Define config structs in `internal/config` with `config.example.yaml` (storage, sources, scheduler).

2) Domain & Repo Skeleton
- Define graph primitives in `internal/domain/graph` (Node, Edge, Labels, Properties).
- Create repository interfaces in `internal/repo/graph` and an initial SQLite/Badger impl.

3) Canonical Citation & IDs
- Implement `internal/pkg/parse/citation.go` to normalize CA code keys (BPC, CIV, PEN, …) and sections.
- Build `CanonicalID(jurisdiction, code, title, chapter, section)`.

4) Ingestion: California Codes (LegInfo)
- Start with bulk downloads (preferred) or HTML scraper under `internal/etl/leginfo`.
- Parse hierarchy: Code → Title → Division/Part → Chapter → Article → Section.
- Emit normalized units with canonical IDs and source metadata.

5) Graph Builder
- Implement `internal/services/graph` to upsert nodes/edges (`PARENT_OF`, `CITES`, `AMENDS`).
- Idempotent writes by canonical ID; attach `sources[]`.

6) Versioning & Change Detection
- Track `version{ fetched_at, effective_date, hash }` on nodes.
- Re-ingest monthly by default; per-source schedule in config.

7) HTTP API (v1)
- Endpoints: `GET /health`, `GET /nodes/:id` (supports `expand=parents|children`), `GET /nodes/:id/children` (labels, sort, pagination), `GET /nodes/:id/parents`, `GET /nodes/:id/citations` (filters/sort/cursor/count_only), `GET /nodes/:id/cites` (filters/sort/cursor/count_only), `GET /graph`, `GET /search`, `GET /diff/:id`, `GET /versions/:id`, `GET /sources`, `GET /topics`, `GET /topics/:id`.
- Pagination metadata: responses include `total`, `next_offset`, and `next_cursor` where applicable.
- Maintain OpenAPI (`API/docs/openapi.yaml`) and examples. Framework-agnostic; Fiber or stdlib acceptable.

8) Testing & Fixtures
- Unit tests for citation parsing and ID generation.
- E2E tests in `internal/test/e2e` with a tiny CA sample in `internal/test/fixtures`.

9) Tooling & Ops
- Scripts in `API/scripts` (`dev.sh`, `lint.sh`, `import_leginfo.sh`).
- Optional Docker: `docker compose -f build/docker-compose.yml up --build`.

10) Extend Coverage (California)
- Add CA Constitution (CONS), Rules of Court (CRC), and Attorney General Opinions.
- Implement classification via `TOPIC` nodes + `HAS_TOPIC` edges.
- Ensure `CITES`/`INTERPRETS` links across codes/opinions.

11) Extend Coverage (Federal)
- Plan US Code (OLRC bulk XML), CFR (GovInfo) + eCFR, and Federal rules (FRCP, FRE, FRCrP, FRAP).
- Add US Constitution (CONST) and U.S. Sentencing Guidelines (USSG).
- Link SCOTUS/Circuit opinions via CourtListener.

12) Tooling & Docs Sync
- Keep OpenAPI and `USAGE.md` aligned with endpoints; CI validates spec and fixtures.
- Add reverse-citations endpoint plan (`GET /nodes/:id/citations`) and source coverage introspection for later phases.

References: See `API/source.md` for authoritative endpoints.
