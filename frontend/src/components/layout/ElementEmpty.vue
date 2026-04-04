<template>
  <div class="element-empty">
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

    <h3 class="element-empty-title">{{ title || $t("common.noElements") }}</h3>

    <p class="element-empty-description">
      <slot> {{ $t("common.getStartedElement") }} </slot>
    </p>

    <div class="element-empty-cta">
      <UiButton :to="route" variant="primary">
        <icon-fa6-solid:plus class="icon" aria-hidden="true" />
        {{ button || $t("common.addElement") }}
      </UiButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

withDefaults(
  defineProps<{
    title?: string;
    button?: string;
    route: RouteLocationRaw;
  }>(),
  {
    title: undefined,
    button: undefined,
  },
);
</script>

<style scoped>
.element-empty {
  text-align: center;
  padding: 2.5rem 1.5rem;
  border-radius: 1rem;
  border: 1px solid var(--panel-border-color);
  background: var(--control-background);
  box-shadow: var(--surface-shadow);
}

.element-empty-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 0.75rem;
  color: var(--text-color-muted);
}

.element-empty-title {
  font-size: 1.25rem;
  line-height: 1.5rem;
  font-weight: 600;
}

.element-empty-description {
  max-width: 34rem;
  margin: 0.5rem auto 0;
  color: var(--text-color-muted);
}

.element-empty-cta {
  margin-top: 1.5rem;
}
</style>
