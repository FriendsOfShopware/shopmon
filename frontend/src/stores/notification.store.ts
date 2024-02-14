import { defineStore } from "pinia";
import type { WebsocketMessage } from "@/types/notification";
import { useShopStore } from "./shop.store";
import { useAlertStore } from "./alert.store";
import { trpcClient, RouterOutput } from "@/helpers/trpc";

export const useNotificationStore = defineStore('notification', {
    state: () => ({
        isLoading: false,
        isRefreshing: false,
        websocket: null as WebSocket | null,
        notifications: [] as RouterOutput['account']['notification']['list'],
    }),
    getters: {
        isConnected(): boolean {
            if (!this.websocket) return false;

            return this.websocket.readyState === WebSocket.OPEN;
        },
        unreadNotificationCount(): number {
            return this.notifications.filter((n) => !n.read).length;
        }
    },
    actions: {
        connect(access_token: string) {
            if (this.isConnected) {
                return;
            }

            const url = new URL(window.location.href);
            url.pathname = '/api/ws';
            url.searchParams.set('auth_token', access_token);

            if (url.protocol === 'https:') {
                url.protocol = 'wss:';
            } else {
                url.protocol = 'ws:';
            }

            this.websocket = new WebSocket(url);
            this.websocket.onmessage = (event) => {
                const data = JSON.parse(event.data) as WebsocketMessage;

                if (data.notification) {
                    for (const notification of this.notifications) {
                        if (notification.title === data.notification.title && notification.message === data.notification.message) {
                            notification.read = false;
                            return;
                        }
                    }

                    // @ts-expect-error
                    this.notifications.unshift(data.notification);
                }

                if (data.shopUpdate) {
                    const shopStore = useShopStore();
                    const alertStore = useAlertStore();

                    if (shopStore.shop?.id === data.shopUpdate.id) {
                        shopStore.loadShop(shopStore.shop.organizationId, data.shopUpdate.id);

                        alertStore.success('Shop has been updated');
                    }
                }
            }
        },

        async loadNotifications() {
            this.isLoading = true;
            const notifications = await trpcClient.account.notification.list.query();
            this.isLoading = false;

            this.notifications = notifications;
        },

        async markAllRead() {
            let allRead = true;
            for (const notification of this.notifications) {
                if (notification.read === false) {
                    allRead = false;
                    break;
                }
            }

            if (allRead) {
                return;
            }

            await trpcClient.account.notification.markAllRead.mutate();

            for (const notification of this.notifications) {
                notification.read = true;
            }
        },

        async deleteAllNotifications() {
            await trpcClient.account.notification.delete.mutate();

            this.notifications = [];
        },

        async deleteNotification(id: number) {
            await trpcClient.account.notification.delete.mutate(id);

            this.notifications = this.notifications.filter(e => e.id !== id);
        },

        disconnect() {
            if (!this.websocket) return;

            this.websocket.close();
            this.websocket = null;
        }
    }
})
