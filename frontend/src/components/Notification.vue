<template>
  <div class="notifications-container">
    <transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="translate-x-0"
      leave-to-class="translate-x-full"
    >
      <div v-if="alert" class="notification" :class="`notification-${alert.type}`">
        <button class="notification-close" type="button" @click="clear()">
          <icon-fa6-solid:xmark aria-hidden="true" />
        </button>
        <div class="notification-icon">
          <icon-fa6-solid:circle-xmark v-if="alert.type === 'error'" class="icon icon-error" />
          <icon-fa6-solid:circle-info
            v-else-if="alert.type === 'warning'"
            class="icon icon-warning"
          />
          <icon-fa6-solid:circle-info v-else-if="alert.type === 'info'" class="icon icon-info" />
          <icon-fa6-solid:circle-check v-else class="icon icon-success" />
        </div>
        <div class="notification-content">
          <div class="notification-title">{{ alert.title }}</div>
          {{ alert.message }}
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";

const alertComposable = useAlert();
const { alert, clear } = alertComposable;
</script>

<style scoped>
.notifications-container {
  position: fixed;
  top: 0.75rem;
  right: 0;
  z-index: 20;
  display: flex;
  max-width: 24rem;
  overflow: hidden;
}

.notification {
  position: relative;
  width: 100vw;
  padding: 1rem 1rem 0.95rem;
  border-radius: 1rem;
  border: 1px solid var(--notification-border-color);
  background-color: var(--panel-background);
  margin: 0 0.75rem 1rem 0;
  display: flex;
  gap: 0.75rem;
  box-shadow:
    inset 0 0 0 1px var(--notification-border-color),
    var(--surface-shadow-strong);
}

.notification-info {
  background: color-mix(in srgb, var(--info-color) 8%, var(--panel-background));
  --notification-accent: var(--info-color);
}

.notification-success {
  background: color-mix(in srgb, var(--success-color) 8%, var(--panel-background));
  --notification-accent: var(--success-color);
}

.notification-warning {
  background: color-mix(in srgb, var(--warning-color) 10%, var(--panel-background));
  --notification-accent: var(--warning-color);
}

.notification-error {
  background: color-mix(in srgb, var(--error-color) 8%, var(--panel-background));
  --notification-accent: var(--error-color);
}

.notification-close {
  position: absolute;
  top: 0.55rem;
  right: 0.55rem;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  color: var(--text-color-muted);

  &:hover {
    background: var(--button-ghost-hover-background);
    color: var(--text-color);
  }
}

.notification-icon {
  display: flex;
  justify-content: center;
  margin-top: 0.1rem;
  color: var(--notification-accent);
}

.notification-content {
  flex: 1;
}

.notification-title {
  font-weight: 600;
  margin-bottom: 0.2rem;
  color: var(--notification-accent);
}
</style>
