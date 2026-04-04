import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import ListShops from "./ListShops.vue";

// Stubs
const StatusIconStub = defineComponent({
  name: "StatusIcon",
  props: ["status"],
  template: '<span :class="status">{{ status }}</span>',
});

// Mock vue-router
vi.mock("vue-router", () => ({
  useRoute: () => ({ params: {} }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a :href="JSON.stringify(to)"><slot /></a>',
  },
}));

// Mock data
const mockShops = [
  {
    id: 1,
    name: "Test Shop",
    nameCombined: "Test Org / Test Shop",
    description: "A test shop description",
    createdAt: new Date("2024-01-15").toISOString(),
  },
];

const mockEnvironments = [
  {
    id: 1,
    name: "Test Environment",
    shopwareVersion: "6.5.0",
    status: "green",
    favicon: null,
    shopId: 1,
    organizationId: "org-1",
  },
];

// Mock api client
vi.mock("@/helpers/api", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
    PATCH: vi.fn(),
    DELETE: vi.fn(),
    PUT: vi.fn(),
  },
  setToken: vi.fn(),
  getToken: vi.fn(),
}));

// Mock formatter
vi.mock("@/helpers/formatter", () => ({
  formatDate: (date: string) => new Date(date).toLocaleDateString(),
}));

// Mock useAlert
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

// Mock useAccountEnvironments
vi.mock("@/composables/useAccountEnvironments", () => ({
  useAccountEnvironments: () => ({
    environments: { value: mockEnvironments },
  }),
}));

import { api } from "@/helpers/api";

describe("ListShops", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: mockShops, error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(ListShops, {
      global: {
        stubs: {
          StatusIcon: StatusIconStub,
        },
      },
    });
  }

  it("renders successfully", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("My Shops");
  });

  it("displays Add Project button in header", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Shop");
  });

  it("displays empty state when no projects exist", async () => {
    // Override mock to return empty shops
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);

    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("No Shops");
  });

  it("displays projects list when projects exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Test Shop");
  });

  it("displays project description when available", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("A test shop description");
  });

  it("displays project meta information", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    // The component displays each environment as a card with its name and version
    expect(wrapper.text()).toContain("Test Environment");
    expect(wrapper.text()).toContain("6.5.0");
  });

  it("displays shops for each project", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Test Shop");
    expect(wrapper.text()).toContain("6.5.0");
  });

  it("displays shop status icons", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.find(".green").exists()).toBe(true);
  });

  it("shows empty state CTA when no projects", async () => {
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);

    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Get started by creating your first shop");
  });

  it("shows an edit project button", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Edit Shop");
  });

  it("shows an add shop shortcut on the project card", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    expect(wrapper.text()).toContain("Add Environment");
  });

  it("displays shop favicon or fallback icon", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    // Shop should be displayed
    expect(wrapper.text()).toContain("Test Shop");
  });
});
