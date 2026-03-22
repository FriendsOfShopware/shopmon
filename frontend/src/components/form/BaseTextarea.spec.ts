import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import BaseTextarea from "./BaseTextarea.vue";

describe("BaseTextarea", () => {
  it("renders a textarea element", () => {
    const wrapper = mount(BaseTextarea);
    expect(wrapper.find("textarea").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mount(BaseTextarea, { props: { label: "Description", name: "desc" } });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Description");
    expect(label.attributes("for")).toBe("desc");
  });

  it("does not render label when not provided", () => {
    const wrapper = mount(BaseTextarea);
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("sets textarea attributes from props", () => {
    const wrapper = mount(BaseTextarea, {
      props: { name: "desc", placeholder: "Enter description", rows: 6 },
    });
    const textarea = wrapper.find("textarea");
    expect(textarea.attributes("name")).toBe("desc");
    expect(textarea.attributes("placeholder")).toBe("Enter description");
    expect(textarea.attributes("rows")).toBe("6");
  });

  it("defaults rows to 4", () => {
    const wrapper = mount(BaseTextarea);
    expect(wrapper.find("textarea").attributes("rows")).toBe("4");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mount(BaseTextarea, { props: { error: "Required" } });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Required");
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mount(BaseTextarea, { props: { error: "Required" } });
    expect(wrapper.find("textarea").classes()).toContain("has-error");
  });

  it("emits update:modelValue on input", async () => {
    const wrapper = mount(BaseTextarea);
    await wrapper.find("textarea").setValue("test content");
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["test content"]);
  });

  it("sets textarea value from modelValue prop", () => {
    const wrapper = mount(BaseTextarea, { props: { modelValue: "hello" } });
    expect((wrapper.find("textarea").element as HTMLTextAreaElement).value).toBe("hello");
  });
});
