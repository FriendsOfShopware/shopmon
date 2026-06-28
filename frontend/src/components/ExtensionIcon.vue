<template>
  <div
    class="flex shrink-0 items-center justify-center overflow-hidden rounded-xl border bg-card"
    :style="{ width: `${size}px`, height: `${size}px` }"
  >
    <img
      v-if="src && !failed"
      :src="src"
      :alt="alt"
      class="size-full object-contain"
      loading="lazy"
      @error="failed = true"
    />
    <div v-else class="flex size-full items-center justify-center bg-primary/10 text-primary">
      <icon-fa6-solid:puzzle-piece :style="{ fontSize: `${Math.round(size * 0.45)}px` }" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

const props = withDefaults(
  defineProps<{
    src?: string | null;
    alt?: string;
    size?: number;
  }>(),
  {
    src: null,
    alt: "",
    size: 44,
  },
);

const failed = ref(false);
// Reset the error flag if the source changes (e.g. list re-renders).
watch(
  () => props.src,
  () => {
    failed.value = false;
  },
);
</script>
