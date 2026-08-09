import { expect, test } from "@playwright/test";

const usesPrismMock = !process.env.API_BASE_URL;

test("builds a graph from the palette, runs it, and shows the result", async ({
  page,
}) => {
  if (usesPrismMock) {
    // Prism's contract example stays in "running". Complete only the polling
    // response so this mock-backed test exercises the UI's terminal state and
    // result rendering. Start→No-op against Temporal is covered by
    // live-run.spec.ts (API create with edges + canvas Run).
    await page.route("**/runs/*", async (route) => {
      if (route.request().method() !== "GET") {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          id: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
          graphId: "550e8400-e29b-41d4-a716-446655440000",
          status: "completed",
          startedAt: "2026-07-16T08:01:00Z",
          completedAt: "2026-07-16T08:01:01Z",
          result: { message: "Workflow completed" },
        }),
      });
    });
  } else {
    test.skip(
      true,
      "Live Temporal create→run is covered by live-run.spec.ts",
    );
  }

  await page.goto("/");
  await expect(page.getByTestId("node-type-start")).toBeVisible();
  await expect(page.getByTestId("node-type-noop")).toBeVisible();

  await page.getByLabel("Graph name").fill("Happy path workflow");
  await page.getByTestId("node-type-start").click();
  await expect(
    page.getByTestId("graph-canvas").getByText("Start", { exact: true }),
  ).toBeVisible();

  // Start-only is a valid Temporal graph. Connecting Start→No-op in the UI is
  // covered manually / via seed-demo; automated Start→No-op uses live-run.
  const createGraphRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" && new URL(request.url()).pathname === "/graphs",
  );
  await page.getByRole("button", { name: "Run" }).click();

  const graphPayload = (await createGraphRequest).postDataJSON();
  expect(graphPayload).toMatchObject({
    name: "Happy path workflow",
    nodes: [{ type: "start", label: "Start", config: {} }],
    edges: [],
  });

  await expect(page.getByTestId("run-status")).toHaveText("Run completed", {
    timeout: 10_000,
  });
  await expect(page.getByTestId("graph-id-chip")).toBeVisible();
  await expect(page.getByTestId("graph-id-chip")).toContainText("Graph:");
  await expect(page.getByTestId("copy-graph-id")).toBeVisible();
  await expect(page.getByTestId("run-history")).toBeVisible();
  await expect(page.getByTestId("run-history")).toContainText("completed");
  await expect(page.getByTestId("run-result-panel")).toBeVisible();
  await expect(page.getByTestId("run-result")).toBeVisible();
  await expect(page.getByTestId("run-result")).toContainText(
    "Workflow completed",
  );

  await page.getByTestId("run-result-collapse").click();
  await expect(page.getByTestId("run-result")).toHaveCount(0);
  await page.getByTestId("run-result-collapse").click();
  await expect(page.getByTestId("run-result")).toBeVisible();

  await page.getByTestId("run-result-dismiss").click();
  await expect(page.getByTestId("run-result-panel")).toHaveCount(0);
  await page.getByTestId("run-status").click();
  await expect(page.getByTestId("run-result-panel")).toBeVisible();
  await expect(page.getByTestId("run-result")).toContainText(
    "Workflow completed",
  );
});
