<script setup lang="ts">
defineProps<{
  label: string;
  value: string;
  /** Monospace the value — used for masked identifiers, where character shape matters. */
  mono?: boolean;
}>();
</script>

<template>
  <div class="flex items-baseline gap-1.5">
    <dt class="text-muted-foreground">{{ label }}:</dt>
    <dd :class="['font-medium text-foreground', mono ? 'font-mono' : '']">{{ value }}</dd>
  </div>
</template>
