# temflowral devlog

Running log of progress. Newest entry on top. One entry per work session —
doesn't need to be daily.

---

## 2026-07-19 — Playwright scaffold against Prism (#18)

**Did:**
- Added Playwright 1.61 with a Chromium project and an isolated smoke spec for
  the graph editor/node palette.
- Configured Playwright to start pinned Prism and Next.js automatically,
  wiring `NEXT_PUBLIC_API_BASE_URL` to the mock. `API_BASE_URL` and
  `PLAYWRIGHT_BASE_URL` opt into pre-existing/real services.
- Added stable `data-testid` hooks for canvas interactions, test output
  ignores, npm/Makefile scripts, and frontend documentation.

**Decided / learned:**
- Prism remains a pinned `npx` command instead of a dev dependency, avoiding
  roughly 185 transitive packages and their extra advisories.
- The CI E2E job remains disabled as planned; this scaffold runs locally and
  is ready for the dedicated testing track to extend/enable later.
- Build and E2E checks must run sequentially because `next build` and
  `next dev` share `.next`; concurrent execution corrupts build manifests.

**Verified:**
- Vitest 15/15, ESLint, production build, and Playwright Chromium 1/1 pass.

**Next:**
- #19 contract conformance checks; #20 full graph → run UI happy path.

---

## 2026-07-19 — Save and run graphs from the canvas (#17)

**Did:**
- Added a typed React Flow → `CreateGraphRequest` serializer, including node
  type/label/position/config and edge handle mapping.
- Added graph-name, Save, and Run controls. Save calls `POST /graphs`; Run
  first saves the current canvas, calls `POST /graphs/{id}/run`, then polls
  `GET /runs/{id}` every 1.5 seconds until a terminal status.
- Added saved graph/run status and API error feedback, with controls disabled
  while requests are in flight or the canvas is empty.
- Added unit coverage for serialization, terminal statuses, and contract error
  extraction.

**Decided / learned:**
- Run always saves current canvas state first, avoiding stale graph IDs after
  edits.
- Polling cleanup cancels timers and ignores responses after unmount or run
  changes.
- Prism integration smoke passed: create `201`, run `202`, poll `200`; browser
  render showed the palette and Save/Run controls.

**Next:**
- #18 Playwright scaffold can exercise the full mocked UI workflow; backend
  execution is available for a real Temporal integration check.

---

## 2026-07-19 — Node palette from registry (#16)

**Did:**
- Added a `NodePalette` sidebar populated from `GET /node-types` via the typed
  client (`useNodeTypes`), grouped by category — no hardcoded node list.
- Added a generic custom `WorkflowNode` renderer (registered as React Flow
  node type `workflow`) driven by node data, plus drag-from-palette and
  click-to-add wiring on the canvas.
- Extended `createNode` to carry the backend node-type id/label/category; added
  a pure `groupByCategory` helper with Vitest coverage.

**Decided / learned:**
- All palette nodes render through one data-driven component rather than a
  per-type switch; type-specific renderers can be added as siblings later
  without touching canvas code.
- Palette fetch degrades gracefully: loading, error, and empty states render
  without breaking the canvas.
- Smoke-tested against the Prism mock (`/node-types` → 200) with the dev server
  pointed at `NEXT_PUBLIC_API_BASE_URL`.

**Next:**
- #17 serialize canvas to the `Graph` schema and save/run via the client.

---

## 2026-07-19 — React Flow canvas (#15)

**Did:**
- Added `@xyflow/react` and a base `GraphCanvas` client component with pan/zoom,
  add node, drag-to-connect, and select-then-Delete removal, plus Background,
  MiniMap, and Controls.
- Replaced the placeholder home page with a full-height canvas shell.
- Extracted pure node helpers (`createNode`/`nextNodeId`) into
  `src/lib/graph-canvas.ts` with Vitest coverage (canvas runtime stays in the
  client component, testable logic stays pure).

**Decided / learned:**
- No custom node types yet — that is #16 (palette driven by `GET /node-types`).
- Delete is wired via React Flow's `deleteKeyCode` (Backspace/Delete); the
  component is wrapped in `ReactFlowProvider` so `useReactFlow` works.

**Next:**
- #16 node palette + custom node rendering from the node-type registry.

---

## 2026-07-19 — Typed frontend API client (#14)

**Did:**
- Added `openapi-typescript` generation of `frontend/src/api/generated/schema.ts`
  from `api/openapi.yaml`, plus an `openapi-fetch` wrapper (`createApiClient`)
  that defaults to the Prism mock via `NEXT_PUBLIC_API_BASE_URL`.
- Wired `npm run generate` and extended root `make generate` to refresh both
  Go and TypeScript contract-derived code.
- Documented usage in `frontend/README.md`; ESLint ignores the generated tree.

**Decided / learned:**
- Prefer `openapi-typescript` + `openapi-fetch` over Orval for a thin typed
  client that matches the contract without generating a heavy SDK.
- Generated sources are committed (same policy as Go oapi-codegen output).

**Next:**
- #15 React Flow canvas; #16/#17 will consume `createApiClient`.

---

## 2026-07-19 — Graph → Temporal workflow translator (#12)

**Did:**
- Added a deterministic graph planner and `temflowral.graph` Temporal workflow
  that walks nodes in topological order and dispatches per-node-type activities.
- Implemented in-memory graph/run storage and wired `CreateGraph`, `GetGraph`,
  `StartGraphRun`, `GetRun`, and `ListNodeTypes` to the generated API.
- Registered a graph-compatible `noop` node activity; `start` remains a control
  node that seeds workflow input without an activity.

**Decided / learned:**
- Unsupported node types (including `http` until #21) and cyclic/unreachable
  graphs return `409` from `POST /graphs/{id}/run`.
- Fan-out is sequential for now, preserving `edges[]` order; conditional
  branching stays with #23.
- Graph/run state is process-local memory only — restarts clear it until a
  durable store lands.

**Next:**
- #21 HTTP activity node (with SSRF validation) once graph execution is in use.

---

## 2026-07-19 — Temporal client and worker wiring (#11)

**Did:**
- Added the Temporal Go SDK and connected the backend to a configurable
  Temporal service during startup.
- Registered a deterministic smoke workflow and activity on the `temflowral`
  task queue, with a workflow test that executes the activity end to end.
- Added graceful HTTP/worker shutdown and a `docker-compose.yml` Temporal dev
  server service, driven by `make temporal-dev` / `temporal-smoke` /
  `temporal-down`, so no local Temporal install is required.

**Decided / learned:**
- The Temporal SDK is pinned to `v1.46.0`, the latest stable release compatible
  with the project's Go baseline.
- Local defaults are `localhost:7233`, namespace `default`, and task queue
  `temflowral`; each can be overridden by environment variable.
- Temporalite is deprecated. Local development uses the CLI dev server via the
  pinned `temporalio/temporal:1.8.0` image in `docker-compose.yml` (issue #25
  extends this file with Postgres, backend, and frontend); a native CLI install
  remains a documented alternative.

**Next:**
- #12 translate saved graphs into Temporal workflow execution and replace the
  run endpoint placeholder.

---

## 2026-07-19 — Next.js frontend bootstrap (#13)

**Did:**
- Scaffolded `frontend/` with Next.js 15 App Router, TypeScript, Tailwind,
  ESLint, and `src/` layout via `create-next-app`.
- Added Vitest plus a small smoke test for `NEXT_PUBLIC_API_BASE_URL` defaults
  (Prism on `:4010`), matching Makefile/CI `npm test` expectations.
- Replaced the create-next-app marketing page with a minimal temflowral shell
  and documented local scripts in `frontend/README.md` / `.env.example`.

**Decided / learned:**
- Stay on Next.js `15.5.9` (patched for CVE-2025-66478). Next 16 breaks the
  FlatCompat ESLint config shipped by create-next-app; revisit when ready to
  adopt Next 16's native flat config.
- Vitest is pinned at `3.2.6+` for the UI-server advisory fix.
- Frontend work lives on `feat/13-frontend-bootstrap` so it can land in
  parallel with backend codegen (#10).

**Next:**
- #14 generate a typed API client from `api/openapi.yaml`.

---

## 2026-07-19 — Generated Go API server contract (#10)

**Did:**
- Added pinned `oapi-codegen` generation for Go models, standard-library HTTP
  routing, and strict request/response server interfaces.
- Wired every generated API route into the backend alongside the existing
  OpenAPI and Swagger UI routes.
- Added a checked-in generated source file, a `make generate` target, route
  integration tests, and a CI drift check.

**Decided / learned:**
- Application handlers implement the generated strict interface in
  `internal/server`; generated code remains isolated in `internal/api`.
- Until endpoint behavior lands in later backend issues, registered API routes
  return the contract's typed `500` response with a `not_implemented` code.
- OpenAPI contract changes now trigger backend CI because they can change the
  generated Go surface.

**Next:**
- #11 connect the Temporal client and local worker, then replace run endpoint
  placeholders as execution behavior lands.

---

## 2026-07-19 — OpenAPI validation in CI (#9)

**Did:**
- Added an `openapi-lint` CI job that runs whenever `api/openapi.yaml` or its
  lint configuration changes.
- Added an explicit Redocly recommended-rules configuration and validated the
  v0.1 contract against it.

**Decided / learned:**
- Redocly CLI is pinned to `2.39.0` for reproducible CI and supports the
  project's Node 20 baseline.
- The `no-server-example.com` rule is disabled intentionally while v0.1
  advertises only the local development backend.

**Next:**
- #10 generate Go server interfaces and types from the contract.

---

## 2026-07-19 — Contract-backed mock server (#8)

**Did:**
- Documented how to launch Prism from `api/openapi.yaml`, verify representative
  endpoints, and point frontend or Playwright work at the mock.
- Defined base-URL environment variable conventions so switching from Prism
  on port 4010 to the backend on port 8080 requires only a value change.
- Corrected the documented Go development version to match `backend/go.mod`.

**Decided / learned:**
- Prism is pinned to `5.14.2`, the newest checked release compatible with the
  project's Node 20 baseline. Prism `5.16.0` requires Node 24.18 or newer.
- Frontend uses `NEXT_PUBLIC_API_BASE_URL`; Node-based tests use
  `API_BASE_URL`.

**Next:**
- #9 add OpenAPI linting to CI.

---

## 2026-07-19 — Backend API documentation (#7)

**Did:**
- Added the first backend HTTP server, serving the raw contract at
  `GET /openapi.yaml` and Swagger UI at `GET /docs`.
- Added handler and contract-loading tests.
- Updated CI to derive its Go version from `backend/go.mod` now that backend
  jobs run against real Go source.
- Migrated the existing golangci-lint configuration to v2 syntax so the
  newly activated lint job can run.
- Fixed the go-lint CI job: `golangci-lint-action@v6` installs golangci-lint
  v1 (built with Go 1.24) which refuses a module targeting Go 1.25.7. Bumped
  to `@v8` pinned at golangci-lint `v2.12.2` (built with Go 1.25), and added
  the same pinned lint run to the pre-commit hook so local and CI agree.

**Decided / learned:**
- Swagger UI assets load from the version-5 unpkg CDN; the API contract remains
  the repository's single `api/openapi.yaml` file.
- The server finds the contract when run from either the repository root or
  `backend/`; deployments can set `OPENAPI_SPEC_PATH` explicitly.

**Next:**
- #8 document the Prism contract mock in CONTRIBUTING.md.
- #9 add OpenAPI linting to CI.

---

## 2026-07-16 — OpenAPI v0.1 contract (#6)

**Did:**
- Added `api/openapi.yaml` v0.1 with five endpoints (`POST/GET /graphs`,
  `POST /graphs/{id}/run`, `GET /runs/{id}`, `GET /node-types`) and core
  schemas (`Graph`, `Node`, `Edge`, `NodeType`, `Run`, `RunStatus`).
- Validated with Redocly CLI lint and Prism mock smoke test.

**Decided / learned:**
- OpenAPI 3.1; no auth in v0.1 (`security: []` declared explicitly).
- `Node.config` and `NodeType.configSchema` use open objects — intentional
  for per-node-type config; validated at runtime per node type registry.
- Prism serves from spec examples; downstream tracks can point at mock on
  port 4010 until #7/#8 land.

**Next:**
- #7 serve spec + Swagger UI from backend
- #8 document Prism mock in CONTRIBUTING.md
- #9 CI spec lint (Spectral/Redocly)

---

## 2026-07-15 — Cursor rules

Added `.cursor/rules/*.mdc` (project overview always-on + path-scoped rules for
`api/**`, `backend/**`, `frontend/**`, and Playwright test files) plus
`CURSOR_GUIDE.md`. Main point encoded in the rules: contract changes go through
`openapi.yaml` first, generated client/server code is never hand-edited, and the
Playwright track defaults to the Prism mock instead of the live backend.

---

## 2026-07-15 — Contract-first pivot

**Decided:** OpenAPI/Swagger spec is now the pivot point, not an afterthought.
Backend publishes `api/openapi.yaml` + Swagger UI before writing real handlers.
Frontend and Playwright (contributor 2) both build against a Prism mock server
generated from the spec instead of waiting on each other or on real backend logic.

**Reordered `ISSUES.md`:** new #2–#5 (author spec, serve it, mock server, CI spec
lint) now sit ahead of backend/frontend implementation work. Backend/frontend/testing
tracks can run in parallel once the contract exists. Old single-track ordering is
gone.

**Next session:** Write `api/openapi.yaml` v0.1 (#2) — `Graph`/`Node`/`Edge`/`Run`
schemas plus the five core endpoints. Fix the issue-template merge conflict (#1)
first since it's a five-minute fix.

---

## 2026-07-15 — Kickoff

**Status:** Day 0. Repo has governance scaffolding only, no application code.

**State of the repo:**
- `backend/go.mod` exists (Go 1.25.7), no `.go` source files yet.
- No `frontend/` directory yet.
- No `docker-compose.yml` yet, though `make run` and the README's own pitch assume it.
- CI (`.github/workflows/ci.yml`) is already built out and correctly skips jobs when
  there's no code to test. Its `e2e` job is commented out until "Day 13" — a pacing
  note worth keeping in mind.
- README/CONTRIBUTING/CHANGELOG/CODE_OF_CONDUCT are outlines, not written yet.

**Found:** `.github/ISSUE_TEMPLATE/bug_report.md` and `feature_request.md` have
unresolved git merge conflict markers committed to `main` (`<<<<<<< HEAD` etc.) —
filed as issue #1, fix before anything else.

**Plan:** Seeded `ISSUES.md` with the initial backlog (backend scaffold, frontend
scaffold, docker-compose, graph→Temporal translator, first node types, testing,
docs). Suggested build order at the bottom of that file.

**Next session:** Start on #1 (fix conflict markers) and #3 (`cmd/server/main.go`).

---

## Template for new entries

```
## YYYY-MM-DD — Short title

**Did:**
-

**Decided / learned:**
-

**Blocked on / open questions:**
-

**Next:**
-
```
