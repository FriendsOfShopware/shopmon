import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, ref } from "vue";

const isLoading = ref(false);
const isRefreshing = ref(false);
const notifications = ref<components["schemas"]["Notification"][]>([]);

export function useNotifications() {
  const { session } = useSession();

  if (session.value?.user && notifications.value.length === 0) {
    loadNotifications();
  }

  const unreadNotificationCount = computed(() => {
    return notifications.value.filter((n) => !n.read).length;
  });

  async function loadNotifications() {
    isLoading.value = true;
    const { data } = await api.GET("/notifications");
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

    await api.POST("/notifications/mark-read");

    for (const notification of notifications.value) {
      notification.read = true;
    }
  }

  async function deleteAllNotifications() {
    await api.DELETE("/notifications");
    notifications.value = [];
  }

  async function deleteNotification(id: number) {
    await api.DELETE("/notifications/{id}", {
      params: { path: { id } },
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
