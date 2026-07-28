import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ListOrganizations from "./ListOrganizations.vue";

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountOrganizations: vi.fn(),
}));

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: { value: { user: { id: "1", name: "Test" } } },
    loading: { value: false },
    fetchSession: vi.fn(),
  }),
  fetchSession: vi.fn(),
}));

vi.mock("vue-router", () => ({
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a :href="JSON.stringify(to)"><slot /></a>',
  },
}));

import { getAccountOrganizations } from "@/api/generated";

const mockOrganizations = [
  { id: "1", name: "Test Organization", role: "owner", createdAt: "2024-01-01" },
  { id: "2", name: "Another Org", role: "member", createdAt: "2024-02-01" },
];

describe("ListOrganizations", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAccountOrganizations).mockResolvedValue({
      data: mockOrganizations,
      error: undefined,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(ListOrganizations);
  }

  it("renders successfully", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("My Organization");
  });

  it("displays Add Organization button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Organization");
  });

  it("displays organizations when data exists", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Organization");
    expect(wrapper.text()).toContain("Another Org");
  });

  it("displays organization names", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Organization");
    expect(wrapper.text()).toContain("Another Org");
  });

  it("displays empty state when no organizations exist", async () => {
    vi.mocked(getAccountOrganizations).mockResolvedValueOnce({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);

    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("No Organization");
  });

  it("shows add organization button in empty state", async () => {
    vi.mocked(getAccountOrganizations).mockResolvedValueOnce({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);

    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Organization");
  });
});
