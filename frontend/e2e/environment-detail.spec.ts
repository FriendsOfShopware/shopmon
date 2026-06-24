import { test, expect } from "./support/test";
import { ENVIRONMENTS } from "./support/constants";

const BASE = `/app/environments/${ENVIRONMENTS.production.id}`;

test.describe("environment detail tabs", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(BASE);
    // Wait for the tab bar so we don't race the session guard / initial load.
    await expect(page.getByRole("link", { name: /Environment informations/ })).toBeVisible();
  });

  test("loads the environment information tab", async ({ page }) => {
    await expect(page).toHaveURL(new RegExp(`${BASE}$`));
    await expect(page.getByRole("link", { name: /Environment informations/ })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Security & Health Checks" })).toBeVisible();
  });

  // Each tab is a link whose href is the tab's route segment.
  const tabs = [
    { name: "Checks", segment: "/checks" },
    { name: "Extensions", segment: "/extensions" },
    { name: "Scheduled tasks", segment: "/tasks" },
    { name: "Queue", segment: "/queue" },
    { name: "Sitespeed", segment: "/sitespeed" },
    { name: "Changelog", segment: "/changelog" },
    { name: "Deployments", segment: "/deployments" },
  ];

  for (const tab of tabs) {
    test(`"${tab.name}" tab navigates to its route`, async ({ page }) => {
      await page.getByRole("link", { name: new RegExp(`^${tab.name}`) }).click();
      await expect(page).toHaveURL(new RegExp(`${BASE}${tab.segment}`));
    });
  }

  test("header exposes storefront and admin links", async ({ page }) => {
    await expect(page.getByRole("link", { name: "Storefront" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Admin" })).toBeVisible();
  });
});
