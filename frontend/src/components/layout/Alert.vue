<template>
  <div :class="['alert', `alert-${type}`]">
    <div class="alert-icon">
      <component :is="getIconComponent(type)" class="icon icon-status" :class="`icon-${type}`" />
    </div>

    <div class="alert-content">
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaCircleInfo from "~icons/fa6-solid/circle-info";
import FaCircleXmark from "~icons/fa6-solid/circle-xmark";

defineProps<{ type: string }>();

function getIconComponent(type: string) {
  switch (type) {
    case "error":
      return FaCircleXmark;
    case "info":
    case "warning":
      return FaCircleInfo;
    default:
      return FaCircleCheck;
  }
}
</script>

<style scoped>
.alert {
  display: flex;
  align-items: flex-start;
  padding: 1rem;
  border-radius: 0.875rem;
  border: 1px solid transparent;
  box-shadow: var(--surface-shadow);
}

.alert-icon {
  flex-shrink: 0;

  svg {
    width: 1.25rem;
    height: 1.25rem;
    vertical-align: -0.4em;
  }
}

.alert-content {
  flex: 1;
  margin-left: 0.75rem;
  padding-top: 0.15rem;
}

.alert-info {
  border-color: color-mix(in srgb, var(--info-color) 30%, var(--panel-border-color));
  background-color: color-mix(in srgb, var(--info-color) 10%, var(--panel-background));
}

.alert-success {
  border-color: color-mix(in srgb, var(--success-color) 30%, var(--panel-border-color));
  background-color: color-mix(in srgb, var(--success-color) 10%, var(--panel-background));
}

.alert-warning {
  border-color: color-mix(in srgb, var(--warning-color) 34%, var(--panel-border-color));
  background-color: color-mix(in srgb, var(--warning-color) 12%, var(--panel-background));
}

.alert-error {
  border-color: color-mix(in srgb, var(--error-color) 30%, var(--panel-border-color));
  background-color: color-mix(in srgb, var(--error-color) 10%, var(--panel-background));
}
</style>
