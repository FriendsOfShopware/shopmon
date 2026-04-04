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
  props: ["title", "id", "variant", "description", "class"],
  setup(props, { slots }) {
    return () =>
      h("div", { class: "panel", id: props.id }, [
        props.title ? h("h2", {}, props.title) : null,
        slots.action?.(),
        slots.default?.(),
      ]);
  },
});

const FormGroupStub = defineComponent({
  name: "FormGroup",
  props: ["title"],
  setup(props, { slots }) {
    return () => h("fieldset", {}, [h("legend", {}, props.title), slots.default?.()]);
  },
});

const ModalStub = defineComponent({
  name: "Modal",
  props: ["show", "closeXMark"],
  setup(props, { slots }) {
    return () =>
      props.show
        ? h("div", { class: "modal" }, [
            slots.title?.(),
            slots.content?.(),
            slots.footer?.(),
            slots.icon?.(),
          ])
        : null;
  },
});

const DeleteConfirmationModalStub = defineComponent({
  name: "DeleteConfirmationModal",
  props: [
    "show",
    "title",
    "entityName",
    "customConsequence",
    "reversedButtons",
    "isLoading",
    "confirmButtonText",
  ],
  template: '<div v-if="show" class="delete-modal" />',
});

const AlertStub = defineComponent({
  name: "Banner",
  props: ["type"],
  setup(_, { slots }) {
    return () => h("div", { class: "alert" }, slots.default?.());
  },
});

const mockShop = {
  id: 1,
  name: "Test Shop",
  description: "A test shop",
  gitUrl: "https://github.com/test/repo",
  organizationId: "org-1",
};

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { shopId: "1" }, hash: "" }),
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

vi.mock("@/composables/useAccountEnvironments", () => ({
  fetchAccountEnvironments: vi.fn(() => Promise.resolve([])),
  useAccountEnvironments: () => ({
    environments: { value: [] },
    fetchAccountEnvironments: vi.fn(),
  }),
}));

vi.mock("@/helpers/formatter", () => ({
  formatDate: (d: string) => d,
  timeAgo: (d: string) => d,
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

import { api } from "@/helpers/api";

describe("EditShop", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: [mockShop], error: null, response: new Response() });
      }
      if (path === "/api-key-scopes") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/packages-token/configuration") {
        return Promise.resolve({
          data: { configured: false, composerUrl: null },
          error: null,
          response: new Response(),
        });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(EditShop, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          FormGroup: FormGroupStub,
          Modal: ModalStub,
          DeleteConfirmationModal: DeleteConfirmationModalStub,
          Banner: AlertStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toContain("Edit");
  });

  it("shows loading state initially", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Loading shop...");
  });

  it("displays shop form after loading", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("has Shop Information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Shop Information");
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="name"]').exists()).toBe(true);
  });

  it("has description textarea", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('textarea[name="description"]').exists()).toBe(true);
  });

  it("has git URL input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="gitUrl"]').exists()).toBe(true);
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.find('button[type="submit"]');
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toContain("Save");
  });

  it("displays API Keys section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("API Keys");
  });

  it("has Create API Key button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Create API Key"));
    expect(btn).toBeTruthy();
  });

  it("displays Danger Zone section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Danger Zone");
  });

  it("has delete shop button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete shop"));
    expect(btn).toBeTruthy();
  });

  it("has Back to Shops link", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Back to Shops");
  });
});
