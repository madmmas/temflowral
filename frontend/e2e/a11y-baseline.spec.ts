import { expect, test } from "@playwright/test";

const usesPrismMock = !process.env.API_BASE_URL;

/**
 * Accessibility baseline (#114): keyboard-operable happy path for primary
 * controls. Not a full WCAG audit — see docs/canvas-accessibility.md.
 */
test("keyboard path reaches palette, config, and Run; Escape closes drawers", async ({
  page,
}) => {
  if (usesPrismMock) {
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
    test.skip(true, "Prism-backed keyboard path only");
  }

  await page.goto("/");

  await expect(page.getByRole("link", { name: "Skip to canvas" })).toHaveCount(
    1,
  );
  await expect(page.getByRole("toolbar", { name: "Graph actions" })).toBeVisible();
  await expect(page.getByRole("complementary", { name: "Node types" })).toBeVisible();

  await page.getByRole("button", { name: "Add Start node" }).focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByTestId("graph-canvas").getByText("Start", { exact: true }),
  ).toBeVisible();

  await page.getByLabel("Graph name").fill("A11y keyboard flow");

  await page
    .getByTestId("graph-canvas")
    .getByText("Start", { exact: true })
    .click();
  await expect(page.getByTestId("node-config-panel")).toBeVisible();
  await expect(page.getByLabel("Node label")).toBeFocused();

  await page.getByLabel("Node label").fill("Entry");
  // Backspace while editing must not delete the canvas node.
  await page.keyboard.press("Backspace");
  await expect(
    page.getByTestId("graph-canvas").getByText("Entr", { exact: true }),
  ).toBeVisible();
  await expect(page.locator('[data-testid^="workflow-node-"]')).toHaveCount(1);

  await page.keyboard.press("Escape");
  await expect(page.getByTestId("node-config-panel")).toHaveCount(0);

  await page.getByRole("button", { name: "Open workflow library" }).click();
  await expect(page.getByTestId("workflow-library")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByTestId("workflow-library")).toHaveCount(0);

  await page.getByRole("button", { name: "Run", exact: true }).press("Enter");
  await expect(page.getByTestId("run-status")).toHaveText("Run completed", {
    timeout: 10_000,
  });
  await expect(page.getByTestId("run-result-panel")).toBeVisible();

  await page.keyboard.press("Escape");
  await expect(page.getByTestId("run-result-panel")).toHaveCount(0);
});
