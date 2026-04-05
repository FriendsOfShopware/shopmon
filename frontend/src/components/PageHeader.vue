<script setup lang="ts">
defineProps<{
  title: string;
  description?: string;
}>();
</script>

<template>
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">{{ title }}</h1>
      <p v-if="description" class="mt-1 text-muted-foreground">{{ description }}</p>
    </div>
    <slot />
  </div>
</template>
