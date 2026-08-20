# ADR-001: Canvas packaging

- **Status:** Accepted (amended by [ADR-002](002-module-federation-canvas.md))
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

**Reference-only for a general npm component SDK. No iframe embed service.**

The supported ways to build (or replace) a canvas are:

1. Treat `api/openapi.yaml` as the HTTP contract (typed clients via codegen).
2. Drive the palette and node metadata from `GET /node-types` (and the shared
   `pkg/nodetype` registry on the worker), not from a hardcoded frontend list.
3. Persist and run graphs with `POST /graphs` and `POST /graphs/{id}/run` using
   the same `Node` / `Edge` shapes the reference UI uses.
4. **(ADR-002)** Optionally load the official **Module Federation remote** that
   exports `WorkflowBuilder` — not a broad npm React SDK.

`frontend/` remains a working reference app. There is no guarantee of a stable
deep React component tree API or CSS contract beyond the documented MF props.

## Alternatives considered

| Option | Why not (now) |
| --- | --- |
| **Importable package** (e.g. `@temflowral/canvas` npm SDK) | Freezing a full component API early; MF remote covers the embed case without a deep public React surface. |
| **Embeddable service** (hosted iframe / micro-frontend product) | Implies auth, tenancy, hosting, and a cross-origin postMessage protocol. Out of scope while the product has no tenant isolation (see SECURITY.md). |

## Consequences

**Positive**

- One integration surface for any UI: OpenAPI + `GET /node-types`.
- Backend extensibility stays useful without waiting on a published canvas SDK.
- MF (ADR-002) gives a practical embed path without iframe tenancy.

**Negative / accepted costs**

- External products that want a designer must use MF, fork, or build their own UI.
- Type-specific config forms remain driven by `configSchema` in the shared builder.

## When to revisit

- Multiple hosts need a documented npm package with a deeper React API.
- A product requirement demands iframe embedding with postMessage.

## Related

- `GET /node-types` in [`api/openapi.yaml`](../../api/openapi.yaml)
- Frontend discovery: `frontend/src/lib/node-types.ts`
- [ADR-002: Module Federation canvas remote](002-module-federation-canvas.md)
- Adding node types: [`docs/adding-a-node-type.md`](../adding-a-node-type.md)
