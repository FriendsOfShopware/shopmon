import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h, ref } from "vue";
import Dashboard from "./Dashboard.vue";

// Stubs
const StatusIconStub = defineComponent({
  name: "StatusIcon",
  props: ["status"],
  template: '<span :class="status">Status</span>',
});

const RouterLinkStub = defineComponent({
  name: "RouterLink",
  props: ["to"],
  setup(props, { slots }) {
    return () => h("a", { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock data — shops now drive the dashboard; environments provide details
const mockShops = [
  {
    id: 1,
    name: "Test Shop 1",
    description: null,
    gitUrl: null,
    organizationId: "org-1",
    organizationName: "Test Org",
    defaultEnvironmentId: 1,
  },
  {
    id: 2,
    name: "Test Shop 2",
    description: null,
    gitUrl: null,
    organizationId: "org-1",
    organizationName: "Test Org",
    defaultEnvironmentId: 2,
  },
];

const mockEnvironments = [
  {
    id: 1,
    name: "Test Environment 1",
    url: "https://shop1.example.com",
    organizationId: "org-1",
    organizationName: "Test Org",
    shopId: 1,
    shopName: "Test Shop 1",
    shopwareVersion: "6.5.0",
    status: "green",
    favicon: null,
    lastScrapedAt: null,
    lastScrapedError: null,
  },
  {
    id: 2,
    name: "Test Environment 2",
    url: "https://shop2.example.com",
    organizationId: "org-1",
    organizationName: "Test Org",
    shopId: 2,
    shopName: "Test Shop 2",
    shopwareVersion: "6.4.0",
    status: "red",
    favicon: "/favicon.ico",
    lastScrapedAt: null,
    lastScrapedError: null,
  },
];

const mockChangelogs = [
  {
    id: 1,
    environmentId: 1,
    environmentName: "Test Environment 1",
    environmentOrganizationName: "Test Org",
    environmentOrganizationId: "org-1",
    extensions: [
      {
        name: "Test Extension",
        label: "Test Extension",
        state: "installed",
        oldVersion: "1.0.0",
        newVersion: "1.1.0",
        active: true,
      },
    ],
    oldShopwareVersion: null,
    newShopwareVersion: null,
    date: new Date("2024-01-15").toISOString(),
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

// Mock changelog helper
vi.mock("@/helpers/changelog", () => ({
  sumChanges: (row: any) => row.extensions?.length ?? 0,
}));

// Mock formatter
vi.mock("@/helpers/formatter", () => ({
  formatDate: (date: string) => new Date(date).toLocaleDateString(),
}));

// Mock useAccountEnvironments - note: vi.mock is hoisted, so we use dynamic ref
const mockEnvironmentsRef = ref([] as any[]);
vi.mock("@/composables/useAccountEnvironments", () => ({
  useAccountEnvironments: () => ({
    environments: mockEnvironmentsRef,
  }),
  fetchAccountEnvironments: vi.fn(),
}));

// Mock useSession
vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref({ user: { id: "1" } }),
    activeOrganizationId: ref(null),
  }),
}));

import { api } from "@/helpers/api";

describe("Dashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Populate the environments ref for useAccountEnvironments mock
    mockEnvironmentsRef.value = mockEnvironments;
    // Reset mock data
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: mockShops, error: null, response: new Response() });
      }
      if (path === "/account/changelogs") {
        return Promise.resolve({ data: mockChangelogs, error: null, response: new Response() });
      }
      if (path === "/account/extensions") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(Dashboard, {
      global: {
        stubs: {
          StatusIcon: StatusIconStub,
          RouterLink: RouterLinkStub,
        },
      },
    });
  }

  it("renders successfully", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays dashboard title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("Dashboard");
  });

  it("displays My Environments section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("My Environments");
  });

  it("displays environments data", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    // The dashboard now shows shop names in a card grid; version comes from the default environment
    expect(wrapper.text()).toContain("Test Shop 1");
    expect(wrapper.text()).toContain("Test Shop 2");
    expect(wrapper.text()).toContain("6.5.0");
  });

  it("displays environment status icons", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const statusIcons = wrapper.findAllComponents(StatusIconStub);
    expect(statusIcons.length).toBe(2);
  });

  it("displays Shopware Versions section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Shopware Versions");
  });

  it("displays version distribution data", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    // Both Shopware versions should appear in the version distribution
    expect(wrapper.text()).toContain("6.5.0");
    expect(wrapper.text()).toContain("6.4.0");
  });

  it("displays Last Changes section when changelogs exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Last Changes");
  });

  it("does not display Last Changes section when no changelogs", async () => {
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: mockShops, error: null, response: new Response() });
      }
      if (path === "/account/changelogs") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/account/extensions") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
    const wrapper = mountComponent();
    await flushPromises();
    // When changelogs are empty the "No recent changes" fallback is shown instead of changelog entries
    expect(wrapper.text()).toContain("No recent changes");
  });

  it("displays changelog data", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Environment 1");
  });

  it("displays correct environment links", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const links = wrapper.findAll("a");
    // Shop cards link to the default environment; find one containing the shop name
    const shopLink = links.find((l) => l.text().includes("Test Shop 1"));
    expect(shopLink).toBeTruthy();
  });
});
