# ADR-002: Module Federation canvas remote

- **Status:** Accepted
- **Date:** 2026-08-19
- **Amends:** [ADR-001](001-canvas-packaging.md)
- **Tags:** `canvas`, `decision`, `sidecar`

## Context

ADR-001 kept the React Flow UI as a reference app without a publishable UI
SDK. Host products that want the designer had to fork `frontend/` or rebuild
against OpenAPI.

Consumers now also want to **import the workflow builder into an existing SPA**
via Module Federation while the engine runs as an HTTP sidecar.

## Decision

**Allow a versioned Module Federation remote** that exports `WorkflowBuilder`.

1. OpenAPI + `GET /node-types` remain the integration contract for persistence
   and execution.
2. The federated remote exposes a props-driven `WorkflowBuilder` (`apiBaseUrl`,
   auth, `initialGraphId`, save/run callbacks) — not Next.js page chrome.
3. The Next.js app in this repo is a **reference host** (and local demo). It
   may import the builder locally for tests; production hosts should load
   `remoteEntry.js` (see [`embedding-canvas-mf.md`](../embedding-canvas-mf.md)).
4. We still do **not** publish a general-purpose npm component SDK with a
   semver React public API beyond the documented MF expose name and props.

## Consequences

**Positive**

- Host apps embed the builder without forking the whole Next app.
- Sidecar + MF matches “run engine next to us, import UI remotely.”

**Costs**

- Shared singleton deps (`react`, `react-dom`, `@xyflow/react`) must align.
- Host must use webpack/Rspack for MF until Turbopack MF is reliable.
- CSS / Tailwind collisions are a documented host concern.

## Related

- [`docs/sidecar.md`](../sidecar.md)
- [`docs/embedding-canvas-mf.md`](../embedding-canvas-mf.md)
- `packages/canvas-remote`
- `WorkflowBuilder` in `frontend/src/components/graph-canvas.tsx`
