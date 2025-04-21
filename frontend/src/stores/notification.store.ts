import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { WebsocketMessage } from '@/types/notification';
import { useShopStore } from './shop.store';
import { useAlertStore } from './alert.store';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

export const useNotificationStore = defineStore('notification', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const websocket = ref<WebSocket | null>(null);
    const notifications = ref<RouterOutput['account']['notification']['list']>([]);

    const isConnected = computed(() => {
        if (!websocket.value) return false;
        return websocket.value.readyState === WebSocket.OPEN;
    });

    const unreadNotificationCount = computed(() => {
        return notifications.value.filter((n) => !n.read).length;
    });

    async function loadNotifications() {
        isLoading.value = true;
        notifications.value = await trpcClient.account.notification.list.query();
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
        notifications.value = notifications.value.filter(e => e.id !== id);
    }

    function connect(token: string) {
        if (websocket.value) {
            websocket.value.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = import.meta.env.VITE_WS_URL || `${protocol}//${window.location.host}`;
        websocket.value = new WebSocket(`${wsUrl}/api/ws?token=${token}`);

        websocket.value.onmessage = async (event) => {
            const message = JSON.parse(event.data) as WebsocketMessage;

            console.log(message);

            if (message.shopUpdate) {
                const shopStore = useShopStore();
                await shopStore.loadShops();
            } else if (message.notification) {
                const alertStore = useAlertStore();
                alertStore.info(message.notification.message);
                await loadNotifications();
            }
        };

        websocket.value.onclose = () => {
            setTimeout(() => {
                connect(token);
            }, 1000);
        };
    }

    function disconnect() {
        if (websocket.value) {
            websocket.value.close();
            websocket.value = null;
        }
    }

    return {
        isLoading,
        isRefreshing,
        websocket,
        notifications,
        isConnected,
        unreadNotificationCount,
        loadNotifications,
        markAllRead,
        deleteAllNotifications,
        deleteNotification,
        connect,
        disconnect
    };
});
