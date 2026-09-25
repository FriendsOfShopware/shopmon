import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent } from "vue";
import EditShop from "./EditShop.vue";

const DeleteConfirmationModalStub = defineComponent({
  name: "DeleteConfirmationModal",
  props: [
    "show",
    "title",
    "entityName",
    "customConsequence",
    "reversedButtons",
    "isLoading",
    "confirmButtonText",
  ],
  template: '<div v-if="show" class="delete-modal" />',
});

const mockShop = {
  id: 1,
  name: "Test Shop",
  description: "A test shop",
  gitUrl: "https://github.com/test/repo",
  organizationId: "org-1",
};

const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { shopId: "1" }, hash: "" }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: "<a><slot /></a>",
  },
}));

vi.mock("@/api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/generated")>()),
  getAccountShops: vi.fn(),
  updateShop: vi.fn(),
  deleteShop: vi.fn(),
  getApiKeys: vi.fn(() => Promise.resolve({ data: [] })),
  getApiKeyScopes: vi.fn(() => Promise.resolve({ data: [] })),
  getPackagesTokenConfiguration: vi.fn(() =>
    Promise.resolve({ data: { configured: false, composerUrl: null } }),
  ),
  getPackagesTokens: vi.fn(() => Promise.resolve({ data: [] })),
}));

vi.mock("@/composables/useAccountEnvironments", () => ({
  fetchAccountEnvironments: vi.fn(() => Promise.resolve([])),
  useAccountEnvironments: () => ({
    environments: { value: [] },
    fetchAccountEnvironments: vi.fn(),
  }),
}));

vi.mock("@/helpers/formatter", () => ({
  formatDate: (d: string) => d,
  timeAgo: (d: string) => d,
}));

vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

import { getAccountShops, updateShop } from "@/api/generated";

describe("EditShop", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getAccountShops).mockResolvedValue({
      data: [mockShop],
      error: undefined,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(EditShop, {
      global: {
        stubs: {
          DeleteConfirmationModal: DeleteConfirmationModalStub,
        },
      },
    });
  }

  it("renders page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toContain("Edit");
  });

  it("shows loading state initially", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("Loading shop...");
  });

  it("displays shop form after loading", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("has Shop Information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Shop Information");
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="name"]').exists()).toBe(true);
  });

  it("prefills shop fields from the loaded shop and saves edits", async () => {
    vi.mocked(updateShop).mockResolvedValue({
      data: undefined,
      error: undefined,
      response: new Response(null, { status: 204 }),
    } as any);

    const wrapper = mountComponent();
    await flushPromises();

    const nameInput = wrapper.find('input[name="name"]');
    const gitInput = wrapper.find('input[name="gitUrl"]');
    const description = wrapper.find("textarea");

    expect((nameInput.element as HTMLInputElement).value).toBe("Test Shop");
    expect((gitInput.element as HTMLInputElement).value).toBe("https://github.com/test/repo");
    expect((description.element as HTMLTextAreaElement).value).toBe("A test shop");

    await nameInput.setValue("Renamed Shop");
    await gitInput.setValue("https://github.com/test/other");
    await description.setValue("Updated description");
    await wrapper.find("form").trigger("submit");

    // vee-validate debounces schema validation before calling the submit handler.
    await vi.waitFor(() => {
      expect(updateShop).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { orgId: "org-1", shopId: 1 },
          body: {
            name: "Renamed Shop",
            description: "Updated description",
            gitUrl: "https://github.com/test/other",
          },
        }),
      );
    });
  });

  it("prefills the name when description and git URL are null", async () => {
    vi.mocked(getAccountShops).mockResolvedValue({
      data: [{ ...mockShop, description: null, gitUrl: null }],
      error: undefined,
      response: new Response(),
    } as any);

    const wrapper = mountComponent();
    await flushPromises();

    expect((wrapper.find('input[name="name"]').element as HTMLInputElement).value).toBe(
      "Test Shop",
    );
    expect((wrapper.find('input[name="gitUrl"]').element as HTMLInputElement).value).toBe("");
    expect((wrapper.find("textarea").element as HTMLTextAreaElement).value).toBe("");
  });

  it("has description textarea", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('textarea[name="description"]').exists()).toBe(true);
  });

  it("has git URL input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('input[name="gitUrl"]').exists()).toBe(true);
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.find('button[type="submit"]');
    expect(btn.exists()).toBe(true);
    expect(btn.text()).toContain("Save");
  });

  it("displays API Keys section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("API Keys");
  });

  it("has Create API Key button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Create API Key"));
    expect(btn).toBeTruthy();
  });

  it("displays Danger Zone section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Danger Zone");
  });

  it("has delete shop button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const btn = wrapper.findAll("button").find((b) => b.text().includes("Delete shop"));
    expect(btn).toBeTruthy();
  });

  it("has Back to Shops link", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Back to Shops");
  });
});
