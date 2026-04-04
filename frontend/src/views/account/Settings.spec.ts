import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h, ref } from "vue";
import Settings from "./Settings.vue";

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
  props: ["title"],
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
  props: ["title", "subTitle", "class"],
  setup(props, { slots }) {
    return () => h("fieldset", {}, [h("legend", {}, props.title), slots.default?.()]);
  },
});

const DataTableStub = defineComponent({
  name: "DataTable",
  props: ["columns", "data"],
  template: "<table><slot /></table>",
});

const ModalStub = defineComponent({
  name: "Modal",
  props: ["show"],
  setup(props, { slots }) {
    return () =>
      props.show
        ? h("div", { class: "modal" }, [slots.title?.(), slots.content?.(), slots.footer?.()])
        : null;
  },
});

const DeleteConfirmationModalStub = defineComponent({
  name: "DeleteConfirmationModal",
  props: ["show", "title", "entityName", "requirePassword"],
  template: '<div v-if="show" class="delete-modal" />',
});

const AlertStub = defineComponent({
  name: "Banner",
  props: ["type"],
  setup(_, { slots }) {
    return () => h("div", { class: "alert" }, slots.default?.());
  },
});

const mockSessionData = {
  user: {
    id: "1",
    name: "Test User",
    email: "test@example.com",
    emailVerified: true,
    image: null,
    role: "user",
  },
  session: { id: "sess-1", userId: "1", expiresAt: "2099-01-01", activeOrganizationId: null },
};

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref(mockSessionData),
    loading: ref(false),
    fetchSession: vi.fn(),
  }),
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
  getToken: vi.fn(() => "test-token"),
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

vi.mock("@simplewebauthn/browser", () => ({
  startRegistration: vi.fn(),
}));

import { api } from "@/helpers/api";

describe("Settings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/auth/passkey/list-user-passkeys") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/auth/list-sessions") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/auth/list-accounts") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/auth/list-organizations") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/account/subscribed-environments") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
  });

  function mountComponent() {
    return mount(Settings, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Panel: PanelStub,
          FormGroup: FormGroupStub,
          DataTable: DataTableStub,
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
    expect(wrapper.find("h1").text()).toBe("Settings");
  });

  it("displays account section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Account");
  });

  it("displays passkey devices section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Passkey Devices");
  });

  it("displays sessions section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Sessions");
  });

  it("displays notifications section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Notifications");
  });

  it("displays delete account section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Deleting your Account");
  });

  it("has save button in account form", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const saveBtn = wrapper.findAll("button").find((b) => b.text().includes("Save"));
    expect(saveBtn).toBeTruthy();
  });

  it("has add passkey button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Add a new Device"));
    expect(btn).toBeTruthy();
  });

  it("has delete account button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete account"));
    expect(btn).toBeTruthy();
  });

  it("shows GitHub link button when not connected", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Link GitHub"));
    expect(btn).toBeTruthy();
  });

  it("shows empty notification state when no subscriptions", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("You are not subscribed to any environment notifications");
  });
});
