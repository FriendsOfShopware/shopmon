import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import ForgotPassword from "./ForgotPassword.vue";

// Mock vue-router for RouterLink
vi.mock("vue-router", () => ({
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a :href="to"><slot /></a>',
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

// Mock useAlert composable
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("ForgotPassword", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(ForgotPassword, {
      global: {
        stubs: {
          RouterLink: {
            props: ["to"],
            template: "<a :href=\"typeof to === 'string' ? to : to?.name\"><slot /></a>",
          },
        },
      },
    });
  }

  it("renders successfully", () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h2").text()).toBe("Forgot password");
  });

  it("displays instructions", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("We will send you a confirmation email");
  });

  it("has email input field", () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[type="email"]');
    expect(emailInput.exists()).toBe(true);
  });

  it("has submit button", () => {
    const wrapper = mountComponent();
    const submitButton = wrapper.find('button[type="submit"]');
    expect(submitButton.exists()).toBe(true);
    expect(submitButton.text()).toContain("Send email");
  });

  it("has cancel link to login", () => {
    const wrapper = mountComponent();
    const cancelLink = wrapper.find('a[href="login"]');
    expect(cancelLink.exists()).toBe(true);
    expect(cancelLink.text()).toBe("Cancel");
  });

  it("has form element with submit handler", () => {
    const wrapper = mountComponent();
    const form = wrapper.find("form");
    expect(form.exists()).toBe(true);
  });

  it("has email input with correct attributes", () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[type="email"]');
    expect(emailInput.attributes("placeholder")).toBe("Email address");
  });
});
