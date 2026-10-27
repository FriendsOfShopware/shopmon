<script setup lang="ts">
import type { PrimitiveProps } from "reka-ui";
import type { ButtonHTMLAttributes, HTMLAttributes } from "vue";
import type { ButtonVariants } from ".";
import { reactiveOmit } from "@vueuse/core";
import { Primitive } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from ".";

const props = withDefaults(
  defineProps<
    PrimitiveProps &
      /* @vue-ignore */ Omit<ButtonHTMLAttributes, "class"> & {
        variant?: ButtonVariants["variant"];
        size?: ButtonVariants["size"];
        class?: HTMLAttributes["class"];
      }
  >(),
  { as: "button" },
);

const forwarded = reactiveOmit(props, "variant", "size", "class");
</script>

<template>
  <Primitive
    data-slot="button"
    :data-variant="variant"
    :data-size="size"
    v-bind="forwarded"
    :class="cn(buttonVariants({ variant, size }), props.class)"
  >
    <slot />
  </Primitive>
</template>
