import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import ListExtensions from "./ListExtensions.vue";

const ExtensionChangelogStub = defineComponent({
  name: "ExtensionChangelog",
  props: ["show", "extension"],
  template: '<div class="extension-changelog" />',
});

const RouterLinkStub = defineComponent({
  name: "RouterLink",
  props: ["to"],
  template:
    '<a class="router-link-stub" :data-route-name="to.name" :data-environment-id="to.params?.environmentId" :data-extension="to.query?.extension"><slot /></a>',
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
        latestVersion: "1.1.0",
        active: true,
        installed: true,
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
          RouterLink: RouterLinkStub,
          "router-link": RouterLinkStub,
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
          RouterLink: RouterLinkStub,
          "router-link": RouterLinkStub,
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

  it("links installed environments to the environment extensions tab", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    const expandButton = wrapper.find("button.size-7");
    expect(expandButton.exists()).toBe(true);
    await expandButton.trigger("click");

    const environmentLink = wrapper.find(".router-link-stub");
    expect(environmentLink.attributes("data-route-name")).toBe(
      "account.environments.detail.extensions",
    );
    expect(environmentLink.attributes("data-environment-id")).toBe("1");
    expect(environmentLink.attributes("data-extension")).toBe("FroshTools");
  });
});
