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

  await expect(page.getByTestId("toggle-minimap")).toBeVisible();
  const minimap = page.locator(".react-flow__minimap");
  await expect(minimap).toBeVisible();
  await page.getByTestId("toggle-minimap").click();
  await expect(minimap).toHaveCount(0);
  await page.getByTestId("toggle-minimap").click();
  await expect(minimap).toBeVisible();
});
