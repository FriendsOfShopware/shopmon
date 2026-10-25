import { beforeEach, expect, it, vi } from "vitest";
import { ref } from "vue";
import { useNotifications } from "./useNotifications";
const error = vi.hoisted(() => vi.fn());
vi.mock("@/composables/useAlert", () => ({ useAlert: () => ({ error }) }));
vi.mock("@/composables/useSession", () => ({ useSession: () => ({ session: ref(null) }) }));
vi.mock("@/api/generated", () => ({
  getNotifications: vi.fn(),
  markNotificationsRead: vi.fn(),
  deleteAllNotifications: vi.fn(),
  deleteNotification: vi.fn(),
}));
import { markNotificationsRead, deleteAllNotifications, deleteNotification } from "@/api/generated";
beforeEach(() => {
  vi.clearAllMocks();
});
it.each([
  ["markAllRead", markNotificationsRead],
  ["deleteAllNotifications", deleteAllNotifications],
  ["deleteNotification", deleteNotification],
] as const)("%s retains notifications on a resolved error", async (action, mutation) => {
  const state = useNotifications();
  state.notifications.value = [{ id: 1, read: false }] as any;
  vi.mocked(mutation)
    .mockResolvedValueOnce({ error: { message: "Request rejected" } } as any)
    .mockResolvedValueOnce({} as any);
  await state[action](1);
  expect(state.notifications.value).toEqual([{ id: 1, read: false }]);
  expect(state.unreadNotificationCount.value).toBe(1);
  expect(error).toHaveBeenCalledWith("Request rejected");
  await state[action](1);
  expect(state.unreadNotificationCount.value).toBe(0);
});
