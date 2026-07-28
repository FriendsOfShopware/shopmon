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

    await apiMarkNotificationsRead();

    for (const notification of notifications.value) {
      notification.read = true;
    }
  }

  async function deleteAllNotifications() {
    await apiDeleteAllNotifications();
    notifications.value = [];
  }

  async function deleteNotification(id: number) {
    await apiDeleteNotification({
      path: { id },
    });
    notifications.value = notifications.value.filter((e) => e.id !== id);
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
