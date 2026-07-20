# temflowral

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Draw a workflow as a graph, and temflowral runs it as a durable
[Temporal](https://temporal.io/) workflow. You design nodes and edges on a
visual canvas; the backend translates the graph into a Temporal workflow and
executes it with retries, timers, and branching handled for you.

## Architecture

```text
┌─────────────────────┐        ┌──────────────────────┐        ┌─────────────┐
│ Frontend            │        │ Backend              │        │ Temporal    │
│ Next.js + React     │  HTTP  │ Go HTTP API          │  gRPC  │ server      │
│ Flow canvas         │ ─────▶ │ + graph translator   │ ─────▶ │ + worker    │
│ (typed OpenAPI      │        │ + node activities    │        │ (Postgres   │
│  client)            │ ◀───── │ + Postgres store     │ ◀───── │  persisted) │
└─────────────────────┘        └──────────────────────┘        └─────────────┘
```

Graph and run API metadata live in a dedicated `temflowral` Postgres database
(`DATABASE_URL`). Temporal workflow history uses separate databases on the same
Postgres instance. The backend refuses to start without `DATABASE_URL`.

`api/openapi.yaml` is the single source of truth for the HTTP API. Both the Go
server types and the TypeScript client are generated from it, so the two sides
never drift by hand. See [`docs/adding-a-node-type.md`](docs/adding-a-node-type.md)
for the contract-first workflow.

## Quickstart

Requires Docker. From the repository root:

```sh
docker compose up   # or: make run
```

First boot builds the backend and frontend images, creates the `temflowral`
application database, and initializes the Temporal schema in Postgres, so allow
a minute. Then open:

- Frontend canvas: <http://localhost:3000>
- Backend API + interactive docs: <http://localhost:8080> / <http://localhost:8080/docs>
- Temporal Web UI: <http://localhost:8233>

Stop with `Ctrl+C`, then `make temporal-down` to remove the containers. The
Postgres volume persists across restarts; `docker compose down -v` wipes it.

Prefer running the backend or frontend outside Docker? See
[`CONTRIBUTING.md`](CONTRIBUTING.md) for the mock-API, backend, and frontend
development flows.

## How it works

You build a graph on the canvas by dragging node types from a palette (fetched
live from `GET /node-types`) and connecting them with edges. Saving the graph
sends it to `POST /graphs`, which validates each node's configuration against
its type before persisting.

Starting a run (`POST /graphs/{graphId}/run`) hands the graph to the backend
translator. It validates the graph, orders the nodes topologically, and starts
a Temporal workflow. The workflow walks the graph, running one activity per
executable node, evaluating condition branches, and sleeping on durable timers
for delay nodes.

Progress and results are durable in Temporal. Poll `GET /runs/{runId}` for the
current status and, once complete, the per-node output — or watch the same run
live in the Temporal Web UI.

## Node types

| Type | Category | Description |
| --- | --- | --- |
| `start` | core | Workflow entry point; carries the run input. |
| `noop` | core | No-op node for smoke-testing execution. |
| `http` | integration | Allowlisted outbound HTTP request (deny-by-default). |
| `delay` | core | Pause with a durable Temporal timer. |
| `condition` | core | Branch on a predecessor field (`true`/`false`). |

The `http` node is the primary attack surface; its outbound policy (host
allowlisting, SSRF protection, and size/time limits) is documented in
[`SECURITY.md`](SECURITY.md).

## Adding a custom node type

Node types are the main extensibility hook. Built-in types follow the
contract-first recipe in
[`docs/adding-a-node-type.md`](docs/adding-a-node-type.md). Custom domain
activities that must not live in this repo register at worker startup via
`backend/pkg/nodetype` (see the external-registration section of that guide).

## Contributing

Contributions are welcome. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md) for
development setup and conventions, and please follow the
[Contributor Covenant](CODE_OF_CONDUCT.md). Notable changes are tracked in
[`CHANGELOG.md`](CHANGELOG.md).

## License

Released under the [MIT License](LICENSE).
