import { describe, it, expect, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { h, defineComponent } from "vue";
import AccountConfirm from "./AccountConfirm.vue";

// Create a stub for Alert component
const AlertStub = defineComponent({
  name: "Alert",
  props: ["type"],
  setup(props, { slots }) {
    return () => h("div", { class: `alert alert-${props.type}` }, slots.default?.());
  },
});

// Create a stub for router-link
const RouterLinkStub = defineComponent({
  name: "RouterLink",
  props: ["to"],
  setup(props, { slots }) {
    return () => h("a", { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock vue-router
const mockRoute = {
  params: {
    token: "test-token-123",
  },
};

vi.mock("vue-router", () => ({
  useRoute: () => mockRoute,
}));

// Track mock calls
const mockCalls = {
  verifyEmail: [] as any[],
  error: [] as string[],
};

// Mock auth client
vi.mock("@/helpers/auth-client", () => ({
  authClient: {
    verifyEmail: vi.fn((...args: any[]) => {
      mockCalls.verifyEmail.push(args);
      return Promise.resolve({ data: {}, error: null });
    }),
  },
}));

// Mock useAlert composable
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn((msg: string) => {
      mockCalls.error.push(msg);
    }),
  }),
}));

// Import after mocks
import { authClient } from "@/helpers/auth-client";

describe("AccountConfirm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCalls.verifyEmail = [];
    mockCalls.error = [];
  });

  it("renders loading state initially", () => {
    // Override the mock to return a pending promise
    vi.mocked(authClient.verifyEmail).mockImplementationOnce(() => new Promise(() => {}));

    const wrapper = mount(AccountConfirm, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
        },
      },
    });

    expect(wrapper.text()).toContain("Confirming your Account Registration");
    expect(wrapper.find(".alert-info").exists()).toBe(true);
    expect(wrapper.text()).toContain("Loading...");
  });

  it("renders success state when email verification succeeds", async () => {
    vi.mocked(authClient.verifyEmail).mockResolvedValueOnce({
      data: {},
      error: null,
    });

    const wrapper = mount(AccountConfirm, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
        },
      },
    });

    await flushPromises();

    expect(wrapper.find(".alert-success").exists()).toBe(true);
    expect(wrapper.text()).toContain("Your email address has been confirmed");
    expect(wrapper.find("a").exists()).toBe(true);
    expect(wrapper.find("a").text()).toBe("Login");
  });

  it("renders error state when token is expired", async () => {
    vi.mocked(authClient.verifyEmail).mockResolvedValueOnce({
      data: null,
      error: { message: "Token expired" },
    });

    const wrapper = mount(AccountConfirm, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
        },
      },
    });

    await flushPromises();

    expect(wrapper.find(".alert-error").exists()).toBe(true);
    expect(wrapper.text()).toContain("The given token has been expired");
    expect(mockCalls.error).toContain("Token expired");
  });

  it("renders error state with default message when error has no message", async () => {
    vi.mocked(authClient.verifyEmail).mockResolvedValueOnce({
      data: null,
      error: {},
    });

    const wrapper = mount(AccountConfirm, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
        },
      },
    });

    await flushPromises();

    expect(wrapper.find(".alert-error").exists()).toBe(true);
    expect(wrapper.text()).toContain("The given token has been expired");
    expect(mockCalls.error).toContain("Failed to verify email");
  });

  it("calls verifyEmail with correct token on mount", async () => {
    vi.mocked(authClient.verifyEmail).mockResolvedValueOnce({
      data: {},
      error: null,
    });

    mount(AccountConfirm, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
        },
      },
    });

    await flushPromises();

    expect(authClient.verifyEmail).toHaveBeenCalledWith({
      query: { token: "test-token-123" },
    });
  });
});
