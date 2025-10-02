# API Overview

HTTP contract for the graph service. See:
- Endpoints: `API/docs/api/graph_endpoints.md`
- DTO Models: `API/docs/api/dto.md`
- Usage & Sources: `API/docs/api/USAGE.md`

Base URL
- Local dev: `http://localhost:8080`

Versioning
- Header `Accept-Version: v1` (optional; defaults to `v1`).

Conventions
- All responses are JSON with UTF-8 encoding.
- Pagination uses `limit` and `cursor` when noted.
- IDs are canonical strings (see `GRAPH_MODEL.md`).
