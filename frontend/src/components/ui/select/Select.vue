<script setup lang="ts" generic="T extends AcceptableValue = string">
import type { AcceptableValue, SelectRootEmits, SelectRootProps } from "reka-ui";
import { SelectRoot, useForwardPropsEmits } from "reka-ui";

const props = defineProps<SelectRootProps<T> & { id?: string }>();
const emits = defineEmits<SelectRootEmits<T>>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <SelectRoot v-slot="slotProps" data-slot="select" v-bind="forwarded">
    <slot v-bind="slotProps" />
  </SelectRoot>
</template>
