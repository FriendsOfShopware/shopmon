import { defineConfig, devices } from "@playwright/test";

// Minimal local typing for env access — the root tsconfig targets the browser
// (Vite types) and does not pull in @types/node, so `process` is otherwise
// untyped here. Scoped to this file to avoid clobbering Vite's import.meta.env.
declare const process: { env: Record<string, string | undefined> };

const BASE_URL = process.env.E2E_BASE_URL ?? "http://localhost:3000";

export default defineConfig({
  testDir: "./e2e",
  // Generated auth state lives here; ignored by git.
  outputDir: "./e2e/.output",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  // Cap local concurrency: the dev API session endpoint is the shared bottleneck,
  // and too many parallel cold-session loads race the router's auth guard.
  workers: process.env.CI ? 1 : 4,
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : "list",

  use: {
    baseURL: BASE_URL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  projects: [
    // Authenticates once and saves storage state for the authenticated projects.
    { name: "setup", testMatch: /auth\.setup\.ts/ },

    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        storageState: "e2e/.auth/user.json",
      },
      dependencies: ["setup"],
      // Public + authenticated specs. Auth-less flows opt out of storageState per-file.
      testIgnore: /auth\.setup\.ts/,
    },
  ],

  // Reuse an already-running dev server (the usual local case); otherwise start one.
  webServer: {
    command: "npm run dev",
    url: BASE_URL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
