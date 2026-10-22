import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, ref } from "vue";
import ListAdvisories from "./ListAdvisories.vue";

const mockSession = ref<{ user: { id: string; name: string; email: string } } | null>(null);
const mockRouteQuery = ref<Record<string, string>>({});
const mockRouterReplace = vi.fn();

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: mockSession,
  }),
}));

vi.mock("vue-router", () => ({
  useRoute: () => ({
    get query() {
      return mockRouteQuery.value;
    },
  }),
  useRouter: () => ({
    replace: mockRouterReplace,
  }),
  RouterLink: defineComponent({
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  }),
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  listAdvisories: vi.fn(),
  listAdvisoryPackages: vi.fn(),
}));

import { listAdvisories, listAdvisoryPackages } from "@/api/generated";

describe("ListAdvisories", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRouteQuery.value = {};
    vi.mocked(listAdvisoryPackages).mockResolvedValue({
      data: ["shopware/core"],
      error: undefined,
      response: new Response(),
    } as never);
    vi.mocked(listAdvisories).mockResolvedValue({
      data: {
        advisories: [
          {
            advisoryId: "CVE-2026-1000",
            title: "Test Security Advisory",
            effectiveSeverity: "high",
            packages: [
              {
                packageName: "shopware/core",
                affectedVersions: ">=6.7.0.0,<6.7.10.1",
                packagistAdvisoryId: "PKSA-1",
              },
            ],
            sources: [{ name: "GitHub", remoteId: "GHSA-1000" }],
            affectedComponents: [],
            tags: [],
            isVisible: true,
            syncedAt: "2026-06-01T00:00:00Z",
            createdAt: "2026-06-01T00:00:00Z",
            updatedAt: "2026-06-01T00:00:00Z",
          },
        ],
        total: 1,
        scopeCounts: { all: 1, affected: 0, suppressed: 0 },
      },
      error: undefined,
      response: new Response(),
    } as never);
  });

  it("hides scope tabs and queries 'all' scope when not logged in", async () => {
    mockSession.value = null;
    mockRouteQuery.value = { scope: "affected" };

    const wrapper = mount(ListAdvisories);
    await flushPromises();

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false);
    expect(listAdvisories).toHaveBeenCalledWith({
      query: expect.objectContaining({
        scope: "all",
      }),
    });
    expect(wrapper.text()).toContain("Test Security Advisory");
  });

  it("shows scope tabs and defaults to 'affected' scope when logged in", async () => {
    mockSession.value = {
      user: { id: "user-1", name: "User", email: "user@example.com" },
    };

    const wrapper = mount(ListAdvisories);
    await flushPromises();

    expect(wrapper.find('[role="tablist"]').exists()).toBe(true);
    expect(listAdvisories).toHaveBeenCalledWith({
      query: expect.objectContaining({
        scope: "affected",
      }),
    });
  });
});
