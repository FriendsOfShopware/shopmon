import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import Docs from "./Docs.vue";

const HeaderContainerStub = defineComponent({
  name: "HeaderContainer",
  props: ["title"],
  template: "<header>{{ title }}</header>",
});

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

const AlertStub = defineComponent({
  name: "Banner",
  props: ["type"],
  setup(_, { slots }) {
    return () => h("div", { class: "alert" }, slots.default?.());
  },
});

describe("Docs", () => {
  function mountComponent() {
    return mount(Docs, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          Banner: AlertStub,
        },
      },
    });
  }

  it("renders page title", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("header").text()).toBe("Documentation");
  });

  it("renders table of contents", () => {
    const wrapper = mountComponent();
    expect(wrapper.find(".docs-nav").exists()).toBe(true);
    expect(wrapper.find(".docs-toc").exists()).toBe(true);
  });

  it("has all TOC links", () => {
    const wrapper = mountComponent();
    const links = wrapper.findAll(".docs-toc a");
    const hrefs = links.map((l) => l.attributes("href"));
    expect(hrefs).toContain("#getting-started");
    expect(hrefs).toContain("#connecting-shop");
    expect(hrefs).toContain("#organizations");
    expect(hrefs).toContain("#projects");
    expect(hrefs).toContain("#dashboard-overview");
    expect(hrefs).toContain("#health-checks");
    expect(hrefs).toContain("#extensions");
    expect(hrefs).toContain("#scheduled-tasks");
    expect(hrefs).toContain("#queue");
    expect(hrefs).toContain("#sitespeed");
    expect(hrefs).toContain("#deployments");
    expect(hrefs).toContain("#changelog");
    expect(hrefs).toContain("#notifications");
    expect(hrefs).toContain("#shop-token");
    expect(hrefs).toContain("#packages-mirror");
    expect(hrefs).toContain("#sso");
  });

  it("renders all documentation sections", () => {
    const wrapper = mountComponent();
    const sections = wrapper.findAll(".docs-section");
    expect(sections.length).toBeGreaterThanOrEqual(16);
  });

  it("has Getting Started section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#getting-started").exists()).toBe(true);
    expect(wrapper.text()).toContain("Getting Started");
  });

  it("has Connecting a Shop section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#connecting-shop").exists()).toBe(true);
    expect(wrapper.text()).toContain("Connecting a Shop");
  });

  it("has Organizations section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#organizations").exists()).toBe(true);
  });

  it("has Health Checks section with check descriptions", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#health-checks").exists()).toBe(true);
    expect(wrapper.text()).toContain("Security");
    expect(wrapper.text()).toContain("Environment");
    expect(wrapper.text()).toContain("Scheduled Tasks");
    expect(wrapper.text()).toContain("Admin Worker");
  });

  it("has Deployments section with CLI info", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#deployments").exists()).toBe(true);
    expect(wrapper.text()).toContain("shopmon-cli");
  });

  it("has SSO section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#sso").exists()).toBe(true);
    expect(wrapper.text()).toContain("Single Sign-On");
  });

  it("has Packages Mirror section", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("#packages-mirror").exists()).toBe(true);
    expect(wrapper.text()).toContain("Packages Mirror");
  });

  it("has How Scraping Works section", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("How Scraping Works");
  });
});
