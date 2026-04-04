<template>
  <div>
    <label v-if="label" :for="id ?? name">{{ label }}</label>
    <div
      v-if="$slots.append"
      class="input-group"
      :class="{ 'has-error': error, 'field-readonly': $attrs.readonly !== undefined }"
    >
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
  width: 100%;
  min-height: 2.25rem;
  overflow: hidden;
  border-radius: 0.5rem;
  background-color: var(--field-group-background);
  box-shadow: inset 0 0 0 1px var(--field-border-color);
  transition:
    background-color 0.15s ease,
    box-shadow 0.15s ease;

  > input {
    flex: 1;
    border-radius: 0;
    box-shadow: none !important;
    background-color: var(--panel-background);
  }

  > :not(input) {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    min-height: inherit;
    border-radius: 0;
    border: 0;
    border-left: 1px solid var(--field-border-color);
    box-shadow: none !important;
  }

  .ui-button {
    min-height: 100%;
    border-radius: 0;
  }

  &:focus-within {
    box-shadow: inset 0 0 0 1px var(--field-focus-ring-color);
  }

  &.has-error {
    box-shadow: inset 0 0 0 1px var(--error-color);
  }

  &.field-readonly {
    background-color: var(--field-readonly-background);

    > :not(input) {
      border-left-color: var(--field-border-color);
    }
  }
}

.field-readonly {
  background-color: var(--field-readonly-background);
  color: var(--text-color-muted);
  cursor: default;

  &:focus {
    box-shadow: inset 0 0 0 1px var(--field-border-color) !important;
  }
}
</style>
