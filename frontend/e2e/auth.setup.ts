import { test as setup, expect } from "@playwright/test";
import { PASSWORD, STORAGE_STATE, USERS } from "./support/constants";

// Logs in once as the org owner and persists the session so authenticated
// specs don't each pay the login cost. Run via the `setup` project.
setup("authenticate", async ({ page }) => {
  // Pre-dismiss the "What's New" modal so it never overlays the app.
  await page.addInitScript(() => {
    try {
      window.localStorage.setItem("shopmon-whats-new", "2026-03-packages-mirror");
    } catch {
      /* storage unavailable — ignore */
    }
  });

  await page.goto("/account/login");

  await page.getByLabel("Email address").fill(USERS.owner.email);
  await page.getByLabel("Password", { exact: true }).fill(PASSWORD);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();

  // Landing on the dashboard confirms the session is established.
  await page.waitForURL("**/app/dashboard");
  await expect(page.getByRole("heading", { name: "Dashboard", level: 1 })).toBeVisible();

  await page.context().storageState({ path: STORAGE_STATE });
});
