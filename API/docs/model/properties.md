# Property Keys

Common Node Properties
- `jurisdiction: string` – `US`, `CA`, etc.
- `code: string` – `CIV`, `PEN`, `USC`, etc.
- `title_num: number` – numeric when applicable.
- `chapter_num: number` – numeric when applicable.
- `section_num: string` – may include letters/subdivisions.
- `citation: string` – human-readable citation (`CIV § 3342`).
- `name: string` – official name/title.
- `text: string` – normalized text of the unit (for `SECTION`, `REGULATION`, `RULE`, `OPINION` excerpts).
- `effective_date: string` – `YYYY-MM-DD` when known.

Versioning (on nodes)
- `version.fetched_at: string`
- `version.effective_date: string`
- `version.hash: string`

Edge Properties
- `order: number` – child ordering within a parent.
- `pin_cite: string` – pinpoint citation for `CITES` (`"§ 3342(b)"`).
- `context: string` – free-form note for `INTERPRETS`/`CITES`.

Source Metadata (on nodes)
- `sources[]: { name: string, url: string, retrieved_at: string }`

Notes
- Prefer normalized numeric fields where applicable and keep human-readable variants as `citation`/`name`.
