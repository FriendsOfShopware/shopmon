import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Notification } from "@apiTypes/notification";

export const useNotificationStore = defineStore('notification', {
    state: () => ({
        isLoading: false,
        isRefreshing: false,
        websocket: null as WebSocket | null,
        notifications: [] as Notification[],
    }),
    getters: {
        isConnected(): boolean {
            if (!this.websocket) return false;

            return this.websocket.readyState === WebSocket.OPEN;
        },
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
                const data = JSON.parse(event.data) as Notification;

                this.notifications.push(data);
            }
        },

        async loadNotifications() {
            this.isLoading = true;
            const notifications = await fetchWrapper.get('/account/me/notifications') as Notification[];
            this.isLoading = false;

            this.notifications = notifications;
        },

        async deleteAllNotifications() {
            await fetchWrapper.delete('/account/me/notifications');

            this.notifications = [];
        },

        disconnect() {
            if (!this.websocket) return;

            this.websocket.close();
            this.websocket = null;
        }
    }
})