import { test, expect } from "./support/test";
import { SHOP } from "./support/constants";

test.describe("dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/app/dashboard");
  });

  test("renders the summary stat cards", async ({ page }) => {
    // Seeded fixtures: 1 shop, 1 warning environment.
    await expect(page.getByTestId("dashboard-stat-shops")).toHaveText("1");
    await expect(page.getByTestId("dashboard-stat-healthy")).toBeVisible();
    await expect(page.getByTestId("dashboard-stat-warnings")).toBeVisible();
    await expect(page.getByTestId("dashboard-stat-errors")).toBeVisible();
  });

  test("shows the seeded shop and its environment widgets", async ({ page }) => {
    await expect(page.getByRole("heading", { name: "Shopware Versions" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Last Changes" })).toBeVisible();
    // The shop appears in the grid as a link to its environment.
    await expect(page.getByRole("link", { name: new RegExp(SHOP.name) }).first()).toBeVisible();
  });

  test("clicking a shop card opens the environment detail", async ({ page }) => {
    await page
      .getByRole("link", { name: new RegExp(SHOP.name) })
      .first()
      .click();
    await expect(page).toHaveURL(/\/app\/environments\/\d+/);
  });
});
