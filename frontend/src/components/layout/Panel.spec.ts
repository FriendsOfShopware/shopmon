import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Panel from "./Panel.vue";

describe("Panel", () => {
  it("renders with Card (data-slot=card)", () => {
    const wrapper = mount(Panel);
    expect(wrapper.find('[data-slot="card"]').exists()).toBe(true);
  });

  it("renders default slot content", () => {
    const wrapper = mount(Panel, { slots: { default: "<p>Content</p>" } });
    expect(wrapper.find("p").text()).toBe("Content");
  });

  it("renders title via CardTitle when no action slot", () => {
    const wrapper = mount(Panel, { props: { title: "My Title" } });
    const title = wrapper.find('[data-slot="card-title"]');
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("My Title");
  });

  it("renders title slot via CardTitle when no action slot", () => {
    const wrapper = mount(Panel, {
      slots: { title: "Slot Title" },
    });
    const title = wrapper.find('[data-slot="card-title"]');
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("Slot Title");
  });

  it("renders CardHeader when title prop and action slot are provided", () => {
    const wrapper = mount(Panel, {
      props: { title: "Header Title" },
      slots: { action: "<button>Action</button>" },
    });
    expect(wrapper.find('[data-slot="card-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-slot="card-title"]').text()).toBe("Header Title");
    expect(wrapper.find("button").text()).toBe("Action");
  });

  it("renders CardHeader when title slot and action slot are provided", () => {
    const wrapper = mount(Panel, {
      slots: {
        title: "Header Title",
        action: "<button>Action</button>",
      },
    });
    expect(wrapper.find('[data-slot="card-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-slot="card-title"]').text()).toBe("Header Title");
    expect(wrapper.find("button").text()).toBe("Action");
  });

  it("renders description via CardDescription", () => {
    const wrapper = mount(Panel, {
      props: { title: "Title", description: "Some description" },
      slots: { action: "<button>Action</button>" },
    });
    const desc = wrapper.find('[data-slot="card-description"]');
    expect(desc.exists()).toBe(true);
    expect(desc.text()).toBe("Some description");
  });

  it("does not render description without description prop", () => {
    const wrapper = mount(Panel, {
      props: { title: "Title" },
    });
    expect(wrapper.find('[data-slot="card-description"]').exists()).toBe(false);
  });

  it("applies overflow-hidden for table variant", () => {
    const wrapper = mount(Panel, { props: { variant: "table" } });
    const card = wrapper.find('[data-slot="card"]');
    expect(card.classes()).toContain("overflow-hidden");
  });

  it("does not apply overflow-hidden for default variant", () => {
    const wrapper = mount(Panel);
    const card = wrapper.find('[data-slot="card"]');
    expect(card.classes()).not.toContain("overflow-hidden");
  });

  it("does not render title or header when neither is provided", () => {
    const wrapper = mount(Panel, { slots: { default: "Just content" } });
    expect(wrapper.find('[data-slot="card-title"]').exists()).toBe(false);
    expect(wrapper.find('[data-slot="card-header"]').exists()).toBe(false);
    expect(wrapper.text()).toBe("Just content");
  });
});
