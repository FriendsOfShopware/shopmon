import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DOMWrapper, enableAutoUnmount, flushPromises, mount } from "@vue/test-utils";
import AdvisoryImpactPreview from "./AdvisoryImpactPreview.vue";
import type { AdminAdvisoryImpact } from "@/api/generated";

const api = vi.hoisted(() => ({ adminPreviewAdvisoryImpact: vi.fn() }));
vi.mock("@/api/generated", () => api);
enableAutoUnmount(afterEach);

const packages = [{ packageName: " vendor/package ", affectedVersions: " <2.0.0 " }];
const empty: AdminAdvisoryImpact = { shopCount: 0, environmentCount: 0, shops: [] };
const impact: AdminAdvisoryImpact = {
  shopCount: 1,
  environmentCount: 2,
  shops: [
    {
      shopId: 10,
      shopName: "Example shop",
      organizationId: "org-1",
      organizationName: "Example org",
      matches: [
        {
          environmentId: 20,
          environmentName: "Production",
          packageName: "vendor/package",
          installedVersion: "1.4.0",
          affectedVersions: "<2.0.0",
        },
        {
          environmentId: 21,
          environmentName: "Staging",
          packageName: "vendor/package",
          installedVersion: "1.5.0",
          affectedVersions: "<2.0.0",
        },
      ],
    },
  ],
};
const body = () => new DOMWrapper(document.body);
async function open(wrapper: ReturnType<typeof mount>) {
  await wrapper.get("button").trigger("click");
  await flushPromises();
}

describe("advisory impact preview", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    api.adminPreviewAdvisoryImpact.mockResolvedValue({ data: impact });
  });

  it("loads only on click and shows affected shops, organizations, and matching versions", async () => {
    const wrapper = mount(AdvisoryImpactPreview, { props: { packages }, attachTo: document.body });
    expect(api.adminPreviewAdvisoryImpact).not.toHaveBeenCalled();
    await open(wrapper);
    expect(api.adminPreviewAdvisoryImpact).toHaveBeenCalledExactlyOnceWith({
      body: { packages: [{ packageName: "vendor/package", affectedVersions: "<2.0.0" }] },
    });
    const dialog = body().get('[role="dialog"]');
    expect(dialog.findAll("dd").map((node) => node.text())).toEqual(["1", "2"]);
    for (const text of [
      "Example shop",
      "Example org",
      "Production",
      "Staging",
      "1.4.0",
      "1.5.0",
      "<2.0.0",
    ]) {
      expect(dialog.text()).toContain(text);
    }
    expect(dialog.findAll("a").map((a) => JSON.parse(a.attributes("href")!))).toContainEqual({
      name: "admin.environments.detail",
      params: { id: 20 },
    });
  });

  it("shows errors and supports retrying with an explicit empty result", async () => {
    api.adminPreviewAdvisoryImpact.mockRejectedValueOnce(new Error("offline"));
    const wrapper = mount(AdvisoryImpactPreview, { props: { packages }, attachTo: document.body });
    await open(wrapper);
    expect(body().get('[role="dialog"]').text()).toContain("Could not check shop impact");
    api.adminPreviewAdvisoryImpact.mockResolvedValueOnce({ data: empty });
    const retry = body()
      .findAll("button")
      .find((button) => button.text() === "Try again")!;
    await retry.trigger("click");
    await flushPromises();
    const dialog = body().get('[role="dialog"]');
    expect(dialog.text()).toContain("No shops match these package constraints.");
    expect(dialog.findAll("dd").map((node) => node.text())).toEqual(["0", "0"]);
  });

  it("displays constraint validation errors without presenting zero impact", async () => {
    api.adminPreviewAdvisoryImpact.mockResolvedValueOnce({
      error: { message: "Invalid version constraint" },
    });
    const wrapper = mount(AdvisoryImpactPreview, { props: { packages }, attachTo: document.body });
    await open(wrapper);
    const dialog = body().get('[role="dialog"]');
    expect(dialog.text()).toContain("Invalid version constraint");
    expect(dialog.find("dl").exists()).toBe(false);
  });

  it("ignores an old response after closing and checking new constraints", async () => {
    let resolve!: (result: { data: AdminAdvisoryImpact }) => void;
    api.adminPreviewAdvisoryImpact.mockReturnValueOnce(
      new Promise((done) => {
        resolve = done;
      }),
    );
    const wrapper = mount(AdvisoryImpactPreview, { props: { packages }, attachTo: document.body });
    await open(wrapper);
    expect(body().get('[role="status"]').text()).toContain("Checking shop impact");
    await body().get('[data-slot="dialog-close"]').trigger("click");
    await wrapper.setProps({
      packages: [{ packageName: "vendor/package", affectedVersions: "<1.0.0" }],
    });
    api.adminPreviewAdvisoryImpact.mockResolvedValueOnce({ data: empty });
    await open(wrapper);
    resolve({ data: impact });
    await flushPromises();
    expect(body().get('[role="dialog"]').text()).toContain(
      "No shops match these package constraints.",
    );
    expect(body().get('[role="dialog"]').text()).not.toContain("Example shop");
    expect(api.adminPreviewAdvisoryImpact).toHaveBeenLastCalledWith({
      body: { packages: [{ packageName: "vendor/package", affectedVersions: "<1.0.0" }] },
    });
  });
});
