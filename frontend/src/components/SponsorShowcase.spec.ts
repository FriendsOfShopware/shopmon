import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import SponsorShowcase from "./SponsorShowcase.vue";

const sponsors = [
  {
    name: "Acme Commerce",
    url: "https://example.com/acme",
    description: "Supporting Shopmon development.",
  },
];

describe("SponsorShowcase", () => {
  it("renders sponsor cards", () => {
    const wrapper = mount(SponsorShowcase, {
      props: {
        sponsors,
        title: "Sponsors",
      },
    });

    expect(wrapper.text()).toContain("Sponsors");
    expect(wrapper.text()).toContain("Acme Commerce");
    expect(wrapper.text()).toContain("Supporting Shopmon development.");
    expect(wrapper.find("a").attributes("href")).toBe("https://example.com/acme");
  });

  it("renders sponsor descriptions in compact mode", () => {
    const wrapper = mount(SponsorShowcase, {
      props: {
        sponsors,
        compact: true,
      },
    });

    expect(wrapper.text()).toContain("Supporting Shopmon development.");
  });

  it("renders nothing without sponsors", () => {
    const wrapper = mount(SponsorShowcase, {
      props: {
        sponsors: [],
      },
    });

    expect(wrapper.find("section").exists()).toBe(false);
  });
});
