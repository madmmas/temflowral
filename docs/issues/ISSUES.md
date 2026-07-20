# temflowral — Issue List

Kickoff issues [#5](https://github.com/madmmas/temflowral/issues/5)–[#27](https://github.com/madmmas/temflowral/issues/27)
are closed (filed 2026-07-15; repo PRs already occupied #1–#4).

Post-kickoff backlog filed 2026-07-19 as
[#55](https://github.com/madmmas/temflowral/issues/55)–[#67](https://github.com/madmmas/temflowral/issues/67).

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

**[#57](https://github.com/madmmas/temflowral/issues/57) — Caller-supplied idempotency key on `StartGraphRun`** `[executor]`
Accept an optional idempotency key on `StartGraphRun`; dedupe against it before
starting a new Temporal workflow. Needed for at-least-once callers (webhooks,
queues, retried API calls). Depends on: #56.

## 7. Signals & execution controls

**[#58](https://github.com/madmmas/temflowral/issues/58) — Signal/wait primitive** `[executor]`
A "wait for signal" node type or run-level primitive that suspends execution
until a named signal arrives, with a timeout fallback. Only timers exist today.
Depends on: #55.

**[#59](https://github.com/madmmas/temflowral/issues/59) — Signal-delivery endpoint** `[executor]`
`POST /runs/{id}/signal` (or similar), validating the run is waiting on that
signal name before forwarding to the Temporal workflow. Depends on: #58.

**[#60](https://github.com/madmmas/temflowral/issues/60) — Per-node ActivityOptions (timeout/retry override)** `[executor]`
Allow a node's config to specify `startToCloseTimeout`, `retryPolicy`, etc.,
overriding engine defaults. Depends on: #55.

**[#61](https://github.com/madmmas/temflowral/issues/61) — Per-node task-queue routing** `[executor]`
Allow a node's config to specify a target Temporal task queue so activities run
only on workers with specific capabilities. Depends on: #55.

## 8. Graph expressiveness

**[#62](https://github.com/madmmas/temflowral/issues/62) — Child Workflow node type** `[node-type][executor]`
Node that spawns a child Temporal workflow and can gate on its result — for
fan-out/fan-in or per-item sub-workflows expressed as a graph. Depends on:
Graph → Temporal translator (#12, done).

**[#63](https://github.com/madmmas/temflowral/issues/63) — Templating syntax for node config** `[executor]`
Minimal templating syntax (e.g. `{{ nodes.foo.output.bar }}`) resolved at
execution time so node config can reference another node's output. Depends on:
#55.

**[#64](https://github.com/madmmas/temflowral/issues/64) — Graph validation before run start** `[executor]`
Validate node types against the registry and detect cycles before a run starts
— reject unknown types and cycles at submission time, not mid-run. Depends on:
#55.

## 9. Product decisions & docs

**[#65](https://github.com/madmmas/temflowral/issues/65) — Canvas packaging decision** `[canvas][decision]`
ADR-style doc: whether the React Flow frontend becomes an importable package,
an embeddable service, or stays reference-only. "No shared package yet — build
against the node-type registry API" is a valid answer.

**[#66](https://github.com/madmmas/temflowral/issues/66) — Document API auth baseline and trust-boundary stance** `[docs][security]`
Minimal service-to-service auth (shared secret or mTLS); extend SECURITY.md
with an explicit trust-boundary statement (no tenant isolation enforced);
short compatibility note for interpreter upgrades.

**[#67](https://github.com/madmmas/temflowral/issues/67) — Extend `docs/adding-a-node-type.md` for external registration** `[docs]`
Document registering a node type from outside this repo once #55 lands.
Depends on: #55.

---

### Suggested order
55 + 56 in parallel (foundation) → 57 (after 56)
then: 58 → 59 · 60 · 61 · 63 · 64 (after 55, can parallelize once #55 lands)
62 can start after #12 (already done); pairs well with #55
65 · 66 anytime · 67 after #55
