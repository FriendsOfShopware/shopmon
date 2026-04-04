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
    environments: [
      {
        environmentId: 1,
        environmentName: "Environment A",
        environmentOrganizationId: "org-a",
        version: "1.0.0",
      },
    ],
  },
];

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

import { api } from "@/helpers/api";

describe("ListExtensions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/extensions") {
        return Promise.resolve({ data: mockExtensions, error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
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
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/extensions") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
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
    expect(wrapper.text()).toContain("Extensions");
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
