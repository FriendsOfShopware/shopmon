import { config } from "@vue/test-utils";

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
  HTMLElement.prototype.showModal ??= function () {
    this.setAttribute("open", "");
  };
}

// Globally stub router-link to prevent Vue warnings
config.global.stubs = {
  "router-link": {
    props: ["to"],
    template: '<a :href="to"><slot /></a>',
  },
};
