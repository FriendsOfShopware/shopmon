import { config } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import en from "./locales/en.json";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackLocale: "en",
  messages: { en },
  // Some locale messages intentionally contain HTML (rendered via v-html in
  // e.g. Docs.vue). Suppress vue-i18n's per-render HTML warning in tests.
  warnHtmlMessage: false,
});

// jsdom does not implement HTMLDialogElement.showModal / .close.
// The `??=` reads the existing prototype method, which oxlint flags as an
// unbound-method reference; that read is intentional here (polyfill guard).
/* oxlint-disable typescript/unbound-method */
if (typeof HTMLDialogElement !== "undefined") {
  HTMLDialogElement.prototype.showModal ??= function () {
    this.setAttribute("open", "");
  };
  HTMLDialogElement.prototype.close ??= function () {
    this.removeAttribute("open");
  };
} else {
  // Fallback when jsdom exposes <dialog> as a generic HTMLElement
  //@ts-expect-error
  HTMLElement.prototype.showModal ??= function () {
    this.setAttribute("open", "");
  };
}
/* oxlint-enable typescript/unbound-method */

// Register i18n plugin globally for tests
config.global.plugins = config.global.plugins || [];
config.global.plugins.push(i18n as any);

// Globally stub router-link/router-view to prevent Vue warnings.
// Components use both kebab-case (<router-link>) and PascalCase (<RouterLink>),
// and Vue resolves stub keys case-sensitively, so register every variant.
const routerLinkStub = {
  props: ["to"],
  template: "<a :href=\"typeof to === 'string' ? to : JSON.stringify(to)\"><slot /></a>",
};
const routerViewStub = {
  template: "<div><slot /></div>",
};
config.global.stubs = {
  "router-link": routerLinkStub,
  RouterLink: routerLinkStub,
  "router-view": routerViewStub,
  RouterView: routerViewStub,
};
