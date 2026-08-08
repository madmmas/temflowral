# Plan: reference product completion

**Status:** proposed  
**Date:** 2026-08-08  
**Context:** Post-backlog (#55–#67 closed). Engine/contract are largely done; remaining gaps are the reference canvas, a few API affordances, and live verification. See also the project-status canvas notes from the same session.

## Goal

Make the compose demo feel like a usable authoring loop: **create → configure → save → reopen → run → observe**, without changing ADR-001 (canvas stays reference-only) or the single-trust-domain security stance.

## Non-goals

- Multi-tenant isolation / per-user ACLs
- Publishing an npm canvas package or embed protocol
- New node types (unless needed as fixtures)
- Replacing Temporal or the OpenAPI contract-first workflow

## Principles

1. **Contract first** — any new HTTP surface lands in `api/openapi.yaml`, then regenerate Go + TS, then implement.
2. **Registry-driven UI** — palette and config forms come from `GET /node-types` / `configSchema`, not hardcoded per-type forms (except thin custom widgets where schema is not enough).
3. **Thin increments** — each phase ships a vertical slice that is demoable alone.
4. **Track in** `docs/issues/ISSUES.md` and GitHub issues [#90](https://github.com/madmmas/temflowral/issues/90)–[#94](https://github.com/madmmas/temflowral/issues/94).

---

## Phase 0 — Baseline hygiene (done)

| Item | Notes |
| --- | --- |
| CORS for local canvas | `API_CORS_ORIGINS`; merged in [#89](https://github.com/madmmas/temflowral/pull/89) |
| Document how to smoke-test | CONTRIBUTING curl create→run→poll (exists) |

**Exit:** Palette loads against compose; curl smoke completes a Start→No-op run.

---

## Phase 1 — Reopen what you saved (P0) — [#90](https://github.com/madmmas/temflowral/issues/90)

**Problem:** Save persists to Postgres, but the UI never loads graphs; refresh loses the canvas; every Save creates a new id.

### 1a. Contract

- Keep `GET /graphs/{graphId}` (already exists).
- Add **`GET /graphs`** — list graphs (id, name, `updatedAt`; optional limit/cursor later).
- Add **`PUT /graphs/{graphId}`** (or `PATCH`) — replace nodes/edges/name for an existing id; 404 if missing.
- Regenerate clients; add Prism examples; contract tests.

### 1b. Backend

- Store: `ListGraphs`, `UpdateGraph` on Postgres + memory.
- Handlers + auth middleware already apply to OpenAPI routes.
- Tests: list empty/non-empty; update round-trip; update unknown → 404.

### 1c. Frontend

- After successful create/update, keep `savedGraphId` and show it (subtle).
- **Open by id** — query `?graph=<uuid>` and/or a small “Open graph…” prompt.
- **Recent list** — sidebar or modal from `GET /graphs`.
- Save: if `savedGraphId` set → `PUT`; else → `POST`.
- Run: save (create or update) then `POST .../run` (current behavior, fixed update path).

### Exit criteria

- Create graph → refresh with `?graph=…` → same nodes/edges reload.
- Second Save does not invent a new id.
- `GET /graphs` returns the saved row.

**Depends on:** nothing blocking.  
**Suggested issue tags:** `[api][canvas][storage]`

---

## Phase 2 — Configure nodes from the registry (P0) — [#91](https://github.com/madmmas/temflowral/issues/91)

**Problem:** Nodes drop with `{}` config; HTTP/delay/condition/wait/childWorkflow cannot be authored correctly from the UI.

### 2a. Generic config panel

- Selecting a node opens a side panel.
- Render fields from that type’s `configSchema` (string/number/enum/object basics first).
- Persist edits into React Flow node `data` / `config` used by `serializeGraph`.
- Unit-test schema→form mapping for `http`, `delay`, `condition`, `wait`.

### 2b. Type-specific polish (minimal)

- **HTTP:** method select + URL (+ optional headers/body later).
- **Delay:** `seconds`.
- **Condition:** `field` + `equals` (JSON-ish input ok).
- **Wait:** `signal` + optional `timeoutSeconds`.
- **childWorkflow:** nested graph editor can wait (Phase 2c or later); until then show read-only JSON or “edit via API”.

### 2c. Validation feedback

- Surface backend 400 messages on Save/Run.
- Optional: client-side required-field hints from schema.

### Exit criteria

- Author Start → Delay(1s) → No-op in the UI, Run completes.
- Author HTTP against an allowlisted host (`HTTP_ALLOWED_HOSTS`) and see activity output on the run.

**Depends on:** Phase 1 optional but recommended (so configs survive reopen).  
**Suggested tags:** `[canvas][frontend]`

---

## Phase 3 — Branching & wait UX (P1) — [#92](https://github.com/madmmas/temflowral/issues/92)

**Problem:** Backend supports `sourceHandle` (`true`/`false`, `received`/`timedOut`); reference node UI does not expose dual handles or signal delivery.

### 3a. Named output handles on canvas

- For types advertising output handles (from registry / OpenAPI), render multiple source handles with labels.
- Edges store `sourceHandle`; serializer already has hooks — verify round-trip with Phase 1 load.

### 3b. Signal delivery UI

- When a run is blocked on wait (`GET /runs/{id}` + wait query info if exposed), show “Send signal” with name + optional payload.
- Calls `POST /runs/{runId}/signal`.
- Happy path: Wait node → Run → Signal → completes via `received` branch.

### Exit criteria

- Condition graph with both branches drawable and only the taken path executed.
- Wait graph completable from the UI without curling the signal endpoint.

**Depends on:** Phase 2 (config for condition/wait).  
**Suggested tags:** `[canvas][executor]`

---

## Phase 4 — Authoring quality & API hygiene (P1) — [#93](https://github.com/madmmas/temflowral/issues/93)

| Item | Detail |
| --- | --- |
| Dirty state | Disable Run/Save when unchanged; warn on navigate away |
| Idempotent Run | Pass optional `idempotencyKey` from UI (e.g. per click uuid) |
| Delete graph | Optional `DELETE /graphs/{id}` if list UX needs cleanup |
| activityOptions / taskQueue | Advanced panel collapsed by default |
| Template hints | Placeholder text for `{{ nodes.<id>.output.<path> }}` on string fields |

**Exit:** Power-user fields reachable without cluttering the default path.

---

## Phase 5 — Verification & demo data (P2) — [#94](https://github.com/madmmas/temflowral/issues/94)

### 5a. Live e2e

- Enable Playwright happy-path against compose/backend (`API_BASE_URL=http://127.0.0.1:8080`).
- Uncomment / add CI job when flakiness is acceptable (or nightly).
- Cover: load palette → add Start+No-op → Run → terminal status.

### 5b. Seeds / fixtures

- `make seed-demo` or compose init script: insert 1–2 sample graphs (Start→No-op, Start→Delay→No-op).
- Document in README quickstart: “open `?graph=<seed-id>`” or list UI.
- Keep seeds idempotent (fixed UUIDs ok for local only).

### 5c. Contract suite on live API

- Already optional via `API_BASE_URL`; wire into CI matrix or docs “verify before release”.

**Exit:** New contributor can `docker compose up`, open a seeded graph, Run, and see success without inventing a graph from scratch.

---

## Phase 6 — Explicitly deferred

| Item | Why deferred |
| --- | --- |
| Canvas npm package / embed | ADR-001 |
| Multi-tenant / ACL | SECURITY.md trust boundary |
| Parallel fan-out beyond childWorkflow depth-1 | Product not requested |
| Rich expression language for templates | Minimal `{{ nodes.*.output.* }}` is enough |
| Dependabot-only PRs | Maintenance, not product plan |

---

## Suggested delivery order

```text
Phase 0 (CORS) ──▶ Phase 1 (list/get/update + reopen UI)
                      │
                      ▼
                   Phase 2 (config panel)
                      │
                      ▼
                   Phase 3 (handles + signal UI)
                      │
                      ├─▶ Phase 4 (polish)     } can overlap
                      └─▶ Phase 5 (e2e/seeds)  }
```

**Parallelism:** Backend Phase 1a/1b can start while frontend still on Phase 2 design. Phase 5a can start once Phase 1+2 happy path exists.

## Effort sketch (rough)

| Phase | Size | Primary track |
| --- | --- | --- |
| 1 | M | Backend + frontend |
| 2 | L | Frontend (+ small API only if schema gaps) |
| 3 | M | Frontend |
| 4 | S–M | Frontend / optional API |
| 5 | S–M | Testing / ops |

## Definition of “reference product complete”

1. User can author a non-trivial graph (delay + condition or wait) entirely in the UI.  
2. Save updates one graph id; reopen after refresh.  
3. Run shows terminal status and result in the UI.  
4. One automated path proves create→run against the real backend.  
5. README quickstart mentions seed or reopen flow.

## Tracking

GitHub issues filed for Phases 1–5: [#90](https://github.com/madmmas/temflowral/issues/90)–[#94](https://github.com/madmmas/temflowral/issues/94).
Keep `docs/issues/ISSUES.md` in sync. Add a DEVLOG entry per phase merge, not only at the end.
