import { beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import type { AdminAdvisory, GhsaAdvisoryPreview } from "@/api/generated";
import CreateAdvisory from "./CreateAdvisory.vue";
import DetailAdvisory from "./DetailAdvisory.vue";

const api = vi.hoisted(() => ({
  adminGetAdvisory: vi.fn(),
  adminUpdateAdvisory: vi.fn(),
  adminLinkAdvisoryOfficial: vi.fn(),
  adminPreviewGhsaImport: vi.fn(),
  adminImportGhsaAdvisory: vi.fn(),
  adminCreateAdvisory: vi.fn(),
}));
vi.mock("@/api/generated", () => api);
vi.mock("vue-router", () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => ({ params: { id: "SHOPMON-2026-0001" } }),
}));

const cwes = [{ id: "CWE-79", name: "Cross-site Scripting" }];
const advisory: AdminAdvisory = {
  advisoryId: "SHOPMON-2026-0001",
  impact: { shopCount: 2, environmentCount: 3, activeShopCount: 1, suppressedEnvironmentCount: 2 },
  title: "Original title",
  origin: "shopmon",
  isVisible: false,
  effectiveSeverity: "high",
  affectedComponents: [],
  tags: [],
  sources: [],
  packages: [{ packageName: "vendor/package", affectedVersions: "<2.0.0" }],
  createdAt: "2026-10-22T00:00:00Z",
  updatedAt: "2026-10-22T00:00:00Z",
  syncedAt: "2026-10-22T00:00:00Z",
  cwes,
  notesInternal: "Saved notes",
  recommendedUpgrade: "2.0.0",
};
function preview(id: string, patchedVersion?: string): GhsaAdvisoryPreview {
  return {
    ghsaId: id,
    title: id,
    cwes,
    externalReferences: [],
    packages: [{ packageName: "vendor/package", affectedVersions: "<2.0.0", patchedVersion }],
  };
}
async function click(wrapper: VueWrapper, text: string, index = 0) {
  const button = wrapper.findAll("button").filter((b) => b.text() === text)[index];
  expect(button).toBeDefined();
  await button!.trigger("click");
  await flushPromises();
}
async function fetchPreview(wrapper: VueWrapper, data: GhsaAdvisoryPreview) {
  api.adminPreviewGhsaImport.mockResolvedValueOnce({ data });
  await wrapper.get('input[aria-label="Import from GitHub"]').setValue(data.ghsaId);
  await click(wrapper, "Fetch advisory");
}

beforeEach(() => {
  vi.resetAllMocks();
  api.adminGetAdvisory.mockResolvedValue({ data: advisory });
  api.adminUpdateAdvisory.mockResolvedValue({ data: { ...advisory, title: "Edited title" } });
  api.adminLinkAdvisoryOfficial.mockResolvedValue({
    data: { ...advisory, ghsaId: "GHSA-2345-2345-2345" },
  });
  api.adminImportGhsaAdvisory.mockResolvedValue({ data: advisory });
});

describe("advisory import form", () => {
  it("updates generated upgrade advice for a new preview and clears it when no single fix exists", async () => {
    const wrapper = mount(CreateAdvisory);
    await fetchPreview(wrapper, preview("GHSA-aaaa-aaaa-aaaa", "2.0.0"));
    const upgrade = wrapper.get<HTMLInputElement>("#create-advisory-form-recommendedUpgrade");
    expect(upgrade.element.value).toBe("2.0.0");
    await fetchPreview(wrapper, preview("GHSA-bbbb-bbbb-bbbb", "3.0.0"));
    expect(upgrade.element.value).toBe("3.0.0");
    await fetchPreview(wrapper, preview("GHSA-cccc-cccc-cccc"));
    expect(upgrade.element.value).toBe("");
    wrapper.unmount();
  });

  it("preserves edited upgrade advice and imported CWE names on submission", async () => {
    const wrapper = mount(CreateAdvisory);
    await fetchPreview(wrapper, preview("GHSA-aaaa-aaaa-aaaa", "2.0.0"));
    const upgrade = wrapper.get<HTMLInputElement>("#create-advisory-form-recommendedUpgrade");
    await upgrade.setValue("Custom guidance");
    await fetchPreview(wrapper, preview("GHSA-bbbb-bbbb-bbbb", "3.0.0"));
    expect(upgrade.element.value).toBe("Custom guidance");
    await wrapper.get("#create-advisory-form-cwes").setValue("cwe-79, CWE-89");
    await wrapper.get("form").trigger("submit");
    await flushPromises();
    expect(api.adminImportGhsaAdvisory).toHaveBeenCalledWith(
      expect.objectContaining({
        body: expect.objectContaining({
          recommendedUpgrade: "Custom guidance",
          cwes: [...cwes, { id: "CWE-89", name: "" }],
        }),
      }),
    );
    wrapper.unmount();
  });
});

describe("advisory editing", () => {
  it("shows distinct shop impact for a draft and refreshes it after saving", async () => {
    const wrapper = mount(DetailAdvisory);
    await flushPromises();
    const metrics = () => wrapper.findAll("dl > div").map((metric) => metric.text());
    expect(metrics()).toEqual([
      "Affected shops2",
      "Affected environments3",
      "Shops with unsuppressed matches1",
      "Suppressed environments2",
    ]);
    expect(wrapper.text()).toContain(
      "Includes drafts; shops without inventory data are not included.",
    );
    api.adminUpdateAdvisory.mockResolvedValueOnce({
      data: {
        ...advisory,
        impact: {
          shopCount: 0,
          environmentCount: 0,
          activeShopCount: 0,
          suppressedEnvironmentCount: 0,
        },
      },
    });
    await wrapper.get("#edit-advisory-form-notesInternal").setValue("Reviewed");
    await click(wrapper, "Save", 1);
    expect(metrics()).toEqual([
      "Affected shops0",
      "Affected environments0",
      "Shops with unsuppressed matches0",
      "Suppressed environments0",
    ]);
    wrapper.unmount();
  });

  it.each([
    { name: "updated disclosure", saved: "https://example.com/new", edited: undefined },
    { name: "cleared disclosure", saved: "", edited: undefined },
    {
      name: "independent linking edit",
      saved: "https://example.com/new",
      edited: "https://example.com/official",
    },
    { name: "independently cleared linking field", saved: "https://example.com/new", edited: "" },
  ])("keeps the linking URL correct after saving $name", async ({ saved, edited }) => {
    api.adminGetAdvisory.mockResolvedValue({
      data: { ...advisory, link: "https://example.com/old" },
    });
    api.adminUpdateAdvisory.mockResolvedValue({ data: { ...advisory, link: saved || undefined } });
    const wrapper = mount(DetailAdvisory);
    await flushPromises();
    const linkInput = wrapper.get<HTMLInputElement>("#edit-advisory-linkForm-link");
    if (edited !== undefined) {
      await linkInput.setValue(edited);
    }
    await wrapper.get("#edit-advisory-contentForm-link").setValue(saved);
    await click(wrapper, "Save");
    const expected = edited ?? saved;
    expect(linkInput.element.value).toBe(expected);
    await wrapper.get("#edit-advisory-linkForm-ghsaId").setValue("GHSA-2345-2345-2345");
    await click(wrapper, "Link official advisory");
    expect(api.adminLinkAdvisoryOfficial).toHaveBeenCalledWith({
      path: { advisoryId: advisory.advisoryId },
      body: { cve: undefined, ghsaId: "GHSA-2345-2345-2345", link: expected || undefined },
    });
    wrapper.unmount();
  });

  it("preserves unsaved enrichment and link inputs when saving content, including CWE names", async () => {
    const wrapper = mount(DetailAdvisory);
    await flushPromises();
    await wrapper.get("#edit-advisory-form-notesInternal").setValue("Unsaved notes");
    await wrapper.get("#edit-advisory-form-recommendedUpgrade").setValue("Unsaved guidance");
    await wrapper.get("#edit-advisory-form-isVisible").trigger("click");
    await wrapper.get("#edit-advisory-linkForm-ghsaId").setValue("GHSA-2345-2345-2345");
    await wrapper.get("#edit-advisory-contentForm-title").setValue("Edited title");
    await click(wrapper, "Save");
    expect(api.adminUpdateAdvisory).toHaveBeenCalledWith(
      expect.objectContaining({ body: expect.objectContaining({ title: "Edited title", cwes }) }),
    );
    expect(wrapper.get<HTMLInputElement>("#edit-advisory-form-notesInternal").element.value).toBe(
      "Unsaved notes",
    );
    expect(
      wrapper.get<HTMLInputElement>("#edit-advisory-form-recommendedUpgrade").element.value,
    ).toBe("Unsaved guidance");
    expect(wrapper.get("#edit-advisory-form-isVisible").attributes("aria-checked")).toBe("true");
    expect(wrapper.get<HTMLInputElement>("#edit-advisory-linkForm-ghsaId").element.value).toBe(
      "GHSA-2345-2345-2345",
    );
    await click(wrapper, "Link official advisory");
    expect(wrapper.get<HTMLInputElement>("#edit-advisory-form-notesInternal").element.value).toBe(
      "Unsaved notes",
    );
    expect(
      wrapper.get<HTMLInputElement>("#edit-advisory-form-recommendedUpgrade").element.value,
    ).toBe("Unsaved guidance");
    expect(wrapper.get("#edit-advisory-form-isVisible").attributes("aria-checked")).toBe("true");
    wrapper.unmount();
  });
});
