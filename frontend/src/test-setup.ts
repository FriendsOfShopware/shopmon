import { config } from "@vue/test-utils";
import { createI18n } from "vue-i18n";
import en from "./locales/en.json";

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackLocale: "en",
  messages: { en },
});

// jsdom does not implement HTMLDialogElement.showModal / .close
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
