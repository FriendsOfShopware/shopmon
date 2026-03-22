import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Form } from "vee-validate";
import TextareaField from "./TextareaField.vue";

function mountField(props: Record<string, unknown> = {}) {
  return mount(
    {
      components: { TextareaField },
      template: `<Form @submit="() => {}"><TextareaField v-bind="props" /></Form>`,
      setup() {
        return { props: { name: "test", ...props } };
      },
    },
    {
      global: {
        components: { Form },
      },
    },
  );
}

describe("TextareaField", () => {
  it("renders a textarea inside a vee-validate form", () => {
    const wrapper = mountField();
    expect(wrapper.find("textarea").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mountField({ label: "Description" });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Description");
  });

  it("does not render label when not provided", () => {
    const wrapper = mountField();
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("sets placeholder from prop", () => {
    const wrapper = mountField({ placeholder: "Enter description" });
    expect(wrapper.find("textarea").attributes("placeholder")).toBe("Enter description");
  });

  it("sets rows from prop", () => {
    const wrapper = mountField({ rows: 8 });
    expect(wrapper.find("textarea").attributes("rows")).toBe("8");
  });

  it("defaults rows to 4", () => {
    const wrapper = mountField();
    expect(wrapper.find("textarea").attributes("rows")).toBe("4");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mountField({ error: "Required" });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Required");
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mountField({ error: "Required" });
    expect(wrapper.find("textarea").classes()).toContain("has-error");
  });
});
