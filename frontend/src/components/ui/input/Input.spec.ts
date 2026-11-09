import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import { describe, expect, it } from "vitest";
import Input from "./Input.vue";
import Textarea from "../textarea/Textarea.vue";

for (const component of [Input, Textarea]) {
  describe(component === Input ? "Input" : "Textarea", () => {
    it("waits for IME composition to finish before updating the model", async () => {
      const wrapper = mount(component, { props: { modelValue: "" } });
      const field = wrapper.get("input, textarea");

      await field.trigger("compositionstart");
      (field.element as HTMLInputElement | HTMLTextAreaElement).value = "に";
      await field.trigger("input");
      expect(wrapper.emitted("update:modelValue")).toBeUndefined();

      (field.element as HTMLInputElement | HTMLTextAreaElement).value = "日本";
      await field.trigger("compositionend");
      await nextTick();
      expect(wrapper.emitted("update:modelValue")).toEqual([["日本"]]);
      wrapper.unmount();
    });
  });
}
