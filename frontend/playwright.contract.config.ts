import { defineConfig } from "@playwright/test";

/**
 * Contract conformance config (issue #19). Separate from playwright.config.ts
 * (UI e2e) on purpose: this suite is HTTP-only, needs no browser and no
 * Next.js app — just an API to validate against.
 *
 * Default target is the Prism mock so it runs in parallel with the backend.
 * Set API_BASE_URL=http://localhost:8080 to validate the live implementation.
 */
const apiBaseUrl = process.env.API_BASE_URL ?? "http://127.0.0.1:4010";

export default defineConfig({
  testDir: "./contract",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : [["list"]],
  use: {
    baseURL: apiBaseUrl,
  },
  // Playwright owns Prism unless a real backend is supplied via API_BASE_URL.
  webServer: process.env.API_BASE_URL
    ? undefined
    : {
        command: "npm run mock:api",
        url: `${apiBaseUrl}/node-types`,
        reuseExistingServer: !process.env.CI,
        timeout: 60_000,
      },
});
