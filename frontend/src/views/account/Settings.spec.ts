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
  name: "Alert",
  props: ["type"],
  setup(_, { slots }) {
    return () => h("div", { class: "alert" }, slots.default?.());
  },
});

const mockSession = {
  value: {
    data: {
      user: { id: "1", name: "Test User", email: "test@example.com" },
      session: { token: "test-token" },
    },
  },
};

vi.mock("@/helpers/auth-client", () => ({
  authClient: {
    useSession: () => mockSession,
    useListOrganizations: () => ref({ data: [] }),
    passkey: {
      listUserPasskeys: vi.fn(() => Promise.resolve({ data: [] })),
      addPasskey: vi.fn(() => Promise.resolve()),
      deletePasskey: vi.fn(() => Promise.resolve()),
    },
    listSessions: vi.fn(() => Promise.resolve({ data: [] })),
    listAccounts: vi.fn(() => Promise.resolve({ data: [] })),
    changeEmail: vi.fn(() => Promise.resolve()),
    updateUser: vi.fn(() => Promise.resolve()),
    changePassword: vi.fn(() => Promise.resolve()),
    deleteUser: vi.fn(() => Promise.resolve({ error: null })),
    linkSocial: vi.fn(() => Promise.resolve()),
    unlinkAccount: vi.fn(() => Promise.resolve()),
  },
}));

vi.mock("@/helpers/trpc", () => ({
  trpcClient: {
    account: {
      subscribedShops: {
        query: vi.fn(() => Promise.resolve([])),
      },
    },
    organization: {
      shop: {
        unsubscribeFromNotifications: {
          mutate: vi.fn(() => Promise.resolve()),
        },
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

describe("Settings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
          Alert: AlertStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("header").text()).toBe("Settings");
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
    expect(wrapper.text()).toContain("You are not subscribed to any shop notifications");
  });
});
