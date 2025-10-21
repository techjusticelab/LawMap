-- Initial schema for LawMap graph database
-- PostgreSQL-optimized schema with JSONB for flexible properties

-- Nodes table: stores all graph nodes (jurisdictions, codes, sections, opinions, etc.)
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    labels TEXT[] NOT NULL,
    title TEXT,
    citation TEXT,
    text TEXT,
    props JSONB DEFAULT '{}'::jsonb,
    version_fetched_at TIMESTAMPTZ,
    version_effective_date DATE,
    version_hash TEXT,
    sources JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for nodes
CREATE INDEX IF NOT EXISTS idx_nodes_labels ON nodes USING GIN(labels);
CREATE INDEX IF NOT EXISTS idx_nodes_props ON nodes USING GIN(props);
CREATE INDEX IF NOT EXISTS idx_nodes_citation ON nodes(citation);
CREATE INDEX IF NOT EXISTS idx_nodes_text_search ON nodes USING GIN(to_tsvector('english', COALESCE(text, '')));

-- Add index for jurisdiction and code filtering
CREATE INDEX IF NOT EXISTS idx_nodes_jurisdiction ON nodes ((props->>'jurisdiction'));
CREATE INDEX IF NOT EXISTS idx_nodes_code ON nodes ((props->>'code'));

-- Edges table: stores relationships between nodes
CREATE TABLE IF NOT EXISTS edges (
    id TEXT PRIMARY KEY,
    edge_type TEXT NOT NULL,
    from_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    to_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    props JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for edges
CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_id);
CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_id);
CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(edge_type);
CREATE INDEX IF NOT EXISTS idx_edges_from_type ON edges(from_id, edge_type);
CREATE INDEX IF NOT EXISTS idx_edges_to_type ON edges(to_id, edge_type);

-- Unique constraint on edge relationships to prevent duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_edges_unique ON edges(from_id, to_id, edge_type);

-- Topics lookup table for fast topic queries
CREATE TABLE IF NOT EXISTS topics (
    id TEXT PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_topics_name ON topics(name);

-- Migration metadata table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    description TEXT
);

-- Insert initial migration record
INSERT INTO schema_migrations (version, description)
VALUES (1, 'Initial schema with nodes, edges, and topics tables')
ON CONFLICT (version) DO NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_nodes_updated_at BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Helper view for PARENT_OF relationships
CREATE OR REPLACE VIEW node_hierarchy AS
SELECT
    e.from_id as parent_id,
    e.to_id as child_id,
    (e.props->>'order')::INTEGER as order_num,
    p.title as parent_title,
    c.title as child_title
FROM edges e
JOIN nodes p ON p.id = e.from_id
JOIN nodes c ON c.id = e.to_id
WHERE e.edge_type = 'PARENT_OF';

-- Helper view for citation relationships
CREATE OR REPLACE VIEW citation_graph AS
SELECT
    e.from_id as citing_id,
    e.to_id as cited_id,
    e.props->>'pin_cite' as pin_cite,
    e.props->>'context' as context,
    citing.title as citing_title,
    citing.citation as citing_citation,
    cited.title as cited_title,
    cited.citation as cited_citation
FROM edges e
JOIN nodes citing ON citing.id = e.from_id
JOIN nodes cited ON cited.id = e.to_id
WHERE e.edge_type = 'CITES';
