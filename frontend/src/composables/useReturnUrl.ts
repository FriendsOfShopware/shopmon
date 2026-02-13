import { ref } from "vue";

// Shared state for return URL after login
const returnUrl = ref<string | null>(null);

export function useReturnUrl() {
  function setReturnUrl(url: string | null) {
    returnUrl.value = url;
  }

  function getReturnUrl(): string | null {
    return returnUrl.value;
  }

  function clearReturnUrl() {
    returnUrl.value = null;
  }

  return {
    returnUrl,
    setReturnUrl,
    getReturnUrl,
    clearReturnUrl,
  };
}
