<template>
  <div :class="['flex w-full flex-col items-center gap-6 rounded-xl border border-dashed bg-card text-center', sizeClasses]">
    <slot name="icon">
      <svg
        class="size-12 text-muted-foreground"
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

    <h2 v-if="resolvedTitle" class="text-2xl font-semibold">{{ resolvedTitle }}</h2>

    <p v-if="$slots.default || descriptionText" class="max-w-sm text-muted-foreground">
      <slot>{{ descriptionText }}</slot>
    </p>

    <div v-if="$slots.contents || route" class="flex w-full items-center justify-center gap-3">
      <slot name="contents">
        <UiButton v-if="route" :to="route" variant="primary">
          <icon-fa6-solid:plus class="mr-1 size-4" aria-hidden="true" />
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

const sizeClasses = computed(() => {
  switch (props.size) {
    case "sm":
      return "gap-4 px-6 py-8";
    case "lg":
      return "gap-8 px-12 py-20";
    default:
      return "gap-6 px-10 py-16";
  }
});
</script>
