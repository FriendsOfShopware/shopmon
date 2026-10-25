import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import en from "@/locales/en.json";
import de from "@/locales/de.json";
import DeleteConfirmationModal from "./DeleteConfirmationModal.vue";

describe("DeleteConfirmationModal", () => {
  const payloads = [
    '<img src=x onerror="alert(document.domain)">',
    '<svg onload="alert(document.domain)"></svg>',
    "</p><script>alert(document.domain)</script><p>",
    'Shop & Co. <b>"quoted"</b> {name}',
  ];

  describe.each(["en", "de"])("%s", (locale) => {
    it.each(payloads)("renders the entity name as literal text: %s", (entityName) => {
      const i18n = createI18n({
        legacy: false,
        locale,
        messages: { en, de },
      });
      const wrapper = mount(DeleteConfirmationModal, {
        props: { show: true, title: "Delete entity", entityName },
        global: {
          plugins: [i18n],
          stubs: {
            Dialog: { template: "<div><slot /></div>" },
            DialogContent: { template: "<div><slot /></div>" },
            DialogHeader: { template: "<div><slot /></div>" },
            DialogTitle: { template: "<h2><slot /></h2>" },
            DialogFooter: { template: "<div><slot /></div>" },
          },
        },
      });

      const paragraphs = wrapper.findAll("p");
      expect(paragraphs[0]!.text()).toBe(
        i18n.global.t("deleteModal.confirmText", { name: entityName }),
      );
      expect(paragraphs[0]!.text()).toContain(entityName);
      // No elements may be created from the name, even if jsdom does not execute handlers.
      expect(paragraphs[0]!.element.children).toHaveLength(0);
      expect(wrapper.find("img, script, [onerror], [onload]").exists()).toBe(false);
      expect(paragraphs[1]!.text()).toBe(i18n.global.t("deleteModal.consequence"));
      wrapper.unmount();
    });
  });
});
