import type { Page } from "@playwright/test";
import { test, expect } from "./support/test";
import { ENVIRONMENTS } from "./support/constants";

const BASE = `/app/environments/${ENVIRONMENTS.production.id}`;

/** Tab bar is a named landmark so queries don't collide with other links. */
const tabNav = (page: Page) => page.getByRole("navigation", { name: "Environment sections" });
const tab = (page: Page, name: string | RegExp) => tabNav(page).getByRole("link", { name });

test.describe("environment detail tabs", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(BASE);
    // Wait for the tab bar so we don't race the session guard / initial load.
    await expect(tab(page, /Environment information/)).toBeVisible();
  });

  test("loads the environment information tab", async ({ page }) => {
    await expect(page).toHaveURL(new RegExp(`${BASE}$`));
    await expect(tab(page, /Environment information/)).toBeVisible();
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

  for (const tabDef of tabs) {
    test(`"${tabDef.name}" tab navigates to its route`, async ({ page }) => {
      await tab(page, new RegExp(`^${tabDef.name}`)).click();
      await expect(page).toHaveURL(new RegExp(`${BASE}${tabDef.segment}`));
    });
  }

  test("header exposes storefront and admin links", async ({ page }) => {
    await expect(page.getByRole("link", { name: "Storefront" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Admin" })).toBeVisible();
  });

  // Charts are drawn on a <canvas> by Chart.js, which only renders correctly once
  // the canvas has been laid out. This regressed when switching to the tab: the
  // canvas stayed blank until a full reload. Assert the charts actually paint
  // when navigating in via the tab bar (the broken path).
  test("Sitespeed tab renders charts when navigated to via the tab bar", async ({ page }) => {
    await tab(page, /^Sitespeed/).click();
    await expect(page).toHaveURL(new RegExp(`${BASE}/sitespeed`));

    // The "not enabled" alert must not be shown — the fixture enables sitespeed.
    const canvases = page.locator("canvas");
    await expect(canvases.first()).toBeVisible();

    // Three charts: performance over time, transfer size, CLS.
    await expect(canvases).toHaveCount(3);

    // The bug renders the chart chrome (axes, gridlines, legend) but never draws
    // the data series, so the canvas isn't blank — checking for "any non-blank
    // pixel" gives a false pass. The data series are saturated colours (blue,
    // pink, orange, yellow) while the axes/gridlines are near-grey, so assert
    // there are coloured pixels inside the plot area. Poll because Chart.js
    // draws on the animation frame after layout.
    for (let i = 0; i < 3; i++) {
      // Note: don't scrollIntoView here — that triggers a Chart.js resize which
      // restarts the entry animation, so a read right after catches a half-drawn
      // frame. The poll below tolerates animation; getImageData works off-screen.
      await expect
        .poll(
          () =>
            canvases.nth(i).evaluate((el) => {
              const canvas = el as HTMLCanvasElement;
              if (!canvas.width || !canvas.height) return 0;
              const ctx = canvas.getContext("2d", { willReadFrequently: true });
              if (!ctx) return 0;
              const { data, width, height } = ctx.getImageData(0, 0, canvas.width, canvas.height);
              // Restrict to the plot area so the coloured legend swatches at the
              // top don't count — they're present even when the data is missing.
              const top = Math.floor(height * 0.25);
              let coloured = 0;
              for (let y = top; y < height; y++) {
                for (let x = 0; x < width; x++) {
                  const o = (y * width + x) * 4;
                  const [r, g, b, a] = [data[o], data[o + 1], data[o + 2], data[o + 3]];
                  if (a === 0) continue;
                  // Saturated == max channel clearly exceeds min channel. Grey
                  // axes/gridlines/text have r≈g≈b and are filtered out.
                  if (Math.max(r, g, b) - Math.min(r, g, b) > 40) coloured++;
                }
              }
              return coloured;
            }),
          { message: `canvas ${i} should have its data series drawn`, timeout: 5000 },
        )
        .toBeGreaterThan(20);
    }
  });

  // The fixtures seed more changelog entries than fit on one page, so the tab
  // must page through them rather than showing the whole history at once.
  test("Changelog tab pages through the history", async ({ page }) => {
    await tab(page, /^Changelog/).click();
    await expect(page).toHaveURL(new RegExp(`${BASE}/changelog`));

    const previous = page.getByRole("button", { name: "Previous" });
    const next = page.getByRole("button", { name: "Next" });

    await expect(page.getByText(/^Page 1 of \d+$/)).toBeVisible();
    await expect(previous).toBeDisabled();
    await expect(next).toBeEnabled();

    // Each entry's header is an expand button labelled with the entry's date, so
    // match on that prefix rather than any button on the page. Capturing the
    // first row proves the list actually changes instead of re-rendering.
    const entries = page.getByRole("button").filter({ hasText: /^\d{2}\.\d{2}\.\d{4}/ });
    const firstPageText = await entries.first().innerText();
    const firstPageCount = await entries.count();
    expect(firstPageCount).toBe(10);

    await next.click();
    await expect(page.getByText("Page 2 of", { exact: false })).toBeVisible();
    await expect(previous).toBeEnabled();
    expect(await entries.first().innerText()).not.toBe(firstPageText);
    // The last page holds the remainder, so it must not be larger than a full page.
    expect(await entries.count()).toBeLessThanOrEqual(firstPageCount);

    await previous.click();
    await expect(page.getByText(/^Page 1 of \d+$/)).toBeVisible();
    await expect(previous).toBeDisabled();
    expect(await entries.first().innerText()).toBe(firstPageText);
  });
});
