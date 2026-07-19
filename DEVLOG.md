# temflowral devlog

Running log of progress. Newest entry on top. One entry per work session —
doesn't need to be daily.

---

## 2026-07-19 — Temporal client and worker wiring (#11)

**Did:**
- Added the Temporal Go SDK and connected the backend to a configurable
  Temporal service during startup.
- Registered a deterministic smoke workflow and activity on the `temflowral`
  task queue, with a workflow test that executes the activity end to end.
- Added graceful HTTP/worker shutdown and documented the local Temporal CLI
  development server and smoke command.

**Decided / learned:**
- The Temporal SDK is pinned to `v1.46.0`, the latest stable release compatible
  with the project's Go baseline.
- Local defaults are `localhost:7233`, namespace `default`, and task queue
  `temflowral`; each can be overridden by environment variable.
- Temporalite is deprecated, so local development uses
  `temporal server start-dev`.

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
