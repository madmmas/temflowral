# temflowral devlog

Running log of progress. Newest entry on top. One entry per work session —
doesn't need to be daily.

---

## 2026-07-19 — Fill in placeholder docs (#27)

**Did:**
- Rewrote `README.md` from an outline into real prose: pitch, architecture
  diagram, `docker compose up` quickstart, how-it-works, a node-types table,
  the extensibility hook, contributing/license sections, and an MIT badge.
- Replaced the numbered outline header in `CONTRIBUTING.md` with prose for
  development setup, git hooks, branch workflow, running tests, and adding a
  node type; appended commit-message conventions, a PR checklist, and a
  questions section.
- Rewrote `CHANGELOG.md` in Keep a Changelog format with an `Unreleased`
  section summarizing everything delivered so far.

**Decided / learned:**
- Kept the already-written CONTRIBUTING sections (linting, mock API, full
  stack, backend) and only converted the outline parts, avoiding churn.
- Node-types table and endpoints were sourced from `ListNodeTypes` and
  `api/openapi.yaml` so the docs match the implementation.
- CHANGELOG stays pre-1.0 under `Unreleased`; `DEVLOG.md` remains the detailed
  per-session log and is linked from both README and CHANGELOG.

**Verified:**
- `markdownlint-cli2` passes for `README.md`, `CONTRIBUTING.md`, and
  `CHANGELOG.md`.
- All relative repository links resolve to existing files.

**Next:**
- Backlog in `docs/issues/ISSUES.md` is complete through #27.

---

## 2026-07-19 — Guide for adding a node type (#26)

**Did:**
- Added `docs/adding-a-node-type.md`, the implementation guide referenced by
  the feature request template.
- Documented the contract-first path from an OpenAPI config schema through
  generated Go/TypeScript types, strict backend validation, node discovery,
  Temporal execution, frontend rendering/configuration, and tests.
- Used the HTTP node as the complete activity-backed example and the delay and
  condition nodes as workflow-native timer/branch examples.
- Added security, verification, and pull-request checklists.

**Decided / learned:**
- `Node.config` remains generic for graph storage, so the OpenAPI `anyOf`
  documents and generates config types while `ValidateNodeConfig` enforces the
  authoritative node-type/config pairing.
- Palette discovery and basic rendering are generic, but configured nodes still
  need editor UI and branching nodes need renderers with named handles.
- The OpenAPI config schema and the manually assembled `GET /node-types`
  `ConfigSchema` are intentional duplication and must remain aligned.

**Verified:**
- Every referenced path, symbol, command, and current frontend limitation was
  checked against the implementation.
- Markdown formatting and repository links were checked locally.

**Next:**
- #27 fill in placeholder project documentation.

---

## 2026-07-19 — Full docker-compose development stack (#25)

**Did:**
- Rewrote `docker-compose.yml` into a one-command stack: Postgres, Temporal
  (`temporalio/auto-setup` backed by Postgres), Temporal Web UI, the backend,
  and the frontend, wired together with healthcheck-gated `depends_on`.
- Added `backend/Dockerfile` (multi-stage Go build → distroless static, bundling
  `api/openapi.yaml` via `OPENAPI_SPEC_PATH`) and a root `.dockerignore` to keep
  the repo-root build context small.
- Added `frontend/Dockerfile` (Next.js standalone output → slim non-root runner)
  plus `frontend/.dockerignore`, and enabled `output: "standalone"` in
  `next.config.ts`.
- Updated `make temporal-dev` to bring up Temporal + Web UI (Postgres follows via
  `depends_on`), refreshed `make run`/`temporal-down` comments, and documented
  the full-stack quickstart in `CONTRIBUTING.md`.

**Decided / learned:**
- Postgres is a real datastore here: Temporal persistence runs on it via
  auto-setup, so workflow state survives restarts. The backend graph store is
  still in-memory — a Postgres-backed store is a separate feature, out of scope
  for this infra issue.
- `NEXT_PUBLIC_API_BASE_URL` is baked at image build time and must point at the
  host-published backend port (`http://localhost:8080`), not the `backend`
  compose service name, because the client bundle runs in the browser.
- distroless "static" ships CA certificates, so outbound HTTPS from HTTP
  activity nodes works without adding a certs layer.

**Verified:**
- `docker compose config` is valid; all pinned image tags resolve.
- `docker compose build` succeeds for backend and frontend.
- `docker compose up` brings all five services healthy; `GET /node-types`,
  the frontend, and the Temporal Web UI all return 200, and a create → run →
  poll graph run completes through Temporal (Postgres-backed).

**Next:**
- #26 `docs/adding-a-node-type.md`.

---

## 2026-07-19 — Reproducible golangci-lint config (#24)

**Did:**
- Strengthened `backend/.golangci.yml` with explicit test/module/timeout
  behavior, full issue reporting, and HTTP/Temporal-relevant checks
  (`bodyclose`, `noctx`, `durationcheck`, and `nilerr`) in addition to the
  existing static analysis and formatting rules.
- Added `scripts/run-golangci-lint.sh`, pinned to the same v2.12.2 used in CI.
  It prefers a matching installed binary and falls back to `go run`, so
  `make lint` works without a separate golangci-lint installation.
- Split `make lint` into backend/frontend targets and made the pre-commit hook
  delegate to the shared runner instead of maintaining duplicate execution
  logic.
- Documented local linting/config verification and marked the CI pin that must
  remain aligned with the runner.

**Decided / learned:**
- `noctx` remains enforced for production code but is excluded for `_test.go`;
  `httptest.NewRequest` intentionally creates synthetic requests and the rule
  produced only test-helper noise there.
- Module downloads are readonly during lint, preventing a lint run from
  modifying `go.mod`/`go.sum`.

**Verified:**
- Pinned `golangci-lint config verify` passes.
- `make lint` passes backend (0 issues) and frontend ESLint.
- `make test` passes Go race tests and frontend Vitest.
- Shell syntax checks pass for the shared runner and hook helper.

**Next:**
- #25 full docker-compose development stack.

---

## 2026-07-19 — Conditional/branch node (#23)

**Did:**
- Added a `condition` node type defined contract-first as `ConditionNodeConfig`
  (`field` + `equals`), referenced from `Node.config`, and regenerated
  Go/TypeScript models plus the live node-type registry entry.
- Taught the planner to require both `true` and `false` outgoing edges via
  `Edge.sourceHandle`, and the workflow to skip nodes on the untaken path while
  still joining after the taken branch.
- Evaluated conditions with JSON equality against the first active
  predecessor's top-level field — no expression language / nested paths, to
  avoid template-injection risk flagged in SECURITY.md.
- Added config, planner, true/false branch, and diamond-join workflow tests.

**Decided / learned:**
- Branching is edge-based (`sourceHandle`), not a second graph model: the topo
  plan still includes every node, but runtime path selection filters active
  inputs so only the taken branch executes.
- Missing fields take the false branch rather than failing the run, which keeps
  incomplete upstream payloads from hard-failing conditional graphs.

**Verified:**
- Redocly OpenAPI lint clean; generated clients reproducible.
- `go vet ./...`, `go test -race ./...`, and golangci-lint v2.12.2 (0 issues).
- Frontend ESLint, Vitest 15/15, production build, and contract conformance 6/6.

**Next:**
- #24/#25 cleanup & docker-compose; frontend condition config UI can follow once
  palette forms exist.

---

## 2026-07-19 — Durable delay/wait node (#22)

**Did:**
- Added a `delay` node type defined contract-first as `DelayNodeConfig`
  (`seconds`, 0–604800), referenced from generic `Node.config`, and regenerated
  Go/TypeScript models plus the live node-type registry entry.
- Handled the delay inside `GraphWorkflow` with `workflow.Sleep` — a durable
  Temporal timer that survives worker restarts — rather than an activity, and
  taught the planner/validator about workflow-handled (non-activity) node types.
- Validated delay config at graph creation and plan build, including explicit
  presence of `seconds` (a required non-pointer field would otherwise decode to
  0 silently).
- Added planner, workflow-timer (asserts a timer actually fires), config, and
  registry tests.

**Decided / learned:**
- Durable timers must run in workflow code, so `delay` is dispatched in the
  workflow switch alongside `start` instead of via `activityByNodeType`;
  `isExecutableNodeType` now gates planning for both activity and
  workflow-handled types.
- `seconds` uses `format: double` so oapi-codegen emits `float64`, matching
  `Position` and allowing sub-second waits.

**Verified:**
- Redocly OpenAPI lint clean; generated clients reproducible.
- `go vet ./...`, `go test -race ./...`, and golangci-lint v2.12.2 (0 issues).
- Frontend ESLint, Vitest 15/15, production build, and contract conformance 6/6.

**Next:**
- #23 conditional/branch node to exercise edge-based branching in the translator.

---

## 2026-07-19 — Allowlisted HTTP activity node (#21)

**Did:**
- Defined `HttpNodeConfig` contract-first in `api/openapi.yaml`, referenced it
  from generic `Node.config`, tightened the registry JSON Schema, and
  regenerated Go/TypeScript models.
- Added HTTP-node config validation at graph creation and execution-plan build,
  registered the activity with Temporal, and returned bounded status/body/
  content-type results for successful 2xx responses.
- Added a deny-by-default `HTTP_ALLOWED_HOSTS` policy with exact hostname
  matching, dial-time DNS/IP validation, private/loopback/link-local blocking,
  redirect revalidation, disabled environment proxies, 20-second requests,
  1 MiB request/response limits, 64 KiB response headers, and restricted
  request headers.
- Expanded planner, workflow dispatch, API registry/config, activity, SSRF,
  payload-bound, and error-redaction tests; documented operator setup and the
  security boundary.

**Decided / learned:**
- Template interpolation is deliberately absent. If added later, the fully
  rendered request must pass the same URL/header policy before execution.
- Temporal activity retries are disabled (`MaximumAttempts: 1`) so POST/PATCH
  requests are not silently replayed.
- Destination errors do not include the underlying `net/http` message because
  it can contain sensitive URL query parameters.
- An empty allowlist keeps the node registered/discoverable but denies all
  outbound requests until an operator explicitly permits hosts.

**Verified:**
- Redocly OpenAPI lint clean; generated clients reproducible.
- `go vet ./...` and `go test -race ./...` pass.
- Contract conformance 6/6, frontend ESLint, Vitest 15/15, and production build
  pass after regeneration.

**Next:**
- #22 delay/wait node. Exercise HTTP against a real allowlisted service
  manually before merge if desired.

---

## 2026-07-19 — Graph run E2E happy path (#20)

**Did:**
- Added an isolated Playwright flow that names a graph, adds a contract-backed
  Start node, runs it, verifies the serialized `POST /graphs` payload, waits
  for terminal status, and asserts the workflow result is visible.
- Added explicit `run-status` and `run-result` test hooks and rendered completed
  run results in the graph editor instead of exposing only the status.
- Kept the test Prism-backed by default. Because Prism's `Run` example remains
  `running`, the spec supplies only the terminal poll response; setting
  `API_BASE_URL` bypasses that route and exercises the real API end to end.

**Decided / learned:**
- A Start-only graph is the smallest useful happy path and avoids coupling this
  test to HTTP-node configuration work planned for #21.
- CI's full E2E job remains disabled exactly as #20 requests. The test is ready
  to enable once the real stack supports the complete path.

**Verified:**
- ESLint clean, Vitest 15/15, and Playwright Chromium 2/2 pass against Prism.

**Next:**
- #21 HTTP activity node; enable CI E2E when the full backend execution path is
  ready and verify this same spec with `API_BASE_URL` pointed at it.

---

## 2026-07-19 — Contract conformance checks (#19)

**Did:**
- Added a contract conformance suite (`frontend/contract/`) that validates API
  responses against the exact `api/openapi.yaml` component schemas using AJV
  (+ ajv-formats), deriving validators straight from the spec so there is no
  hand-written schema to drift.
- Covered the success paths for `/node-types`, `POST /graphs`, `GET /graphs/{id}`,
  `POST /graphs/{id}/run`, and `GET /runs/{id}`, plus a guard test proving the
  validator rejects a non-conforming payload (so it can't degrade into a no-op).
- Gave it a dedicated `playwright.contract.config.ts` (HTTP-only, no browser,
  Prism by default), `npm run test:contract`, a `make test-contract` target,
  and a `contract-conformance` CI job gated on api/frontend changes.

**Decided / learned:**
- Kept conformance in its own config/dir, separate from the UI e2e specs, per
  the testing conventions — it talks HTTP only and never launches Next.js.
- Default target is the Prism mock (keeps it parallel-friendly and validates
  the spec's own examples/schemas + the harness); real-backend drift detection
  is opt-in via `API_BASE_URL=http://localhost:8080`.
- Registered the spec's `components.schemas` under a single AJV root id so
  intra-spec `$ref`s resolve without pre-dereferencing; `strict:false` lets
  OpenAPI-only keywords and the `double` format pass through.

**Verified:**
- ESLint clean, Vitest 15/15, and `npm run test:contract` 6/6 pass against Prism.

**Next:**
- #20 full graph → run UI happy path; wire the conformance suite at `API_BASE_URL`
  once the real backend endpoints land.

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
