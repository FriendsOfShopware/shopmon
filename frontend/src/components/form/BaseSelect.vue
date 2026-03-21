<template>
  <div>
    <label v-if="label" :for="id ?? name">{{ label }}</label>
    <select
      :id="id ?? name"
      :name="name"
      class="field"
      :class="{ 'has-error': error }"
      :value="modelValue"
      v-bind="$attrs"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot />
    </select>
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
    error?: string;
    modelValue?: string | number;
  }>(),
  {
    name: undefined,
    id: undefined,
    label: undefined,
    error: undefined,
    modelValue: undefined,
  },
);

defineEmits<{
  "update:modelValue": [value: string];
}>();
</script>
