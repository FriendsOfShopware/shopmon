import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import Register from "./Register.vue";

// Mock router
vi.mock("@/router", () => ({
  router: {
    push: vi.fn(),
  },
}));

// Mock vue-router for RouterLink
vi.mock("vue-router", () => ({
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

// Mock useAlert composable
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

describe("Register", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(Register);
  }

  it("renders successfully", () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h2").text()).toBe("Create account");
  });

  it("has display name input field", () => {
    const wrapper = mountComponent();
    // The display name input is the first text input
    const inputs = wrapper.findAll("input");
    expect(inputs.length).toBeGreaterThanOrEqual(2);
  });

  it("has email input field", () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[type="email"]');
    expect(input.exists()).toBe(true);
  });

  it("has password input field", () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[type="password"]');
    expect(input.exists()).toBe(true);
  });

  it("has register button", () => {
    const wrapper = mountComponent();
    const button = wrapper.find('button[type="submit"]');
    expect(button.exists()).toBe(true);
    expect(button.text()).toContain("Register");
  });

  it("has cancel link to login", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Cancel");
  });

  it("has form element with submit handler", () => {
    const wrapper = mountComponent();
    const form = wrapper.find("form");
    expect(form.exists()).toBe(true);
  });

  it("has display name input with correct placeholder", () => {
    const wrapper = mountComponent();
    const inputs = wrapper.findAll("input");
    const displayNameInput = inputs.find(
      (i) => i.attributes("placeholder") === "Display Name",
    );
    expect(displayNameInput).toBeTruthy();
  });

  it("has email input with correct placeholder", () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[type="email"]');
    expect(input.attributes("placeholder")).toBe("Email address");
  });
});
