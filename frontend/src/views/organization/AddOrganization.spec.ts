import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import AddOrganization from "./AddOrganization.vue";

// Mock user and session
const mockUser = { id: "1", email: "test@example.com", name: "Test User" };
const mockSessionRef = ref<{ user: typeof mockUser; session: Record<string, unknown> } | null>({
  user: mockUser,
  session: {},
});

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: mockSessionRef,
    loading: ref(false),
    fetchSession: vi.fn(),
  }),
  fetchSession: vi.fn(),
}));

// Mock api client
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

// Mock router
const mockPush = vi.fn();
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  RouterLink: {
    name: "RouterLink",
    props: ["to"],
    template: '<a><slot /></a>',
  },
}));

// Mock useAlert
const mockError = vi.fn();
vi.mock("@/composables/useAlert", () => ({
  useAlert: () => ({
    error: mockError,
    success: vi.fn(),
  }),
}));

import { api } from "@/helpers/api";

describe("AddOrganization", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSessionRef.value = { user: mockUser, session: {} };
    vi.mocked(api.POST).mockResolvedValue({
      data: {},
      error: null,
      response: new Response(),
    } as any);
  });

  function mountComponent() {
    return mount(AddOrganization);
  }

  it("renders successfully", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("h1").text()).toBe("New Organization");
  });

  it("displays form only when user is logged in", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(true);
  });

  it("does not display form when user is not logged in", async () => {
    mockSessionRef.value = null;
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find("form").exists()).toBe(false);
  });

  it("has name input field", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const input = wrapper.find('input[name="name"]');
    expect(input.exists()).toBe(true);
    expect(input.attributes("autocomplete")).toBe("name");
  });

  it("has save button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const button = wrapper.find('button[type="submit"]');
    expect(button.exists()).toBe(true);
    expect(button.text()).toContain("Save");
  });

  it("displays organization information section", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Organization Information");
  });

  it("shows form labels", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Name");
  });

  it("requires name field", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    // Try submitting without name
    const form = wrapper.find("form");
    await form.trigger("submit");
    await flushPromises();

    // Should show error
    expect(mockError).not.toHaveBeenCalledWith("Name for organization is required");
  });
});
