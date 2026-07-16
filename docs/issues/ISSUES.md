# temflowral — Kickoff Issue List

Repo audit (2026-07-15): the project is a governance-only scaffold. `backend/go.mod`
exists (Go 1.25.7, module `github.com/madmmas/temflowral/backend`) but has no source
files. There is no `frontend/` directory, no `docker-compose.yml`, no `cmd/server/main.go`.
README, CONTRIBUTING, CHANGELOG, CODE_OF_CONDUCT are outline placeholders, not written.
CI (`.github/workflows/ci.yml`) is already built and correctly no-ops until code exists;
its `e2e` job is commented out with "Disabled until Day 13."

**Contract-first, updated 2026-07-15:** the OpenAPI/Swagger spec is now the pivot point.
Backend publishes it before writing real handlers; frontend and the Playwright
contributor both build against it in parallel instead of waiting on each other.

GitHub issues filed 2026-07-15 as [#5](https://github.com/madmmas/temflowral/issues/5)–[#27](https://github.com/madmmas/temflowral/issues/27)
(repo PRs already occupied #1–#4).

## Team split
- **You:** backend contract + implementation, frontend scaffold + integration.
- **Contributor 2:** Playwright automated testing, working off the contract/mock
  server rather than the live backend.

## 0. Contract first — blocks everything else

**[#5](https://github.com/madmmas/temflowral/issues/5) — Fix unresolved merge conflict markers in issue templates** `[bug][priority:high]`
`.github/ISSUE_TEMPLATE/bug_report.md` and `feature_request.md` both have raw
`<<<<<<< HEAD / ======= / >>>>>>> 979a133` conflict markers committed straight to `main`.
Keep the temflowral-specific version (references `docs/adding-a-node-type.md`, asks for
Go/Node versions), delete the other half, commit clean. Do this before anything else —
five minutes, and it's visible to every future contributor.

**[#6](https://github.com/madmmas/temflowral/issues/6) — Author `openapi.yaml` v0.1** `[api][priority:high]`
The contract every other track builds against. Minimum surface for kickoff:
- `POST /graphs` — create a graph (nodes + edges)
- `GET /graphs/{id}` — fetch a graph
- `POST /graphs/{id}/run` — start a Temporal workflow run from a graph
- `GET /runs/{id}` — poll run status/result
- `GET /node-types` — registry of available node types + their config schema

Schemas: `Graph`, `Node`, `Edge`, `NodeType`, `Run`, `RunStatus`. Keep it thin — this
is v0.1, expand as node types land. Live in `api/openapi.yaml` at repo root so it's
not buried under `backend/` or `frontend/`.

**[#7](https://github.com/madmmas/temflowral/issues/7) — Serve the spec + Swagger UI from the backend** `[backend][api]`
`GET /openapi.yaml` (raw file) and `GET /docs` (Swagger UI pointed at it). This can
ship before any real handlers exist — it's static-file serving plus a CDN'd
swagger-ui bundle. Gives the frontend and Playwright contributor a live reference
immediately, not just a file in the repo.

**[#8](https://github.com/madmmas/temflowral/issues/8) — Mock server from the contract** `[api][infra][priority:high]`
Document (in CONTRIBUTING.md) how to run a spec-backed mock, e.g.
`npx @stoplight/prism-cli mock api/openapi.yaml`. This is what unblocks frontend
and Playwright work before real backend logic lands — both tracks point at the mock
until real endpoints come online, then flip a base-URL env var.

**[#9](https://github.com/madmmas/temflowral/issues/9) — Spec validation in CI** `[infra][api]`
Add a CI job that lints `openapi.yaml` (Spectral or equivalent) on every PR that
touches it, so a broken contract never lands on `main`. Small addition to the
existing `.github/workflows/ci.yml`.

## 1. Backend — implements the contract

**[#10](https://github.com/madmmas/temflowral/issues/10) — Generate Go server interfaces from the spec** `[backend]`
Use `oapi-codegen` (or hand-write if you'd rather not add codegen yet) to produce
request/response types + a server interface from `openapi.yaml`. `cmd/server/main.go`
wires the generated interface to real handlers — this replaces writing the HTTP
layer from scratch.

**[#11](https://github.com/madmmas/temflowral/issues/11) — Temporal client + local worker wiring** `[backend]`
Connect to Temporal (temporalite for local dev). Register a no-op workflow +
activity to prove the wiring end to end before `POST /graphs/{id}/run` does
anything real.

**[#12](https://github.com/madmmas/temflowral/issues/12) — Graph → Temporal workflow translator** `[backend][core]`
Core engine: walk the `Graph` schema from #6 and drive a generic Temporal workflow
that dispatches to the right activity per node type, respecting edge order/branching.

## 2. Frontend — implements against the contract

**[#13](https://github.com/madmmas/temflowral/issues/13) — Next.js app bootstrap in `frontend/`** `[frontend]`
`create-next-app`, TypeScript, matches Node 20 / npm assumptions in Makefile and CI.

**[#14](https://github.com/madmmas/temflowral/issues/14) — Generate a typed API client from the spec** `[frontend][api]`
`openapi-typescript` or `orval` against `api/openapi.yaml` — gives you typed
request/response objects and a client, regenerated whenever the contract changes.
Point it at the Prism mock (#8) until real backend endpoints exist.

**[#15](https://github.com/madmmas/temflowral/issues/15) — React Flow (xyflow) canvas** `[frontend]`
Base canvas: pan/zoom, add/connect nodes, delete nodes/edges. No custom node types yet.

**[#16](https://github.com/madmmas/temflowral/issues/16) — Node palette + custom node rendering, driven by `GET /node-types`** `[frontend]`
Sidebar populated from the node-type registry endpoint rather than hardcoded, so
adding a node type on the backend doesn't require a matching frontend PR.

**[#17](https://github.com/madmmas/temflowral/issues/17) — Save/run graph against the generated client** `[frontend][integration]`
Serialize canvas state to the `Graph` schema, call `POST /graphs` and
`POST /graphs/{id}/run` via the generated client, poll `GET /runs/{id}` for status.

## 3. Automated testing (Contributor 2 — Playwright)

**[#18](https://github.com/madmmas/temflowral/issues/18) — Playwright scaffold pointed at the mock server** `[testing]`
Set up Playwright in `frontend/`, config to target the Prism mock (#8) by default so
this contributor isn't blocked on real backend work landing.

**[#19](https://github.com/madmmas/temflowral/issues/19) — Contract conformance checks** `[testing][api]`
Assert backend responses actually match `openapi.yaml` schemas (e.g. Dredd, or a
lightweight AJV-based check against captured responses) — catches drift between the
spec and the real implementation before it reaches the frontend.

**[#20](https://github.com/madmmas/temflowral/issues/20) — E2E happy-path test: build graph → run → see result** `[testing]`
The real target once #12 and #17 exist. This is the one CI's `e2e` job comment marks
"Disabled until Day 13" — keep it off until the stack can actually support it, but
write it against the mock in the meantime so it's ready to flip on.

## 4. First node types

**[#21](https://github.com/madmmas/temflowral/issues/21) — HTTP activity node** `[node-type]`
First real node type; add its config schema to `openapi.yaml`'s `NodeType`/`Node`
shapes as part of the PR. SECURITY.md already flags this as the primary attack
surface (arbitrary requests, SSRF via user-supplied URLs, template injection in
config fields) — build the allowlist/validation story alongside the happy path.

**[#22](https://github.com/madmmas/temflowral/issues/22) — Delay/wait node** `[node-type]`
Proves Temporal's durable timers work through the graph translator.

**[#23](https://github.com/madmmas/temflowral/issues/23) — Conditional/branch node** `[node-type]`
Exercises edge-based branching in the translator beyond a linear chain.

## 5. Cleanup & docs

**[#24](https://github.com/madmmas/temflowral/issues/24) — `golangci-lint` config** `[infra]`
`.golangci.yml` so `make lint` (and CI's go-lint job) enforces real rules.

**[#25](https://github.com/madmmas/temflowral/issues/25) — `docker-compose.yml`** `[infra]`
Temporal server + Postgres + backend + frontend, one `docker compose up`.

**[#26](https://github.com/madmmas/temflowral/issues/26) — `docs/adding-a-node-type.md`** `[docs]`
Referenced by the feature request issue template. Write once #21 exists, so it
documents both the backend node implementation and the `openapi.yaml` schema change.

**[#27](https://github.com/madmmas/temflowral/issues/27) — Fill in placeholder docs** `[docs]`
README.md, CONTRIBUTING.md, CHANGELOG.md are numbered outlines, not prose.

---

### Suggested order
5 → 6 → 7 → 8 → 9 (contract track, can start immediately)
then in parallel: 10 → 11 → 12 (backend) · 13 → 14 → 15 → 16 → 17 (frontend) · 18 → 19 → 20 (Playwright, off the mock until backend catches up)
then: 21 → 22/23 → 24 → 25 → 26 → 27
