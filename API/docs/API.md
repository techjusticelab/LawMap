# API Reference

Overview
- REST endpoints for browsing and traversing a legal knowledge graph.
- Start with California codes; expand to other jurisdictions.

Quick Start
- Base: `http://localhost:8080`
- Health: `GET /health`
- Example: `GET /nodes/CA:CIV:T02:CH02:§3342`

Docs Map
- Endpoints: `API/docs/api/graph_endpoints.md`
- Models/DTOs: `API/docs/api/dto.md`
- Graph model: `API/docs/GRAPH_MODEL.md`

Notes
- Pagination: `limit`, `cursor` when provided by an endpoint.
- IDs: canonical strings as defined in `GRAPH_MODEL.md`.
