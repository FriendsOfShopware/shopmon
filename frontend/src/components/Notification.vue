<template>
  <div class="fixed right-0 top-3 z-20 flex max-w-sm overflow-hidden">
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="translate-x-0"
      leave-to-class="translate-x-full"
    >
      <div
        v-if="alert"
        :class="[
          'relative mr-3 mb-4 flex w-screen gap-3 rounded-xl border bg-card p-4 shadow-lg',
          variantBorder,
        ]"
      >
        <button
          class="absolute right-2 top-2 flex size-6 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground"
          type="button"
          @click="clear()"
        >
          <icon-fa6-solid:xmark aria-hidden="true" />
        </button>
        <div class="mt-0.5 shrink-0" :class="variantColor">
          <icon-fa6-solid:circle-xmark v-if="alert.type === 'error'" class="size-4" />
          <icon-fa6-solid:circle-info v-else-if="alert.type === 'warning'" class="size-4" />
          <icon-fa6-solid:circle-info v-else-if="alert.type === 'info'" class="size-4" />
          <icon-fa6-solid:circle-check v-else class="size-4" />
        </div>
        <div class="flex-1">
          <div class="font-semibold" :class="variantColor">{{ alert.title }}</div>
          {{ alert.message }}
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useAlert } from "@/composables/useAlert";

const alertComposable = useAlert();
const { alert, clear } = alertComposable;

const variantColor = computed(() => {
  switch (alert.value?.type) {
    case "error":
      return "text-destructive";
    case "warning":
      return "text-warning";
    case "info":
      return "text-info";
    default:
      return "text-success";
  }
});

const variantBorder = computed(() => {
  switch (alert.value?.type) {
    case "error":
      return "border-destructive/30";
    case "warning":
      return "border-warning/30";
    case "info":
      return "border-info/30";
    default:
      return "border-success/30";
  }
});
</script>
