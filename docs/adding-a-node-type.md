# Adding a node type

temflowral is contract-first: define a node's public configuration in
`api/openapi.yaml` before implementing it in Go or adding frontend behavior.
Generated Go and TypeScript files are outputs, not editing points.

The existing HTTP node is the complete reference for an activity-backed node.
The delay and condition nodes show when behavior belongs directly in Temporal
workflow code.

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

Register the parser in `ValidateNodeConfig`. This validation is called when a
graph is created and again while `BuildExecutionPlan` validates a run.

OpenAPI constraints, the parser, and the discovery schema must agree. A request
passing the HTTP transport's OpenAPI validation is not proof that a specific
`Node.type` has the right config; `ValidateNodeConfig` is the authoritative
type-specific check.

## 5. Publish discovery metadata

Add an `api.NodeType` entry to `API.ListNodeTypes` in
`backend/internal/server/api.go` with:

- `Id`: exactly the backend node-type constant;
- `Name`, `Description`, and `Category`;
- `ConfigSchema`: JSON Schema matching the OpenAPI config schema.

The frontend fetches `GET /node-types` and groups the palette by category.
Currently `ConfigSchema` is assembled in Go, so this is intentional duplication
that must be kept in sync with `api/openapi.yaml`. Update
`TestListNodeTypes` whenever this registry changes.

## 6. Add execution behavior

### Activity-backed nodes

Most nodes should be activities. Model them on the HTTP node:

1. Define a stable activity name, such as `temflowral.node.http`.
2. Implement an activity accepting `NodeActivityInput` and returning
   `NodeResult`.
3. Add `Node.type -> activity name` to `activityByNodeType` in
   `backend/internal/temporal/graph_workflow.go`.
4. Register the implementation under that exact name in
   `backend/internal/temporal/runtime.go`.

Adding the type to `activityByNodeType` also makes `isExecutableNodeType`
recognize it during planning.

The graph workflow currently sets `MaximumAttempts: 1` for node activities.
Do not increase retries globally: side-effecting activities such as HTTP POST
may not be safe to replay. If a node needs retries, design and test its
idempotency policy explicitly.

### Workflow-native control nodes

Only modify `GraphWorkflow` and planner invariants when the node changes
orchestration itself:

- delay uses `workflow.Sleep`, producing a durable timer;
- condition evaluates deterministic JSON values and routes edges using
  `Edge.sourceHandle`;
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

The current frontend does **not** generate a config form from `ConfigSchema`;
new nodes are created with `{}` config. A configured node needs a form or
editor that writes validated values to `CanvasNodeData.config` before save.
A node with named handles, such as condition's `true` and `false` outputs,
also needs a custom renderer exposing those handle IDs. Keep type-specific UI
in sibling components rather than adding node-specific logic throughout the
canvas.

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
- avoid template or expression evaluation unless it has a separately reviewed
  safe design;
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
