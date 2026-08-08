import { type APIRequestContext, expect, test } from "@playwright/test";

/**
 * Live-backend happy path (#94): create Start→No-op via the API, open it in the
 * canvas, Run, and wait for a terminal completed status.
 *
 * Skipped unless API_BASE_URL points at a real server (compose backend).
 * Prism cannot execute Temporal workflows.
 */
const apiBaseUrl = process.env.API_BASE_URL;
const usesLiveAPI = Boolean(apiBaseUrl);

test.describe("live backend create → run", () => {
  test.skip(!usesLiveAPI, "Set API_BASE_URL to exercise the live backend");

  test("opens an API-created Start→No-op graph and completes a run", async ({
    page,
    request,
  }) => {
    const graphId = await createStartNoopGraph(request);

    await page.goto(`/?graph=${graphId}`);
    await expect(page.getByTestId("graph-editor")).toBeVisible();
    await expect(page.getByTestId("open-graph")).toHaveValue(graphId, {
      timeout: 15_000,
    });
    await expect(
      page.getByTestId("graph-canvas").getByText("Start", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByTestId("graph-canvas").getByText("No-op", { exact: true }),
    ).toBeVisible();

    await page.getByRole("button", { name: "Run" }).click();
    await expect(page.getByTestId("run-status")).toHaveText("Run completed", {
      timeout: 60_000,
    });
    await expect(page.getByTestId("run-result")).toBeVisible();
  });
});

async function createStartNoopGraph(
  request: APIRequestContext,
): Promise<string> {
  const response = await request.post(`${apiBaseUrl}/graphs`, {
    data: {
      name: `e2e live ${Date.now()}`,
      nodes: [
        {
          id: "start-1",
          type: "start",
          label: "Start",
          position: { x: 0, y: 0 },
          config: {},
        },
        {
          id: "noop-1",
          type: "noop",
          label: "No-op",
          position: { x: 220, y: 0 },
          config: {},
        },
      ],
      edges: [{ id: "e1", source: "start-1", target: "noop-1" }],
    },
  });
  expect(response.status(), await response.text()).toBe(201);
  const body = (await response.json()) as { id: string };
  expect(body.id).toBeTruthy();
  return body.id;
}
