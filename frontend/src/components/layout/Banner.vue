<template>
  <Alert :variant="alertVariant" :class="variantClasses" :role="normalizedVariant === 'error' ? 'alert' : 'status'">
    <slot name="icon">
      <component
        :is="getIconComponent(normalizedVariant)"
        class="size-5"
      />
    </slot>

    <div class="flex flex-1 items-start justify-between gap-3">
      <div class="min-w-0 flex-1">
        <AlertTitle v-if="title || $slots.title">
          <slot name="title">{{ title }}</slot>
        </AlertTitle>

        <AlertDescription v-if="description || $slots.description">
          <slot name="description">{{ description }}</slot>
        </AlertDescription>

        <AlertDescription v-if="$slots.default">
          <slot />
        </AlertDescription>
      </div>

      <div v-if="$slots.action" class="flex shrink-0 items-center gap-2">
        <slot name="action" />
      </div>
    </div>
  </Alert>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaCircleInfo from "~icons/fa6-solid/circle-info";
import FaCircleXmark from "~icons/fa6-solid/circle-xmark";

const props = withDefaults(
  defineProps<{
    variant?: "default" | "alert" | "error" | "success";
    title?: string;
    description?: string;
  }>(),
  {
    variant: "default",
    title: undefined,
    description: undefined,
  },
);

const normalizedVariant = computed(() => props.variant);

const alertVariant = computed(() => {
  return props.variant === "error" ? "destructive" : "default";
});

const variantClasses = computed(() => {
  switch (props.variant) {
    case "error":
      return "border-destructive/30 bg-destructive/10 text-destructive";
    case "success":
      return "border-success/30 bg-success/10 text-success";
    case "alert":
      return "border-warning/30 bg-warning/10 text-warning";
    default:
      return "border-info/30 bg-info/10 text-info";
  }
});

function getIconComponent(variant: string) {
  switch (variant) {
    case "error":
      return FaCircleXmark;
    case "default":
    case "alert":
      return FaCircleInfo;
    default:
      return FaCircleCheck;
  }
}
</script>
