import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import EditProject from "./EditProject.vue";

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
  name: "Alert",
  props: ["type"],
  setup(_, { slots }) {
    return () => h("div", { class: "alert" }, slots.default?.());
  },
});

const mockProject = {
  id: 1,
  name: "Test Project",
  description: "A test project",
  gitUrl: "https://github.com/test/repo",
  organizationId: "org-1",
};

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { projectId: "1" }, hash: "" }),
}));

vi.mock("@/helpers/trpc", () => ({
  trpcClient: {
    account: {
      currentUserProjects: {
        query: vi.fn(() => Promise.resolve([mockProject])),
      },
      currentUserShops: {
        query: vi.fn(() => Promise.resolve([])),
      },
    },
    organization: {
      project: {
        update: { mutate: vi.fn(() => Promise.resolve()) },
        delete: { mutate: vi.fn(() => Promise.resolve()) },
      },
      apiKey: {
        list: { query: vi.fn(() => Promise.resolve([])) },
        scopes: { query: vi.fn(() => Promise.resolve([])) },
        create: { mutate: vi.fn(() => Promise.resolve({ token: "test-token" })) },
        delete: { mutate: vi.fn(() => Promise.resolve()) },
      },
      packagesToken: {
        configuration: {
          query: vi.fn(() => Promise.resolve({ configured: false, composerUrl: null })),
        },
        list: { query: vi.fn(() => Promise.resolve([])) },
        create: { mutate: vi.fn(() => Promise.resolve()) },
        delete: { mutate: vi.fn(() => Promise.resolve()) },
        sync: { mutate: vi.fn(() => Promise.resolve()) },
      },
    },
  },
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

describe("EditProject", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(EditProject, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          FormGroup: FormGroupStub,
          Modal: ModalStub,
          DeleteConfirmationModal: DeleteConfirmationModalStub,
          Alert: AlertStub,
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
    expect(wrapper.text()).toContain("Loading project...");
  });

  it("displays project form after loading", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("has Project Information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Project Information");
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

  it("has delete project button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete project"));
    expect(btn).toBeTruthy();
  });

  it("has Back to Projects link", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Back to Projects");
  });
});
