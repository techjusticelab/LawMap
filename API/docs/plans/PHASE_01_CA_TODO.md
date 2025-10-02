# Phase 01 – California Pilot (TODO)

- [ ] Implement LegInfo ingestion (`internal/etl/leginfo`): bulk preferred; fallback HTML scraper.
- [ ] Parse hierarchy (Code → Title → Div/Part → Chapter → Article → Section) and emit canonical IDs.
- [ ] Build graph upserts (`internal/services/graph`): `PARENT_OF`, `AMENDS`.
- [ ] Dedup by canonical citation; attach `sources[]` and effective dates.
- [ ] Expose endpoints: `GET /nodes/:id`, `GET /nodes/:id/children`, `GET /nodes/:id/parents`, `GET /graph`, `GET /search`.
- [ ] E2E tests with small CA sample; measure coverage.
- [ ] Basic diff endpoint `GET /diff/:id`.
- [ ] Schedule monthly refresh with per-source overrides.
- [ ] Add CA Constitution (CONS) and Rules of Court (CRC) ingestion.
- [ ] Add Attorney General Opinions (CA) ingestion; link via `CITES`/`INTERPRETS`.
- [ ] Classification: define initial topics and attach via `HAS_TOPIC`.
- [ ] Add `GET /sources` and `GET /topics*` endpoints + docs.
