<template>
  <div :class="['element-empty', `element-empty-${size}`]">
    <slot name="icon">
      <svg
        class="element-empty-icon"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        aria-hidden="true"
      >
        <path
          vector-effect="non-scaling-stroke"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
        />
      </svg>
    </slot>

    <h2 v-if="resolvedTitle" class="element-empty-title">{{ resolvedTitle }}</h2>

    <p v-if="$slots.default || descriptionText" class="element-empty-description">
      <slot>{{ descriptionText }}</slot>
    </p>

    <div v-if="$slots.contents || route" class="element-empty-contents">
      <slot name="contents">
        <UiButton v-if="route" :to="route" variant="primary">
          <icon-fa6-solid:plus class="icon" aria-hidden="true" />
          {{ resolvedButton }}
        </UiButton>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { RouteLocationRaw } from "vue-router";
import { useI18n } from "vue-i18n";

const props = withDefaults(
  defineProps<{
    title?: string;
    button?: string;
    route?: RouteLocationRaw;
    size?: "sm" | "base" | "lg";
  }>(),
  {
    title: undefined,
    button: undefined,
    route: undefined,
    size: "base",
  },
);

const { t } = useI18n();

const resolvedTitle = computed(() => {
  return props.title ?? t("common.noElements");
});

const resolvedButton = computed(() => {
  return props.button ?? t("common.addElement");
});

const descriptionText = computed(() => {
  return t("common.getStartedElement");
});
</script>

<style scoped>
.element-empty {
  display: flex;
  width: 100%;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  text-align: center;
  border-radius: 0.75rem;
  border: 1px solid var(--field-border-color);
  background: var(--control-background);
  color: var(--text-color);
}

.element-empty-sm {
  padding: 2rem 1.5rem;
  gap: 1rem;
}

.element-empty-base {
  padding: 4rem 2.5rem;
  gap: 1.5rem;
}

.element-empty-lg {
  padding: 5rem 3rem;
  gap: 2rem;
}

.element-empty-icon {
  width: 3rem;
  height: 3rem;
  color: var(--text-color-muted);
  flex-shrink: 0;
}

.element-empty-title {
  margin: 0;
  font-size: 1.5rem;
  line-height: 1.2;
  font-weight: 600;
}

.element-empty-description {
  margin: 0;
  max-width: 34rem;
  color: var(--text-color-muted);
  line-height: 1.5;
}

.element-empty-contents {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  width: 100%;
}
</style>
