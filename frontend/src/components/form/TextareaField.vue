<template>
  <BaseTextarea
    :name="name"
    :label="label"
    :placeholder="placeholder"
    :rows="rows"
    :error="error"
    :model-value="value"
    v-bind="$attrs"
    @update:model-value="handleChange"
  />
</template>

<script setup lang="ts">
import { useField } from "vee-validate";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    name: string;
    label?: string;
    placeholder?: string;
    rows?: number;
    error?: string;
  }>(),
  {
    label: undefined,
    placeholder: undefined,
    rows: 4,
    error: undefined,
  },
);

const { value, handleChange } = useField(() => props.name);
</script>
