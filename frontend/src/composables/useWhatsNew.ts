import { ref } from "vue";

const WHATS_NEW_STORAGE_KEY = "shopmon-whats-new";

export const WHATS_NEW_VERSION = "2026-03-packages-mirror";

const isOpen = ref(false);
let initialized = false;

export function useWhatsNew() {
  if (!initialized) {
    initialized = true;
    initializeWhatsNew();
  }

  function open() {
    isOpen.value = true;
  }

  function dismiss() {
    if (!isOpen.value) return;
    markAsSeen();
    isOpen.value = false;
  }

  return {
    isOpen,
    open,
    dismiss,
  };
}

function initializeWhatsNew() {
  try {
    const seenVersion = localStorage.getItem(WHATS_NEW_STORAGE_KEY);
    isOpen.value = seenVersion !== WHATS_NEW_VERSION;
  } catch {
    isOpen.value = true;
  }
}

function markAsSeen() {
  try {
    localStorage.setItem(WHATS_NEW_STORAGE_KEY, WHATS_NEW_VERSION);
  } catch {
    // Ignore storage failures and keep the modal dismissible for the session.
  }
}
