import { authClient } from '@/helpers/auth-client';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { computed, ref } from 'vue';

const isLoading = ref(false);
const isRefreshing = ref(false);
const notifications = ref<RouterOutput['account']['notification']['list']>([]);

export function useNotifications() {
    const session = authClient.useSession();

    if (session.value.data?.user && notifications.value.length === 0) {
        loadNotifications();
    }

    const unreadNotificationCount = computed(() => {
        return notifications.value.filter((n) => !n.read).length;
    });

    async function loadNotifications() {
        isLoading.value = true;
        notifications.value =
            await trpcClient.account.notification.list.query();
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

        await trpcClient.account.notification.markAllRead.mutate();

        for (const notification of notifications.value) {
            notification.read = true;
        }
    }

    async function deleteAllNotifications() {
        await trpcClient.account.notification.delete.mutate();
        notifications.value = [];
    }

    async function deleteNotification(id: number) {
        await trpcClient.account.notification.delete.mutate(id);
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
