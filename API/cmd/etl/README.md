# ETL Command

The ETL command populates the LawMap database with legal data from various sources.

## Usage

```bash
go run ./cmd/etl [flags]
```

## Flags

- `--source` - ETL source (default: "leginfo")
  - `leginfo` - California Legislative Information

- `--jurisdiction` - Jurisdiction code (default: "CA")
  - `CA` - California
  - `US` - Federal

- `--code` - Specific code to fetch (default: "" for all)
  - `CIV` - Civil Code
  - `PEN` - Penal Code
  - `BPC` - Business and Professions Code
  - etc.

- `--dry-run` - Parse and validate without loading to database (default: false)

## Examples

### Dry run to test parsing (no database required)
```bash
go run ./cmd/etl --dry-run --source leginfo --jurisdiction CA
```

### Load California Civil Code into PostgreSQL
```bash
# Ensure DATABASE_URL is set in .env file
go run ./cmd/etl --source leginfo --jurisdiction CA --code CIV
```

### Load all California codes
```bash
go run ./cmd/etl --source leginfo --jurisdiction CA
```

## Environment Variables

The ETL command reads from the `.env` file in the API directory:

```bash
DATABASE_URL='postgresql://user:pass@host/db?sslmode=require'
LOG_LEVEL=info
```

## How It Works

1. **Extract** - Fetches raw data from the source (e.g., leginfo.legislature.ca.gov)
2. **Transform** - Parses HTML/XML into graph nodes and edges
3. **Load** - Upserts nodes and edges into PostgreSQL

The ETL respects rate limits and includes retry logic for reliability.

## Current Status

**Note**: The LegInfo fetcher and parser are currently placeholder implementations.
They demonstrate the ETL interface but don't yet parse real data.

To populate your database with example data, see `POSTGRES_SETUP.md`.

## Next Steps

- Implement HTML parsing for LegInfo sections
- Add bulk download support
- Implement incremental updates (detect changes)
- Add more sources (US Code, CFR, court opinions)
