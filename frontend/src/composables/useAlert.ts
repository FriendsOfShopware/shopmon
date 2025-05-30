import { readonly, ref } from 'vue';

interface Alert {
    title: string;
    message: string;
    type: 'success' | 'info' | 'error' | 'warning';
}

// Shared alert state across all components
const alert = ref<Alert | null>(null);
let timeoutId: ReturnType<typeof setTimeout> | null = null;

export function useAlert() {
    function success(message: string) {
        clearExistingTimeout();
        alert.value = { title: 'Action success', message, type: 'success' };
        timeoutId = setTimeout(() => {
            clear();
        }, 5000);
    }

    function info(message: string) {
        clearExistingTimeout();
        alert.value = { title: 'Information', message, type: 'info' };
        timeoutId = setTimeout(() => {
            clear();
        }, 5000);
    }

    function error(message: string) {
        clearExistingTimeout();
        alert.value = { title: 'Something went wrong', message, type: 'error' };
        // Error alerts don't auto-dismiss
    }

    function warning(message: string) {
        clearExistingTimeout();
        alert.value = {
            title: 'Additional information',
            message,
            type: 'warning',
        };
        // Warning alerts don't auto-dismiss
    }

    function clear() {
        clearExistingTimeout();
        alert.value = null;
    }

    function clearExistingTimeout() {
        if (timeoutId) {
            clearTimeout(timeoutId);
            timeoutId = null;
        }
    }

    return {
        // Use readonly to prevent direct mutation
        alert: readonly(alert),
        success,
        info,
        error,
        warning,
        clear,
    };
}
