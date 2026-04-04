import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import ListOrganizations from "./ListOrganizations.vue";

vi.mock("@/helpers/api", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
  },
  getToken: vi.fn(() => "test-token"),
  setToken: vi.fn(),
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

import { api } from "@/helpers/api";

const mockOrganizations = [
  { id: "1", name: "Test Organization", role: "owner", createdAt: "2024-01-01" },
  { id: "2", name: "Another Org", role: "member", createdAt: "2024-02-01" },
];

describe("ListOrganizations", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/auth/list-organizations") {
        return Promise.resolve({ data: mockOrganizations, error: null, response: new Response() });
      }
      return Promise.resolve({
        data: null,
        error: { message: "not found" },
        response: new Response(),
      });
    }) as any);
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
    vi.mocked(api.GET).mockImplementation((() =>
      Promise.resolve({ data: [], error: null, response: new Response() })) as any);

    const wrapper = mountComponent();
    await flushPromises();
    // The empty state now uses inline Card content with a "No Organizations" message
    expect(wrapper.text()).toContain("No Organization");
  });

  it("shows add organization button in empty state", async () => {
    vi.mocked(api.GET).mockImplementation((() =>
      Promise.resolve({ data: [], error: null, response: new Response() })) as any);

    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Organization");
  });
});
