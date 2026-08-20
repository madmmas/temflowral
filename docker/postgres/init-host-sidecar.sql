-- Host / sidecar Postgres bootstrap for temflowral.
-- Run against an existing Postgres instance when you do not use the bundled
-- `postgresql` compose service (see docs/sidecar.md).
--
-- Application metadata (graphs / runs) MUST live in a dedicated database.
-- Do not create these tables inside your primary application database.

SELECT 'CREATE DATABASE temflowral'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temflowral')\gexec

-- Temporal auto-setup also needs its own databases when you self-host Temporal
-- against this same Postgres instance (skip if using Temporal Cloud).
SELECT 'CREATE DATABASE temporal'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temporal')\gexec

SELECT 'CREATE DATABASE temporal_visibility'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temporal_visibility')\gexec
