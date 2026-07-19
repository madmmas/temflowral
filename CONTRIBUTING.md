# Contributing to temflowral

Thanks for your interest in improving temflowral. This guide covers local
setup, the development workflow, and the conventions we expect on pull
requests.

## Development setup

You need:

- **Go 1.25.7** for the backend.
- **Node 20** (or newer) for the frontend and tooling.
- **Docker** for Temporal, Postgres, and the full-stack compose flow.

Clone the repository and enable the versioned git hooks once:

```sh
make hooks
```

The hooks block direct commits and pushes to `main` and run the backend linter
before each commit.

## Branch workflow

Never commit on `main` locally. Create a branch first:

```sh
git checkout -b your-branch
```

We link branches to issues (for example via `gh issue develop`) and open a pull
request against `main`. Keep each PR focused on a single issue.

## Running tests

Run the full backend and frontend test suites from the repository root:

```sh
make test
```

Other useful targets: `make lint` (see below), `make e2e` (Playwright), and
`make test-contract` (contract conformance against the OpenAPI spec).

## Adding a new node type

Node types are the primary extension point. Follow the contract-first,
step-by-step recipe in
[`docs/adding-a-node-type.md`](docs/adding-a-node-type.md): edit
`api/openapi.yaml`, regenerate both clients, then implement backend validation,
execution, frontend behavior, tests, and security controls.

## Linting

Run both backend and frontend linters from the repository root:

```sh
make lint
```

The backend runner prefers an installed golangci-lint v2.12.2 and otherwise
uses `go run` at that pinned version, so no separate linter installation is
required. The pre-commit hook delegates to the same runner, and CI uses the
same version.

Backend rules live in `backend/.golangci.yml`. Validate configuration changes
before opening a PR:

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 \
  config verify -c backend/.golangci.yml
```

## Run the contract-backed mock API

The frontend and Playwright tracks can develop against a local mock before the
backend implements the API. Prism reads `api/openapi.yaml`, validates requests,
and returns the response examples defined in the contract.

From the repository root, with Node 20 or newer installed:

```sh
npx --yes @stoplight/prism-cli@5.14.2 mock api/openapi.yaml
```

Prism listens at `http://127.0.0.1:4010` by default. The version is pinned
because Prism 5.16 and newer require Node 24, while this project currently
targets Node 20.

In another terminal, verify the mock:

```sh
curl --fail-with-body http://127.0.0.1:4010/node-types
curl --fail-with-body \
  http://127.0.0.1:4010/graphs/550e8400-e29b-41d4-a716-446655440000
```

Use these base URLs while developing:

- Frontend: `NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:4010`
- Playwright and other Node-based tests:
  `API_BASE_URL=http://127.0.0.1:4010`

Keep API calls behind the generated client and read the base URL from the
appropriate environment variable. To test against the real local backend
later, change only the value to `http://127.0.0.1:8080`; do not change endpoint
paths or hand-written response types. Local values can go in `.env.local`,
which is ignored by Git.

Stop Prism with `Ctrl+C`.

## Run the full stack with one command

`docker-compose.yml` runs the whole stack — Postgres, Temporal (server + Web
UI), the backend, and the frontend — from the repository root:

```sh
docker compose up   # or: make run
```

First boot builds the backend and frontend images and initializes the Temporal
schema in Postgres, so allow a minute. Once healthy:

- Frontend: `http://localhost:3000`
- Backend API + docs: `http://localhost:8080` / `http://localhost:8080/docs`
- Temporal Web UI: `http://localhost:8233`

The frontend image bakes `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080` at
build time (browsers reach the backend via its published port, not the compose
service name). To allow HTTP activity nodes to reach an external host, set
`HTTP_ALLOWED_HOSTS` before `docker compose up` (see `SECURITY.md`). Stop with
`Ctrl+C`, then `make temporal-down` to remove the containers; the Postgres
volume persists across restarts. Use `docker compose down -v` to wipe it.

## Run the backend with Temporal

Prefer developing the backend outside a container? Temporal still runs from
`docker-compose.yml` — server, Web UI, and Postgres persistence via a pinned
`temporalio/auto-setup` image. You do not need to install Temporal locally.

Start Temporal (Postgres is pulled in automatically) from the repository root:

```sh
make temporal-dev
```

This serves Temporal at `localhost:7233` and its Web UI at
`http://localhost:8233`. Stop it with `Ctrl+C`, then `make temporal-down` to
remove the containers.

In a second terminal, start the backend:

```sh
make run-backend
```

The backend connects to Temporal before listening on port 8080 and shuts down
its worker cleanly on `SIGINT` or `SIGTERM`. Override local defaults when
needed:

- `TEMPORAL_ADDRESS` (default `localhost:7233`)
- `TEMPORAL_NAMESPACE` (default `default`)
- `TEMPORAL_TASK_QUEUE` (default `temflowral`)
- `HTTP_ALLOWED_HOSTS` (default empty/deny all; comma-separated exact
  hostnames permitted for HTTP activity nodes)

For example, allow the contract's HTTP-node example while developing:

```sh
HTTP_ALLOWED_HOSTS=httpbin.org make run-backend
```

Schemes, ports, wildcards, localhost, and private IPs are not valid allowlist
entries. See `SECURITY.md` for the full outbound-request policy.

To prove the client and worker path end to end, run the registered smoke
workflow from a third terminal. This uses the CLI already inside the dev
container, so nothing extra is installed:

```sh
make temporal-smoke
```

The workflow runs one activity and returns `"hello"`.

To exercise the graph translator over HTTP (create → run → poll):

```sh
GRAPH_ID=$(curl -sS -X POST http://127.0.0.1:8080/graphs \
  -H 'content-type: application/json' \
  -d '{
    "name":"smoke",
    "nodes":[
      {"id":"start-1","type":"start","position":{"x":0,"y":0},"config":{}},
      {"id":"noop-1","type":"noop","position":{"x":200,"y":0},"config":{}}
    ],
    "edges":[{"id":"e1","source":"start-1","target":"noop-1"}]
  }' | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

RUN_ID=$(curl -sS -X POST "http://127.0.0.1:8080/graphs/${GRAPH_ID}/run" \
  -H 'content-type: application/json' \
  -d '{"input":{"message":"hello"}}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

curl -sS "http://127.0.0.1:8080/runs/${RUN_ID}"
```

Poll until `status` is `completed`. Stop the backend and development server with
`Ctrl+C`, then `make temporal-down`.

Prefer a native install instead of Docker? Install the
[Temporal CLI](https://docs.temporal.io/cli/setup-cli) and substitute
`temporal server start-dev` for `make temporal-dev`, and `temporal workflow
execute ...` for `make temporal-smoke`.

## Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/). Use a
type prefix and a concise, imperative subject, for example:

```text
feat(backend): add durable delay/wait node
fix(frontend): guard against empty run result
docs: explain how to add node types
chore(lint): make golangci checks reproducible
```

Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `infra`.
Reference the issue in the body or subject, and use `Closes #NN` in the PR so
the issue closes on merge.

## Pull request checklist

Before requesting review, confirm:

- [ ] `make test` passes (backend and frontend).
- [ ] `make lint` passes.
- [ ] API changes started in `api/openapi.yaml`, and generated code was
      regenerated with `make generate` (not hand-edited).
- [ ] Documentation is updated, including a new `DEVLOG.md` entry for the work
      session.
- [ ] The PR is scoped to one issue and links it with `Closes #NN`.

## Questions

Open a thread in the repository's **Discussions** tab, or comment on the
relevant issue. For security concerns, follow [`SECURITY.md`](SECURITY.md)
instead of filing a public issue.
