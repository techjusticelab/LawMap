# Readiness Checklist (Pre‑Implementation)

## Scope & Acceptance
- [ ] Pilot scope fixed (CA Codes + minimal opinions) and out‑of‑scope noted.
- [ ] Definition of Done, success metrics, and demo scenarios written.

## Sources & Compliance
- [ ] Primary sources confirmed (bulk vs HTML) with URLs and formats.
- [ ] Robots.txt/terms reviewed; attribution/disclaimer language drafted.

## Canonicalization
- [ ] ID scheme finalized with examples and collision rules.
- [ ] Normalization rules for citations/sections (e.g., §, (a)(1)(A)(i)).
- [ ] Crosswalk for code key aliases and renumbers (initial CA set).

## Versioning & Dedup
- [ ] Change detection per source (ETag/Last‑Modified or hashing).
- [ ] Version record fields agreed; retention policy defined.
- [ ] Dedup/merge policy by canonical citation + text hash; source precedence.

## Storage & Ops
- [ ] Initial store chosen (SQLite/Badger) + indexes (id, label, code, section).
- [ ] Backup/export approach documented.

## Scheduling & Resilience
- [ ] Monthly refresh default; per‑source overrides.
- [ ] Idempotent, resumable ETL approach; checkpoints/runbooks.

## API & Contracts
- [ ] OpenAPI reviewed (endpoints, params, error model, pagination).
- [ ] Breaking change policy and versioning approach agreed.

## Quality & Observability
- [ ] Fixture set selected for CA; golden tests planned.
- [ ] Logging, metrics, and basic SLOs defined.

## Security
- [ ] Config keys settled; no secrets in repo; env strategy defined.
- [ ] Rate limits and request size/time caps documented.
