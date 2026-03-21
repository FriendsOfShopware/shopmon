<template>
  <BaseInput
    :name="name"
    :label="label"
    :type="type"
    :autocomplete="autocomplete"
    :placeholder="placeholder"
    :error="error"
    :model-value="value"
    v-bind="$attrs"
    @update:model-value="onChange"
  >
    <template v-if="$slots.append" #append>
      <slot name="append" />
    </template>
  </BaseInput>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    name: string;
    label?: string;
    type?: string;
    autocomplete?: string;
    placeholder?: string;
    error?: string;
  }>(),
  {
    label: undefined,
    type: "text",
    autocomplete: undefined,
    placeholder: undefined,
    error: undefined,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const { value, handleChange } = useField(() => props.name);

function onChange(val: string) {
  handleChange(val);
  emit("update:modelValue", val);
}
</script>
