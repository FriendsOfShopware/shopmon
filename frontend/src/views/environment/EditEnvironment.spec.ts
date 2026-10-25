import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import EditEnvironment from "./EditEnvironment.vue";

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

const DeleteConfirmationModalStub = defineComponent({
  name: "DeleteConfirmationModal",
  props: ["show", "title", "entityName"],
  emits: ["confirm"],
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
  useRoute: () => ({ params: { environmentId: "1" } }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  },
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountShops: vi.fn(),
  getEnvironment: vi.fn(),
  updateEnvironment: vi.fn(),
  deleteEnvironment: vi.fn(),
  updateSitespeedSettings: vi.fn(),
}));

const { mockError, mockSuccess, mockFetchShops, mockFetchEnvironments } = vi.hoisted(() => ({
  mockError: vi.fn(),
  mockSuccess: vi.fn(),
  mockFetchShops: vi.fn(),
  mockFetchEnvironments: vi.fn(),
}));

vi.mock("@/composables/useAccountShops", () => ({
  fetchAccountShops: mockFetchShops,
}));

vi.mock("@/composables/useAccountEnvironments", () => ({
  fetchAccountEnvironments: mockFetchEnvironments,
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: mockError,
    success: mockSuccess,
  }),
}));

import {
  getAccountShops,
  getEnvironment,
  updateEnvironment,
  deleteEnvironment,
  updateSitespeedSettings,
} from "@/api/generated";

describe("EditEnvironment", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAccountShops).mockResolvedValue({
      data: mockShops,
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getEnvironment).mockResolvedValue({
      data: mockEnvironment,
      error: undefined,
      response: new Response(),
    } as any);
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

  it("shows failed saves without navigating or refreshing the sidebar", async () => {
    vi.mocked(updateEnvironment).mockResolvedValue({
      data: undefined,
      error: { message: "Cannot reach shop with new credentials." },
      response: new Response(null, { status: 400 }),
    });
    const wrapper = mountComponent();
    await flushPromises();
    await wrapper.find('input[name="clientId"]').setValue("new-client");
    await wrapper.find('input[name="clientSecret"]').setValue("new-secret");
    await wrapper.find("form").trigger("submit");
    await vi.waitFor(() =>
      expect(mockError).toHaveBeenCalledWith("Cannot reach shop with new credentials."),
    );
    expect(mockPush).not.toHaveBeenCalled();
    expect(mockFetchEnvironments).not.toHaveBeenCalled();
    expect(wrapper.find('input[name="clientSecret"]').element).toHaveProperty(
      "value",
      "new-secret",
    );
  });

  it("preserves stored credentials when inputs are blank and refreshes after saving", async () => {
    vi.mocked(updateEnvironment).mockResolvedValue({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    });
    const wrapper = mountComponent();
    await flushPromises();
    await wrapper.find('input[name="clientId"]').setValue("");
    await wrapper.find('input[name="clientSecret"]').setValue("");
    await wrapper.find('input[name="name"]').setValue("Renamed Environment");
    await wrapper.find("form").trigger("submit");
    await vi.waitFor(() => expect(mockPush).toHaveBeenCalled());
    expect(updateEnvironment).toHaveBeenCalledWith(
      expect.objectContaining({
        body: expect.objectContaining({
          name: "Renamed Environment",
          clientId: undefined,
          clientSecret: undefined,
        }),
      }),
    );
    expect(mockFetchShops).toHaveBeenCalledOnce();
    expect(mockFetchEnvironments).toHaveBeenCalledOnce();
  });

  it("shows deletion errors and stays on the edit screen", async () => {
    vi.mocked(deleteEnvironment).mockResolvedValue({
      data: undefined,
      error: { message: "failed to delete environment" },
      response: new Response(null, { status: 500 }),
    });
    const wrapper = mountComponent();
    await flushPromises();
    wrapper.findComponent(DeleteConfirmationModalStub).vm.$emit("confirm");
    await flushPromises();
    expect(mockError).toHaveBeenCalledWith("failed to delete environment");
    expect(mockPush).not.toHaveBeenCalled();
    expect(mockFetchEnvironments).not.toHaveBeenCalled();
  });

  it("refreshes sidebar environments after deletion", async () => {
    vi.mocked(deleteEnvironment).mockResolvedValue({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    });
    const wrapper = mountComponent();
    await flushPromises();
    wrapper.findComponent(DeleteConfirmationModalStub).vm.$emit("confirm");
    await flushPromises();
    expect(mockFetchShops).toHaveBeenCalledOnce();
    expect(mockFetchEnvironments).toHaveBeenCalledOnce();
    expect(mockPush).toHaveBeenCalledWith({ name: "account.shop.list" });
  });

  it("does not show success when sitespeed settings fail to save", async () => {
    vi.mocked(updateSitespeedSettings).mockResolvedValue({
      data: undefined,
      error: { message: "failed to save sitespeed settings" },
      response: new Response(null, { status: 500 }),
    });
    const wrapper = mountComponent();
    await flushPromises();
    await wrapper.findAll("form")[1]!.trigger("submit");
    await flushPromises();
    expect(mockError).toHaveBeenCalledWith("failed to save sitespeed settings");
    expect(mockSuccess).not.toHaveBeenCalled();
  });
  it("removes blank sitespeed rows before saving", async () => {
    vi.mocked(updateSitespeedSettings).mockResolvedValue({ error: undefined } as any);
    const wrapper = mountComponent();
    await flushPromises();
    const vm = wrapper.vm as any;
    vm.sitespeedEnabled = true;
    vm.sitespeedUrls = ["  ", " https://example.com/path ", ""];
    await vm.onSitespeedSubmit();
    expect(updateSitespeedSettings).toHaveBeenCalledWith({
      path: { environmentId: 1 },
      body: { enabled: true, urls: ["https://example.com/path"] },
    });
    wrapper.unmount();
  });
  it("keeps sitespeed settings open when all URL rows are blank", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    vi.mocked(updateSitespeedSettings).mockClear();
    const vm = wrapper.vm as any;
    vm.sitespeedEnabled = true;
    vm.sitespeedUrls = ["  ", ""];
    await vm.onSitespeedSubmit();
    expect(updateSitespeedSettings).not.toHaveBeenCalled();
    expect(mockError).toHaveBeenCalled();
    expect(mockSuccess).not.toHaveBeenCalled();
    expect(vm.sitespeedUrls).toEqual(["  ", ""]);
    expect(vm.isSitespeedSubmitting).toBe(false);
    wrapper.unmount();
  });
});
