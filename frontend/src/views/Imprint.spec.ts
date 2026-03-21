import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Imprint from "./Imprint.vue";

// Mock components
vi.mock("@/components/layout/HeaderContainer.vue", () => ({
  default: {
    name: "HeaderContainer",
    props: ["title"],
    template: "<header>{{ title }}</header>",
  },
}));

vi.mock("@/components/layout/Panel.vue", () => ({
  default: {
    name: "Panel",
    template: "<div class='panel'><slot /></div>",
  },
}));

describe("Imprint", () => {
  it("renders successfully", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.exists()).toBe(true);
  });

  it("displays the correct page title", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find("header").text()).toBe("Legal notice");
  });

  it("contains Contact section", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain("Contact");
  });

  it("displays non-profit disclaimer", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain("Non-Profit Website");
    expect(wrapper.text()).toContain("it's not required to provide a legal representative");
  });

  it("displays email contact information", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain("E-Mail: shopmon at fos.gg");
  });

  it("uses Panel component", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find(".panel").exists()).toBe(true);
  });

  it("has container class on wrapper", () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find(".container").exists()).toBe(true);
  });
});
