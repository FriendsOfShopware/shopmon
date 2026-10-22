import { mount } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { describe, expect, it, vi } from "vitest";
import { ref, unref } from "vue";

import App from "./App.vue";
import { routes, setupRouterGuards } from "./router";

const useHead = vi.hoisted(() => vi.fn());
const mockSession = vi.hoisted(() => ({ value: null as unknown }));

vi.mock("@unhead/vue", () => ({
  useHead,
}));

vi.mock("./composables/useDarkMode", () => ({
  useDarkMode: vi.fn(),
}));

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref(mockSession.value),
    loading: ref(false),
  }),
  fetchSession: vi.fn().mockResolvedValue(null),
}));

vi.mock("@/composables/useInstanceConfig", () => ({
  useInstanceConfig: () => ({
    config: ref({ registrationEnabled: true }),
    load: vi.fn().mockResolvedValue(undefined),
  }),
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

  it("allows unauthenticated users to visit advisories list and detail routes", async () => {
    mockSession.value = null;
    const router = createRouter({
      history: createMemoryHistory(),
      routes,
    });
    setupRouterGuards(router);

    await router.push("/app/advisories");
    expect(router.currentRoute.value.name).toBe("account.advisories.list");

    await router.push("/advisories");
    expect(router.currentRoute.value.name).toBe("account.advisories.list");

    await router.push("/app/advisories/CVE-2026-1234");
    expect(router.currentRoute.value.name).toBe("account.advisories.detail");
    expect(router.currentRoute.value.params.id).toBe("CVE-2026-1234");

    await router.push("/advisories/CVE-2026-1234");
    expect(router.currentRoute.value.name).toBe("account.advisories.detail");
    expect(router.currentRoute.value.params.id).toBe("CVE-2026-1234");

    // Restricted pages still redirect to login when unauthenticated.
    await router.push("/app/dashboard");
    expect(router.currentRoute.value.name).toBe("account.login");
  });
});
