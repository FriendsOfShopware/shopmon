import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import WhatsNewModal from "./WhatsNewModal.vue";

vi.mock("@/data/sponsors", () => ({
  sponsors: [
    {
      name: "Acme Commerce",
      url: "https://example.com/acme",
      description: "Supporting Shopmon development.",
    },
  ],
}));

describe("WhatsNewModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders packages mirror and sponsor content", () => {
    const wrapper = mount(WhatsNewModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          SponsorShowcase: {
            props: ["sponsors", "compact"],
            template: '<div class="sponsor-stub">{{ sponsors.map(s => s.name).join(", ") }}</div>',
          },
        },
      },
    });

    expect(wrapper.text()).toContain("What's new: Packages Mirror");
    expect(wrapper.text()).toContain("75x faster Shopware store packages via Global CDN");
    expect(wrapper.text()).toContain("Acme Commerce");
  });

  it("emits close when the close button is clicked", async () => {
    const wrapper = mount(WhatsNewModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          SponsorShowcase: true,
        },
      },
    });

    await wrapper.find("button.ui-button--primary").trigger("click");

    expect(wrapper.emitted("close")).toHaveLength(1);
  });
});
