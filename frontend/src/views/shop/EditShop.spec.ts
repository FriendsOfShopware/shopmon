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
    template: '<a><slot /></a>',
  },
}));

vi.mock("@/helpers/api", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
    PATCH: vi.fn(),
    DELETE: vi.fn(),
    PUT: vi.fn(),
  },
  setToken: vi.fn(),
  getToken: vi.fn(),
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

import { api } from "@/helpers/api";

describe("EditShop", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/account/shops") {
        return Promise.resolve({ data: [mockShop], error: null, response: new Response() });
      }
      if (path === "/api-key-scopes") {
        return Promise.resolve({ data: [], error: null, response: new Response() });
      }
      if (path === "/packages-token/configuration") {
        return Promise.resolve({
          data: { configured: false, composerUrl: null },
          error: null,
          response: new Response(),
        });
      }
      return Promise.resolve({ data: null, error: null, response: new Response() });
    }) as any);
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
