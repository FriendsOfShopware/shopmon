<template>
  <div>
    <label v-if="label" :for="id ?? name">{{ label }}</label>
    <textarea
      :id="id ?? name"
      :name="name"
      :rows="rows"
      :placeholder="placeholder"
      class="field"
      :class="{ 'has-error': error }"
      :value="modelValue"
      v-bind="$attrs"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
    <div v-if="error" class="field-error-message">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false });

withDefaults(
  defineProps<{
    name?: string;
    id?: string;
    label?: string;
    placeholder?: string;
    rows?: number;
    error?: string;
    modelValue?: string;
  }>(),
  {
    name: undefined,
    id: undefined,
    label: undefined,
    placeholder: undefined,
    rows: 4,
    error: undefined,
    modelValue: undefined,
  },
);

defineEmits<{
  "update:modelValue": [value: string];
}>();
</script>
