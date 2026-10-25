import { requireMutation } from "@/helpers/apiError";
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import {
  deleteAllNotifications as apiDeleteAllNotifications,
  deleteNotification as apiDeleteNotification,
  getNotifications,
  markNotificationsRead as apiMarkNotificationsRead,
  type Notification,
} from "@/api/generated";
import { computed, ref } from "vue";

const isLoading = ref(false);
const isRefreshing = ref(false);
const notifications = ref<Notification[]>([]);

export function useNotifications() {
  const { session } = useSession();
  const { error } = useAlert();

  if (session.value?.user && notifications.value.length === 0) {
    void loadNotifications();
  }

  const unreadNotificationCount = computed(() => {
    return notifications.value.filter((n) => !n.read).length;
  });

  async function loadNotifications() {
    isLoading.value = true;
    const { data } = await getNotifications();
    notifications.value = data ?? [];
    isLoading.value = false;
  }

  async function markAllRead() {
    try {
      let allRead = true;
      for (const notification of notifications.value) {
        if (notification.read === false) {
          allRead = false;
          break;
        }
      }

      if (allRead) {
        return;
      }

      requireMutation(
        await apiMarkNotificationsRead(),
        "Failed to update notifications. Please try again.",
      );

      for (const notification of notifications.value) {
        notification.read = true;
      }
    } catch (err) {
      error(
        err instanceof Error ? err.message : "Failed to update notifications. Please try again.",
      );
    }
  }

  async function deleteAllNotifications() {
    try {
      requireMutation(
        await apiDeleteAllNotifications(),
        "Failed to update notifications. Please try again.",
      );
      notifications.value = [];
    } catch (err) {
      error(
        err instanceof Error ? err.message : "Failed to update notifications. Please try again.",
      );
    }
  }

  async function deleteNotification(id: number) {
    try {
      requireMutation(
        await apiDeleteNotification({
          path: { id },
        }),
        "Failed to update notifications. Please try again.",
      );
      notifications.value = notifications.value.filter((e) => e.id !== id);
    } catch (err) {
      error(
        err instanceof Error ? err.message : "Failed to update notifications. Please try again.",
      );
    }
  }

  return {
    isLoading,
    isRefreshing,
    notifications,
    unreadNotificationCount,
    loadNotifications,
    markAllRead,
    deleteAllNotifications,
    deleteNotification,
  };
}
