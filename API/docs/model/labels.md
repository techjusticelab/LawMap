# Node Labels

Structural
- `JURISDICTION` – US federal or state (e.g., `US`, `CA`).
- `CODE` – Named body of law (e.g., `CIV`, `PEN`, `USC`).
- `TITLE` – Title/Division/Part layer (normalized as needed).
- `CHAPTER` – Chapter/Article layer.
- `SECTION` – The atomic statutory/regulatory unit.

Judicial/Regulatory
- `OPINION` – Court opinion/document.
- `RULE` – State rules of court or similar.
- `REGULATION` – Administrative rule (e.g., CFR/CCR section).

Metadata
- `TOPIC` – Optional taxonomy node for grouping.

Notes
- Many jurisdictions have multiple intermediate layers (e.g., Division, Part, Article). Normalize to `TITLE` and `CHAPTER` where possible and preserve original layer names in `props`.
