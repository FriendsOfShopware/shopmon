import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Form } from "vee-validate";
import InputField from "./InputField.vue";

function mountInForm(props: Record<string, unknown> = {}, slots: Record<string, string> = {}) {
  return mount(Form, {
    props: { onSubmit: () => {} },
    slots: {
      default: () => mount(InputField, { props: { name: "test", ...props }, slots }).vm.$el,
    },
  });
}

function mountField(props: Record<string, unknown> = {}, slots: Record<string, string> = {}) {
  return mount(
    {
      components: { InputField },
      template: `<Form @submit="() => {}"><InputField v-bind="props" /></Form>`,
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

describe("InputField", () => {
  it("renders an input inside a vee-validate form", () => {
    const wrapper = mountField();
    expect(wrapper.find("input").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mountField({ label: "Email" });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Email");
  });

  it("does not render label when not provided", () => {
    const wrapper = mountField();
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("sets input type from prop", () => {
    const wrapper = mountField({ type: "email" });
    expect(wrapper.find("input").attributes("type")).toBe("email");
  });

  it("sets placeholder from prop", () => {
    const wrapper = mountField({ placeholder: "Enter value" });
    expect(wrapper.find("input").attributes("placeholder")).toBe("Enter value");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mountField({ error: "Field is required" });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Field is required");
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mountField({ error: "Required" });
    expect(wrapper.find("input").classes()).toContain("has-error");
  });

  it("renders append slot", () => {
    const wrapper = mount(
      {
        components: { InputField },
        template: `<Form @submit="() => {}"><InputField name="test"><template #append><button class="test-btn">Go</button></template></InputField></Form>`,
      },
      { global: { components: { Form } } },
    );
    expect(wrapper.find(".input-group").exists()).toBe(true);
    expect(wrapper.find(".test-btn").text()).toBe("Go");
  });
});
