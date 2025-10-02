# Endpoint Improvements (TODO)

- [x] Add label filtering, sorting, pagination to `GET /nodes/:id/children`.
- [x] Add reverse citations `GET /nodes/:id/citations` with filters, sorting, `count_only`, and `cursor`.
- [x] Add outgoing citations `GET /nodes/:id/cites` with filters, sorting, `count_only`, and `cursor`.
- [x] Return pagination metadata: `total`, `next_offset`, `next_cursor`.
- [x] Support `expand=parents|children` on `GET /nodes/:id`.
- [ ] Add `fields` selector to trim NodeDTO payload (e.g., `fields=id,title,citation`).
- [ ] Add `X-Total-Count` header mirror where `total` is present.
- [ ] Add `sort` to `/search` and include `next_cursor`.
- [ ] Add `GET /nodes/:id/citers` alias (human-friendly); keep `/citations`.
- [ ] Validate params and return `400` on invalid `sort`/`limit`.
- [ ] Add OpenAPI examples and contract tests to validate docs vs handlers.
