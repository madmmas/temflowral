import { expect, test } from "@playwright/test";

test("loads the graph editor with the contract-backed node palette", async ({
  page,
}) => {
  await page.goto("/");

  await expect(
    page.getByRole("heading", { name: "temflowral", level: 1 }),
  ).toBeVisible();
  await expect(page.getByTestId("graph-editor")).toBeVisible();
  await expect(page.getByTestId("graph-canvas")).toBeVisible();
  await expect(page.getByTestId("node-palette")).toBeVisible();

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
  await expect(page.getByTestId("unsaved-indicator")).toHaveText("Unsaved");
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
