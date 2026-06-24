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

  // Charts are drawn on a <canvas> by Chart.js, which only renders correctly once
  // the canvas has been laid out. This regressed when switching to the tab: the
  // canvas stayed blank until a full reload. Assert the charts actually paint
  // when navigating in via the tab bar (the broken path).
  test("Sitespeed tab renders charts when navigated to via the tab bar", async ({ page }) => {
    await page.getByRole("link", { name: /^Sitespeed/ }).click();
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
});
