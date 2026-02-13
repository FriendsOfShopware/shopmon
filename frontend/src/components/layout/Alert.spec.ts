import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Alert from "./Alert.vue";

describe("Alert", () => {
  it("renders successfully with error type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "error",
      },
      slots: {
        default: "Error message",
      },
    });
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.find(".alert-error").exists()).toBe(true);
  });

  it("renders successfully with success type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "success",
      },
      slots: {
        default: "Success message",
      },
    });
    expect(wrapper.find(".alert-success").exists()).toBe(true);
  });

  it("renders successfully with info type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "info",
      },
      slots: {
        default: "Info message",
      },
    });
    expect(wrapper.find(".alert-info").exists()).toBe(true);
  });

  it("renders successfully with warning type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "warning",
      },
      slots: {
        default: "Warning message",
      },
    });
    expect(wrapper.find(".alert-warning").exists()).toBe(true);
  });

  it("displays slot content correctly", () => {
    const message = "This is an important alert";
    const wrapper = mount(Alert, {
      props: {
        type: "info",
      },
      slots: {
        default: message,
      },
    });
    expect(wrapper.find(".alert-content").text()).toBe(message);
  });

  it("applies correct icon class for error type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "error",
      },
      slots: {
        default: "Error",
      },
    });
    expect(wrapper.find(".icon-error").exists()).toBe(true);
  });

  it("applies correct icon class for success type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "success",
      },
      slots: {
        default: "Success",
      },
    });
    expect(wrapper.find(".icon-success").exists()).toBe(true);
  });

  it("applies correct icon class for info type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "info",
      },
      slots: {
        default: "Info",
      },
    });
    expect(wrapper.find(".icon-info").exists()).toBe(true);
  });

  it("applies correct icon class for warning type", () => {
    const wrapper = mount(Alert, {
      props: {
        type: "warning",
      },
      slots: {
        default: "Warning",
      },
    });
    expect(wrapper.find(".icon-warning").exists()).toBe(true);
  });
});
