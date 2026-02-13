import { config } from "@vue/test-utils";

// Globally stub router-link to prevent Vue warnings
config.global.stubs = {
  "router-link": {
    props: ["to"],
    template: '<a :href="to"><slot /></a>',
  },
};
