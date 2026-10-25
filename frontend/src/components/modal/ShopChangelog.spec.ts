import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import en from "@/locales/en.json";
import de from "@/locales/de.json";
import type { AccountChangelog } from "@/api/generated";
import ShopChangelog from "./ShopChangelog.vue";

describe("ShopChangelog", () => {
  describe.each(["en", "de"])("%s", (locale) => {
    it.each(["oldShopwareVersion", "newShopwareVersion"] as const)(
      "renders %s as literal text",
      (field) => {
        const payload = '<img src=x onerror="alert(document.domain)"><svg onload="alert(1)"></svg>';
        const changelog: AccountChangelog = {
          id: 1,
          date: "2026-10-25T00:00:00Z",
          environmentId: 1,
          environmentName: "Production",
          environmentOrganizationId: "org-1",
          environmentOrganizationName: "Organization",
          environmentShopName: "Shop",
          extensions: [],
          oldShopwareVersion: "6.6.0.0",
          newShopwareVersion: "6.7.0.0",
          [field]: payload,
        };
        const i18n = createI18n({
          legacy: false,
          locale,
          messages: { en, de },
        });
        const wrapper = mount(ShopChangelog, {
          props: { show: true, changelog },
          global: {
            plugins: [i18n],
            stubs: {
              Dialog: { template: "<div><slot /></div>" },
              DialogContent: { template: "<div><slot /></div>" },
              DialogHeader: { template: "<div><slot /></div>" },
              DialogTitle: { template: "<h2><slot /></h2>" },
            },
          },
        });

        const update = wrapper.get("span");
        expect(update.text()).toBe(
          i18n.global.t("changelogModal.shopwareUpdate", {
            oldVersion: changelog.oldShopwareVersion,
            newVersion: changelog.newShopwareVersion,
          }),
        );
        expect(update.text()).toContain(payload);
        expect(update.element.children).toHaveLength(0);
        expect(wrapper.find("img, [onerror], [onload]").exists()).toBe(false);
        wrapper.unmount();
      },
    );
  });
});
