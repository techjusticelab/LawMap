# PostgreSQL / Neon Setup Guide

This guide shows you how to set up LawMap API with your Neon PostgreSQL database.

## Quick Start with Neon

1. **Create `.env` file** in the `API/` directory:

```bash
cd API
cp .env.example .env
```

2. **Add your Neon connection string** to `.env`:

```bash
DATABASE_URL='postgresql://neondb_owner:your_password@ep-spring-field-a84psy6d-pooler.eastus2.azure.neon.tech/neondb?sslmode=require'
```

3. **Run the API** - it will automatically:
   - Detect the DATABASE_URL
   - Connect to PostgreSQL
   - Run migrations to set up the schema

```bash
go run ./cmd/api
```

## What Happens on First Run

The system will:
1. Connect to your Neon database
2. Run migrations from `migrations/001_initial_schema.sql`
3. Create tables: `nodes`, `edges`, `topics`, `schema_migrations`
4. Set up indexes for fast queries
5. Start the HTTP server on port 8080

## Database Schema

The migration creates the following tables:

**nodes** - Stores all graph nodes (jurisdictions, codes, sections, opinions)
- Indexed on labels, props (JSONB), jurisdiction, code
- Full-text search on text content

**edges** - Stores relationships between nodes
- Indexed on from_id, to_id, edge_type
- Unique constraint prevents duplicate edges

**topics** - Quick topic lookup table

**schema_migrations** - Tracks which migrations have been applied

## Environment Variables

### Required

- `DATABASE_URL` - Full PostgreSQL connection string from Neon

### Optional

- `PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: info)
- `MIGRATIONS_DIR` - Path to migrations directory (default: migrations)
- `SOURCES_FILE` - Path to sources config (default: configs/sources.example.json)

## Development Mode (No Database)

If you don't set `DATABASE_URL`, the API will use an in-memory store with example data:

```bash
# No DATABASE_URL = in-memory mode
unset DATABASE_URL
go run ./cmd/api
```

This is great for:
- Local development
- Testing
- Demos

## Adding Data to PostgreSQL

### Option 1: Use the HTTP API

```bash
# The API will be ready to accept data via POST endpoints (coming soon)
curl -X POST http://localhost:8080/nodes -d '{ "id": "CA", "labels": ["JURISDICTION"], ...}'
```

### Option 2: Use the ETL Pipeline

```go
// Run a LegInfo ETL job
store := graphrepo.NewPostgresStore(db)
fetcher := leginfo.NewFetcher(config)
parser := leginfo.NewParser()
loader := leginfo.NewLoader(store)

pipeline := &etl.Pipeline{
    Extractor: fetcher,
    Transformer: parser,
    Loader: loader,
}

result, err := pipeline.Run(ctx)
```

### Option 3: Load Example Data (For Testing)

Set this environment variable to load example data on startup:

```bash
LOAD_EXAMPLES=docs/EXAMPLES.graph.jsonl
```

## Verifying the Setup

### 1. Check the database connection

```bash
export DATABASE_URL='your_neon_connection_string'
go run ./cmd/api
```

You should see:
```
Using PostgreSQL backend (DATABASE_URL detected)
Connected to PostgreSQL database
Running database migrations...
Migrations completed successfully
LawMap API starting at 2025-...
Listening on :8080
```

### 2. Test the API

```bash
curl http://localhost:8080/health
# {"ok":true}
```

### 3. Check migrations were applied

Connect to your Neon database and run:

```sql
SELECT * FROM schema_migrations;
```

You should see version `1` applied.

## Troubleshooting

### "database password required"

Make sure your `DATABASE_URL` includes the password, or set individual environment variables:

```bash
DB_HOST=your-host.neon.tech
DB_USER=neondb_owner
DB_PASSWORD=your_password
DB_NAME=neondb
DB_SSLMODE=require
```

### "relation does not exist"

Migrations didn't run. Check:
1. `migrations/` directory exists
2. `001_initial_schema.sql` file is present
3. No errors in the startup logs

### Connection timeout

Neon databases may sleep after inactivity. The first request might be slow as the database wakes up - this is normal.

## Production Deployment

### 1. Set environment variables on your platform

**Heroku**:
```bash
heroku config:set DATABASE_URL='your_neon_url'
```

**Railway**:
Add `DATABASE_URL` in the Variables tab

**Docker**:
```bash
docker run -e DATABASE_URL='your_neon_url' lawmap-api
```

### 2. Ensure migrations run on deploy

The API runs migrations automatically on startup, so each deploy will apply any new migrations.

### 3. Monitor your database

- Check query performance in Neon console
- Monitor connection pool usage
- Set up alerts for slow queries

## Next Steps

- [ ] Implement POST/PUT endpoints for adding nodes/edges
- [ ] Set up ETL jobs to populate from LegInfo
- [ ] Configure connection pooling for production
- [ ] Add database backups
- [ ] Set up read replicas for scaling

## Security Notes

- ⚠️ **Never commit `.env` file** - it's in `.gitignore`
- ⚠️ **Never expose DATABASE_URL** in client-side code
- ⚠️ Use SSL mode `require` for production
- ⚠️ Rotate database credentials periodically
