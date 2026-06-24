import { test, expect } from "./support/test";
import { PASSWORD, USERS } from "./support/constants";

// These flows exercise login/logout themselves, so they run without the
// pre-authenticated storage state.
test.use({ storageState: { cookies: [], origins: [] } });

test.describe("authentication", () => {
  test("logs in with valid credentials and lands on the dashboard", async ({ page }) => {
    await page.goto("/account/login");

    await page.getByLabel("Email address").fill(USERS.owner.email);
    await page.getByLabel("Password", { exact: true }).fill(PASSWORD);
    await page.getByRole("button", { name: "Sign in", exact: true }).click();

    await expect(page).toHaveURL(/\/app\/dashboard/);
    await expect(page.getByRole("heading", { name: "Dashboard", level: 1 })).toBeVisible();
  });

  test("logs out and returns to the login page", async ({ page }) => {
    // Fresh login so the logout only tears down this throwaway session,
    // never the shared storage-state token other specs reuse.
    await page.goto("/account/login");
    await page.getByLabel("Email address").fill(USERS.owner.email);
    await page.getByLabel("Password", { exact: true }).fill(PASSWORD);
    await page.getByRole("button", { name: "Sign in", exact: true }).click();
    await expect(page.getByRole("button", { name: "User menu" })).toBeVisible();

    await page.getByRole("button", { name: "User menu" }).click();
    await expect(page.getByRole("menu")).toBeVisible();
    await page.getByRole("menuitem", { name: "Logout" }).click();

    // Logout drops us out of the app (home or login) and the session is gone.
    await expect(page).toHaveURL(/localhost:3000\/(account\/login)?$/);
    await expect(page.getByRole("button", { name: "User menu" })).toBeHidden();

    // The session is really cleared: app routes now redirect to login.
    await page.goto("/app/dashboard");
    await expect(page).toHaveURL(/\/account\/login/);
  });

  test("rejects invalid credentials and stays on login", async ({ page }) => {
    await page.goto("/account/login");

    await page.getByLabel("Email address").fill(USERS.owner.email);
    await page.getByLabel("Password", { exact: true }).fill("wrong-password");
    await page.getByRole("button", { name: "Sign in", exact: true }).click();

    await expect(page).toHaveURL(/\/account\/login/);
    // Heading stays put — we never reached the app.
    await expect(page.getByRole("heading", { name: "Sign in to your account" })).toBeVisible();
  });

  test("password visibility toggle reveals and hides the value", async ({ page }) => {
    await page.goto("/account/login");

    const password = page.getByLabel("Password", { exact: true });
    await password.fill(PASSWORD);
    await expect(password).toHaveAttribute("type", "password");

    await page.getByRole("button", { name: "Show password" }).click();
    await expect(password).toHaveAttribute("type", "text");

    await page.getByRole("button", { name: "Hide password" }).click();
    await expect(password).toHaveAttribute("type", "password");
  });

  test("redirects unauthenticated users away from app routes", async ({ page }) => {
    await page.goto("/app/dashboard");
    await expect(page).toHaveURL(/\/account\/login/);
  });

  test("can navigate to register and back to login", async ({ page }) => {
    await page.goto("/account/login");

    await page.getByRole("link", { name: "Create an account" }).click();
    await expect(page).toHaveURL(/\/account\/register/);
    await expect(page.getByRole("heading", { name: "Create account" })).toBeVisible();
  });
});
