import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Privacy from "./Privacy.vue";

// Mock components
vi.mock("@/components/layout/HeaderContainer.vue", () => ({
  default: {
    name: "HeaderContainer",
    props: ["title"],
    template: "<header>{{ title }}</header>",
  },
}));

vi.mock("@/components/layout/MainContainer.vue", () => ({
  default: {
    name: "MainContainer",
    template: "<main><slot /></main>",
  },
}));

describe("Privacy", () => {
  it("renders successfully", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.exists()).toBe(true);
  });

  it("displays the correct page title", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.find("header").text()).toBe("Privacy Policy");
  });

  it("contains Data We Collect section", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Data We Collect");
  });

  it("contains Account Information subsection", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Account Information");
    expect(wrapper.text()).toContain("Email address and name for account creation");
    expect(wrapper.text()).toContain("Password (hashed using bcrypt)");
  });

  it("contains Shop Monitoring Data subsection", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Shop Monitoring Data");
    expect(wrapper.text()).toContain("Shopware instance URLs");
  });

  it("contains Technical Data subsection", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Technical Data");
  });

  it("contains How We Use Your Data section", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("How We Use Your Data");
    expect(wrapper.text()).toContain("To provide shop monitoring services");
  });

  it("contains Data Storage and Security section", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Data Storage and Security");
    expect(wrapper.text()).toContain("Oracle Cloud at Frankfurt in Germany");
  });

  it("contains Cookies and Tracking section", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Cookies and Tracking");
  });

  it("contains Contact section", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.text()).toContain("Contact");
    expect(wrapper.text()).toContain("shopmon at fos.gg");
  });

  it("uses main-container component", () => {
    const wrapper = mount(Privacy);
    expect(wrapper.find("main").exists()).toBe(true);
  });
});
