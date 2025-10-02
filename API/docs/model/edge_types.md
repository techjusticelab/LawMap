# Edge Types

Hierarchy
- `PARENT_OF(from: parent, to: child)` – containment (e.g., Chapter → Section).

Lineage
- `AMENDS(from: newer, to: older)` – newer text amends older.
- `REPEALS(from: repealer, to: repealed)` – repealing relationship.

Citations & Semantics
- `CITES(from: citing, to: cited)` – textual citation.
- `INTERPRETS(from: opinion, to: section)` – judicial interpretation of a section.
- `SAME_AS(from: a, to: b)` – canonical equivalence across sources.
- `HAS_TOPIC(from: item, to: topic)` – classification linking a node to a `TOPIC`.

Notes
- Directionality matters for lineage; for hierarchy, traversal inverts easily.
- Keep edges sparse and typed; decorate with `props` when ordering or context is needed (e.g., `order`, `pin_cite`).
