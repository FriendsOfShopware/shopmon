import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { enableAutoUnmount, flushPromises, mount } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { ref } from "vue";
import { listAdvisories, listAdvisoryPackages, type Advisory } from "@/api/generated";
import FilterSelect from "@/components/FilterSelect.vue";
import ListAdvisories from "./ListAdvisories.vue";

const mockSession = ref<{ user: { id: string } } | null>({ user: { id: "user-1" } });
vi.mock("@/composables/useSession", () => ({ useSession: () => ({ session: mockSession }) }));

vi.mock("@/api/generated", () => ({
  listAdvisories: vi.fn(),
  listAdvisoryPackages: vi.fn(),
}));

enableAutoUnmount(afterEach);

const advisory: Advisory = {
  advisoryId: "PKSA-test",
  origin: "upstream",
  title: "Stored cross-site scripting in storefront",
  effectiveSeverity: "high",
  cvssScore: 0,
  packages: [
    {
      packageName: "shopware/core",
      affectedVersions: "<6.6.10.0",
      packagistAdvisoryId: "PKSA-test",
    },
  ],
  affectedComponents: [],
  sources: [],
  tags: [],
  isVisible: true,
  createdAt: "2026-01-01T00:00:00Z",
  syncedAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

function response(advisories: Advisory[] = [advisory], total = advisories.length) {
  return {
    data: { advisories, total, scopeCounts: { all: 40, affected: 2, suppressed: 1 } },
    response: new Response(),
  } as Awaited<ReturnType<typeof listAdvisories<false>>>;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

async function setup(query = "") {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/advisories", name: "account.advisories.list", component: ListAdvisories },
      {
        path: "/advisories/:id",
        name: "account.advisories.detail",
        component: { template: "<div />" },
      },
    ],
  });
  await router.push(`/advisories${query}`);
  await router.isReady();
  const wrapper = mount(ListAdvisories, { global: { plugins: [router] } });
  await flushPromises();
  return { wrapper, router };
}

describe("advisory listing", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockSession.value = { user: { id: "user-1" } };
    vi.mocked(listAdvisories).mockResolvedValue(response());
    vi.mocked(listAdvisoryPackages).mockResolvedValue({
      data: ["shopware/core"],
      response: new Response(),
    } as Awaited<ReturnType<typeof listAdvisoryPackages<false>>>);
  });

  it("hides shop scopes and affected sorting for anonymous visitors", async () => {
    mockSession.value = null;
    const { wrapper, router } = await setup("?scope=affected&sort=affected");
    expect(wrapper.find('[role="group"][aria-label="Advisory scope"]').exists()).toBe(false);
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 0, scope: "all", sort: "reported" },
    });
    const sortFilter = wrapper.findAllComponents(FilterSelect)[2].props() as {
      options: { value: string }[];
    };
    expect(sortFilter.options.map((option) => option.value)).not.toContain("affected");
    expect(wrapper.text()).toContain(advisory.title);
    await wrapper.get("input").setValue("storefront");
    await wrapper.get("form").trigger("submit");
    await flushPromises();
    expect(router.currentRoute.value.query).toEqual({ q: "storefront" });
  });

  it("shows shop scopes and defaults to affected for signed-in users", async () => {
    const { wrapper } = await setup();
    expect(wrapper.find('[role="group"][aria-label="Advisory scope"]').exists()).toBe(true);
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 0, scope: "affected", sort: "reported" },
    });
  });

  it("keeps Shopmon-origin advisories labelled", async () => {
    vi.mocked(listAdvisories).mockResolvedValue(response([{ ...advisory, origin: "shopmon" }]));
    const { wrapper } = await setup();
    expect(wrapper.text()).toContain("Shopmon Advisory");
  });

  it("restores shared filters and shows the result range and a zero CVSS score", async () => {
    vi.mocked(listAdvisories).mockResolvedValue(response([advisory], 26));
    const { wrapper } = await setup(
      "?scope=all&q=CVE&package=shopware/core&severity=high&sort=cvss&page=2",
    );
    expect(listAdvisories).toHaveBeenCalledExactlyOnceWith({
      query: {
        limit: 25,
        offset: 25,
        scope: "all",
        sort: "cvss",
        q: "CVE",
        package: "shopware/core",
        severity: "high",
      },
    });
    expect(wrapper.text()).toContain("Showing 26–26 of 26 advisories");
    expect(wrapper.text()).toContain("CVSS 0.0");
    expect(wrapper.get("#advisory-search").element).toHaveProperty("value", "CVE");
  });

  it("resets pagination on search and restores the previous view with browser history", async () => {
    vi.mocked(listAdvisories).mockResolvedValue(response([advisory], 40));
    const { wrapper, router } = await setup("?scope=all&page=2");
    await wrapper.get("input").setValue("  storefront  ");
    await wrapper.get("form").trigger("submit");
    await flushPromises();
    expect(router.currentRoute.value.query).toEqual({ scope: "all", q: "storefront" });
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 0, scope: "all", sort: "reported", q: "storefront" },
    });
    router.back();
    await flushPromises();
    expect(wrapper.get("input").element).toHaveProperty("value", "");
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 25, scope: "all", sort: "reported" },
    });
  });

  it("applies select filters immediately and clears them without changing scope or sort", async () => {
    const { wrapper, router } = await setup("?scope=all&sort=cvss");
    wrapper.findAllComponents(FilterSelect)[1].vm.$emit("update:modelValue", "high");
    await flushPromises();
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 0, scope: "all", sort: "cvss", severity: "high" },
    });
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "Clear filters")!
      .trigger("click");
    await flushPromises();
    expect(router.currentRoute.value.query).toEqual({ scope: "all", sort: "cvss" });
  });

  it.each(["api", "network"])(
    "shows a retryable %s failure without an all-clear message or stale cards",
    async (kind) => {
      const { wrapper, router } = await setup();
      if (kind === "api") {
        vi.mocked(listAdvisories).mockResolvedValueOnce({
          error: { message: "Unavailable" },
          response: new Response(),
        } as Awaited<ReturnType<typeof listAdvisories<false>>>);
      } else {
        vi.mocked(listAdvisories).mockRejectedValueOnce(new Error("Offline"));
      }
      await router.push("?scope=all");
      await flushPromises();
      expect(wrapper.get('[role="alert"]').text()).toContain("Failed to load advisories");
      expect(wrapper.text()).not.toContain("None of your shops are affected");
      expect(wrapper.text()).not.toContain(advisory.title);
      expect(wrapper.find("nav").exists()).toBe(false);
      await wrapper.get('[role="alert"] button').trigger("click");
      await flushPromises();
      expect(wrapper.text()).toContain(advisory.title);
      expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    },
  );

  it("ignores an older response that finishes after a newer request", async () => {
    const older = deferred<Awaited<ReturnType<typeof listAdvisories<false>>>>();
    vi.mocked(listAdvisories).mockReturnValueOnce(older.promise);
    const { wrapper, router } = await setup();
    await router.push("?scope=all");
    await flushPromises();
    older.resolve(response([{ ...advisory, title: "Stale advisory" }], 100));
    await flushPromises();
    expect(wrapper.text()).toContain(advisory.title);
    expect(wrapper.text()).not.toContain("Stale advisory");
    expect(wrapper.text()).toContain("Showing 1–1 of 1 advisories");
  });

  it("keeps loading until the latest request finishes", async () => {
    const older = deferred<Awaited<ReturnType<typeof listAdvisories<false>>>>();
    const newer = deferred<Awaited<ReturnType<typeof listAdvisories<false>>>>();
    vi.mocked(listAdvisories).mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise);
    const { wrapper, router } = await setup();
    await router.push("?scope=all");
    await flushPromises();
    older.resolve(response());
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Loading advisories…");
    newer.resolve(response());
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("1 advisory");
  });

  it("loads advisories without waiting for package suggestions", async () => {
    const packages = deferred<Awaited<ReturnType<typeof listAdvisoryPackages<false>>>>();
    vi.mocked(listAdvisoryPackages).mockReturnValueOnce(packages.promise);
    const { wrapper } = await setup();
    expect(wrapper.text()).toContain(advisory.title);
    packages.resolve({ data: [], error: undefined, response: new Response() } as Awaited<
      ReturnType<typeof listAdvisoryPackages<false>>
    >);
    await flushPromises();
  });

  it("recovers an out-of-range bookmarked page", async () => {
    vi.mocked(listAdvisories)
      .mockResolvedValueOnce(response([], 26))
      .mockResolvedValue(response([advisory], 26));
    const { wrapper, router } = await setup("?scope=all&page=99");
    expect(router.currentRoute.value.query.page).toBe("2");
    expect(listAdvisories).toHaveBeenLastCalledWith({
      query: { limit: 25, offset: 25, scope: "all", sort: "reported" },
    });
    expect(wrapper.text()).toContain("Showing 26–26 of 26 advisories");
  });

  it("does not treat an unsubmitted search draft as an active filter", async () => {
    vi.mocked(listAdvisories).mockResolvedValue(response([]));
    const { wrapper } = await setup();
    await wrapper.get("input").setValue("draft");
    expect(wrapper.text()).toContain("None of your shops are affected");
    expect(wrapper.text()).not.toContain("No advisories match your filters");
    expect(listAdvisories).toHaveBeenCalledTimes(1);
  });
});
