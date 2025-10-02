# Repository Guidelines

This repository hosts a graph-based Go API in `API/` (Fiber HTTP server) with space reserved for a future client in `Web/`.

## Project Structure & Module Organization
- `API/cmd/api/` – entrypoint (`main.go`).
- `API/internal/` – application, domain, services, http, repo, etl, pkg, and test helpers.
- `API/configs/` – configuration files (see `config.example.yaml`).
- `API/scripts/` – dev helpers (`dev.sh`, `lint.sh`, imports/seeding).
- `API/build/` – Dockerfile and compose resources.
- `API/docs/` – API and graph model docs.
- `Web/` – placeholder for UI; not yet implemented.

## Build, Test, and Development Commands
- Initialize deps: `cd API && go mod tidy`
- Run locally: `cd API && go run ./cmd/api`
- Build binary: `cd API && go build -o bin/api ./cmd/api`
- Test all: `cd API && go test ./... -race -cover`
- Lint (recommended): `golangci-lint run` or `API/scripts/lint.sh`
- Docker (optional): `cd API && docker compose -f build/docker-compose.yml up --build`

## Coding Style & Naming Conventions
- Language: Go 1.21+; format with `gofmt -s -w .` and `go vet ./...`.
- Indentation: tabs (default Go style). Line width ~100 cols.
- Packages: short, lower-case, no underscores (e.g., `graph`, `httpx`).
- Files: `feature.go`, tests as `feature_test.go`.
- Errors: return `error` as last result; avoid panics in libraries; wrap with context.

## Testing Guidelines
- Use standard `testing` with table-driven tests.
- Name tests `TestXxx` per package; keep unit vs e2e separated (`API/internal/test/e2e`).
- Run coverage locally (`-cover`); target ≥80% for touched packages.
- Place fixtures under `API/internal/test/fixtures`.

## Commit & Pull Request Guidelines
- Commit messages: follow Conventional Commits.
  - Examples: `feat(api): add nodes traversal`, `fix(repo): handle empty edges`.
- PRs should include: clear description, linked issues, test coverage for changes, and updates to docs/configs as needed. Add screenshots or sample payloads for API changes.

## Security & Configuration Tips
- Start from `API/configs/config.example.yaml`; do not commit secrets. Prefer env vars or untracked `config.yaml`.
- Be cautious with large data/fixtures; keep samples minimal.

## Agent-Specific Instructions
- Keep changes focused; respect the directory layout above.
- Prefer small, composable packages in `API/internal/*` and avoid cross-layer leaks (domain ↔ repo ↔ http).
- When adding commands/scripts, place them in `API/scripts` and document usage in `API/README.md`.

