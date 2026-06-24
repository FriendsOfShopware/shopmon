import { test as base, expect } from "@playwright/test";

// Keep this in sync with `WHATS_NEW_VERSION` in src/composables/useWhatsNew.ts.
const WHATS_NEW_KEY = "shopmon-whats-new";
const WHATS_NEW_VERSION = "2026-03-packages-mirror";

// Extended test that suppresses the "What's New" modal before every navigation,
// so it never overlays the page under test. The modal is gated on a localStorage
// flag; we set it via an init script that runs before app code on each document.
export const test = base.extend({
  page: async ({ page }, use) => {
    await page.addInitScript(
      ([key, version]) => {
        try {
          window.localStorage.setItem(key, version);
        } catch {
          // Ignore — storage may be unavailable on some origins.
        }
      },
      [WHATS_NEW_KEY, WHATS_NEW_VERSION] as const,
    );
    await use(page);
  },
});

export { expect };
