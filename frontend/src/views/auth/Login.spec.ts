import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import Login from "./Login.vue";

// Mock router
const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a :href="typeof to === \'string\' ? to : to?.name"><slot /></a>',
  },
}));

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

// Mock useSession composable
vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: { value: { user: { id: "1", email: "test@example.com" }, session: {} } },
    loading: { value: false },
    fetchSession: vi.fn(),
  }),
  fetchSession: vi.fn(),
}));

// Mock useAlert composable
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

// Mock useReturnUrl composable
vi.mock("@/composables/useReturnUrl", () => ({
  useReturnUrl: () => ({
    returnUrl: { value: null },
    clearReturnUrl: vi.fn(),
  }),
}));

describe("Login", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(Login);
  }

  it("renders successfully", () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays sign in heading", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h2").text()).toBe("Sign in to your account");
  });

  it("displays link to create account", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("New to Shopmon?");
    expect(wrapper.text()).toContain("Create an account");
  });

  it("has email input field", () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[type="email"]');
    expect(emailInput.exists()).toBe(true);
    expect(emailInput.attributes("placeholder")).toBe("Email address");
  });

  it("has password input field", () => {
    const wrapper = mountComponent();
    const passwordInput = wrapper.find('input[type="password"]');
    expect(passwordInput.exists()).toBe(true);
    expect(passwordInput.attributes("placeholder")).toBe("Password");
  });

  it("has sign in button", () => {
    const wrapper = mountComponent();
    const signInButton = wrapper.find('button[type="submit"]');
    expect(signInButton.exists()).toBe(true);
    expect(signInButton.text()).toContain("Sign in");
  });

  it("has forgot password link", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Forgot your password?");
  });

  it("has passkey login button", () => {
    const wrapper = mountComponent();
    const passkeyButton = wrapper.findAll("button").find((b) => b.text().includes("Passkey"));
    expect(passkeyButton).toBeTruthy();
  });

  it("has GitHub login button", () => {
    const wrapper = mountComponent();
    const githubButton = wrapper.findAll("button").find((b) => b.text().includes("GitHub"));
    expect(githubButton).toBeTruthy();
  });

  it("has SSO section with email input", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Enterprise SSO");
    expect(wrapper.text()).toContain("Use your company email to sign in with SSO");

    const ssoInput = wrapper
      .findAll("input")
      .find((i) => i.attributes("placeholder")?.includes("work email"));
    expect(ssoInput).toBeTruthy();
  });

  it("has form element", () => {
    const wrapper = mountComponent();
    const form = wrapper.find("form");
    expect(form.exists()).toBe(true);
  });

  it("displays divider between login methods", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Or");
  });
});
