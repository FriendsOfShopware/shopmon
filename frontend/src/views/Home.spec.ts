import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
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

vi.mock("@/data/sponsors", () => ({
  sponsors: [
    {
      name: "Acme Commerce",
      url: "https://example.com/acme",
      description: "Supporting Shopmon development.",
    },
  ],
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
    expect(wrapper.find("h1").text()).toContain(
      "The monitoring dashboard",
    );
    expect(wrapper.find("h1").text()).toContain(
      "your Shopware shops deserve",
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

  it("displays Health Checks showcase section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Health Checks");
    expect(wrapper.text()).toContain("Frosh Tools performance checks");
  });

  it("displays Frosh Tools section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Frosh Tools performance checks");
  });

  it("displays Performance showcase section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Performance");
    expect(wrapper.text()).toContain("Automatic sitespeed.io checks");
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
    expect(wrapper.text()).toContain("Free");
    expect(wrapper.text()).toContain("Community Driven");
    expect(wrapper.text()).toContain("Open Source");
  });

  it("displays the sponsors section", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("Sponsors");
    expect(wrapper.text()).toContain("Acme Commerce");
  });

  it("displays GitHub repository link", () => {
    const wrapper = mount(Home);
    expect(wrapper.text()).toContain("GitHub");
  });

  it("uses theme-aware images", () => {
    mount(Home);
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-dashboard.png");
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-performance-checks.png");
    expect(mockGetThemeImage).toHaveBeenCalledWith("/home/shopmon-sitespeed.png");
  });

  it("has main heading in h1", () => {
    const wrapper = mount(Home);
    expect(wrapper.find("h1").exists()).toBe(true);
  });

  it("has feature cards for bento features", () => {
    const wrapper = mount(Home);
    expect(wrapper.findAll(".bg-card").length).toBeGreaterThan(0);
  });

  it("has feature images", () => {
    const wrapper = mount(Home);
    const images = wrapper.findAll("img");
    expect(images.length).toBeGreaterThan(0);
  });

  it("has primary CTA section with gradient background", () => {
    const wrapper = mount(Home);
    // CTA section uses a gradient with primary color
    expect(wrapper.find(".bg-gradient-to-br").exists()).toBe(true);
  });
});
