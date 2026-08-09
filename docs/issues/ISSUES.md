# temflowral — Issue List

Kickoff issues [#5](https://github.com/madmmas/temflowral/issues/5)–[#27](https://github.com/madmmas/temflowral/issues/27)
are closed (filed 2026-07-15; repo PRs already occupied #1–#4).

Post-kickoff backlog filed 2026-07-19 as
[#55](https://github.com/madmmas/temflowral/issues/55)–[#67](https://github.com/madmmas/temflowral/issues/67).

Canvas UI/UX backlog filed 2026-08-09 as
[#102](https://github.com/madmmas/temflowral/issues/102)–[#110](https://github.com/madmmas/temflowral/issues/110);
follow-ups [#111](https://github.com/madmmas/temflowral/issues/111)–[#114](https://github.com/madmmas/temflowral/issues/114).

## Team split
- **You:** backend contract + implementation, frontend scaffold + integration.
- **Contributor 2:** Playwright automated testing, working off the contract/mock
  server rather than the live backend.

---

## Kickoff — done ([#5](https://github.com/madmmas/temflowral/issues/5)–[#27](https://github.com/madmmas/temflowral/issues/27))

Contract (#5–#9), backend (#10–#12), frontend (#13–#17), Playwright (#18–#20),
first node types (#21–#23), cleanup & docs (#24–#27). All closed.

---

## 6. Extensibility & durable execution

**[#55](https://github.com/madmmas/temflowral/issues/55) — External node-type & activity registration (extensibility hook)** `[executor][extensibility]` ✅
Interface/SDK for registering a custom node type (config schema + output
handles) and its backing Temporal activity, resolvable at worker startup,
independent of built-in node types. Schema must support output handles derived
from config (not just a fixed list). Without this, adopters fork temflowral.
Depends on: Graph → Temporal translator (#12, done). **Shipped:**
`backend/pkg/nodetype` + OpenAPI handle fields.

**[#56](https://github.com/madmmas/temflowral/issues/56) — Durable storage backend for graph/run store** `[executor][storage]` ✅
Pluggable durable store (Postgres to start; keep it an interface). Startup
check that fails loudly if a durable store isn't configured, instead of
silently defaulting to in-memory. A worker restart today loses every in-flight
run. Depends on: Graph → Temporal translator (#12, done). **Shipped:**
`backend/internal/store` with Postgres via `DATABASE_URL`.

**[#57](https://github.com/madmmas/temflowral/issues/57) — Caller-supplied idempotency key on `StartGraphRun`** `[executor]` ✅
Accept an optional idempotency key on `StartGraphRun`; dedupe against it before
starting a new Temporal workflow. Needed for at-least-once callers (webhooks,
queues, retried API calls). Depends on: #56. **Shipped:** optional
`idempotencyKey` on `StartRunRequest`, unique per graph in the durable store.

## 7. Signals & execution controls

**[#58](https://github.com/madmmas/temflowral/issues/58) — Signal/wait primitive** `[executor]` ✅
A "wait for signal" node type or run-level primitive that suspends execution
until a named signal arrives, with a timeout fallback. Only timers exist today.
Depends on: #55. **Shipped:** `wait` node (`WaitNodeConfig`) racing a Temporal
signal channel against a durable timeout; branches via `received`/`timedOut`.

**[#59](https://github.com/madmmas/temflowral/issues/59) — Signal-delivery endpoint** `[executor]` ✅
`POST /runs/{id}/signal` (or similar), validating the run is waiting on that
signal name before forwarding to the Temporal workflow. Depends on: #58.
**Shipped:** `POST /runs/{runId}/signal` with Temporal `currentWait` query
validation, then `SignalWorkflow`.

**[#60](https://github.com/madmmas/temflowral/issues/60) — Per-node ActivityOptions (timeout/retry override)** `[executor]` ✅
Allow a node's config to specify `startToCloseTimeout`, `retryPolicy`, etc.,
overriding engine defaults. Depends on: #55. **Shipped:** optional
`Node.activityOptions` (`ActivityOptions` / `RetryPolicy`) applied on
`KindActivity` nodes; rejected on workflow-native types.

**[#61](https://github.com/madmmas/temflowral/issues/61) — Per-node task-queue routing** `[executor]` ✅
Allow a node's config to specify a target Temporal task queue so activities run
only on workers with specific capabilities. Depends on: #55. **Shipped:**
optional `Node.taskQueue` applied via Temporal `ActivityOptions.TaskQueue` on
`KindActivity` nodes; rejected on workflow-native types.

## 8. Graph expressiveness

**[#62](https://github.com/madmmas/temflowral/issues/62) — Child Workflow node type** `[node-type][executor]` ✅
Node that spawns a child Temporal workflow and can gate on its result — for
fan-out/fan-in or per-item sub-workflows expressed as a graph. Depends on:
Graph → Temporal translator (#12, done). **Shipped:** `childWorkflow` node with
inline `NestedGraph`; runs `ExecuteChildWorkflow(GraphWorkflow)` and waits for
the result. Nested `childWorkflow` nodes are rejected (depth 1).

**[#63](https://github.com/madmmas/temflowral/issues/63) — Templating syntax for node config** `[executor]` ✅
Minimal templating syntax (e.g. `{{ nodes.foo.output.bar }}`) resolved at
execution time so node config can reference another node's output. Depends on:
#55. **Shipped:** `{{ nodes.<id>.output.<path> }}` resolved in GraphWorkflow
from active predecessors; HTTP revalidates rendered URL/headers/body; wait
configs reject templates.

**[#64](https://github.com/madmmas/temflowral/issues/64) — Graph validation before run start** `[executor]` ✅
Validate node types against the registry and detect cycles before a run starts
— reject unknown types and cycles at submission time, not mid-run. Depends on:
#55. **Shipped:** `ValidateGraph` on `StartGraphRun` (409) before Temporal
start; create rejects unregistered types (400); registry + cycle/unreachable
covered by API and plan tests.

## 9. Product decisions & docs

**[#65](https://github.com/madmmas/temflowral/issues/65) — Canvas packaging decision** `[canvas][decision]` ✅
ADR-style doc: whether the React Flow frontend becomes an importable package,
an embeddable service, or stays reference-only. "No shared package yet — build
against the node-type registry API" is a valid answer. **Decided:**
reference-only; integrate via OpenAPI + `GET /node-types` —
[`docs/adr/001-canvas-packaging.md`](../adr/001-canvas-packaging.md).

**[#66](https://github.com/madmmas/temflowral/issues/66) — Document API auth baseline and trust-boundary stance** `[docs][security]` ✅
Minimal service-to-service auth (shared secret or mTLS); extend SECURITY.md
with an explicit trust-boundary statement (no tenant isolation enforced);
short compatibility note for interpreter upgrades. **Shipped:** opt-in
`API_AUTH_TOKEN` Bearer gate on OpenAPI routes; SECURITY.md trust boundary +
mTLS-at-proxy + upgrade compatibility; OpenAPI `BearerAuth` / `401`.

**[#67](https://github.com/madmmas/temflowral/issues/67) — Extend `docs/adding-a-node-type.md` for external registration** `[docs]` ✅
Document registering a node type from outside this repo once #55 lands.
Depends on: #55. **Shipped:** full external-registration walkthrough in
`docs/adding-a-node-type.md` (custom binary, `WithRegistry`, handles, OpenAPI
vs registry authority, verify steps).

---

### Suggested order
55 + 56 in parallel (foundation) → 57 (after 56)
then: 58 → 59 · 60 · 61 · 63 · 64 (after 55, can parallelize once #55 lands)
62 can start after #12 (already done); pairs well with #55
65 · 66 anytime · 67 after #55

---

## Post-#67 — Reference product completion

Filed backlog (#5–#27, #55–#67) is closed. Remaining work is mostly the
reference canvas and a few API affordances. Full plan:

→ [`docs/plans/reference-product-completion.md`](../plans/reference-product-completion.md)

| Phase | Focus | Priority | Issue |
| --- | --- | --- | --- |
| 0 | CORS / local compose hygiene | done | [#89](https://github.com/madmmas/temflowral/pull/89) |
| 1 | List + update graphs; reopen in UI | P0 | [#90](https://github.com/madmmas/temflowral/issues/90) (done) |
| 2 | Config panel from `configSchema` | P0 | [#91](https://github.com/madmmas/temflowral/issues/91) (done) |
| 3 | Named handles + signal UI | P1 | [#92](https://github.com/madmmas/temflowral/issues/92) (done) |
| 4 | Dirty state, idempotent Run, advanced fields | P1 | [#93](https://github.com/madmmas/temflowral/issues/93) (done) |
| 5 | Live e2e + demo seeds | P2 | [#94](https://github.com/madmmas/temflowral/issues/94) (done) |
| 6 | Deferred (ADR-001 package, multi-tenant, …) | — | — |

---

## Post-#94 — Canvas UI/UX backlog

Filed 2026-08-09 after a live-canvas UX audit (`localhost:3000`). Label: `ux` (+
`canvas`). Suggested build order: P0 first, then P1, then P2.

| Pri | Issue | Focus |
| --- | --- | --- |
| P0 | [#102](https://github.com/madmmas/temflowral/issues/102) (done) | Collapsible / resizable / dismissible **run result** panel |
| P0 | [#103](https://github.com/madmmas/temflowral/issues/103) (done) | Replace saved-graph **dropdown** with searchable workflow library |
| P1 | [#104](https://github.com/madmmas/temflowral/issues/104) (done) | Theme **MiniMap** for dark canvas + hide/toggle |
| P1 | [#105](https://github.com/madmmas/temflowral/issues/105) (done) | **Toolbar** layout; separate Delete from Run; dirty indicator; **stable on config open** |
| P1 | [#106](https://github.com/madmmas/temflowral/issues/106) (done) | **Run history** + clearer status / short graph id |
| P1 | [#111](https://github.com/madmmas/temflowral/issues/111) | **Fit viewport** after async graph load |
| P1 | [#112](https://github.com/madmmas/temflowral/issues/112) | Node **config form** quality (order, headers builder, templates) |
| P1 | [#113](https://github.com/madmmas/temflowral/issues/113) | **Per-node** visual execution status on the canvas |
| P2 | [#107](https://github.com/madmmas/temflowral/issues/107) (done) | Graph **naming** UX (discourage duplicate Untitled) |
| P2 | [#108](https://github.com/madmmas/temflowral/issues/108) (done) | Surface load/API **errors** as banners (not footer-only) |
| P2 | [#109](https://github.com/madmmas/temflowral/issues/109) (done) | Empty canvas guidance + dismissible authoring tip |
| P2 | [#110](https://github.com/madmmas/temflowral/issues/110) | Dedicated **wait-signal** panel + resizable node config |
| P2 | [#114](https://github.com/madmmas/temflowral/issues/114) | **Accessibility** baseline for the reference canvas |

### Detail

**[#102](https://github.com/madmmas/temflowral/issues/102) — Collapsible / resizable / dismissible run result panel** `[canvas][ux]` ✅
Result JSON currently lives in the bottom `flex-wrap` status strip and steals
canvas height after Run completes. Move to a dedicated panel with collapse,
resize, hide, and copy. **Shipped:** `RunResultPanel` bottom drawer; footer
keeps a status chip that reopens the panel.

**[#103](https://github.com/madmmas/temflowral/issues/103) — Replace saved-graph dropdown with searchable workflow library** `[canvas][ux]` ✅
`<select>` does not scale; duplicate names + truncated UUIDs. Modal/side library
with search, sort by `updatedAt`, and metadata. Depends on #90 (done).
**Shipped:** `WorkflowLibrary` modal (search, sort Updated/Name, short id +
updated time); toolbar **Open…** button keeps `data-testid="open-graph"`.

**[#104](https://github.com/madmmas/temflowral/issues/104) — Theme MiniMap for dark canvas and add hide/toggle** `[canvas][ux]` ✅
Bright white MiniMap covers nodes; add dark styling and a show/hide control.
**Shipped:** scheme-aware MiniMap colors, compact size, Show/Hide map toggle
with `localStorage` preference.

**[#105](https://github.com/madmmas/temflowral/issues/105) — Toolbar layout: stable actions and separate Delete from Run** `[canvas][ux]` ✅
Toolbar wrap puts Run next to Delete; add explicit dirty/unsaved indicator.
Also: keep hit targets stable when the node config panel opens/closes (no
jump into the graph-name field). **Shipped:** nowrap toolbar with Unsaved
badge, Save/Run primary group, Delete in a separated danger zone; config
panel overlays the canvas instead of shrinking the toolbar column.

**[#106](https://github.com/madmmas/temflowral/issues/106) — Run history and clearer run status presentation** `[canvas][ux]` ✅
Only the latest run is kept; footer dumps full graph UUID. Session (or API)
history + status chips + short id with copy. May need OpenAPI if durable.
**Shipped:** session-local recent-run list (no `GET …/runs` yet), compact
status chip (incl. waiting), short graph id + Copy; clicking history restores
result/error in the drawer.

**[#107](https://github.com/madmmas/temflowral/issues/107) — Graph naming UX: discourage duplicate Untitled workflows** `[canvas][ux]` ✅
Prompt/nudge on first Save; optional duplicate-name warning. Pairs with #103.
**Shipped:** toolbar hint for placeholder/duplicate names; first Save prompts
away from `Untitled workflow`; soft confirm when the name already exists in
`GET /graphs` (API uniqueness unchanged).

**[#108](https://github.com/madmmas/temflowral/issues/108) — Surface load/API errors as banners instead of footer-only text** `[canvas][ux]` ✅
Invalid `?graph=` is easy to miss in the footer. Dismissible banner/toast.
**Shipped:** top `role="alert"` banner for open/save/run/list failures with
Dismiss + New; failed deep links clear `?graph=` from the URL.

**[#109](https://github.com/madmmas/temflowral/issues/109) — Empty canvas guidance and dismissible authoring tip** `[canvas][ux]` ✅
First-run empty state; dismissible tip remembered in `localStorage`.
**Shipped:** centered empty-canvas guide; authoring tip dismissible + persisted;
slight palette-click offset so stacked center drops fan out.

**[#110](https://github.com/madmmas/temflowral/issues/110) — Dedicated wait-signal panel and resizable node config drawer** `[canvas][ux]`
Pull signal UI out of the footer; allow config panel width resize. Builds on #92.

**[#111](https://github.com/madmmas/temflowral/issues/111) — Fit viewport after async graph load** `[canvas][ux]`
`fitView` on mount often runs with empty nodes; call `fitView()` after open/
hydrate. Distinguish off-viewport from sparse graph data (lone Start, etc.).

**[#112](https://github.com/madmmas/temflowral/issues/112) — Improve node config form quality** `[canvas][ux]`
Stable field order (HTTP currently `body, headers, method, url` from Go map
JSON), headers key/value builder, template autocomplete for upstream nodes.
Complements #110 (drawer chrome only).

**[#113](https://github.com/madmmas/temflowral/issues/113) — Per-node visual execution status on the canvas** `[canvas][executor][ux]`
Decorate nodes pending/running/completed/failed/waiting during or after runs.
May need contract progress beyond #106 footer/history.

**[#114](https://github.com/madmmas/temflowral/issues/114) — Accessibility baseline for the reference canvas** `[canvas][ux]`
Keyboard path, focus rings, dark-theme contrast; umbrella placeholder.