<template>
  <div :class="bannerClasses" :role="normalizedVariant === 'error' ? 'alert' : 'status'">
    <span class="banner-icon">
      <slot name="icon">
        <component
          :is="getIconComponent(normalizedVariant)"
          class="icon icon-status"
          :class="iconClasses"
        />
      </slot>
    </span>

    <div class="banner-body">
      <div class="banner-content">
        <p v-if="title || $slots.title" class="banner-title">
          <slot name="title">{{ title }}</slot>
        </p>

        <div v-if="description || $slots.description" class="banner-description">
          <slot name="description">{{ description }}</slot>
        </div>

        <div v-if="$slots.default" class="banner-description banner-description-default">
          <slot />
        </div>
      </div>

      <div v-if="$slots.action" class="banner-action">
        <slot name="action" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, useSlots } from "vue";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaCircleInfo from "~icons/fa6-solid/circle-info";
import FaCircleXmark from "~icons/fa6-solid/circle-xmark";

const props = withDefaults(
  defineProps<{
    variant?: "default" | "alert" | "error" | "success";
    title?: string;
    description?: string;
  }>(),
  {
    variant: "default",
    title: undefined,
    description: undefined,
  },
);

const slots = useSlots();

const normalizedVariant = computed(() => {
  return props.variant;
});

const bannerClasses = computed(() => {
  return [
    "banner",
    "kumo-banner",
    `banner-${props.variant}`,
    `banner-${normalizedVariant.value}`,
    {
      "banner-structured":
        Boolean(props.title) || Boolean(props.description) || Boolean(slots.title) || Boolean(slots.action),
    },
  ];
});

const iconClasses = computed(() => {
  return [`icon-${props.variant}`, `icon-${normalizedVariant.value}`];
});

function getIconComponent(variant: string) {
  switch (variant) {
    case "error":
      return FaCircleXmark;
    case "default":
    case "alert":
      return FaCircleInfo;
    default:
      return FaCircleCheck;
  }
}
</script>

<style scoped>
.kumo-banner {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: 0.5rem;
  border: 1px solid transparent;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  line-height: 1.4;
}

.banner-icon {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  min-height: 1.375rem;

  svg {
    width: 1.25rem;
    height: 1.25rem;
  }
}

.banner-body {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.banner-content {
  min-width: 0;
  flex: 1;

  :deep(p) {
    margin: 0;
  }

  :deep(p + p) {
    margin-top: 0.125rem;
  }
}

.banner-title {
  margin: 0;
  font-weight: 500;
  line-height: 1.35;
}

.banner-description {
  font-size: 0.875rem;
  line-height: 1.35;
}

.banner-description-default {
  font-size: inherit;
}

.banner-structured .banner-description-default {
  margin-top: 0.125rem;
}

.banner-action {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 0.5rem;
}

.banner-default {
  color: var(--info-color);
  border-color: color-mix(in srgb, var(--info-color) 30%, transparent);
  background-color: color-mix(in srgb, var(--info-color) 10%, var(--panel-background));
}

.banner-success {
  color: var(--success-color);
  border-color: color-mix(in srgb, var(--success-color) 30%, transparent);
  background-color: color-mix(in srgb, var(--success-color) 10%, var(--panel-background));
}

.banner-alert {
  color: var(--warning-color);
  border-color: color-mix(in srgb, var(--warning-color) 30%, transparent);
  background-color: color-mix(in srgb, var(--warning-color) 12%, var(--panel-background));
}

.banner-error {
  color: var(--error-color);
  border-color: color-mix(in srgb, var(--error-color) 30%, transparent);
  background-color: color-mix(in srgb, var(--error-color) 10%, var(--panel-background));
}
</style>
