import { test, expect } from "./support/test";
import { SHOP } from "./support/constants";

test.describe("dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/app/dashboard");
  });

  test("renders the summary stat cards", async ({ page }) => {
    await expect(page.getByRole("button", { name: /Healthy/ })).toBeVisible();
    await expect(page.getByRole("button", { name: /Warning/ })).toBeVisible();
    await expect(page.getByRole("button", { name: /Critical/ })).toBeVisible();
  });

  test("shows the seeded shop and its environment widgets", async ({ page }) => {
    await expect(page.getByRole("heading", { name: /Extension Updates/ })).toBeVisible();
    await expect(page.getByRole("heading", { name: /Recent Changes/ })).toBeVisible();
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
