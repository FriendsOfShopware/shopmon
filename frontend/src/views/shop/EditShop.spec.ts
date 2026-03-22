import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import EditShop from "./EditShop.vue";

const HeaderContainerStub = defineComponent({
  name: "HeaderContainer",
  props: ["title"],
  setup(props, { slots }) {
    return () => h("header", {}, [props.title, slots.default?.()]);
  },
});

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

const PanelStub = defineComponent({
  name: "Panel",
  props: ["title", "variant"],
  setup(props, { slots }) {
    return () =>
      h("div", { class: "panel" }, [
        props.title ? h("h2", {}, props.title) : null,
        slots.default?.(),
      ]);
  },
});

const FormGroupStub = defineComponent({
  name: "FormGroup",
  props: ["title"],
  setup(props, { slots }) {
    return () =>
      h("fieldset", {}, [h("legend", {}, props.title), slots.info?.(), slots.default?.()]);
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

const mockShop = {
  id: 1,
  name: "Test Shop",
  url: "https://test.shop",
  organizationId: "org-1",
  projectId: 1,
  clientId: "client-1",
  clientSecret: "",
  sitespeedEnabled: false,
  sitespeedUrls: [],
};

const mockProjects = [
  { id: 1, nameCombined: "Org / Project A" },
  { id: 2, nameCombined: "Org / Project B" },
];

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { slug: "org-1", shopId: "1" } }),
}));

vi.mock("@/helpers/trpc", () => ({
  trpcClient: {
    account: {
      currentUserProjects: {
        query: vi.fn(() => Promise.resolve(mockProjects)),
      },
    },
    organization: {
      shop: {
        get: { query: vi.fn(() => Promise.resolve(mockShop)) },
        update: { mutate: vi.fn(() => Promise.resolve()) },
        delete: { mutate: vi.fn(() => Promise.resolve()) },
        updateSitespeedSettings: { mutate: vi.fn(() => Promise.resolve()) },
      },
    },
  },
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

describe("EditShop", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(EditShop, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          FormGroup: FormGroupStub,
          DeleteConfirmationModal: DeleteConfirmationModalStub,
          PluginConnectionModal: PluginConnectionModalStub,
        },
      },
    });
  }

  it("renders page title with shop name", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toContain("Edit Test Shop");
  });

  it("has cancel link", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toContain("Cancel");
  });

  it("displays shop information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Shop information");
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="name"]').exists()).toBe(true);
  });

  it("has project select", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('select[name="projectId"]').exists()).toBe(true);
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

  it("displays delete shop section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Deleting shop");
  });

  it("has delete shop button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete shop"));
    expect(btn).toBeTruthy();
  });
});
