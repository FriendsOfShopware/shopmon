import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import EditEnvironment from "./EditEnvironment.vue";


const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => {
      const { h } = require("vue");
      return h("main", {}, slots.default?.());
    };
  },
});

const DeleteConfirmationModalStub = defineComponent({
  name: "DeleteConfirmationModal",
  props: ["show", "title", "entityName"],
  template: '<div v-if="show" class="delete-modal" />',
});

const PluginConnectionModalStub = defineComponent({
  name: "PluginConnectionModal",
  props: ["show", "base64", "error"],
  emits: ["close", "import", "update:base64"],
  template: '<div v-if="show" class="plugin-modal" />',
});

const mockEnvironment = {
  id: 1,
  name: "Test Environment",
  url: "https://test.shop",
  shopUrl: "https://test.shop",
  organizationId: "org-1",
  shopId: 1,
  clientId: "client-1",
  clientSecret: "",
  sitespeedEnabled: false,
  sitespeedUrls: [],
};

const mockShops = [
  { id: 1, name: "Shop A", organizationName: "Org" },
  { id: 2, name: "Shop B", organizationName: "Org" },
];

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { organizationId: "org-1", environmentId: "1" } }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a><slot /></a>',
  },
}));

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

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

import { api } from "@/helpers/api";

describe("EditEnvironment", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: mockShops, error: null, response: new Response() });
      }
      if (path.startsWith("/environments/")) {
        return Promise.resolve({ data: mockEnvironment, error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(EditEnvironment, {
      global: {
        stubs: {
          MainContainer: MainContainerStub,
          DeleteConfirmationModal: DeleteConfirmationModalStub,
          PluginConnectionModal: PluginConnectionModalStub,
        },
      },
    });
  }

  it("renders page title with environment name", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toContain("Edit Test Environment");
  });

  it("has cancel link", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const cancelLink = wrapper.findAll("a").find((a) => a.text().includes("Cancel"));
    expect(cancelLink).toBeTruthy();
  });

  it("displays environment information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Environment information");
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="name"]').exists()).toBe(true);
  });

  it("has shop select area", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    // shadcn Select renders a trigger button, not a native <select>
    expect(wrapper.text()).toContain("Shop");
  });

  it("has URL input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="shopUrl"]').exists()).toBe(true);
  });

  it("displays integration section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Integration");
  });

  it("has client ID field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="clientId"]').exists()).toBe(true);
  });

  it("has client secret field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="clientSecret"]').exists()).toBe(true);
  });

  it("has connect using plugin button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper
      .findAll("button")
      .find((b) => b.text().includes("Connect using Shopmon Plugin"));
    expect(btn).toBeTruthy();
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.find('button[type="submit"]');
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toContain("Save");
  });

  it("displays sitespeed section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Sitespeed");
  });

  it("displays delete environment section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Deleting environment");
  });

  it("has delete environment button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete environment"));
    expect(btn).toBeTruthy();
  });
});
