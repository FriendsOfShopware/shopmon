import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h, ref } from "vue";
import AddProject from "./AddProject.vue";

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
  setup(_, { slots }) {
    return () => h("div", { class: "panel" }, slots.default?.());
  },
});

const FormGroupStub = defineComponent({
  name: "FormGroup",
  props: ["title"],
  setup(props, { slots }) {
    return () => h("fieldset", {}, [h("legend", {}, props.title), slots.default?.()]);
  },
});

const mockOrganizations = [
  { id: "org-1", name: "Organization A" },
  { id: "org-2", name: "Organization B" },
];

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
}));

vi.mock("@/helpers/auth-client", () => ({
  authClient: {
    useListOrganizations: () => ref({ data: mockOrganizations, isPending: false }),
  },
}));

vi.mock("@/helpers/trpc", () => ({
  trpcClient: {
    organization: {
      project: {
        create: { mutate: vi.fn(() => Promise.resolve()) },
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

describe("AddProject", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(AddProject, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          FormGroup: FormGroupStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toBe("New Project");
  });

  it("renders form when organizations are loaded", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("displays Project information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Project information");
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

  it("has organization select with options", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const select = wrapper.find('select[name="organizationId"]');
    expect(select.exists()).toBe(true);
    const options = select.findAll("option");
    expect(options.length).toBe(3); // empty + 2 orgs
    expect(options[1].text()).toBe("Organization A");
    expect(options[2].text()).toBe("Organization B");
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.find('button[type="submit"]');
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toContain("Save");
  });
});
