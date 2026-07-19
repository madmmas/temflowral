# temflowral devlog

Running log of progress. Newest entry on top. One entry per work session —
doesn't need to be daily.

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
