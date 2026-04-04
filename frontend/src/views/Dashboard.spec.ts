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

const DataTableStub = defineComponent({
  name: "DataTable",
  props: ["columns", "data"],
  template: '<div class="data-table"><slot v-for="item in data" :row="item" /></div>',
});

const RouterLinkStub = defineComponent({
  name: "RouterLink",
  props: ["to"],
  setup(props, { slots }) {
    return () => h("a", { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock data
const mockEnvironments = [
  {
    id: "1",
    name: "Test Environment 1",
    organizationId: "org-1",
    organizationName: "Test Org",
    shopName: "Test Shop",
    shopwareVersion: "6.5.0",
    status: "green",
    favicon: null,
  },
  {
    id: "2",
    name: "Test Environment 2",
    organizationId: "org-1",
    organizationName: "Test Org",
    shopName: "Another Shop",
    shopwareVersion: "6.4.0",
    status: "red",
    favicon: "/favicon.ico",
  },
];

const mockOrganizations = [
  {
    id: "1",
    name: "Test Organization",
    memberCount: 5,
    environmentCount: 2,
  },
];

const mockChangelogs = [
  {
    id: "1",
    environmentId: "1",
    environmentName: "Test Environment 1",
    environmentOrganizationName: "Test Org",
    environmentOrganizationId: "org-1",
    extensions: [{ name: "Test Extension", oldVersion: "1.0.0", newVersion: "1.1.0" }],
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
  formatDateTime: (date: string) => new Date(date).toLocaleString(),
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
      if (path === "/account/organizations") {
        return Promise.resolve({ data: mockOrganizations, error: null, response: new Response() });
      }
      if (path === "/account/environments") {
        return Promise.resolve({ data: mockEnvironments, error: null, response: new Response() });
      }
      if (path === "/account/changelogs") {
        return Promise.resolve({ data: mockChangelogs, error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(Dashboard, {
      global: {
        stubs: {
          StatusIcon: StatusIconStub,
          DataTable: DataTableStub,
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
    expect(wrapper.text()).toContain("Test Environment 1");
    expect(wrapper.text()).toContain("Test Environment 2");
    expect(wrapper.text()).toContain("6.5.0");
  });

  it("displays environment status icons", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const statusIcons = wrapper.findAll(".green, .red");
    expect(statusIcons.length).toBe(2);
  });

  it("displays My Organizations section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("My Organizations");
  });

  it("displays organizations data", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Organization");
    // The i18n key "{count} Members" renders with the count value
    expect(wrapper.text()).toContain("5");
    expect(wrapper.text()).toContain("2");
  });

  it("displays Last Changes section when changelogs exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Last Changes");
  });

  it("does not display Last Changes section when no changelogs", async () => {
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/organizations") {
        return Promise.resolve({ data: mockOrganizations, error: null, response: new Response() });
      }
      if (path === "/account/environments") {
        return Promise.resolve({ data: mockEnvironments, error: null, response: new Response() });
      }
      if (path === "/account/changelogs") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).not.toContain("Last Changes");
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
    const envLink = links.find((l) => l.text().includes("Test Environment 1"));
    expect(envLink).toBeTruthy();
  });
});
