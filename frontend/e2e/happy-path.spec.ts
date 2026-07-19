import { expect, test } from "@playwright/test";

const usesPrismMock = !process.env.API_BASE_URL;

test("builds a graph, runs it, and shows the result", async ({ page }) => {
  if (usesPrismMock) {
    // Prism's contract example stays in "running". Complete only the polling
    // response so this mock-backed test exercises the UI's terminal state and
    // result rendering. With API_BASE_URL set, the real API remains untouched.
    await page.route("**/runs/*", async (route) => {
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
  }

  await page.goto("/");
  await expect(page.getByTestId("node-type-start")).toBeVisible();

  await page.getByLabel("Graph name").fill("Happy path workflow");
  await page.getByTestId("node-type-start").click();
  await expect(
    page.getByTestId("graph-canvas").getByText("Start", { exact: true }),
  ).toBeVisible();

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
  await expect(page.getByTestId("run-result")).toBeVisible();
  if (usesPrismMock) {
    await expect(page.getByTestId("run-result")).toContainText(
      "Workflow completed",
    );
  }
});
