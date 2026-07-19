import { defineConfig, devices } from "@playwright/test";

const appBaseUrl =
  process.env.PLAYWRIGHT_BASE_URL ?? "http://127.0.0.1:3000";
const apiBaseUrl = process.env.API_BASE_URL ?? "http://127.0.0.1:4010";

const webServer = [
  // By default Playwright owns Prism. Set API_BASE_URL to test another API.
  ...(!process.env.API_BASE_URL
    ? [
        {
          command: "npm run mock:api",
          url: `${apiBaseUrl}/node-types`,
          reuseExistingServer: !process.env.CI,
          timeout: 60_000,
        },
      ]
    : []),
  // By default Playwright owns Next.js. Set PLAYWRIGHT_BASE_URL for an existing app.
  ...(!process.env.PLAYWRIGHT_BASE_URL
    ? [
        {
          command: "npm run dev -- --hostname 127.0.0.1",
          url: appBaseUrl,
          reuseExistingServer: !process.env.CI,
          timeout: 120_000,
          env: {
            NEXT_PUBLIC_API_BASE_URL: apiBaseUrl,
          },
        },
      ]
    : []),
];

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [["list"], ["html", { open: "never" }]]
    : [["list"]],
  use: {
    baseURL: appBaseUrl,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer,
});
