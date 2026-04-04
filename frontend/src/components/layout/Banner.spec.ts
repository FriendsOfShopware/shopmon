import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Banner from "./Banner.vue";

describe("Banner", () => {
  it("renders successfully with error variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "error",
      },
      slots: {
        default: "Error message",
      },
    });
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find(".text-destructive").exists()).toBe(true);
  });

  it("renders successfully with success variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "success",
      },
      slots: {
        default: "Success message",
      },
    });
    expect(wrapper.find(".text-success").exists()).toBe(true);
  });

  it("renders successfully with default variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "default",
      },
      slots: {
        default: "Info message",
      },
    });
    expect(wrapper.find(".text-info").exists()).toBe(true);
  });

  it("renders successfully with alert variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "alert",
      },
      slots: {
        default: "Warning message",
      },
    });
    expect(wrapper.find(".text-warning").exists()).toBe(true);
  });

  it("displays slot content correctly", () => {
    const message = "This is an important alert";
    const wrapper = mount(Banner, {
      props: {
        variant: "default",
      },
      slots: {
        default: message,
      },
    });
    expect(wrapper.text()).toContain(message);
  });

  it("renders correct icon for error variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "error",
      },
      slots: {
        default: "Error",
      },
    });
    // Error variant uses text-destructive color class on the Alert wrapper
    expect(wrapper.find('[data-slot="alert"]').classes()).toContain("text-destructive");
  });

  it("renders correct icon for success variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "success",
      },
      slots: {
        default: "Success",
      },
    });
    expect(wrapper.find('[data-slot="alert"]').classes()).toContain("text-success");
  });

  it("renders correct classes for default variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "default",
      },
      slots: {
        default: "Info",
      },
    });
    expect(wrapper.find('[data-slot="alert"]').classes()).toContain("text-info");
  });

  it("renders correct classes for alert variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "alert",
      },
      slots: {
        default: "Warning",
      },
    });
    expect(wrapper.find('[data-slot="alert"]').classes()).toContain("text-warning");
  });
});
