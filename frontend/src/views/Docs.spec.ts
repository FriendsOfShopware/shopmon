import { describe, it, expect, beforeAll } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import Docs from "./Docs.vue";

beforeAll(() => {
  globalThis.IntersectionObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as any;
});

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

describe("Docs", () => {
  function mountComponent() {
    return mount(Docs, {
      global: {
        stubs: {
          MainContainer: MainContainerStub,
        },
      },
    });
  }

  it("renders page title", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h1").text()).toBe("Documentation");
  });

  it("renders table of contents", () => {
    const wrapper = mountComponent();
    // TOC is now inside a Card component
    const tocLinks = wrapper.findAll('a[href^="#"]');
    expect(tocLinks.length).toBeGreaterThan(0);
  });

  it("has all TOC links", () => {
    const wrapper = mountComponent();
    const links = wrapper.findAll('a[href^="#"]');
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

  it("renders all documentation sections as Card components", () => {
    const wrapper = mountComponent();
    // Each section is now a Card with data-slot="card"
    const cards = wrapper.findAll('[data-slot="card"]');
    expect(cards.length).toBeGreaterThanOrEqual(1);
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
