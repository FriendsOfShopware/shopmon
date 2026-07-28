import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import AddShop from "./AddShop.vue";

const mockOrganizations = [
  { id: "org-1", name: "Organization A" },
  { id: "org-2", name: "Organization B" },
];

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ query: {} }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  },
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountOrganizations: vi.fn(),
  createShop: vi.fn(),
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

vi.mock("@/composables/useAccountShops", () => ({
  fetchAccountShops: vi.fn(() => Promise.resolve([])),
}));

vi.mock("@/composables/useAccountEnvironments", () => ({
  fetchAccountEnvironments: vi.fn(() => Promise.resolve([])),
}));

import { createShop, getAccountOrganizations } from "@/api/generated";

describe("AddShop", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAccountOrganizations).mockResolvedValue({
      data: mockOrganizations,
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(createShop).mockResolvedValue({
      data: {},
      error: undefined,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(AddShop);
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("New Shop");
  });

  it("renders form when organizations are loaded", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("displays Shop Information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Shop Information");
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="name"]').exists()).toBe(true);
  });

  it("has description textarea", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('textarea[name="description"]').exists()).toBe(true);
  });

  it("has git URL input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="gitUrl"]').exists()).toBe(true);
  });

  it("has organization select area", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Organization");
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.find('button[type="submit"]');
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toContain("Create Shop");
  });
});
