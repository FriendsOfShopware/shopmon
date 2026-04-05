<script setup lang="ts">
import type { Component } from "vue";

withDefaults(
  defineProps<{
    icon: Component;
    title: string;
    description?: string;
    size?: "default" | "sm";
  }>(),
  {
    size: "default",
  },
);
</script>

<template>
  <div
    class="flex flex-col items-center text-center"
    :class="
      size === 'sm'
        ? 'gap-2 rounded-xl border border-dashed py-8'
        : 'gap-4 rounded-xl border border-dashed py-16'
    "
  >
    <div
      v-if="size === 'default'"
      class="flex size-14 items-center justify-center rounded-2xl bg-primary/10"
    >
      <component :is="icon" class="size-6 text-primary" />
    </div>
    <component v-else :is="icon" class="size-8 text-muted-foreground" />

    <h2 v-if="size === 'default'" class="text-xl font-semibold">{{ title }}</h2>
    <h3 v-else class="text-lg font-semibold">{{ title }}</h3>

    <p
      v-if="description"
      :class="size === 'sm' ? 'text-sm text-muted-foreground' : 'max-w-md text-muted-foreground'"
    >
      {{ description }}
    </p>

    <slot />
  </div>
</template>
