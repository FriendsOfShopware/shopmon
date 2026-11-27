import { beforeEach, expect, it, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import EnvironmentDestinationPreferences from "./EnvironmentDestinationPreferences.vue";
vi.mock("@/api/generated", () => ({
  listNotificationDestinations: vi.fn(),
  listShopNotificationDestinations: vi.fn(),
}));
import { listNotificationDestinations, listShopNotificationDestinations } from "@/api/generated";
const mount = () =>
  shallowMount(EnvironmentDestinationPreferences, {
    props: { organizationId: "org", shopId: 3, environmentId: 7 },
  });
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(listNotificationDestinations).mockResolvedValue({
    data: {
      canManage: true,
      destinations: [
        { id: "org-dest", name: "Team", provider: "slack", summary: "Incoming webhook" },
      ],
    },
  } as any);
  vi.mocked(listShopNotificationDestinations).mockResolvedValue({
    data: {
      canManage: true,
      destinations: [{ id: "shop-dest", name: "Shop", provider: "telegram", summary: "123" }],
    },
  } as any);
});
it("offers both organization and shop destinations for environment overrides", async () => {
  const w = mount();
  await flushPromises();
  expect(w.findAll("option")).toHaveLength(2);
  const matrix = () => w.findComponent({ name: "NotificationMatrix" });
  expect(matrix().props("destination")).toEqual({ id: "org-dest", organizationId: "org" });
  expect(matrix().props("environmentId")).toBe(7);
  await w.get("select").setValue("shop-dest");
  expect(matrix().props("destination").id).toBe("shop-dest");
  w.unmount();
});
it("does not display partial controls when one scope fails to load", async () => {
  vi.mocked(listShopNotificationDestinations).mockRejectedValue(new Error("offline"));
  const w = mount();
  await flushPromises();
  expect(w.find('[role="alert"]').exists()).toBe(true);
  expect(w.find("select").exists()).toBe(false);
  w.unmount();
});
it("makes the matrix read-only for a member", async () => {
  vi.mocked(listNotificationDestinations).mockResolvedValue({
    data: {
      canManage: false,
      destinations: [
        { id: "org-dest", name: "Team", provider: "slack", summary: "Incoming webhook" },
      ],
    },
  } as any);
  const w = mount();
  await flushPromises();
  expect(w.findComponent({ name: "NotificationMatrix" }).props("disabled")).toBe(true);
  w.unmount();
});
