# ADR-001: Canvas packaging

- **Status:** Accepted
- **Date:** 2026-07-21
- **Issue:** [#65](https://github.com/madmmas/temflowral/issues/65)
- **Tags:** `canvas`, `decision`

## Context

temflowral ships a Next.js + React Flow authoring UI under `frontend/`. Adopters
who want a visual designer face an open question: should that UI become an
importable npm package, an embeddable hosted service, or remain a reference
implementation?

Leaving this undecided encourages every consumer to either fork the whole app
or re-implement a canvas from scratch without a clear integration contract.

## Decision

**Reference-only for now. No shared canvas package and no embeddable canvas
service.**

The supported way to build (or replace) a canvas is:

1. Treat `api/openapi.yaml` as the HTTP contract (typed clients via codegen).
2. Drive the palette and node metadata from `GET /node-types` (and the shared
   `pkg/nodetype` registry on the worker), not from a hardcoded frontend list.
3. Persist and run graphs with `POST /graphs` and `POST /graphs/{id}/run` using
   the same `Node` / `Edge` shapes the reference UI uses.

`frontend/` is a working reference app in this repository. It may be copied or
forked as a starting point. It is **not** a versioned, publishable UI SDK.
There is no guarantee of a stable React component API, CSS contract, or iframe
embedding protocol.

## Alternatives considered

| Option | Why not (now) |
| --- | --- |
| **Importable package** (e.g. `@temflowral/canvas`) | The canvas is still tightly coupled to this Next.js app (routing, env, save/run chrome). Extracting a stable component surface would force API design work we do not need yet and would freeze UI choices prematurely. |
| **Embeddable service** (hosted iframe / micro-frontend) | Implies auth, tenancy, hosting, and a cross-origin postMessage protocol. Out of scope while the product has no tenant isolation and is a demonstration stack (see SECURITY.md / upcoming #66). |

"No shared package yet — build against the node-type registry API" is the
intentional product answer, not a deferral.

## Consequences

**Positive**

- One integration surface for any UI: OpenAPI + `GET /node-types`.
- Backend extensibility (#55) stays useful without waiting on a published canvas.
- We can evolve the reference UI without semver commitments to external apps.

**Negative / accepted costs**

- External products that want a designer must build or fork their own UI.
- Type-specific config forms and multi-handle renderers remain reference-app
  concerns (see `docs/adding-a-node-type.md` §7).

## When to revisit

Reopen this ADR if any of the following become true:

- Two or more first-party or partner UIs need the same canvas behavior.
- A product requirement demands embedding the designer in a third-party app.
- The reference UI has a clear, tested component boundary (palette + canvas +
  graph serialization) with no Next.js-only dependencies.

Until then, do not publish an npm canvas package or document an embed protocol.

## Related

- `GET /node-types` in [`api/openapi.yaml`](../api/openapi.yaml)
- Frontend discovery: `frontend/src/lib/node-types.ts`
- Adding node types (incl. UI notes): [`docs/adding-a-node-type.md`](adding-a-node-type.md)
- External node-type registration: `backend/pkg/nodetype` (#55)
