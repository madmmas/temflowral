# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Detailed per-session progress lives in [`DEVLOG.md`](DEVLOG.md).

## [Unreleased]

### Added

- API auth baseline: optional `API_AUTH_TOKEN` requires `Authorization: Bearer`
  on OpenAPI routes (401 otherwise); SECURITY.md documents trust boundary (no
  tenant isolation), mTLS-at-proxy, and interpreter upgrade compatibility.
- ADR-001: canvas packaging decision — reference-only UI; consumers build
  against OpenAPI + `GET /node-types` (no npm package / embed service yet).
- Pre-run graph validation (`ValidateGraph`): `StartGraphRun` rejects unknown
  registry node types, cycles, and unreachable nodes with 409 before Temporal
  starts; `CreateGraph` rejects unregistered types with 400.
- Minimal node-config templating: `{{ nodes.<nodeId>.output.<path> }}` resolved
  at run time from active predecessor outputs (HTTP url/headers/body supported;
  rendered requests revalidated; wait configs reject templates).
- `childWorkflow` node type (`ChildWorkflowNodeConfig` / `NestedGraph`): run an
  inline nested graph as a Temporal child workflow and wait for its result.
  Nested `childWorkflow` nodes are rejected (depth capped at one).
- Optional `Node.taskQueue`: route an activity-backed node to a Temporal task
  queue other than the workflow default (for specialized workers). Rejected on
  workflow-native nodes.
- Optional `Node.activityOptions` (`ActivityOptions` / `RetryPolicy`): per-node
  Temporal timeout and retry overrides for activity-backed nodes (engine
  defaults remain 30s start-to-close and `maximumAttempts: 1`). Rejected on
  workflow-native nodes.
- `POST /runs/{runId}/signal`: deliver a named Temporal signal to a running
  graph. Rejects with 409 unless the workflow is currently blocked on a wait
  node whose signal name matches (Temporal `temflowral.currentWait` query).
- `wait` node type (`WaitNodeConfig`): suspend until a named Temporal signal
  arrives or a durable timeout elapses; branch via `received` / `timedOut`
  source handles.
- Optional `idempotencyKey` on `StartGraphRun` (`StartRunRequest`): repeating
  the same key for a graph returns the original run without starting another
  Temporal workflow.
- Durable graph/run store (`backend/internal/store`): Postgres via required
  `DATABASE_URL`, pluggable `Store` interface, in-memory only when
  `STORE_ALLOW_MEMORY=1` (tests/experiments). Compose creates a `temflowral`
  database alongside Temporal's.
- External node-type registry (`backend/pkg/nodetype`): register custom node
  types and Temporal activities at worker startup, shared with `GET /node-types`.
  `NodeType` now advertises fixed `outputHandles` and config-derived
  `outputHandlesFromConfig`.
- Contract-first API: `api/openapi.yaml` as the source of truth, with a
  generated Go server (`oapi-codegen`) and a generated TypeScript client
  (`openapi-typescript`).
- Go backend HTTP API for graph CRUD, run start/poll, and a node-type registry
  (`GET /node-types`), served with interactive docs at `/docs`.
- Temporal integration: a graph translator that validates and topologically
  orders nodes, plus a workflow and worker that execute the graph durably.
- Node types: `start`, `noop`, an allowlisted `http` activity node, a durable
  `delay` timer node, a `condition` branch node with `true`/`false` handles,
  a `wait` signal/timeout node with `received`/`timedOut` handles, and a
  `childWorkflow` nested-graph node.
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
