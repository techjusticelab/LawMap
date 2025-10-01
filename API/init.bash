#!/usr/bin/env bash
set -euo pipefail

# Project root is current directory
ROOT="."
cd "$ROOT"

mkd() { mkdir -p "$1"; }
mkf() { mkdir -p "$(dirname "$1")"; : > "$1"; }
mkreadme() {
  local dir="$1"
  local desc="$2"
  mkd "$dir"
  {
    echo "# ${dir##*/}"
    echo
    echo "$desc"
  } > "$dir/README.md"
}

# Top-level files
mkf "Makefile"
mkf "go.mod"
mkf "go.sum"
mkf ".gitignore"

mkreadme "." "Root of the Go Fiber graph-based API project. Provides a content graph of legal sources (nodes/edges) from which trees can be derived at query-time."

# cmd/api
mkreadme "cmd" "Entrypoints for executables."
mkreadme "cmd/api" "Fiber HTTP API entrypoint (main.go)."
mkf "cmd/api/main.go"

# internal
mkreadme "internal" "Private application code per Go convention."

# internal/app
mkreadme "internal/app" "Application bootstrap: DI wiring, loading the content graph, route mounting, lifecycle (start/stop)."

# internal/config
mkreadme "internal/config" "Configuration structs and loaders (env/yaml). Supports graph store/index configuration and feature flags."

# internal/http
mkreadme "internal/http" "Transport layer: handlers, DTOs, middleware, and error mapping."

mkreadme "internal/http/fiber" "Fiber-specific handlers and middleware. Handlers accept graph-centric params (node IDs, edge filters) and return graph responses."
mkreadme "internal/http/routes" "Route registration grouped by modules (graph, nodes, edges, search, admin)."
mkreadme "internal/http/dto" "DTOs for graph responses: NodeDTO, EdgeDTO, GraphSliceDTO, PathDTO; request models for traversals."

# internal/domain
mkreadme "internal/domain" "Domain layer. Core graph abstractions independent of transport and storage."

mkreadme "internal/domain/graph" "Graph primitives and services: Node, Edge, Labels, Properties; traversal/query contracts; tree materialization logic."
mkreadme "internal/domain/sources" "Domain-specific labels and schemas for legal sources (e.g., AuthorityType, Jurisdiction). Validation of node/edge semantics."
mkreadme "internal/domain/documents" "Document entity with source_id references, versions, citations; edges linking documents to source nodes."
mkreadme "internal/domain/search" "Search models for querying nodes/documents with filters (labels, properties, citations)."

# internal/repo
mkreadme "internal/repo" "Persistence adapters. Graph store, document DB, blob storage, and search index."

mkreadme "internal/repo/graph" "Graph repository interface and adapters (e.g., SQL+tables, SQLite, Badger, or external graph DB). Provides CRUD for nodes/edges and traversals."
mkreadme "internal/repo/db" "Relational persistence for documents/versions if needed (Postgres/SQLite) and migration helpers."
mkreadme "internal/repo/blob" "Blob storage (filesystem/S3) for source PDFs/HTML."
mkreadme "internal/repo/index" "Search index adapters (e.g., Meilisearch/Elasticsearch/Bleve) for full-text over documents."

# internal/etl
mkreadme "internal/etl" "ETL jobs ingesting sources and producing nodes/edges/documents. Jobs are idempotent and version-aware."

mkreadme "internal/etl/leginfo" "Ingest California LegInfo: create/update Code Title/Chapter/Section nodes; edges: PARENT_OF, CITES, AMENDS."
mkreadme "internal/etl/courts" "Ingest CA opinions: Opinion nodes; edges: INTERPRETS (to code sections), CITES (to cases/statutes)."
mkreadme "internal/etl/cfr" "Ingest CFR/eCFR: Regulation nodes; edges to Federal Register documents and US Code references."
mkreadme "internal/etl/regs" "Ingest CCR/OAL and agency policies where permissible; link to relevant code sections and topics."
mkreadme "internal/etl/scheduler" "Cron-like scheduler to run ETL jobs on intervals; supports ad-hoc runs."

# internal/services
mkreadme "internal/services" "Application services orchestrating graph operations, documents, search, and citations."

mkreadme "internal/services/graph" "High-level graph service: node/edge CRUD, traversals (parents/children/siblings), path finding, subgraph extraction, tree building."
mkreadme "internal/services/nodes" "Node-centric helpers: create/update nodes with label validation and property schemas."
mkreadme "internal/services/edges" "Edge-centric helpers: create/update edges, enforce allowed relations (e.g., PARENT_OF, CITES)."
mkreadme "internal/services/sources" "Domain service for legal sources: build trees from the content graph on demand; breadcrumbs; summaries."
mkreadme "internal/services/documents" "Document retrieval, sectioning, version diffs; link documents to source nodes."
mkreadme "internal/services/search" "Search across nodes/documents; filter by labels (statute/regulation/opinion) and properties."
mkreadme "internal/services/citations" "Citation parsing/normalization; resolve citations to node IDs and create CITES edges."

# internal/pkg
mkreadme "internal/pkg" "Shared utilities."

mkreadme "internal/pkg/log" "Logger utilities (slog/zerolog) with request-scoped fields."
mkreadme "internal/pkg/httpx" "HTTP client with retry/backoff for ETL."
mkreadme "internal/pkg/parse" "Parsers and text extraction helpers."
mkreadme "internal/pkg/util" "Generic helpers (IDs, slices, maps)."

# internal/test
mkreadme "internal/test" "Testing utilities and fixtures."

mkreadme "internal/test/fixtures" "Static fixtures: small LegInfo extracts, sample opinions, and a tiny graph snapshot."
mkreadme "internal/test/e2e" "End-to-end/API tests that spin Fiber in-memory and validate graph endpoints."

# migrations
mkreadme "migrations" "SQL migrations for relational stores (if using Postgres/SQLite for documents/indices)."

# scripts
mkreadme "scripts" "Dev scripts for linting, running, building, seeding, and importing graph data."
mkf "scripts/dev.sh"
mkf "scripts/lint.sh"
mkf "scripts/seed_graph.sh"
mkf "scripts/import_leginfo.sh"

# configs
mkreadme "configs" "Configuration files (YAML), prod/staging/dev overrides."
mkf "configs/config.example.yaml"

# build
mkreadme "build" "Containerization and deployment artifacts."
mkf "build/Dockerfile"
mkf "build/docker-compose.yml"

# docs
mkreadme "docs" "Documentation and samples."
mkf "docs/README.md"
mkf "docs/API.md"
mkf "docs/GRAPH_MODEL.md"
mkf "docs/EXAMPLES.graph.jsonl"

# Graph model scaffolding files
mkreadme "docs/model" "Reference for graph labels, edge types, and property keys used in the domain."

mkf "docs/model/labels.md"
mkf "docs/model/edge_types.md"
mkf "docs/model/properties.md"

# API contract stubs
mkreadme "docs/api" "HTTP API contract: graph endpoints and DTOs."
mkf "docs/api/graph_endpoints.md"
mkf "docs/api/dto.md"

echo "Graph-based API scaffold complete."
