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
    expect(wrapper.find(".banner-error").exists()).toBe(true);
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
    expect(wrapper.find(".banner-success").exists()).toBe(true);
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
    expect(wrapper.find(".banner-default").exists()).toBe(true);
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
    expect(wrapper.find(".banner-alert").exists()).toBe(true);
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
    expect(wrapper.find(".banner-content").text()).toBe(message);
  });

  it("applies correct icon class for error variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "error",
      },
      slots: {
        default: "Error",
      },
    });
    expect(wrapper.find(".icon-error").exists()).toBe(true);
  });

  it("applies correct icon class for success variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "success",
      },
      slots: {
        default: "Success",
      },
    });
    expect(wrapper.find(".icon-success").exists()).toBe(true);
  });

  it("applies correct icon class for default variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "default",
      },
      slots: {
        default: "Info",
      },
    });
    expect(wrapper.find(".icon-default").exists()).toBe(true);
  });

  it("applies correct icon class for alert variant", () => {
    const wrapper = mount(Banner, {
      props: {
        variant: "alert",
      },
      slots: {
        default: "Warning",
      },
    });
    expect(wrapper.find(".icon-alert").exists()).toBe(true);
  });
});
