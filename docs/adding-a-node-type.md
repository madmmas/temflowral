# Adding a node type

temflowral is contract-first: define a node's public configuration in
`api/openapi.yaml` before implementing it in Go or adding frontend behavior.
Generated Go and TypeScript files are outputs, not editing points.

The existing HTTP node is the complete reference for an activity-backed node.
The delay, condition, wait, and childWorkflow nodes show when behavior belongs
directly in Temporal workflow code.

## 1. Design the node

Before changing code, decide:

- the stable `Node.type` identifier, display name, category, and description;
- its input and output shape;
- required and optional configuration fields, limits, and defaults;
- whether it has one output or named output handles;
- whether retries are safe and the operation is idempotent;
- which inputs, outputs, and errors may contain secrets;
- whether execution is an activity or workflow-native behavior.

Use a **Temporal activity** for I/O and non-deterministic work such as HTTP,
database, filesystem, clock, or random operations. Use workflow code only for
durable orchestration primitives such as timers and branch selection. Workflow
code must remain deterministic: do not perform network calls, use system time,
start goroutines, or generate random values there.

## 2. Update the OpenAPI contract first

Add a named config schema under `components.schemas` in `api/openapi.yaml`.
`HttpNodeConfig` is the main example:

```yaml
HttpNodeConfig:
  type: object
  required: [method, url]
  additionalProperties: false
  properties:
    method:
      type: string
      enum: [GET, POST, PUT, PATCH, DELETE]
    url:
      type: string
      format: uri
      maxLength: 2048
  example:
    method: GET
    url: https://httpbin.org/get
```

Prefer explicit limits and `additionalProperties: false`. Document security and
runtime semantics next to the affected field. Give the schema a realistic
example so the generated docs and Prism mock remain useful.

Then:

1. Add the schema reference to `Node.config.anyOf`.
2. Update the `Node.config` description.
3. Add or update examples that use the node.
4. Add the node to the `NodeTypeList` example returned by `GET /node-types`;
   the Prism-backed frontend and E2E tests consume this example.

Keep `Node.config`'s existing `x-go-type: map[string]interface{}`. It preserves
the generic graph representation while the named config schema still generates
a strongly typed Go type such as `api.HttpNodeConfig`.

The broad object fallback in `Node.config.anyOf` permits multiple node types in
the generic graph model. It does not replace type-specific runtime validation.

## 3. Regenerate both clients

From the repository root:

```sh
make generate
```

This updates:

- `backend/internal/api/openapi.gen.go`
- `frontend/src/api/generated/schema.ts`

Do not edit either generated file by hand. Review the generated diff and verify
that the new named config type exists on both sides.

## 4. Implement strict backend validation

Add a focused file in `backend/internal/temporal/`, for example
`http_node.go`, and define a stable node-type constant:

```go
const HTTPNodeType = "http"
```

Decode the generic `Node.Config` into the generated config type. Follow
`parseHTTPNodeConfig`, `parseDelayNodeConfig`, and
`parseConditionNodeConfig`:

- reject a missing config when the node requires one;
- use `json.Decoder.DisallowUnknownFields`;
- validate semantic rules that OpenAPI alone cannot enforce;
- explicitly detect required fields whose missing JSON value could decode to a
  valid Go zero value;
- bound strings, collections, payloads, and durations;
- return useful errors without exposing secrets or user-controlled URLs.

Register the parser in `ValidateConfig` on the registry definition (wired through
`ValidateNodeConfig`). Config validation runs when a graph is created and again
while `ValidateGraph` / `BuildExecutionPlan` gates a run. Create also rejects
node types missing from the registry; full topology checks (cycles, reachability,
single start) run at `POST .../run` before Temporal starts.

OpenAPI constraints, the parser, and the discovery schema must agree. A request
passing the HTTP transport's OpenAPI validation is not proof that a specific
`Node.type` has the right config; `ValidateNodeConfig` is the authoritative
type-specific check.

## 5. Publish discovery metadata

Built-in types are registered in `RegisterBuiltins`
(`backend/internal/temporal/builtins.go`). That registration is the source of
`GET /node-types`: each `nodetype.Definition` supplies `Id`, `Name`,
`Description`, `Category`, `ConfigSchema`, and optional output-handle
metadata (`OutputHandles` or `OutputHandlesFromConfig`).

`API.ListNodeTypes` maps the shared registry — do not hand-maintain a parallel
slice in `api.go`. Update `TestListNodeTypes` when built-in discovery metadata
changes.

The frontend fetches `GET /node-types` and groups the palette by category.
`ConfigSchema` is still assembled in Go beside the OpenAPI named config schema,
so keep those in sync.

### External registration (outside this repo)

Adopters register custom activity-backed types without forking. Share one
`*nodetype.Registry` between worker startup and the HTTP API:

```go
registry := nodetype.NewRegistry()
if err := temporal.RegisterBuiltins(registry, temporal.BuiltinOptions{
    HTTPAllowedHosts: []string{"api.example.com"},
}); err != nil {
    log.Fatal(err)
}
if err := registry.Register(nodetype.Definition{
    ID:           "billing.charge",
    Name:         "Charge",
    Kind:         nodetype.KindActivity,
    ConfigSchema: map[string]any{ /* JSON Schema */ },
    // Fixed handles, or OutputHandlesFromConfig for config-derived handles:
    OutputHandlesFromConfig: &nodetype.HandlesFromConfig{Path: "branches"},
    ActivityName: "billing.activity.charge",
    Activity:     ChargeActivity, // func(ctx, nodetype.ActivityInput) (nodetype.Result, error)
    ValidateConfig: func(nodeID string, config map[string]any) error {
        // reject bad config; never put secrets in the error
        return nil
    },
}); err != nil {
    log.Fatal(err)
}

runtime, err := temporal.Start(config, temporal.WithRegistry(registry))
apiServer := server.NewAPI(store, runtime, registry)
```

Multi-output activity nodes select a branch by setting
`result.Value["branch"]` to a handle ID. The planner validates
`Edge.sourceHandle` against fixed or config-derived handles. Workflow-native
kinds (`KindWorkflow`) remain reserved for built-in orchestration (start,
delay, condition, wait, childWorkflow).

A fuller external-package walkthrough is tracked as issue #67.

## 6. Add execution behavior

### Activity-backed nodes

Most nodes should be activities. Model them on the HTTP node:

1. Define a stable activity name, such as `temflowral.node.http`.
2. Implement an activity accepting `nodetype.ActivityInput` and returning
   `nodetype.Result` (aliased as `NodeActivityInput` / `NodeResult` in the
   temporal package).
3. Register a `nodetype.Definition` with `KindActivity`, that activity name,
   and the implementation (built-ins: `RegisterBuiltins`; in-repo additions:
   extend `RegisterBuiltins`).
4. `temporal.Start` registers every `KindActivity` on the worker from the
   shared registry — do not add a one-off `RegisterActivityWithOptions` call
   for each new built-in.

Planning and `GraphWorkflow` resolve activity names through the registry.
Before a node runs, GraphWorkflow resolves `{{ nodes.<id>.output.<path> }}`
templates in that node's `config` string leaves using active predecessor
outputs (the nested `graph` key on `childWorkflow` is skipped).

The graph workflow defaults node activities to `StartToCloseTimeout: 30s` and
`MaximumAttempts: 1`. Do not raise retries globally: side-effecting activities
such as HTTP POST may not be safe to replay. Callers may override per activity
node via optional `Node.activityOptions` (timeouts + `retryPolicy`) and
`Node.taskQueue` (route the activity to a specialized worker queue); only raise
retries for idempotent work. Workflow-native nodes (start, delay, condition,
wait, childWorkflow) reject both fields.

### Workflow-native control nodes

Only modify `GraphWorkflow` and planner invariants when the node changes
orchestration itself:

- delay uses `workflow.Sleep`, producing a durable timer;
- condition evaluates deterministic JSON values and routes edges using
  `Edge.sourceHandle`;
- wait races `workflow.GetSignalChannel` against a durable timer and routes
  via `received` / `timedOut` handles (HTTP delivery is
  `POST /runs/{runId}/signal`, which queries `temflowral.currentWait` before
  calling Temporal `SignalWorkflow`);
- childWorkflow runs `workflow.ExecuteChildWorkflow` with an inline nested
  graph (same `GraphWorkflow`); nesting another `childWorkflow` is rejected;
- branch-specific graph rules live in planner validation.

Add workflow-native types to `isExecutableNodeType`. For named handles, validate
the graph shape and ensure `activeInputs` includes only edges on the selected
path.

## 7. Add frontend behavior

Basic discovery and rendering are generic:

- `src/lib/node-types.ts` fetches and groups node types;
- `src/components/node-palette.tsx` renders the palette;
- `src/components/nodes/workflow-node.tsx` renders a generic node;
- `src/lib/graph-canvas.ts` serializes generic config and edge handles.

Therefore, an unconfigured node with one input and one output normally appears
without a frontend code change after `GET /node-types` exposes it.

The `frontend/` app is a reference canvas only (not a published package or
embed). External UIs should use the same discovery + graph APIs — see
[`docs/adr/001-canvas-packaging.md`](adr/001-canvas-packaging.md).

The current frontend does **not** generate a config form from `ConfigSchema`;
new nodes are created with `{}` config. A configured node needs a form or
editor that writes validated values to `CanvasNodeData.config` before save.
A node with named handles, such as condition's `true`/`false` or wait's
`received`/`timedOut` outputs, also needs a custom renderer exposing those
handle IDs. Keep type-specific UI in sibling components rather than adding
node-specific logic throughout the canvas.

Continue using the generated API client and generated schema types. Do not add
handwritten fetch calls or duplicate request/response interfaces.

## 8. Test every layer

At minimum, add or update:

- `<node>_test.go`: valid, missing, malformed, unknown-field, and boundary
  configurations;
- `plan_test.go`: acceptance, invalid config, and node-specific graph
  invariants;
- `graph_workflow_test.go`: activity dispatch or workflow-native behavior;
- `server/api_test.go`: discovery metadata and API-level config rejection;
- frontend unit/component tests for custom editors, renderers, config
  serialization, and named handles;
- Playwright coverage for palette visibility and a representative graph path;
- contract tests/examples when the endpoint payload shapes change.

For an I/O node, test policy boundaries, timeouts, size limits, redirects,
error redaction, and failure behavior—not only the happy path.

## 9. Security checklist

Read `SECURITY.md` before adding an integration node. In particular:

- deny external destinations by default and use explicit operator policy;
- prevent SSRF after DNS resolution and again after redirects;
- reject unknown fields and bound all user-controlled data;
- avoid inventing a second template engine per node type — use the shared
  `{{ nodes.<id>.output.<path> }}` resolver in GraphWorkflow and revalidate
  rendered HTTP requests through the existing URL/header policy;
- never include credentials, request bodies, or sensitive URLs in errors;
- enforce deployment-specific checks inside the activity, not only at graph
  save time;
- document retry and idempotency behavior.

The HTTP implementation in `http_node.go` is the reference for allowlisting,
private-address blocking, bounded bodies/headers, redirect checks, disabled
proxies, timeouts, and redacted errors.

## 10. Verify the change

Run from the repository root:

```sh
make generate
npx --yes @redocly/cli@2.39.0 lint api/openapi.yaml --config redocly.yaml
make test
make lint
make test-contract
```

Confirm generation is clean after it has been committed:

```sh
make generate
git diff --exit-code -- \
  backend/internal/api/openapi.gen.go \
  frontend/src/api/generated/schema.ts
```

For a live execution check, run the full stack:

```sh
docker compose up
```

Create a graph containing the new node, start it through
`POST /graphs/{graphId}/run`, poll `GET /runs/{runId}`, and inspect the same run
at `http://localhost:8233`. Test both success and a representative failure.

Finally, add a work-session entry to `DEVLOG.md` describing the contract,
implementation decisions, security considerations, and verification performed.

## Pull request checklist

- [ ] `api/openapi.yaml` was changed before implementation.
- [ ] Generated Go and TypeScript files were regenerated, not hand-edited.
- [ ] OpenAPI schema, backend parser, and `GET /node-types` schema agree.
- [ ] The node is executable through an activity or deterministic workflow
      behavior.
- [ ] Config, planner, workflow, API, frontend, and contract tests were updated
      as applicable.
- [ ] Security bounds, error redaction, and retry semantics were reviewed.
- [ ] `DEVLOG.md` was updated.
- [ ] The verification commands above pass.
