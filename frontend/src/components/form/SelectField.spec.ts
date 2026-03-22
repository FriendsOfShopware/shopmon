import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Form } from "vee-validate";
import SelectField from "./SelectField.vue";

function mountField(props: Record<string, unknown> = {}, slotContent = "") {
  const options =
    slotContent || '<option value="a">Option A</option><option value="b">Option B</option>';
  return mount(
    {
      components: { SelectField },
      template: `<Form @submit="() => {}"><SelectField v-bind="props">${options}</SelectField></Form>`,
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

describe("SelectField", () => {
  it("renders a select inside a vee-validate form", () => {
    const wrapper = mountField();
    expect(wrapper.find("select").exists()).toBe(true);
  });

  it("renders label when provided", () => {
    const wrapper = mountField({ label: "Role" });
    const label = wrapper.find("label");
    expect(label.exists()).toBe(true);
    expect(label.text()).toBe("Role");
  });

  it("does not render label when not provided", () => {
    const wrapper = mountField();
    expect(wrapper.find("label").exists()).toBe(false);
  });

  it("renders options from slot", () => {
    const wrapper = mountField();
    const options = wrapper.findAll("option");
    expect(options).toHaveLength(2);
    expect(options[0].text()).toBe("Option A");
    expect(options[1].text()).toBe("Option B");
  });

  it("shows error message when error prop is set", () => {
    const wrapper = mountField({ error: "Required" });
    const error = wrapper.find(".field-error-message");
    expect(error.exists()).toBe(true);
    expect(error.text()).toBe("Required");
  });

  it("applies has-error class when error is set", () => {
    const wrapper = mountField({ error: "Required" });
    expect(wrapper.find("select").classes()).toContain("has-error");
  });
});
