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

  it("renders deployment and sponsor content", () => {
    const wrapper = mount(WhatsNewModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          Modal: {
            props: ["show"],
            template:
              '<div class="modal-stub" :data-show="String(show)"><slot name="title" /><slot name="content" /><slot name="footer" /></div>',
          },
          RouterLink: {
            props: ["to"],
            template: "<a><slot /></a>",
          },
        },
      },
    });

    expect(wrapper.find(".modal-stub").attributes("data-show")).toBe("true");
    expect(wrapper.text()).toContain("What's new: Deployment tracking");
    expect(wrapper.text()).toContain("Track deployments directly inside each shop");
    expect(wrapper.text()).toContain("Acme Commerce");
    expect(wrapper.text()).toContain("New sponsors on the start page");
  });

  it("emits close when the primary button is clicked", async () => {
    const wrapper = mount(WhatsNewModal, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          Modal: {
            props: ["show"],
            template:
              '<div class="modal-stub" :data-show="String(show)"><slot name="title" /><slot name="content" /><slot name="footer" /></div>',
          },
          RouterLink: {
            props: ["to"],
            template: "<a><slot /></a>",
          },
        },
      },
    });

    await wrapper.find("button").trigger("click");

    expect(wrapper.emitted("close")).toHaveLength(1);
  });
});
