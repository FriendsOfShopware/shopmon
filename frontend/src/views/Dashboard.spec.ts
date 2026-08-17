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
    environmentShopName: "Test Shop 1",
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

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountChangelogs: vi.fn(),
  getAccountExtensions: vi.fn(),
  getAccountShops: vi.fn(),
}));

// Mock changelog helper
vi.mock("@/helpers/changelog", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/helpers/changelog")>()),
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

// Mock useSession — keep activeOrganizationId as a shared ref so tests can
// simulate org switches (which should clear the dashboard status filter).
const mockActiveOrganizationId = ref<string | null>(null);
vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref({ user: { id: "1" } }),
    activeOrganizationId: mockActiveOrganizationId,
  }),
}));

import { getAccountChangelogs, getAccountExtensions, getAccountShops } from "@/api/generated";

describe("Dashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockActiveOrganizationId.value = null;
    // Populate the environments ref for useAccountEnvironments mock
    mockEnvironmentsRef.value = mockEnvironments;
    // Reset mock data
    vi.mocked(getAccountShops).mockResolvedValue({
      data: mockShops,
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getAccountChangelogs).mockResolvedValue({
      data: mockChangelogs,
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getAccountExtensions).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
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
    expect(wrapper.text()).toContain("6.5.0");
    expect(wrapper.text()).toContain("6.4.0");
  });

  it("displays Last Changes section when changelogs exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Last Changes");
  });

  it("does not display Last Changes section when no changelogs", async () => {
    vi.mocked(getAccountChangelogs).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("No recent changes");
  });

  it("displays changelog data", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Shop 1 · Test Environment 1");
  });

  it("displays correct environment links", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const links = wrapper.findAll("a");
    const shopLink = links.find((l) => l.text().includes("Test Shop 1"));
    expect(shopLink).toBeTruthy();
  });

  it("clears the status filter when the active organization changes", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    const errorFilter = wrapper
      .findAll("button")
      .find((b) => b.text().includes("Errors") && b.attributes("aria-pressed") !== undefined);
    expect(errorFilter).toBeTruthy();
    await errorFilter!.trigger("click");
    await flushPromises();
    expect(errorFilter!.attributes("aria-pressed")).toBe("true");
    const shopGrid = wrapper.find("section");
    expect(shopGrid.text()).not.toContain("Test Shop 1");
    expect(shopGrid.text()).toContain("Test Shop 2");

    mockActiveOrganizationId.value = "org-2";
    await flushPromises();
    expect(errorFilter!.attributes("aria-pressed")).toBe("false");
    expect(shopGrid.text()).toContain("Test Shop 1");
    expect(shopGrid.text()).toContain("Test Shop 2");
  });

  it("ignores stale dashboard responses after an organization switch", async () => {
    type Deferred<T> = {
      promise: Promise<T>;
      resolve: (value: T) => void;
    };
    function deferred<T>(): Deferred<T> {
      let resolve!: (value: T) => void;
      const promise = new Promise<T>((r) => {
        resolve = r;
      });
      return { promise, resolve };
    }

    const firstShops = deferred<any>();
    const secondShops = deferred<any>();
    let callCount = 0;
    vi.mocked(getAccountShops).mockImplementation(() => {
      callCount++;
      if (callCount === 1) return firstShops.promise;
      return secondShops.promise;
    });

    const wrapper = mountComponent();
    mockActiveOrganizationId.value = "org-2";
    await flushPromises();

    const laterOrgShops = [
      {
        id: 99,
        name: "Later Org Shop",
        description: null,
        gitUrl: null,
        organizationId: "org-2",
        organizationName: "Org 2",
        defaultEnvironmentId: null,
      },
    ];
    const earlierOrgShops = [
      {
        id: 1,
        name: "Stale Org Shop",
        description: null,
        gitUrl: null,
        organizationId: "org-1",
        organizationName: "Org 1",
        defaultEnvironmentId: null,
      },
    ];

    secondShops.resolve({ data: laterOrgShops, error: undefined, response: new Response() });
    await flushPromises();
    expect(wrapper.text()).toContain("Later Org Shop");

    firstShops.resolve({ data: earlierOrgShops, error: undefined, response: new Response() });
    await flushPromises();
    expect(wrapper.text()).toContain("Later Org Shop");
    expect(wrapper.text()).not.toContain("Stale Org Shop");
  });
});
