import { beforeEach, describe, expect, it, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
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

// Stub Dialog components to avoid Teleport issues in tests
const DialogStub = defineComponent({
  props: ["open"],
  setup(props, { slots }) {
    return () => (props.open ? h("div", { class: "dialog" }, slots.default?.()) : null);
  },
});

const DialogContentStub = defineComponent({
  setup(_, { slots }) {
    return () => h("div", { class: "dialog-content" }, slots.default?.());
  },
});

const DialogHeaderStub = defineComponent({
  setup(_, { slots }) {
    return () => h("div", { class: "dialog-header" }, slots.default?.());
  },
});

const DialogTitleStub = defineComponent({
  setup(_, { slots }) {
    return () => h("h2", {}, slots.default?.());
  },
});

const DialogFooterStub = defineComponent({
  setup(_, { slots }) {
    return () => h("div", { class: "dialog-footer" }, slots.default?.());
  },
});

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
          Dialog: DialogStub,
          DialogContent: DialogContentStub,
          DialogHeader: DialogHeaderStub,
          DialogTitle: DialogTitleStub,
          DialogFooter: DialogFooterStub,
          SponsorShowcase: {
            props: ["sponsors", "compact"],
            template: '<div class="sponsor-stub">{{ sponsors.map(s => s.name).join(", ") }}</div>',
          },
          UiButton: {
            props: ["to", "variant", "type"],
            emits: ["click"],
            template: '<button :data-variant="variant" @click="$emit(\'click\', $event)"><slot /></button>',
          },
        },
      },
    });

    expect(wrapper.text()).toContain("New in March 2026");
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
          Dialog: DialogStub,
          DialogContent: DialogContentStub,
          DialogHeader: DialogHeaderStub,
          DialogTitle: DialogTitleStub,
          DialogFooter: DialogFooterStub,
          SponsorShowcase: true,
          UiButton: {
            props: ["to", "variant", "type"],
            emits: ["click"],
            template: '<button :data-variant="variant" @click="$emit(\'click\', $event)"><slot /></button>',
          },
        },
      },
    });

    // The close button has variant="primary" and text "Close"
    const buttons = wrapper.findAll("button");
    const closeButton = buttons.find((b) => b.text().includes("Close"));
    expect(closeButton).toBeTruthy();
    await closeButton!.trigger("click");

    expect(wrapper.emitted("close")).toHaveLength(1);
  });
});
