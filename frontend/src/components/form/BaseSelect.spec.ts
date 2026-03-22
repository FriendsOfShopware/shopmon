import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import BaseSelect from "./BaseSelect.vue";

describe("BaseSelect", () => {
  it("renders a select element", () => {
    const wrapper = mount(BaseSelect);
    expect(wrapper.find("select").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mount(BaseSelect, { props: { label: "Role", name: "role" } });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Role");
    expect(label.attributes("for")).toBe("role");
  });

  it("does not render label when not provided", () => {
    const wrapper = mount(BaseSelect);
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("renders slot content as options", () => {
    const wrapper = mount(BaseSelect, {
      slots: {
        default: '<option value="a">Option A</option><option value="b">Option B</option>',
      },
    });
    const options = wrapper.findAll("option");
    expect(options).toHaveLength(2);
    expect(options[0].text()).toBe("Option A");
    expect(options[1].text()).toBe("Option B");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mount(BaseSelect, { props: { error: "Required" } });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Required");
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mount(BaseSelect, { props: { error: "Required" } });
    expect(wrapper.find("select").classes()).toContain("has-error");
  });

  it("emits update:modelValue on change", async () => {
    const wrapper = mount(BaseSelect, {
      slots: {
        default: '<option value="a">A</option><option value="b">B</option>',
      },
    });
    await wrapper.find("select").setValue("b");
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["b"]);
  });

  it("passes through extra attributes to select", () => {
    const wrapper = mount(BaseSelect, { attrs: { required: true } });
    expect(wrapper.find("select").attributes("required")).toBeDefined();
  });
});
