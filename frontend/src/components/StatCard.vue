<script setup lang="ts">
import { Card, CardContent } from "@/components/ui/card";
import type { Component } from "vue";

defineProps<{
  icon: Component;
  value: string | number;
  label: string;
  color?: "primary" | "success" | "warning" | "destructive" | "muted";
}>();

const colorClasses: Record<string, { bg: string; text: string }> = {
  primary: { bg: "bg-primary/10", text: "text-primary" },
  success: { bg: "bg-success/10", text: "text-success" },
  warning: { bg: "bg-warning/10", text: "text-warning" },
  destructive: { bg: "bg-destructive/10", text: "text-destructive" },
  muted: { bg: "bg-muted", text: "text-muted-foreground" },
};
</script>

<template>
  <Card>
    <CardContent class="flex items-center gap-3 p-4">
      <div
        class="flex size-9 items-center justify-center rounded-lg"
        :class="colorClasses[color ?? 'primary'].bg"
      >
        <component :is="icon" class="size-4" :class="colorClasses[color ?? 'primary'].text" />
      </div>
      <div>
        <div class="text-2xl font-bold tabular-nums">{{ value }}</div>
        <div class="text-xs text-muted-foreground">{{ label }}</div>
      </div>
    </CardContent>
  </Card>
</template>
