<template>
  <BaseSelect
    :name="name"
    :label="label"
    :error="error"
    :model-value="value"
    v-bind="$attrs"
    @update:model-value="handleChange"
  >
    <slot />
  </BaseSelect>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    name: string;
    label?: string;
    error?: string;
  }>(),
  {
    label: undefined,
    error: undefined,
  },
);

const { value, handleChange } = useField(() => props.name);
</script>
