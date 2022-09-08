import { defineStore } from 'pinia';

interface Alert {
    message: string;
    type: 'alert-success' | 'alert-danger';
}

export const useAlertStore = defineStore({
    id: 'alert',
    state: () => ({
        alert: null as Alert|null
    }),
    actions: {
        success(message: string) {
            this.alert = { message, type: 'alert-success' };
        },

        error(message: string) {
            this.alert = { message, type: 'alert-danger' };
        },

        clear() {
            this.alert = null;
        }
    }
});
