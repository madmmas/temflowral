import { type APIRequestContext, expect, test } from "@playwright/test";
import { formatErrors, validatorFor } from "./schema-validator";

/**
 * Contract conformance suite (issue #19).
 *
 * Asserts that API responses actually match the schemas declared in
 * api/openapi.yaml. It runs against the Prism mock by default so it can move
 * in parallel with the backend; point it at the live server with
 * `API_BASE_URL=http://localhost:8080 npm run test:contract` to catch real
 * drift between the implementation and the spec.
 *
 * Kept deliberately separate from the UI e2e specs in ../e2e per the testing
 * conventions: this suite talks HTTP only and never opens a browser.
 */

function assertMatchesSchema(schemaName: string, data: unknown): void {
  const validate = validatorFor(schemaName);
  const valid = validate(data);
  expect(valid, formatErrors(schemaName, validate.errors)).toBe(true);
}

const sampleGraph = {
  name: "Conformance workflow",
  nodes: [
    { id: "start-1", type: "start", label: "Start", position: { x: 0, y: 0 }, config: {} },
    {
      id: "http-1",
      type: "http",
      label: "Fetch data",
      position: { x: 200, y: 0 },
      config: { method: "GET", url: "https://httpbin.org/get" },
    },
  ],
  edges: [{ id: "e-start-http", source: "start-1", target: "http-1" }],
};

async function createGraph(
  request: APIRequestContext,
): Promise<{ id: string }> {
  const response = await request.post("/graphs", { data: sampleGraph });
  expect(response.status(), await response.text()).toBe(201);
  const body = await response.json();
  assertMatchesSchema("Graph", body);
  return body;
}

test.describe("contract conformance", () => {
  // Guards against the validator silently degrading into a no-op: a payload
  // that clearly violates the Run schema must be rejected.
  test("schema validator rejects contract violations", () => {
    const validate = validatorFor("Run");
    expect(validate({ id: "not-a-uuid", status: "bogus" })).toBe(false);
  });

  test("GET /node-types returns a valid NodeTypeList", async ({ request }) => {
    const response = await request.get("/node-types");
    expect(response.status(), await response.text()).toBe(200);
    assertMatchesSchema("NodeTypeList", await response.json());
  });

  test("POST /graphs returns a valid Graph", async ({ request }) => {
    await createGraph(request);
  });

  test("GET /graphs/{id} returns a valid Graph", async ({ request }) => {
    const { id } = await createGraph(request);
    const response = await request.get(`/graphs/${id}`);
    expect(response.status(), await response.text()).toBe(200);
    assertMatchesSchema("Graph", await response.json());
  });

  test("GET /graphs returns a valid GraphList", async ({ request }) => {
    await createGraph(request);
    const response = await request.get("/graphs");
    expect(response.status(), await response.text()).toBe(200);
    assertMatchesSchema("GraphList", await response.json());
  });

  test("PUT /graphs/{id} returns a valid Graph", async ({ request }) => {
    const { id } = await createGraph(request);
    const response = await request.put(`/graphs/${id}`, {
      data: {
        name: "Updated conformance workflow",
        nodes: sampleGraph.nodes,
        edges: sampleGraph.edges,
      },
    });
    expect(response.status(), await response.text()).toBe(200);
    const body = await response.json();
    assertMatchesSchema("Graph", body);
    // Prism may return the OpenAPI example; live API preserves the path id.
    if (process.env.API_BASE_URL) {
      expect(body.id).toBe(id);
      expect(body.name).toBe("Updated conformance workflow");
    }
  });

  test("DELETE /graphs/{id} returns 204", async ({ request }) => {
    const { id } = await createGraph(request);
    const response = await request.delete(`/graphs/${id}`);
    expect(response.status(), await response.text()).toBe(204);
    if (process.env.API_BASE_URL) {
      const missing = await request.get(`/graphs/${id}`);
      expect(missing.status()).toBe(404);
    }
  });

  test("POST /graphs/{id}/run returns a valid Run", async ({ request }) => {
    const { id } = await createGraph(request);
    const response = await request.post(`/graphs/${id}/run`, {
      data: { input: { message: "hello" } },
    });
    expect(response.status(), await response.text()).toBe(202);
    assertMatchesSchema("Run", await response.json());
  });

  test("GET /runs/{id} returns a valid Run", async ({ request }) => {
    const { id: graphId } = await createGraph(request);
    const started = await request.post(`/graphs/${graphId}/run`, {
      data: { input: {} },
    });
    expect(started.status(), await started.text()).toBe(202);
    const run = await started.json();
    assertMatchesSchema("Run", run);

    const response = await request.get(`/runs/${run.id}`);
    expect(response.status(), await response.text()).toBe(200);
    assertMatchesSchema("Run", await response.json());
  });

  test("POST /runs/{id}/signal returns a valid SignalRunResponse", async ({
    request,
  }) => {
    const { id: graphId } = await createGraph(request);
    const started = await request.post(`/graphs/${graphId}/run`, {
      data: { input: {} },
    });
    expect(started.status(), await started.text()).toBe(202);
    const run = await started.json();

    const response = await request.post(`/runs/${run.id}/signal`, {
      data: {
        signal: "approval.granted",
        payload: { approvedBy: "alice" },
      },
    });
    expect(response.status(), await response.text()).toBe(202);
    assertMatchesSchema("SignalRunResponse", await response.json());
  });
});
