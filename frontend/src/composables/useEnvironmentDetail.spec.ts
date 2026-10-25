import { beforeEach, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { mount, flushPromises } from "@vue/test-utils";
import { useEnvironmentDetail } from "./useEnvironmentDetail";
const mocks = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }));
vi.mock("@/composables/useAlert", () => ({ useAlert: () => mocks }));
vi.mock("@unhead/vue", () => ({ useHead: vi.fn() }));
vi.mock("vue-router", () => ({ useRoute: () => ({ params: { environmentId: "1" }, meta: {} }) }));
vi.mock("@/helpers/api", () => ({ apiLanguage: () => "en" }));
vi.mock("@/api/generated", () => ({
  getEnvironment: vi.fn(),
  getShopwareVersions: vi.fn(),
  refreshEnvironment: vi.fn(),
  clearEnvironmentCache: vi.fn(),
  subscribeToEnvironment: vi.fn(),
  unsubscribeFromEnvironment: vi.fn(),
}));
import {
  getEnvironment,
  getShopwareVersions,
  refreshEnvironment,
  clearEnvironmentCache,
  subscribeToEnvironment,
  unsubscribeFromEnvironment,
} from "@/api/generated";
let state: ReturnType<typeof useEnvironmentDetail>;
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getEnvironment).mockResolvedValue({
    data: { id: 1, organizationId: "org", subscribed: false },
  } as any);
  vi.mocked(getShopwareVersions).mockResolvedValue({ data: [] } as any);
});
it.each([
  ["onRefresh", refreshEnvironment],
  ["onCacheClear", clearEnvironmentCache],
  ["toggleNotificationSubscription", subscribeToEnvironment],
  ["toggleNotificationSubscription", unsubscribeFromEnvironment],
] as const)("%s stops on resolved errors", async (action, mutation) => {
  const wrapper = mount(
    defineComponent({
      setup() {
        state = useEnvironmentDetail();
        return () => null;
      },
    }),
  );
  await flushPromises();
  state.isSubscribed.value = mutation === unsubscribeFromEnvironment;
  const subscribedBefore = state.isSubscribed.value;
  state.showEnvironmentRefreshModal.value = true;
  vi.mocked(mutation).mockResolvedValue({ error: { message: "Request rejected" } } as any);
  await state[action](false);
  expect(mocks.error).toHaveBeenCalledWith("Request rejected");
  expect(mocks.success).not.toHaveBeenCalled();
  expect(state.isSubscribed.value).toBe(subscribedBefore);
  expect(state.showEnvironmentRefreshModal.value).toBe(true);
  expect(getEnvironment).toHaveBeenCalledTimes(1);
  expect(state.isRefreshing.value).toBe(false);
  expect(state.isCacheClearing.value).toBe(false);
  expect(state.isSubscribing.value).toBe(false);
  wrapper.unmount();
});
