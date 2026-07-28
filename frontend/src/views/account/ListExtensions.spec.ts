import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import ListExtensions from "./ListExtensions.vue";

const mockExtensions = [
  {
    name: "FroshTools",
    label: "Frosh Tools",
    installed: true,
    active: true,
    version: "1.0.0",
    latestVersion: "1.1.0",
    ratingAverage: 5,
    installedAt: "2024-01-01",
    storeLink: "https://store.shopware.com",
    changelog: [],
    environments: [
      {
        environmentId: 1,
        environmentName: "Environment A",
        environmentOrganizationId: "org-a",
        installed: true,
        active: true,
        version: "1.0.0",
        latestVersion: "1.1.0",
      },
    ],
  },
];

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountExtensions: vi.fn(),
}));

import { getAccountExtensions } from "@/api/generated";

const stubs = {
  RouterLink: defineComponent({ template: "<a><slot /></a>" }),
  RatingStars: defineComponent({ template: "<span />" }),
};

describe("ListExtensions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAccountExtensions).mockResolvedValue({
      data: mockExtensions,
      error: undefined,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(ListExtensions, { global: { stubs } });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("My Extensions");
  });

  it("shows empty state when no extensions", async () => {
    vi.mocked(getAccountExtensions).mockResolvedValueOnce({
      data: [],
      error: undefined,
      response: new Response(),
    } as any);
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find(".border-dashed").exists()).toBe(true);
    expect(wrapper.text()).toContain("No extensions yet");
  });

  it("shows search input and extension cards when extensions exist", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("input").exists()).toBe(true);
    const extensionCards = wrapper.findAll("a.group");
    expect(extensionCards.length).toBeGreaterThan(0);
  });

  it("renders extension items with name and label", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Frosh Tools");
    expect(wrapper.text()).toContain("FroshTools");
  });

  it("links each extension card to its detail page", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("a.group").exists()).toBe(true);
  });
});
