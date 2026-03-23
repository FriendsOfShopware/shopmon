import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import ListOrganizations from "./ListOrganizations.vue";

vi.mock("@/helpers/api", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
  },
  getToken: vi.fn(() => "test-token"),
  setToken: vi.fn(),
}));

vi.mock("@/composables/useSession", () => ({
  useSession: () => ({
    session: { value: { user: { id: "1", name: "Test" } } },
    loading: { value: false },
    fetchSession: vi.fn(),
  }),
  fetchSession: vi.fn(),
}));

import { api } from "@/helpers/api";

const HeaderContainerStub = defineComponent({
  name: "HeaderContainer",
  props: ["title"],
  setup(props, { slots }) {
    return () => h("header", {}, [props.title, slots.default?.()]);
  },
});

const MainContainerStub = defineComponent({
  name: "MainContainer",
  setup(_, { slots }) {
    return () => h("main", {}, slots.default?.());
  },
});

const ElementEmptyStub = defineComponent({
  name: "ElementEmpty",
  props: ["title", "route", "button"],
  template:
    '<div class="element-empty"><h3>{{ title }}</h3><slot /><a :href="route?.name">{{ button }}</a></div>',
});

const RouterLinkStub = defineComponent({
  name: "RouterLink",
  props: ["to"],
  setup(props, { slots }) {
    return () => h("a", { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

const mockOrganizations = [
  { id: "1", name: "Test Organization", role: "owner", createdAt: "2024-01-01" },
  { id: "2", name: "Another Org", role: "member", createdAt: "2024-02-01" },
];

describe("ListOrganizations", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.GET).mockImplementation(((path: string) => {
      if (path === "/auth/list-organizations") {
        return Promise.resolve({ data: mockOrganizations, error: null, response: new Response() });
      }
      return Promise.resolve({
        data: null,
        error: { message: "not found" },
        response: new Response(),
      });
    }) as any);
  });

  function mountComponent() {
    return mount(ListOrganizations, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          ElementEmpty: ElementEmptyStub,
          RouterLink: RouterLinkStub,
          Panel: { template: "<div><slot /><slot name='title' /></div>" },
        },
      },
    });
  }

  it("renders successfully", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it("displays page title", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("My Organization");
  });

  it("displays Add Organization button", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Organization");
  });

  it("displays organizations when data exists", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Organization");
    expect(wrapper.text()).toContain("Another Org");
  });

  it("displays organization names", async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Test Organization");
    expect(wrapper.text()).toContain("Another Org");
  });

  it("displays empty state when no organizations exist", async () => {
    vi.mocked(api.GET).mockImplementation((() =>
      Promise.resolve({ data: [], error: null, response: new Response() })) as any);

    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find(".element-empty").exists()).toBe(true);
  });

  it("shows add organization button in empty state", async () => {
    vi.mocked(api.GET).mockImplementation((() =>
      Promise.resolve({ data: [], error: null, response: new Response() })) as any);

    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain("Add Organization");
  });
});
