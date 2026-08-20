# Running temflowral as a sidecar

Other repositories can run the engine next to their app and call the OpenAPI
HTTP API. The React canvas can be embedded via Module Federation (see
[`embedding-canvas-mf.md`](embedding-canvas-mf.md)).

## What you need

| Piece | Required? | Notes |
| --- | --- | --- |
| Postgres | **Yes** (databases, not necessarily our container) | Dedicated DB `temflowral` for graphs/runs; Temporal needs `temporal` + `temporal_visibility` if self-hosted |
| Temporal | **Yes** | Compose Temporal service, your own Temporal, or Temporal Cloud |
| Backend image / compose `backend` | **Yes** | HTTP API + worker |
| Reference frontend | Optional | Demo UI; prefer MF remote for host apps |
| Bundled `postgresql` container | Optional | Omit when the host already runs Postgres |

`STORE_ALLOW_MEMORY=1` is for tests only — do not use it as a sidecar data store.

## Bring-your-own Postgres

1. Create databases (idempotent script):

```sh
psql "$HOST_POSTGRES_URL" -f docker/postgres/init-host-sidecar.sql
```

2. Point the sidecar at them:

```sh
export DATABASE_URL='postgres://USER:PASS@host.docker.internal:5432/temflowral?sslmode=disable'
export TEMPORAL_POSTGRES_SEEDS=host.docker.internal
export TEMPORAL_POSTGRES_USER=USER
export TEMPORAL_POSTGRES_PWD=PASS
# Optional hardening:
# export API_AUTH_TOKEN='…'
# export API_CORS_ORIGINS='https://your-app.example'
# export HTTP_ALLOWED_HOSTS='api.example.com'
```

3. Start the sidecar compose file (no bundled Postgres container):

```sh
make run-sidecar-external-db
# equivalent:
# docker compose -f docker-compose.sidecar.yml up
```

Optional Module Federation remote container:

```sh
docker compose -f docker-compose.sidecar.yml --profile canvas-remote up
# remoteEntry at http://localhost:3002/remoteEntry.js
```

On Linux, `host.docker.internal` is provided via `extra_hosts: host-gateway` in
the overlay. From the host machine (backend run with `go run`), use
`127.0.0.1` instead of `host.docker.internal`.

The backend creates `graphs` / `runs` on first connect (`ensureSchema`). Do not
put those tables in your primary app database — names collide and Temporal must
stay separate from application metadata.

## Bundled Postgres (local demo)

```sh
make run
# docker compose --profile bundled-db up
```

Creates `temflowral` via [`docker/postgres/init-temflowral-db.sql`](../docker/postgres/init-temflowral-db.sql)
and Temporal schemas via the `temporalio/auto-setup` image.

## Environment reference

| Variable | Default / example | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | required for external PG | App metadata DSN |
| `TEMPORAL_ADDRESS` | `temporal:7233` | Worker + client |
| `TEMPORAL_NAMESPACE` | `default` | |
| `TEMPORAL_TASK_QUEUE` | `temflowral` | |
| `TEMPORAL_POSTGRES_*` | see overlay | Only when compose Temporal uses your PG |
| `API_AUTH_TOKEN` | unset (open) | Bearer shared secret |
| `API_CORS_ORIGINS` | localhost:3000 | Browser canvas / MF host origins |
| `HTTP_ALLOWED_HOSTS` | empty (deny) | HTTP node allowlist |

## Health / readiness

- API: `GET http://localhost:8080/node-types` (401 if auth required without token)
- Temporal gRPC: `:7233`
- Temporal UI (compose): `http://localhost:8233`

Authorize access in the host app before exposing temflowral — the engine is a
single-trust-domain (see [`SECURITY.md`](../SECURITY.md)).

## External Temporal (no compose Temporal)

Run only the backend against Temporal Cloud or a shared cluster:

```sh
TEMPORAL_ADDRESS=your-namespace.tmprl.cloud:7233 \
DATABASE_URL='postgres://…/temflowral?sslmode=disable' \
  docker compose up backend
```

(You may still need a compose file that omits `temporal` / `postgresql`; a
minimal override can be added as needed.)

## Go module for custom node types

Import `github.com/madmmas/temflowral/backend/pkg/nodetype` and register
activities into a temflowral-owned worker binary — see
[`adding-a-node-type.md`](adding-a-node-type.md) (External registration). Prefer
tagged module versions (`go get …@vX.Y.Z`) once releases are cut.
