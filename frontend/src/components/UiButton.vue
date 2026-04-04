<template>
  <Button
    v-if="!to && !href"
    :variant="mappedVariant"
    :size="mappedSize"
    :class="[block ? 'w-full' : '', $attrs.class]"
    :disabled="disabled"
    :type="type"
    @click="emit('click', $event)"
  >
    <slot />
  </Button>

  <Button
    v-else
    :variant="mappedVariant"
    :size="mappedSize"
    :class="[block ? 'w-full' : '', $attrs.class]"
    :disabled="disabled"
    as-child
  >
    <RouterLink v-if="to" :to="to" @click="handleClick">
      <slot />
    </RouterLink>
    <a v-else :href="href" @click="handleClick">
      <slot />
    </a>
  </Button>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { RouteLocationRaw } from "vue-router";
import { Button } from "@/components/ui/button";

defineOptions({ inheritAttrs: false });

type ButtonVariant =
  | "primary"
  | "secondary"
  | "ghost"
  | "outline-primary"
  | "destructive"
  | "secondary-destructive"
  | "github"
  | "warning";

const props = withDefaults(
  defineProps<{
    to?: RouteLocationRaw;
    href?: string;
    variant?: ButtonVariant;
    size?: "base" | "sm";
    block?: boolean;
    iconOnly?: boolean;
    disabled?: boolean;
    type?: "button" | "submit" | "reset";
  }>(),
  {
    to: undefined,
    href: undefined,
    variant: undefined,
    size: undefined,
    block: false,
    iconOnly: false,
    disabled: false,
    type: "button",
  },
);

const emit = defineEmits<{
  click: [event: MouseEvent];
}>();

const mappedVariant = computed(() => {
  switch (props.variant) {
    case "primary":
      return "default";
    case "outline-primary":
      return "outline";
    case "destructive":
    case "secondary-destructive":
      return "destructive";
    case "ghost":
      return "ghost";
    case "github":
      return "secondary";
    case "warning":
      return "outline";
    default:
      return "secondary";
  }
});

const mappedSize = computed(() => {
  if (props.iconOnly) return "icon";
  return props.size === "sm" ? "sm" : "default";
});

function handleClick(event: MouseEvent) {
  if (props.disabled) {
    event.preventDefault();
    event.stopPropagation();
    return;
  }
  emit("click", event);
}
</script>
