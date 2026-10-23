import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, ref } from "vue";
import DetailAdvisory from "./DetailAdvisory.vue";

const mockSession = ref<{ user: { id: string; name: string; email: string } } | null>(null);

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: mockSession,
  }),
}));

vi.mock("vue-router", () => ({
  useRoute: () => ({
    params: { id: "CVE-2026-1000" },
  }),
  RouterLink: defineComponent({
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  }),
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAdvisory: vi.fn(),
  listAdvisoryAffectedEnvironments: vi.fn(),
  createAdvisorySuppression: vi.fn(),
}));

import { getAdvisory, listAdvisoryAffectedEnvironments } from "@/api/generated";

describe("DetailAdvisory", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAdvisory).mockResolvedValue({
      data: {
        advisoryId: "CVE-2026-1000",
        title: "Test Security Advisory",
        summary: "Short summary of the advisory",
        descriptionHtml: "<p>Detailed write-up</p>",
        effectiveSeverity: "high",
        cvssScore: 7.5,
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
      error: undefined,
      response: new Response(),
    } as never);
    vi.mocked(listAdvisoryAffectedEnvironments).mockResolvedValue({
      data: {
        environments: [],
        total: 0,
      },
      error: undefined,
      response: new Response(),
    } as never);
  });

  it("renders advisory details without fetching affected environments when logged out", async () => {
    mockSession.value = null;

    const wrapper = mount(DetailAdvisory);
    await flushPromises();

    expect(wrapper.find("h1").text()).toBe("Test Security Advisory");
    expect(wrapper.text()).toContain("Detailed write-up");
    expect(wrapper.text()).not.toContain("Affected shops");
    expect(listAdvisoryAffectedEnvironments).not.toHaveBeenCalled();
  });

  it("fetches and renders affected shops section when logged in", async () => {
    mockSession.value = {
      user: { id: "user-1", name: "User", email: "user@example.com" },
    };

    const wrapper = mount(DetailAdvisory);
    await flushPromises();

    expect(wrapper.find("h1").text()).toBe("Test Security Advisory");
    expect(wrapper.text()).toContain("Affected shops");
    expect(listAdvisoryAffectedEnvironments).toHaveBeenCalledWith({
      path: { advisoryId: "CVE-2026-1000" },
    });
  });
});
