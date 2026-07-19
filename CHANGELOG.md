# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Detailed per-session progress lives in [`DEVLOG.md`](DEVLOG.md).

## [Unreleased]

### Added

- Contract-first API: `api/openapi.yaml` as the source of truth, with a
  generated Go server (`oapi-codegen`) and a generated TypeScript client
  (`openapi-typescript`).
- Go backend HTTP API for graph CRUD, run start/poll, and a node-type registry
  (`GET /node-types`), served with interactive docs at `/docs`.
- Temporal integration: a graph translator that validates and topologically
  orders nodes, plus a workflow and worker that execute the graph durably.
- Node types: `start`, `noop`, an allowlisted `http` activity node, a durable
  `delay` timer node, and a `condition` branch node with `true`/`false`
  handles.
- Frontend: a Next.js + React Flow canvas with an API-driven node palette and
  save/run against the typed client.
- Testing: Go unit tests, frontend Vitest, Playwright E2E, and OpenAPI contract
  conformance checks.
- Local stack: one-command `docker compose up` running Postgres, Temporal
  (server + Web UI), the backend, and the frontend.
- Tooling: GitHub Actions CI, a pinned reproducible `golangci-lint` setup, and
  versioned git hooks.
- Documentation: contributor guide, security policy, and a contract-first
  `docs/adding-a-node-type.md` guide.

[Unreleased]: https://github.com/madmmas/temflowral/commits/main
