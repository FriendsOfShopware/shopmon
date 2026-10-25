import { beforeEach, expect, it, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { ref } from "vue";
import ManageSSO from "./ManageSSO.vue";
const mocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => mocks }));
vi.mock("@/composables/useSession", () => ({
  useSession: () => ({ activeOrganizationId: ref("org") }),
}));
vi.mock("@/composables/usePermissions", () => ({ usePermissions: () => ref(true) }));
vi.mock("@/api/generated", () => ({
  getFullOrganization: vi.fn(),
  getSsoProviders: vi.fn(),
  discoverSso: vi.fn(),
  updateSsoProvider: vi.fn(),
  deleteSsoProvider: vi.fn(),
  registerSsoProvider: vi.fn(),
}));
import {
  getFullOrganization,
  getSsoProviders,
  updateSsoProvider,
  deleteSsoProvider,
} from "@/api/generated";
const provider = {
  providerId: "provider",
  id: "provider",
  domain: "example.com",
  issuer: "https://example.com",
  oidcConfig: {
    clientId: "client",
    authorizationEndpoint: "https://example.com/auth",
    tokenEndpoint: "https://example.com/token",
    jwksEndpoint: "https://example.com/jwks",
  },
};
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getFullOrganization).mockResolvedValue({ data: { id: "org" } } as any);
  vi.mocked(getSsoProviders).mockResolvedValue({ data: [provider] } as any);
});
it("keeps provider form and values when update resolves with error", async () => {
  vi.mocked(updateSsoProvider).mockResolvedValue({
    error: { message: "Provider rejected" },
  } as any);
  const wrapper = shallowMount(ManageSSO);
  await flushPromises();
  const vm = wrapper.vm as any;
  vm.openEditProviderModal(provider);
  await flushPromises();
  await vm.onSubmitProvider();
  await flushPromises();
  expect(updateSsoProvider).toHaveBeenCalled();
  expect(vm.showAddProviderModal).toBe(true);
  expect(vm.editingProvider.providerId).toBe("provider");
  expect(vm.isSubmitting).toBe(false);
  expect(mocks.success).not.toHaveBeenCalled();
  expect(mocks.error).toHaveBeenCalledWith("Provider rejected");
  expect(getSsoProviders).toHaveBeenCalledTimes(1);
  wrapper.unmount();
});
it("preserves deletion target and allows retry after resolved error", async () => {
  vi.mocked(deleteSsoProvider)
    .mockResolvedValueOnce({ error: "Deletion rejected" } as any)
    .mockResolvedValueOnce({} as any);
  const wrapper = shallowMount(ManageSSO);
  await flushPromises();
  const vm = wrapper.vm as any;
  vm.confirmDeleteProvider(provider);
  await vm.deleteProvider();
  expect(vm.showDeleteModal).toBe(true);
  expect(vm.deletingProvider.providerId).toBe("provider");
  expect(vm.isDeleting).toBe(false);
  expect(mocks.success).not.toHaveBeenCalled();
  expect(getSsoProviders).toHaveBeenCalledTimes(1);
  expect(mocks.error).toHaveBeenCalledWith("Deletion rejected");
  await vm.deleteProvider();
  expect(vm.showDeleteModal).toBe(false);
  expect(vm.deletingProvider).toBe(null);
  wrapper.unmount();
});
