import { defineStore } from 'pinia';
import { ref } from 'vue';

interface Alert {
    title: string;
    message: string;
    type: 'success' | 'info' | 'error' | 'warning';
}

export const useAlertStore = defineStore('alert', () => {
    const alert = ref<Alert | null>(null);

    function success(message: string) {
        alert.value = { title: 'Action success', message, type: 'success' };
        setTimeout(() => {
            clear();
        }, 5000);
    }

    function info(message: string) {
        alert.value = { title: 'Information', message, type: 'info' };
        setTimeout(() => {
            clear();
        }, 5000);
    }

    function error(message: string) {
        alert.value = { title: 'Something went wrong', message, type: 'error' };
    }

    function warning(message: string) {
        alert.value = {
            title: 'Additional information',
            message,
            type: 'warning',
        };
    }

    function clear() {
        alert.value = null;
    }

    return {
        alert,
        success,
        info,
        error,
        warning,
        clear,
    };
});
