<template>
  <component
    :is="componentTag"
    v-bind="componentAttrs"
    :class="buttonClasses"
    @click="handleClick"
  >
    <slot />
  </component>
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import type { RouteLocationRaw } from "vue-router";

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

type ButtonSize = "base" | "sm";

const props = withDefaults(
  defineProps<{
    to?: RouteLocationRaw;
    href?: string;
    variant?: ButtonVariant;
    size?: ButtonSize;
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

const attrs = useAttrs();

const resolvedVariant = computed<ButtonVariant>(() => {
  return props.variant ?? "secondary";
});

const resolvedSize = computed<ButtonSize>(() => {
  return props.size ?? "base";
});

const isBlock = computed(() => props.block);

const isIconOnly = computed(() => props.iconOnly);

const componentTag = computed(() => {
  if (props.to !== undefined) {
    return "RouterLink";
  }

  if (props.href) {
    return "a";
  }

  return "button";
});

const forwardedAttrs = computed(() => {
  const values = { ...attrs };
  delete values.class;
  return values;
});

const componentAttrs = computed(() => {
  if (props.to !== undefined) {
    return {
      ...forwardedAttrs.value,
      to: props.to,
      "aria-disabled": props.disabled || undefined,
      tabindex: props.disabled ? -1 : forwardedAttrs.value.tabindex,
    };
  }

  if (props.href) {
    return {
      ...forwardedAttrs.value,
      href: props.href,
      "aria-disabled": props.disabled || undefined,
      tabindex: props.disabled ? -1 : forwardedAttrs.value.tabindex,
    };
  }

  return {
    ...forwardedAttrs.value,
    type: props.type,
    disabled: props.disabled || undefined,
  };
});

const buttonClasses = computed(() => [
  "ui-button",
  `ui-button--${resolvedVariant.value}`,
  `ui-button--${resolvedSize.value}`,
  {
    "ui-button--block": isBlock.value,
    "ui-button--icon-only": isIconOnly.value,
    "ui-button--disabled": props.disabled,
  },
  attrs.class,
]);

function handleClick(event: MouseEvent) {
  if (props.disabled && componentTag.value !== "button") {
    event.preventDefault();
    event.stopPropagation();
    return;
  }

  emit("click", event);
}
</script>
