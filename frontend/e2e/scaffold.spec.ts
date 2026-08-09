import { expect, test } from "@playwright/test";

test("loads the graph editor with the contract-backed node palette", async ({
  page,
}) => {
  await page.goto("/");
  await page.evaluate(() => {
    window.localStorage.removeItem("temflowral.authoringTip.dismissed");
  });
  await page.reload();

  await expect(
    page.getByRole("heading", { name: "temflowral", level: 1 }),
  ).toBeVisible();
  await expect(page.getByTestId("graph-editor")).toBeVisible();
  await expect(page.getByTestId("graph-canvas")).toBeVisible();
  await expect(page.getByTestId("node-palette")).toBeVisible();
  await expect(page.getByTestId("empty-canvas-guide")).toBeVisible();
  await expect(page.getByTestId("empty-canvas-guide")).toContainText(
    "Build a workflow",
  );
  await expect(page.getByTestId("authoring-tip")).toHaveCount(0);

  // These values come from api/openapi.yaml's Prism response example, not a
  // hardcoded frontend registry.
  await expect(page.getByTestId("node-type-start")).toBeVisible();
  await expect(page.getByTestId("node-type-http")).toBeVisible();

  await expect(page.getByRole("button", { name: "Save", exact: true })).toBeDisabled();
  await expect(page.getByRole("button", { name: "Run", exact: true })).toBeDisabled();

  const primary = page.getByTestId("primary-actions");
  const danger = page.getByTestId("danger-actions");
  await expect(primary.getByRole("button", { name: "Run", exact: true })).toBeVisible();
  await expect(danger.getByTestId("delete-graph")).toBeVisible();
  await expect(primary.getByTestId("delete-graph")).toHaveCount(0);

  await expect(page.getByTestId("toggle-minimap")).toBeVisible();
  const minimap = page.locator(".react-flow__minimap");
  await expect(minimap).toBeVisible();
  await page.getByTestId("toggle-minimap").click();
  await expect(minimap).toHaveCount(0);
  await page.getByTestId("toggle-minimap").click();
  await expect(minimap).toBeVisible();

  await page.getByTestId("node-type-start").click();
  await expect(page.getByTestId("empty-canvas-guide")).toHaveCount(0);
  await expect(page.getByTestId("authoring-tip")).toBeVisible();
  await expect(page.getByTestId("unsaved-indicator")).toHaveText("Unsaved");
  await expect(page.getByTestId("graph-name-hint")).toHaveText(
    "Name this workflow before first save",
  );

  await page.getByLabel("Graph name").fill("Scaffold named flow");
  await expect(page.getByTestId("graph-name-hint")).toHaveCount(0);

  await page
    .getByTestId("graph-canvas")
    .getByText("Start", { exact: true })
    .click();
  await expect(page.getByTestId("node-config-panel")).toBeVisible();
  await expect(page.getByTestId("node-config-editing")).toContainText(
    "Editing: Start",
  );
  await expect(page.getByTestId("node-config-resize")).toBeVisible();
});

test("remembers dismissed authoring tip across reloads", async ({ page }) => {
  await page.goto("/");
  await page.evaluate(() => {
    window.localStorage.removeItem("temflowral.authoringTip.dismissed");
  });
  await page.reload();

  await page.getByTestId("node-type-start").click();
  await expect(page.getByTestId("authoring-tip")).toBeVisible();
  await page.getByTestId("authoring-tip-dismiss").click();
  await expect(page.getByTestId("authoring-tip")).toHaveCount(0);

  await page.reload();
  await page.getByTestId("node-type-start").click();
  await expect(page.getByTestId("authoring-tip")).toHaveCount(0);
});

test("prompts for a name on first Save when still Untitled", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByTestId("node-type-start").click();

  page.once("dialog", async (dialog) => {
    expect(dialog.type()).toBe("prompt");
    expect(dialog.message()).toMatch(/Name this workflow/);
    await dialog.accept("Named from prompt");
  });

  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByLabel("Graph name")).toHaveValue("Named from prompt");
  await expect(page.getByTestId("saved-indicator")).toBeVisible({
    timeout: 10_000,
  });
});

test("opens the searchable workflow library from Open…", async ({ page }) => {
  await page.goto("/");
  await page.getByTestId("open-graph").click();
  await expect(page.getByTestId("workflow-library")).toBeVisible();
  await expect(page.getByText("Hello workflow")).toBeVisible();
  await expect(
    page.getByTestId(
      "workflow-library-item-550e8400-e29b-41d4-a716-446655440000",
    ),
  ).toContainText(/Updated/);

  await page.getByTestId("workflow-library-search").fill("hello");
  await expect(
    page.getByTestId(
      "workflow-library-item-550e8400-e29b-41d4-a716-446655440000",
    ),
  ).toBeVisible();

  await page.getByTestId("workflow-library-search").fill("zzz-no-match");
  await expect(page.getByText("No workflows match your search.")).toBeVisible();

  await page.getByTestId("workflow-library-close").click();
  await expect(page.getByTestId("workflow-library")).toHaveCount(0);
});

test("shows a dismissible banner when opening an invalid ?graph= deep link", async ({
  page,
}) => {
  const missingId = "00000000-0000-4000-8000-000000000099";
  await page.route(`**/graphs/${missingId}`, async (route) => {
    if (route.request().method() !== "GET") {
      await route.continue();
      return;
    }
    await route.fulfill({
      status: 404,
      contentType: "application/json",
      body: JSON.stringify({ message: "graph not found" }),
    });
  });

  await page.goto(`/?graph=${missingId}`);
  await expect(page.getByTestId("canvas-error-banner")).toBeVisible();
  await expect(page.getByTestId("canvas-error-message")).toContainText(
    "graph not found",
  );
  await expect(page).not.toHaveURL(/graph=/);

  await page.getByTestId("canvas-error-dismiss").click();
  await expect(page.getByTestId("canvas-error-banner")).toHaveCount(0);
  await expect(page.getByTestId("graph-canvas")).toBeVisible();
});

test("fits the viewport after opening a graph with far-off nodes", async ({
  page,
}) => {
  // Distinguishes off-viewport (#111) from sparse “lone Start” data: without a
  // post-load fitView, these nodes sit far outside the default camera.
  const graphId = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee";
  await page.route(`**/graphs/${graphId}`, async (route) => {
    if (route.request().method() !== "GET") {
      await route.continue();
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        id: graphId,
        name: "Far away workflow",
        createdAt: "2026-08-09T00:00:00Z",
        updatedAt: "2026-08-09T00:00:00Z",
        nodes: [
          {
            id: "start-far",
            type: "start",
            label: "Far Start",
            position: { x: 8000, y: 8000 },
            config: {},
          },
        ],
        edges: [],
      }),
    });
  });

  await page.goto(`/?graph=${graphId}`);
  const canvas = page.getByTestId("graph-canvas");
  const node = canvas.locator(".react-flow__node").filter({
    hasText: "Far Start",
  });
  await expect(node).toBeVisible({ timeout: 10_000 });

  const canvasBox = await canvas.boundingBox();
  const nodeBox = await node.boundingBox();
  expect(canvasBox).toBeTruthy();
  expect(nodeBox).toBeTruthy();
  expect(nodeBox!.x + nodeBox!.width).toBeGreaterThan(canvasBox!.x);
  expect(nodeBox!.x).toBeLessThan(canvasBox!.x + canvasBox!.width);
  expect(nodeBox!.y + nodeBox!.height).toBeGreaterThan(canvasBox!.y);
  expect(nodeBox!.y).toBeLessThan(canvasBox!.y + canvasBox!.height);
});
