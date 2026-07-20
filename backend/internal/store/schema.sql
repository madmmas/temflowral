-- Application schema for graph/run metadata. Temporal uses separate databases
-- (temporal / temporal_visibility) on the same Postgres instance; do not mix.

CREATE TABLE IF NOT EXISTS graphs (
    id UUID PRIMARY KEY,
    name TEXT,
    nodes JSONB NOT NULL,
    edges JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS runs (
    id UUID PRIMARY KEY,
    graph_id UUID NOT NULL REFERENCES graphs (id),
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    result JSONB,
    error TEXT,
    temporal_workflow_id TEXT NOT NULL,
    temporal_run_id TEXT NOT NULL
);
