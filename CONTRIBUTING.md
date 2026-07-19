1. Development setup (Go 1.25.7, Node 20, Docker)
2. Install git hooks once per clone: `make hooks` (blocks commits/pushes to `main`)
3. Branch workflow: never commit on `main` locally — `git checkout -b your-branch` first
4. Running tests: make test
5. Adding a new node type (step-by-step)
6. PR checklist: tests pass, docs updated
7. Commit message format (conventional commits)
8. Where to ask questions (Discussions tab)

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

## Run the backend with Temporal

Install the
[Temporal CLI](https://docs.temporal.io/cli/setup-cli), then start its local
development server:

```sh
temporal server start-dev
```

This provides Temporal at `localhost:7233` and its Web UI at
`http://localhost:8233`. Temporalite is deprecated; the CLI development server
is its supported replacement.

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

To prove the client and worker path end to end, execute the registered smoke
workflow from a third terminal:

```sh
temporal workflow execute \
  --workflow-id "temflowral-smoke-$(date +%s)" \
  --type temflowral.noop \
  --task-queue temflowral \
  --input '"hello"'
```

The workflow runs one activity and returns `"hello"`. Stop the backend and
development server with `Ctrl+C`.
