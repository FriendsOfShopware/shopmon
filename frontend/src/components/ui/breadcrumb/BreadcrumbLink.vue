<script lang="ts" setup>
import type { PrimitiveProps } from "reka-ui";
import type { AnchorHTMLAttributes, HTMLAttributes } from "vue";
import { reactiveOmit } from "@vueuse/core";
import { Primitive } from "reka-ui";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<
    PrimitiveProps &
      /* @vue-ignore */ Omit<AnchorHTMLAttributes, "class"> & {
        class?: HTMLAttributes["class"];
      }
  >(),
  { as: "a" },
);

const forwarded = reactiveOmit(props, "class");
</script>

<template>
  <Primitive
    data-slot="breadcrumb-link"
    v-bind="forwarded"
    :class="cn('hover:text-foreground transition-colors', props.class)"
  >
    <slot />
  </Primitive>
</template>
