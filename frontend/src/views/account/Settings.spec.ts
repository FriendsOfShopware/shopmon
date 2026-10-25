import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, ref } from "vue";
import Settings from "./Settings.vue";

const MainContainerStub = defineComponent({ template: "<div><slot /></div>" });
const FormGroupStub = defineComponent({ template: "<div><slot /></div>" });
const DataTableStub = defineComponent({ template: "<div><slot /></div>" });
const DeleteConfirmationModalStub = defineComponent({ template: "<div><slot /></div>" });

const mockSessionData = {
  user: {
    id: "user-1",
    name: "John Doe",
    email: "john@example.com",
    emailVerified: true,
    image: null,
    role: "user",
  },
  session: {
    id: "session-1",
    userId: "user-1",
    expiresAt: "2099-01-01",
    activeOrganizationId: "org-1",
  },
};

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: ref(mockSessionData),
    loading: ref(false),
    fetchSession: vi.fn(),
  }),
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  deletePasskey: vi.fn(),
  revokeSession: vi.fn(),
  unsubscribeFromEnvironment: vi.fn(),
  listUserPasskeys: vi.fn(),
  listSessions: vi.fn(),
  listAccounts: vi.fn(),
  getAccountOrganizations: vi.fn(),
  getAccountSubscribedEnvironments: vi.fn(),
  getNotificationPreferences: vi.fn(),
  getNotificationEventTypes: vi.fn(),
}));

const mockError = vi.hoisted(() => vi.fn());
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: mockError,
    success: vi.fn(),
  }),
}));

vi.mock("@simplewebauthn/browser", () => ({
  startRegistration: vi.fn(),
}));

import {
  deletePasskey,
  revokeSession,
  unsubscribeFromEnvironment,
  getAccountOrganizations,
  getAccountSubscribedEnvironments,
  getNotificationEventTypes,
  getNotificationPreferences,
  listAccounts,
  listSessions,
  listUserPasskeys,
} from "@/api/generated";

describe("Settings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listUserPasskeys).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(listSessions).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(listAccounts).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getAccountOrganizations).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getAccountSubscribedEnvironments).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getNotificationPreferences).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    vi.mocked(getNotificationEventTypes).mockResolvedValue({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(Settings, {
      global: {
        stubs: {
          MainContainer: MainContainerStub,
          FormGroup: FormGroupStub,
          DataTable: DataTableStub,
          DeleteConfirmationModal: DeleteConfirmationModalStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("Settings");
  });

  it("populates form with user session name", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const nameInput = wrapper.find('input[type="text"]');
    expect(nameInput.exists()).toBe(true);
    expect((nameInput.element as HTMLInputElement).value).toBe("John Doe");
  });

  it("populates form with user session email", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const emailInput = wrapper.find('input[type="email"]');
    expect(emailInput.exists()).toBe(true);
    expect((emailInput.element as HTMLInputElement).value).toBe("john@example.com");
  });

  it("shows security section with passkeys and sessions", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Passkey Devices");
    expect(wrapper.text()).toContain("Sessions");
  });

  it("shows danger zone section with delete account button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Delete account");
  });

  it("calls list APIs on mount", async () => {
    mountComponent();
    await flushPromises();
    expect(listUserPasskeys).toHaveBeenCalled();
    expect(listSessions).toHaveBeenCalled();
    expect(listAccounts).toHaveBeenCalled();
    expect(getAccountOrganizations).toHaveBeenCalled();
  });
  it.each([
    ["removePasskey", deletePasskey, listUserPasskeys, "passkey"],
    ["removeSession", revokeSession, listSessions, { id: "session" }],
    ["unsubscribeFromEnvironment", unsubscribeFromEnvironment, getAccountSubscribedEnvironments, 1],
  ] as const)(
    "%s reports resolved errors without reloading state",
    async (action, mutation, loader, arg) => {
      vi.mocked(mutation).mockResolvedValue({ error: { message: "Request rejected" } } as any);
      const wrapper = mountComponent();
      await flushPromises();
      await (wrapper.vm as any)[action](arg);
      expect(mockError).toHaveBeenCalledWith("Request rejected");
      expect(loader).toHaveBeenCalledTimes(1);
      wrapper.unmount();
    },
  );
});
