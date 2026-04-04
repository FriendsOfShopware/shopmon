import { readonly, ref } from "vue";
import { i18n } from "@/i18n";

interface Alert {
  title: string;
  message: string;
  type: "success" | "info" | "error" | "warning";
}

// Shared alert state across all components
const alert = ref<Alert | null>(null);
let timeoutId: ReturnType<typeof setTimeout> | null = null;

function t(key: string): string {
  return i18n.global.t(key);
}

export function useAlert() {
  function success(message: string) {
    clearExistingTimeout();
    alert.value = { title: t("alert.successTitle"), message, type: "success" };
    timeoutId = setTimeout(() => {
      clear();
    }, 5000);
  }

  function info(message: string) {
    clearExistingTimeout();
    alert.value = { title: t("alert.infoTitle"), message, type: "info" };
    timeoutId = setTimeout(() => {
      clear();
    }, 5000);
  }

  function error(message: string) {
    clearExistingTimeout();
    alert.value = { title: t("alert.errorTitle"), message, type: "error" };
    // Error alerts don't auto-dismiss
  }

  function warning(message: string) {
    clearExistingTimeout();
    alert.value = {
      title: t("alert.warningTitle"),
      message,
      type: "warning",
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
