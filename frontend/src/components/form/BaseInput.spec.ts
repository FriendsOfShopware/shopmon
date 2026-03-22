import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import BaseInput from "./BaseInput.vue";

describe("BaseInput", () => {
  it("renders an input element", () => {
    const wrapper = mount(BaseInput);
    expect(wrapper.find("input").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mount(BaseInput, { props: { label: "Email", name: "email" } });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Email");
    expect(label.attributes("for")).toBe("email");
  });

  it("does not render label when not provided", () => {
    const wrapper = mount(BaseInput);
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("sets input attributes from props", () => {
    const wrapper = mount(BaseInput, {
      props: {
        name: "email",
        type: "email",
        placeholder: "Enter email",
        autocomplete: "email",
      },
    });
    const input = wrapper.find("input");
    expect(input.attributes("name")).toBe("email");
    expect(input.attributes("type")).toBe("email");
    expect(input.attributes("placeholder")).toBe("Enter email");
    expect(input.attributes("autocomplete")).toBe("email");
  });

  it("defaults type to text", () => {
    const wrapper = mount(BaseInput);
    expect(wrapper.find("input").attributes("type")).toBe("text");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mount(BaseInput, { props: { error: "Required" } });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Required");
  });

  it("does not show error message when no error", () => {
    const wrapper = mount(BaseInput);
    expect(wrapper.find(".field-error-message").exists()).toBe(false);
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mount(BaseInput, { props: { error: "Required" } });
    expect(wrapper.find("input").classes()).toContain("has-error");
  });

  it("emits update:modelValue on input", async () => {
    const wrapper = mount(BaseInput);
    await wrapper.find("input").setValue("test");
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["test"]);
  });

  it("sets input value from modelValue prop", () => {
    const wrapper = mount(BaseInput, { props: { modelValue: "hello" } });
    expect((wrapper.find("input").element as HTMLInputElement).value).toBe("hello");
  });

  it("uses id prop over name for input id", () => {
    const wrapper = mount(BaseInput, { props: { id: "custom-id", name: "field" } });
    expect(wrapper.find("input").attributes("id")).toBe("custom-id");
  });

  it("applies field-readonly class when readonly attribute is set", () => {
    const wrapper = mount(BaseInput, { attrs: { readonly: true } });
    expect(wrapper.find("input").classes()).toContain("field-readonly");
  });

  it("renders append slot with input-group wrapper", () => {
    const wrapper = mount(BaseInput, {
      slots: { append: "<button>Go</button>" },
    });
    expect(wrapper.find(".input-group").exists()).toBe(true);
    expect(wrapper.find(".input-group button").text()).toBe("Go");
  });

  it("does not render input-group wrapper without append slot", () => {
    const wrapper = mount(BaseInput);
    expect(wrapper.find(".input-group").exists()).toBe(false);
  });

  it("passes through extra attributes to input", () => {
    const wrapper = mount(BaseInput, { attrs: { required: true, maxlength: "10" } });
    const input = wrapper.find("input");
    expect(input.attributes("required")).toBeDefined();
    expect(input.attributes("maxlength")).toBe("10");
  });
});
