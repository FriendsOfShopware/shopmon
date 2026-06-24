import { test, expect } from "./support/test";

// Public marketing + entry pages — no authentication.
test.use({ storageState: { cookies: [], origins: [] } });

test.describe("public pages", () => {
  test("home page renders the hero and primary CTAs", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.getByRole("main")).toBeVisible();
  });

  test("home links through to the login page", async ({ page }) => {
    await page.goto("/");
    // Desktop nav exposes a Login entry.
    await page.getByRole("link", { name: "Login", exact: true }).first().click();
    await expect(page).toHaveURL(/\/account\/login/);
  });

  test("theme toggle switches the document class", async ({ page }) => {
    await page.goto("/account/login");

    const html = page.locator("html");
    const wasDark = await html.evaluate((el) => el.classList.contains("dark"));

    await page.getByRole("button", { name: "Toggle theme" }).click();

    await expect.poll(() => html.evaluate((el) => el.classList.contains("dark"))).toBe(!wasDark);
  });

  test("footer legal links resolve", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: "Privacy" }).click();
    await expect(page).toHaveURL(/\/privacy/);

    await page.goto("/");
    await page.getByRole("link", { name: "Legal Notice" }).click();
    await expect(page).toHaveURL(/\/imprint/);
  });
});
