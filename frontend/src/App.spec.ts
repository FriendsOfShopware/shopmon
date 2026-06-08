import { mount } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { describe, expect, it, vi } from "vitest";
import { unref } from "vue";

import App from "./App.vue";

const useHead = vi.hoisted(() => vi.fn());

vi.mock("@unhead/vue", () => ({
  useHead,
}));

vi.mock("./composables/useDarkMode", () => ({
  useDarkMode: vi.fn(),
}));

describe("App", () => {
  it("sets route titles through useHead", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          path: "/dashboard",
          component: { template: "<div />" },
          meta: { titleKey: "nav.dashboard" },
        },
      ],
    });
    await router.push("/dashboard");
    await router.isReady();

    mount(App, {
      global: {
        plugins: [router],
      },
    });

    const head = useHead.mock.calls[0]?.[0];

    expect(unref(head.title)).toBe("Dashboard");
    expect(head.titleTemplate("Dashboard")).toBe("Dashboard | Shopmon");
    expect(head.titleTemplate()).toBe("Shopmon");
  });
});
