import { test, expect } from "./support/test";
import { USERS } from "./support/constants";

// Runs authenticated via the shared storage state (see playwright.config.ts).
test.describe("primary navigation", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/app/dashboard");
    // Wait for the authenticated shell so we don't race the session guard.
    await expect(page.getByRole("button", { name: "User menu" })).toBeVisible();
  });

  const links = [
    { name: "Dashboard", url: /\/app\/dashboard/, heading: "Dashboard" },
    { name: "My Shops", url: /\/app\/shops/, heading: "My Shops" },
    { name: "My Extensions", url: /\/app\/extensions/, heading: "My Extensions" },
    { name: "Ecosystem", url: /\/app\/ecosystem/ },
    { name: "My Organisation", url: /\/app\/organization/ },
  ];

  for (const link of links) {
    test(`sidebar "${link.name}" navigates to its page`, async ({ page }) => {
      await page.getByRole("link", { name: link.name, exact: true }).click();
      await expect(page).toHaveURL(link.url);
      if (link.heading) {
        await expect(page.getByRole("heading", { name: link.heading, level: 1 })).toBeVisible();
      }
    });
  }

  test("user menu opens and shows the signed-in account", async ({ page }) => {
    await page.getByRole("button", { name: "User menu" }).click();

    const menu = page.getByRole("menu");
    await expect(menu).toBeVisible();
    await expect(menu.getByText(USERS.owner.email)).toBeVisible();
    await expect(page.getByRole("menuitem", { name: "Settings" })).toBeVisible();
    await expect(page.getByRole("menuitem", { name: "Logout" })).toBeVisible();
  });

  // Note: logout lives in auth.spec.ts so it runs against a throwaway session
  // and doesn't invalidate the shared storage-state token used by other specs.
});
