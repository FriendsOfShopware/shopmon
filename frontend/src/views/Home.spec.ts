import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { ref } from "vue";
import Home from "./Home.vue";

// Mock vue-router
vi.mock("vue-router", () => ({
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a :href="to.name"><slot /></a>',
  },
}));

// Mock useDarkMode composable
const mockGetThemeImage = vi.fn((path: string) => path);
vi.mock("@/composables/useDarkMode", () => ({
  useDarkMode: () => ({
    getThemeImage: mockGetThemeImage,
  }),
}));

describe("Home", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders successfully", () => {
    const wrapper = mount(Home);
    expect(wrapper.exists()).toBe(true);
  });

  it("displays main heading", () => {
    const wrapper = mount(Home);
    expect(wrapper.find("h1").text()).toBe(
      "Shopmon - The Open-Source Dashboard for Shopware Developers",
    );
  });

  it("displays subheading about dashboard overview", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Everything at a glance");
  });

  it("displays feature highlights", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Dashboard overview");
    expect(wrapper.text()).toContain("Automatic performance checks");
    expect(wrapper.text()).toContain("Performance monitoring");
  });

  it("displays Shopware versions section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Shopware versions & environment state dashboard");
  });

  it("displays Frosh Tools section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Frosh Tools performance checks");
  });

  it("displays Extensions section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Extensions and updates");
  });

  it("displays Sitespeed section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Automatic sitespeed.io checks");
  });

  it("displays CTA section with register link", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Ready to monitor your Shopware?");
    expect(wrapper.text()).toContain("Start monitoring for free");
  });

  it("displays value propositions", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Forever Free");
    expect(wrapper.text()).toContain("Community Driven");
    expect(wrapper.text()).toContain("Open Source");
  });

  it("displays GitHub repository link", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("GitHub");
  });

  it("uses theme-aware images", () => {
    mount(Home);
    // Verify that getThemeImage was called for theme-aware images
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-dashboard.png");
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-performance-checks.png");
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-extensions.png");
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-sitespeed.png");
  });

  it("has header section", () => {
    const wrapper = mount(Home);
    expect(wrapper.find(".header").exists()).toBe(true);
  });

  it("has panel-intro section", () => {
    const wrapper = mount(Home);
    expect(wrapper.find(".panel-intro").exists()).toBe(true);
  });

  it("has multiple row sections", () => {
    const wrapper = mount(Home);
    const rows = wrapper.findAll(".row");
    expect(rows.length).toBeGreaterThan(0);
  });

  it("has primary CTA section", () => {
    const wrapper = mount(Home);
    expect(wrapper.find(".section-primary").exists()).toBe(true);
  });
});
