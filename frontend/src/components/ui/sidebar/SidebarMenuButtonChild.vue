<script setup lang="ts">
import type { PrimitiveProps } from "reka-ui";
import type { ButtonHTMLAttributes, HTMLAttributes } from "vue";
import type { SidebarMenuButtonVariants } from ".";
import { reactiveOmit } from "@vueuse/core";
import { Primitive } from "reka-ui";
import { cn } from "@/lib/utils";
import { sidebarMenuButtonVariants } from ".";

export interface SidebarMenuButtonProps
  extends PrimitiveProps, /* @vue-ignore */ Omit<ButtonHTMLAttributes, "class"> {
  variant?: SidebarMenuButtonVariants["variant"];
  size?: SidebarMenuButtonVariants["size"];
  isActive?: boolean;
  class?: HTMLAttributes["class"];
}

const props = withDefaults(defineProps<SidebarMenuButtonProps>(), {
  as: "button",
  variant: "default",
  size: "default",
});

const forwarded = reactiveOmit(props, "variant", "size", "isActive", "class");
</script>

<template>
  <Primitive
    data-slot="sidebar-menu-button"
    data-sidebar="menu-button"
    :data-size="size"
    :data-active="isActive"
    v-bind="forwarded"
    :class="cn(sidebarMenuButtonVariants({ variant, size }), props.class)"
  >
    <slot />
  </Primitive>
</template>
