import { describe, expect, it, vi } from "vitest";
import type { RouteLocationNormalized, Router } from "vue-router";

vi.mock("@/i18n", async () => {
  const { createI18n } = await import("vue-i18n");
  const { default: en } = await import("@/locales/en.json");
  return {
    i18n: createI18n({
      legacy: false,
      locale: "en",
      fallbackLocale: "en",
      messages: { en },
    }),
  };
});

import { setupRouterGuards } from "./index";

type AfterEachHook = Parameters<Router["afterEach"]>[0];

function route(meta: RouteLocationNormalized["meta"] = {}) {
  return { meta } as RouteLocationNormalized;
}

describe("setupRouterGuards", () => {
  it("formats translated document titles with a spaced separator", async () => {
    let afterEachHook: AfterEachHook | undefined;
    const router = {
      beforeEach: vi.fn(),
      afterEach: vi.fn((callback: AfterEachHook) => {
        afterEachHook = callback;
        return () => {};
      }),
    } as unknown as Router;

    setupRouterGuards(router);

    if (!afterEachHook) {
      throw new Error("router.afterEach was not registered");
    }

    await afterEachHook(route({ titleKey: "nav.dashboard" }), route(), undefined);

    expect(document.title).toBe("Dashboard | Shopmon");
  });
});
