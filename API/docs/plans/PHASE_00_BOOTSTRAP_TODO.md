# Phase 00 – Bootstrap & Infra (TODO)

- [ ] Define Go version; init `go.mod` with modules.
- [ ] Implement config loader (`internal/config`) and `config.example.yaml` fields.
- [ ] Set up logging (`internal/pkg/log`) and error handling patterns.
- [ ] Define domain model (`internal/domain/graph`: Node, Edge, Labels, Properties).
- [ ] Create repo interfaces + minimal impl (`internal/repo/graph`).
- [ ] Wire app startup (`internal/app`) and HTTP scaffolding (Fiber).
- [ ] Developer scripts in `API/scripts` (`dev.sh`, `lint.sh`).
- [ ] Sample fixtures under `internal/test/fixtures`.
- [ ] Author OpenAPI (`API/docs/openapi.yaml`) + `USAGE.md`; add CI to lint spec and validate fixtures.
