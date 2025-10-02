# Phase 03 – Federal Coverage (TODO)

- [ ] Ingest US Code bulk XML (OLRC) and map to canonical IDs.
- [ ] Ingest CFR annual editions and eCFR snapshots; handle versioning across updates.
- [ ] Add Federal Register metadata for rulemaking context.
- [ ] Integrate opinions via CourtListener; link `CITES`/`INTERPRETS` to code sections.
- [ ] Provide cross-jurisdiction citation resolution (`USC` ↔ `CFR` ↔ CA codes where applicable).
- [ ] Optimize storage/indexes; plan search integration.
- [ ] Expand docs and examples in `API/docs`.
- [ ] Add Federal rules (FRCP, FRE, FRCrP, FRAP) and US Constitution (CONST) coverage.
- [ ] Add U.S. Sentencing Guidelines (USSG) and link to relevant statutes.
- [ ] Implement reverse citations: `GET /nodes/:id/citations`.
- [ ] Make `/sources` dynamic from config; expose coverage metadata.
