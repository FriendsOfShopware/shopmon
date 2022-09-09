import { defineStore } from 'pinia';

interface Alert {
    title: string,
    message: string;
    type: 'success' | 'error' | 'warning';
}

export const useAlertStore = defineStore({
    id: 'alert',
    state: () => ({
        alert: null as Alert|null
    }),
    actions: {
        success(message: string) {
            this.alert = { title: 'Action success', message, type: 'success' };
            setTimeout(() => {
                this.clear()
            }, 5000);
        },

        error(message: string) {
            this.alert = { title: 'Something went wrong', message, type: 'error' };
        },

        warning(message: string) {
            this.alert = { title: 'Additional information', message, type: 'warning' };
        },

        clear() {
            this.alert = null;
        }
    }
});
