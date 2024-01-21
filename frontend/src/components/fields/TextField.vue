<template>
    <label
        v-if="label"
        :for="name"
        class="block text-sm font-medium mb-1"
    >
        {{ label }}
    </label>
    <input
        :id="name"
        :name="name"
        :type="type"
        :value="inputValue"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :class="{ 'has-error': !!errorMessage }"
        class="field"
        @input="handleChange"
        @blur="handleBlur"
    >

    <p
        v-show="errorMessage"
        class="text-red-700"
    >
        {{ errorMessage }}
    </p>
</template>

<script setup lang="ts">
import { useField } from 'vee-validate';

const props = withDefaults(defineProps<{
    name: string;
    type?: string;
    label?: string;
    placeholder?: string;
    autocomplete?: string;
    successMessage?: string;
}>(), {
    type: 'text',
    label: undefined,
    placeholder: undefined,
    autocomplete: undefined,
    successMessage: undefined,
});

const {
    value: inputValue,
    errorMessage,
    handleBlur,
    handleChange,
} = useField(() => props.name);
</script>
