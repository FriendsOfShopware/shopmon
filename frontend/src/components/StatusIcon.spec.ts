import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import StatusIcon from "./StatusIcon.vue";

describe("StatusIcon", () => {
  it("renders successfully with default props", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "green",
      },
    });
    expect(wrapper.exists()).toBe(true);
  });

  it("renders with tooltip when tooltip prop is true", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "green",
        tooltip: true,
      },
    });
    // When tooltip is true, the component wraps in a <span> with title attribute
    expect(wrapper.find('[title="green"]').exists()).toBe(true);
  });

  it("renders without tooltip when tooltip prop is false", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "green",
        tooltip: false,
      },
    });
    expect(wrapper.find("[title]").exists()).toBe(false);
  });

  it("applies text-success class for green status", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "green",
      },
    });
    expect(wrapper.find(".text-success").exists()).toBe(true);
  });

  it("applies text-destructive class for red status", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "red",
      },
    });
    expect(wrapper.find(".text-destructive").exists()).toBe(true);
  });

  it("applies text-warning class for yellow status", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "yellow",
      },
    });
    expect(wrapper.find(".text-warning").exists()).toBe(true);
  });

  it("applies text-muted-foreground class for inactive status", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "inactive",
      },
    });
    expect(wrapper.find(".text-muted-foreground").exists()).toBe(true);
  });

  it("applies text-muted-foreground class for not installed status", () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: "not installed",
      },
    });
    expect(wrapper.find(".text-muted-foreground").exists()).toBe(true);
  });
});
