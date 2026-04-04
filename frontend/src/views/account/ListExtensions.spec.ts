import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import ListExtensions from "./ListExtensions.vue";

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
          RouterLink: defineComponent({
            template: "<a><slot /></a>",
          }),
          ExtensionChangelog: ExtensionChangelogStub,
          StatusIcon: defineComponent({ template: "<span />" }),
          RatingStars: defineComponent({ template: "<span />" }),
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("My Extensions");
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
          RouterLink: defineComponent({
            template: "<a><slot /></a>",
          }),
          ExtensionChangelog: ExtensionChangelogStub,
          StatusIcon: defineComponent({ template: "<span />" }),
          RatingStars: defineComponent({ template: "<span />" }),
        },
      },
    });
    await flushPromises();
    expect(wrapper.find(".border-dashed").exists()).toBe(true);
    expect(wrapper.text()).toContain("Extensions");
  });

  it("shows search input and extension list when extensions exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("input").exists()).toBe(true);
    // Component renders extensions as bordered div cards, not a table
    const extensionCards = wrapper.findAll(".rounded-xl.border");
    expect(extensionCards.length).toBeGreaterThan(0);
  });

  it("renders extension items with name and label", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Frosh Tools");
    expect(wrapper.text()).toContain("FroshTools");
  });

  it("renders extension changelog modal", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find(".extension-changelog").exists()).toBe(true);
  });
});
