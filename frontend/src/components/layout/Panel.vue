<template>
  <Card :class="[variant === 'table' ? 'p-0 overflow-hidden' : '', 'mb-8']">
    <CardHeader v-if="$slots.title || title" class="flex-row items-start justify-between space-y-0">
      <div>
        <CardTitle>
          <slot name="title">{{ title }}</slot>
        </CardTitle>
        <CardDescription v-if="description">{{ description }}</CardDescription>
      </div>
      <div v-if="$slots.action">
        <slot name="action" />
      </div>
    </CardHeader>

    <CardContent :class="[!$slots.title && !title ? 'pt-6' : '']">
      <slot />
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

withDefaults(
  defineProps<{
    variant?: "default" | "table";
    title?: string;
    description?: string;
  }>(),
  {
    variant: "default",
    title: undefined,
    description: undefined,
  },
);
</script>
