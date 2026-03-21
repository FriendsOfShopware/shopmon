<template>
  <div>
    <label v-if="label" :for="id ?? name">{{ label }}</label>
    <div v-if="$slots.append" class="input-group">
      <input
        :id="id ?? name"
        :type="type"
        :name="name"
        :autocomplete="autocomplete"
        :placeholder="placeholder"
        class="field"
        :class="{ 'has-error': error, 'field-readonly': $attrs.readonly !== undefined }"
        :value="modelValue"
        v-bind="$attrs"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <slot name="append" />
    </div>
    <input
      v-else
      :id="id ?? name"
      :type="type"
      :name="name"
      :autocomplete="autocomplete"
      :placeholder="placeholder"
      class="field"
      :class="{ 'has-error': error, 'field-readonly': $attrs.readonly !== undefined }"
      :value="modelValue"
      v-bind="$attrs"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
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
    type?: string;
    autocomplete?: string;
    placeholder?: string;
    error?: string;
    modelValue?: string;
  }>(),
  {
    name: undefined,
    id: undefined,
    label: undefined,
    type: "text",
    autocomplete: undefined,
    placeholder: undefined,
    error: undefined,
    modelValue: undefined,
  },
);

defineEmits<{
  "update:modelValue": [value: string];
}>();
</script>

<style>
.input-group {
  display: flex;
  gap: 0.5rem;

  > input {
    flex: 1;
  }
}

.field-readonly {
  background-color: var(--panel-background-color);
  color: var(--text-color-muted);
  cursor: default;

  &:focus {
    box-shadow: none;
    border-color: var(--input-border-color);
  }
}
</style>
