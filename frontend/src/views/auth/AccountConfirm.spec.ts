import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import AccountConfirm from "./AccountConfirm.vue";

// Mock vue-router
const mockRoute = {
  params: {
    token: "test-token-123",
  },
};

vi.mock("vue-router", () => ({
  useRoute: () => mockRoute,
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  },
}));

// Track mock calls
const mockCalls = {
  error: [] as string[],
};

// Mock api client
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

// Mock useAlert composable
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn((msg: string) => {
      mockCalls.error.push(msg);
    }),
  }),
}));

// Import after mocks
import { api } from "@/helpers/api";

describe("AccountConfirm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCalls.error = [];
  });

  it("renders loading state initially", () => {
    // Override the mock to return a pending promise
    vi.mocked(api.GET).mockImplementationOnce(() => new Promise(() => {}));

    const wrapper = mount(AccountConfirm);

    expect(wrapper.text()).toContain("Confirming your Account Registration");
    // The loading state uses Alert (data-slot="alert") with an info icon
    expect(wrapper.find('[data-slot="alert"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Loading...");
  });

  it("renders success state when email verification succeeds", async () => {
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: {},
      error: undefined,
      response: new Response(),
    } as any);

    const wrapper = mount(AccountConfirm, {
      global: {
        stubs: {
          RouterLink: { props: ["to"], template: "<a><slot /></a>" },
        },
      },
    });

    await flushPromises();

    // Success state uses Alert with green border class
    const alerts = wrapper.findAll('[data-slot="alert"]');
    const successAlert = alerts.find((a) => a.classes().some((c) => c.includes("green")));
    expect(successAlert).toBeTruthy();
    expect(wrapper.text()).toContain("Your email address has been confirmed");
    expect(wrapper.find("a").exists()).toBe(true);
    expect(wrapper.find("a").text()).toBe("Login");
  });

  it("renders error state when token is expired", async () => {
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: undefined,
      error: { message: "Token expired" },
      response: new Response(),
    } as any);

    const wrapper = mount(AccountConfirm);

    await flushPromises();

    // Error state uses Alert with variant="destructive"
    const alert = wrapper.find('[data-slot="alert"]');
    expect(alert.exists()).toBe(true);
    expect(wrapper.text()).toContain("The given token has been expired");
    expect(mockCalls.error).toContain("Token expired");
  });

  it("renders error state with default message when error has no message", async () => {
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: undefined,
      error: {},
      response: new Response(),
    } as any);

    const wrapper = mount(AccountConfirm);

    await flushPromises();

    const alert = wrapper.find('[data-slot="alert"]');
    expect(alert.exists()).toBe(true);
    expect(wrapper.text()).toContain("The given token has been expired");
    expect(mockCalls.error).toContain("Failed to verify email");
  });

  it("calls api.GET with correct path and token on mount", async () => {
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: {},
      error: undefined,
      response: new Response(),
    } as any);

    mount(AccountConfirm);

    await flushPromises();

    expect(api.GET).toHaveBeenCalledWith("/auth/verify-email", {
      params: { query: { token: "test-token-123" } },
    });
  });
});
