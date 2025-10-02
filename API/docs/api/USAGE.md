# API Usage & Sources

Base URL: `http://localhost:8080`

## Quick Calls
- Health: `GET /health`
- One node: `GET /nodes/{id}`
  - Expand parents: `GET /nodes/{id}?expand=parents`
  - Expand children: `GET /nodes/{id}?expand=children`
- Children: `GET /nodes/{id}/children`
  - Filter/limit: `GET /nodes/{id}/children?labels=SECTION&limit=10&offset=0`
- Parents: `GET /nodes/{id}/parents`
- Reverse citations: `GET /nodes/{id}/citations`
- Graph slice: `GET /graph?root={id}&depth=1[&labels=SECTION,CHAPTER]`
- Search: `GET /search?q=text[&jurisdiction=CA|US][&code=CIV|USC|CFR|...]`
- Versions/Diff: `GET /versions/{id}`, `GET /diff/{id}`
- Sources: `GET /sources` (enumerates configured/target sources)
- Topics: `GET /topics` and `GET /topics/{id}` (classification)

Note: encode `§` as `%C2%A7` in URLs.

## Examples (curl)
- CA Civil Code section:
  - `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342"`
- US Code (18 USC § 924(e)):
  - `curl "http://localhost:8080/nodes/US:USC:T18:%C2%A7924(e)"`
- CFR section (28 CFR § 600.4):
  - `curl "http://localhost:8080/nodes/US:CFR:T28:%C2%A7600.4"`
- CA Rules of Court (Rule 1.1):
  - `curl "http://localhost:8080/nodes/CA:CRC:rule_1.1"`
- US Sentencing Guidelines (§2D1.1):
  - `curl "http://localhost:8080/nodes/US:USSG:%C2%A72D1.1"`
- US Constitution (Amendment IV):
  - `curl "http://localhost:8080/nodes/US:CONST:AmdIV"`
- Federal Rules of Evidence (FRE 401):
  - `curl "http://localhost:8080/nodes/US:FRE:Rule_401"`
- Federal Rules of Criminal Procedure (Rule 11):
  - `curl "http://localhost:8080/nodes/US:FRCRP:Rule_11"`
- Federal Rules of Appellate Procedure (Rule 4):
  - `curl "http://localhost:8080/nodes/US:FRAP:Rule_4"`
- Graph slice with only sections below a chapter:
  - `curl "http://localhost:8080/graph?root=CA:CIV:T02:CH02&depth=1&labels=SECTION"`
- List topics and show associations:
  - `curl "http://localhost:8080/topics"`
  - `curl "http://localhost:8080/topics/TOPIC:Dogs"`
- Search with filters:
  - `curl "http://localhost:8080/search?q=dog+bite&jurisdiction=CA&code=CIV"`
- Search sort and cursor:
  - `curl "http://localhost:8080/search?q=dog&sort=title&limit=1"` → reuse `next_cursor` for next page
- Fields selection (trim payload):
  - `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342?fields=id,title,citation"`
- Reverse citations:
  - `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342/citations"`
  - Filter only opinions: `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?labels=OPINION"`
  - Paginate: `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?limit=1&offset=1"`
  - Cursor: `curl "http://localhost:8080/nodes/CA:CIV:T02:CH02:%C2%A73342/citations?limit=1"` → reuse `next_cursor` for next page
- Outgoing citations: `GET /nodes/{id}/cites`
  - `curl "http://localhost:8080/nodes/CA:OPN:People_v_Smith_2020_1/cites"`

## Agencies & Official Sources (ingestion targets)
- US Code: Office of the Law Revision Counsel (OLRC); GovInfo
- eCFR/CFR: eCFR; GovInfo (CFR annual editions)
- Federal Register: federalregister.gov
- U.S. Sentencing Commission (USSC): ussc.gov
- U.S. Courts: Supreme Court and appellate portals; CourtListener (opinions + RECAP)
- National Archives: Constitution (founding documents)
- California LegInfo: codes, constitution, bills/analyses
- California OAL: regulations and publications; CCR host
- California Courts: opinions, Rules of Court (Judicial Council)
- Agencies with policy/regulatory materials: CDCR, DMV, and others
  - Attorney General Opinions (CA): curated opinions; licensing varies

Use `jurisdiction` and `code` filters to scope queries:
- Jurisdictions: `US`, `CA`
- Common codes: `USC`, `CFR`, `FRCP`, `CONST`, `USSG`, `CIV`, `PEN`, `CRC`, `CCR`, `CONS`

See also: `API/docs/API.md`, `API/docs/GRAPH_MODEL.md`, and `API/source.md`.
