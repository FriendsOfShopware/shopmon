<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { Component, HTMLAttributes } from "vue";
import { cn } from "@/lib/utils";

defineProps<{
  icon: Component;
  title: string;
  description?: string;
  class?: HTMLAttributes["class"];
}>();
</script>

<template>
  <Card :class="cn($props.class)">
    <CardHeader class="pb-3">
      <div class="flex items-center justify-between">
        <CardTitle class="flex items-center gap-2 text-base">
          <component :is="icon" class="size-4 text-muted-foreground" />
          {{ title }}
        </CardTitle>
        <slot name="action" />
      </div>
      <CardDescription v-if="description">{{ description }}</CardDescription>
    </CardHeader>
    <CardContent>
      <slot />
    </CardContent>
  </Card>
</template>
