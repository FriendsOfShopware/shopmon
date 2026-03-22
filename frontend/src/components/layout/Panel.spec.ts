import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Panel from "./Panel.vue";

describe("Panel", () => {
  it("renders with panel class", () => {
    const wrapper = mount(Panel);
    expect(wrapper.find(".panel").exists()).toBe(true);
  });

  it("renders default slot content", () => {
    const wrapper = mount(Panel, { slots: { default: "<p>Content</p>" } });
    expect(wrapper.find("p").text()).toBe("Content");
  });

  it("renders title as panel-title when no action slot", () => {
    const wrapper = mount(Panel, { props: { title: "My Title" } });
    const title = wrapper.find(".panel-title");
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("My Title");
  });

  it("renders title slot as panel-title when no action slot", () => {
    const wrapper = mount(Panel, {
      slots: { title: "Slot Title" },
    });
    const title = wrapper.find(".panel-title");
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("Slot Title");
  });

  it("renders panel-header when title prop and action slot are provided", () => {
    const wrapper = mount(Panel, {
      props: { title: "Header Title" },
      slots: { action: "<button>Action</button>" },
    });
    expect(wrapper.find(".panel-header").exists()).toBe(true);
    expect(wrapper.find(".panel-header h3").text()).toBe("Header Title");
    expect(wrapper.find(".panel-header button").text()).toBe("Action");
  });

  it("renders panel-header when title slot and action slot are provided", () => {
    const wrapper = mount(Panel, {
      slots: {
        title: "Header Title",
        action: "<button>Action</button>",
      },
    });
    expect(wrapper.find(".panel-header").exists()).toBe(true);
    expect(wrapper.find(".panel-header h3").text()).toBe("Header Title");
    expect(wrapper.find(".panel-header button").text()).toBe("Action");
  });

  it("renders description in panel-header", () => {
    const wrapper = mount(Panel, {
      props: { title: "Title", description: "Some description" },
      slots: { action: "<button>Action</button>" },
    });
    const desc = wrapper.find(".panel-description");
    expect(desc.exists()).toBe(true);
    expect(desc.text()).toBe("Some description");
  });

  it("does not render description without action slot", () => {
    const wrapper = mount(Panel, {
      props: { title: "Title", description: "Some description" },
    });
    expect(wrapper.find(".panel-description").exists()).toBe(false);
  });

  it("applies panel-table class for table variant", () => {
    const wrapper = mount(Panel, { props: { variant: "table" } });
    expect(wrapper.find(".panel-table").exists()).toBe(true);
  });

  it("does not apply panel-table class for default variant", () => {
    const wrapper = mount(Panel);
    expect(wrapper.find(".panel-table").exists()).toBe(false);
  });

  it("does not render title or header when neither is provided", () => {
    const wrapper = mount(Panel, { slots: { default: "Just content" } });
    expect(wrapper.find(".panel-title").exists()).toBe(false);
    expect(wrapper.find(".panel-header").exists()).toBe(false);
    expect(wrapper.text()).toBe("Just content");
  });
});
