import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import ListExtensions from "./ListExtensions.vue";

const HeaderContainerStub = defineComponent({
  name: "HeaderContainer",
  props: ["title"],
  template: "<header>{{ title }}</header>",
});

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

const PanelStub = defineComponent({
  name: "Panel",
  props: ["variant", "title"],
  setup(props, { slots }) {
    return () =>
      h(
        "div",
        { class: `panel ${props.variant === "table" ? "panel-table" : ""}` },
        slots.default?.(),
      );
  },
});

const ElementEmptyStub = defineComponent({
  name: "ElementEmpty",
  props: ["title", "button", "route"],
  setup(props, { slots }) {
    return () => h("div", { class: "empty" }, [h("h3", {}, props.title), slots.default?.()]);
  },
});

const DataTableStub = defineComponent({
  name: "DataTable",
  props: ["columns", "data", "defaultSort", "searchTerm"],
  template: "<table><slot /></table>",
});

const ExtensionChangelogStub = defineComponent({
  name: "ExtensionChangelog",
  props: ["show", "extension"],
  template: '<div class="extension-changelog" />',
});

const mockExtensions = [
  {
    name: "FroshTools",
    label: "Frosh Tools",
    installed: true,
    active: true,
    version: "1.0.0",
    latestVersion: "1.1.0",
    ratingAverage: 5,
    installedAt: "2024-01-01",
    storeLink: "https://store.shopware.com",
    changelog: [],
    shops: [{ id: 1, name: "Shop A", organizationSlug: "org-a", version: "1.0.0" }],
  },
];

vi.mock("@/helpers/trpc", () => ({
  trpcClient: {
    account: {
      currentUserExtensions: {
        query: vi.fn(() => Promise.resolve(mockExtensions)),
      },
    },
  },
}));

vi.mock("@/helpers/formatter", () => ({
  formatDateTime: (d: string) => d,
}));

vi.mock("@/composables/useExtensionChangelogModal", () => ({
  useExtensionChangelogModal: () => ({
    viewExtensionChangelogDialog: { value: false },
    dialogExtension: { value: null },
    openExtensionChangelog: vi.fn(),
    closeExtensionChangelog: vi.fn(),
  }),
}));

describe("ListExtensions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(ListExtensions, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          ElementEmpty: ElementEmptyStub,
          DataTable: DataTableStub,
          ExtensionChangelog: ExtensionChangelogStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toBe("My Extensions");
  });

  it("shows empty state when no extensions", async () => {
    const trpc = await import("@/helpers/trpc");
    vi.mocked(trpc.trpcClient.account.currentUserExtensions.query).mockResolvedValueOnce([]);
    const wrapper = mount(ListExtensions, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          ElementEmpty: ElementEmptyStub,
          DataTable: DataTableStub,
          ExtensionChangelog: ExtensionChangelogStub,
        },
      },
    });
    await flushPromises();
    expect(wrapper.find(".empty").exists()).toBe(true);
    expect(wrapper.text()).toContain("No Extensions");
  });

  it("shows search input and data table when extensions exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("input").exists()).toBe(true);
    expect(wrapper.find("table").exists()).toBe(true);
  });

  it("shows table in panel-table variant", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find(".panel-table").exists()).toBe(true);
  });

  it("renders extension changelog modal", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find(".extension-changelog").exists()).toBe(true);
  });
});
