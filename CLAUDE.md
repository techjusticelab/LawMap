# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LawMap is a graph-based API for US law, starting with California as a pilot. The project ingests legal documents from official sources, builds a citation graph with versioning, and exposes HTTP endpoints for traversal and search. The goal is to create open-source tooling for public defenders and improve access to legal information.

**Tech Stack**: Go 1.21+, Fiber HTTP framework
**Working Directory**: All commands run from `API/` unless noted otherwise

## Development Commands

```bash
# Setup
cd API && go mod tidy

# Run locally (default :8080)
cd API && go run ./cmd/api
# or use the helper script
cd API && ./scripts/dev.sh

# Build binary
cd API && go build -o bin/api ./cmd/api

# Testing
cd API && go test ./... -race -cover

# Linting
cd API && golangci-lint run
# or use the helper
cd API && ./scripts/lint.sh

# Docker (optional)
cd API && docker compose -f build/docker-compose.yml up --build
```

## Architecture

The codebase follows clean architecture with clear separation of concerns:

**Entry Point**: `cmd/api/main.go` → `internal/app` bootstraps dependencies and starts the server

**Core Layers**:
- `internal/domain/` — Core entities and business logic (graph, documents, sources, search)
- `internal/services/` — Business workflows (citations, nodes, edges, graph operations)
- `internal/repo/` — Repository interfaces and implementations (graph, blob, db, index)
- `internal/http/` — HTTP handlers and routing (Fiber-based)
- `internal/etl/` — Extract-Transform-Load pipelines per source (leginfo, cfr, courts, regs, scheduler)
- `internal/pkg/` — Shared utilities (parse, log, httpx, util)
- `internal/config/` — Configuration management
- `internal/test/` — Testing helpers, fixtures, e2e tests

**Data Flow**: ETL → Domain → Services → Repositories → Persistence
**API Flow**: HTTP → Handlers → Services → Repositories → Graph/DB

## Graph Model

The system represents legal documents as a labeled property graph:

**Canonical ID Format**: `jurisdiction:code:title:chapter:section`
Examples: `CA:CIV:T02:CH02:§3342`, `US:USC:T18:§924(e)`

**Node Labels**: `JURISDICTION`, `CODE`, `TITLE`, `CHAPTER`, `SECTION`, `OPINION`, `RULE`, `REGULATION`, `TOPIC`

**Edge Types**:
- Hierarchy: `PARENT_OF` (Chapter → Section)
- Lineage: `AMENDS`, `REPEALS` (newer → older)
- Citations: `CITES`, `INTERPRETS`, `HAS_TOPIC`
- Equivalence: `SAME_AS` (deduplication across sources)

**Versioning**: Each node tracks `version{fetched_at, effective_date, hash}` for change detection and diffs.

See `API/docs/model/` for detailed schemas.

## Key HTTP Endpoints

Base URL: `http://localhost:8080`

- `GET /health` — Health check
- `GET /nodes/:id` — Fetch single node (supports `?expand=parents|children`)
- `GET /nodes/:id/children` — Get child nodes (pagination, filtering by labels)
- `GET /nodes/:id/parents` — Get parent nodes
- `GET /nodes/:id/citations` — Reverse citations (what cites this node)
- `GET /nodes/:id/cites` — Outgoing citations (what this node cites)
- `GET /graph?root=:id&depth=N` — Graph slice with depth control
- `GET /search?q=text&jurisdiction=CA&code=CIV` — Search with filters
- `GET /topics` — List classification topics
- `GET /sources` — Enumerate configured sources

Query parameters: `labels`, `limit`, `offset`, `cursor`, `sort`, `fields`, `count_only`

Note: URL-encode `§` as `%C2%A7`

## Legal Sources & ETL

The system ingests from official sources documented in `API/source.md`:

**California**:
- Codes: LegInfo (leginfo.legislature.ca.gov) bulk downloads
- Constitution: CONS
- Rules of Court: CRC (courts.ca.gov)
- Case Law: CA appellate opinions, CourtListener
- Regulations: CCR (via OAL or Westlaw)

**Federal**:
- US Code: OLRC bulk XML, GovInfo
- CFR/eCFR: GovInfo, ecfr.gov
- Federal Rules: FRCP, FRE, FRCrP, FRAP
- Constitution: CONST
- Sentencing Guidelines: USSG (ussc.gov)
- Case Law: SCOTUS, Circuit courts, CourtListener

**ETL Philosophy**:
- Prefer official bulk downloads over scraping
- Respect robots.txt and rate limits
- Idempotent writes keyed on canonical IDs
- Track all `sources[]` that refer to same canonical citation
- Scheduler runs periodic refreshes (default monthly)

## Configuration

Start from `configs/config.example.yaml`. Configuration covers:
- Storage backends (graph, blob, index)
- Source definitions and schedules
- HTTP server settings

Never commit secrets; use environment variables or untracked `config.yaml`.

## Testing Strategy

- Unit tests: `*_test.go` files alongside implementation
- Table-driven tests for parsing/normalization logic
- E2E tests: `internal/test/e2e/` with fixtures in `internal/test/fixtures/`
- Target ≥80% coverage for touched packages
- Run with `-race` to catch concurrency issues

## Code Style

- Go standard formatting: `gofmt -s -w .` and `go vet ./...`
- Package names: lowercase, no underscores (graph, httpx, not graph_utils)
- Error handling: return `error` as last result, wrap with context, avoid panics in libraries
- Indentation: tabs (Go default), ~100 col line width
- Commits: Conventional Commits format (e.g., `feat(api): add reverse citations`, `fix(repo): handle empty edges`)

## Important Patterns

**Canonical ID Construction**: Use `internal/pkg/parse` for citation normalization. Always uppercase code keys (BPC, CIV, PEN), normalize section markers (`§`, ranges, subdivisions).

**Deduplication**: Merge nodes with same canonical ID; attach all `sources[]` for provenance. Prefer most recent `effective_date` for content.

**Cross-Layer Boundaries**: Respect clean architecture. Domain should not import repo or http. Services orchestrate across domain and repo. HTTP layer only maps requests/responses.

**ETL Idempotency**: Safe to re-run ingestion jobs. Writes keyed on canonical IDs; updates replace existing content with newer version metadata.

## Common Workflows

**Adding a new source**:
1. Create ETL module in `internal/etl/<source>/`
2. Implement fetch, parse, normalize to canonical IDs
3. Add source config to `configs/config.example.yaml`
4. Register in scheduler if periodic refresh needed
5. Add source metadata to `API/source.md`
6. Write unit tests for parser, e2e test for sample document

**Adding a new endpoint**:
1. Define request/response DTOs in `internal/http/dto/`
2. Add handler in `internal/http/routes/`
3. Wire service dependency from `internal/services/`
4. Update OpenAPI spec (if exists) and `docs/api/USAGE.md`
5. Add example curl command to docs

**Extending graph model**:
1. Add label to `docs/model/labels.md` or edge type to `docs/model/edge_types.md`
2. Update `internal/domain/graph/types.go`
3. Add properties to `docs/model/properties.md` if new metadata needed
4. Update repository interface and implementation
5. Add migration if schema changes required

## Reference Documentation

- `API/docs/DESIGN.md` — Overall architecture and design decisions
- `API/docs/GRAPH_MODEL.md` — Graph schema overview
- `API/docs/IMPLEMENTATION_STEPS.md` — Phased rollout plan
- `API/docs/api/USAGE.md` — API endpoint examples
- `API/docs/model/` — Detailed label, edge, property schemas
- `API/source.md` — Authoritative legal source URLs and notes
- `AGENTS.md` — Repository-level coding guidelines
