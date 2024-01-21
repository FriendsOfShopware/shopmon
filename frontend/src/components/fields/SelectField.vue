<template>
    <label
        v-if="label"
        :for="name"
        class="block text-sm font-medium mb-1"
    >
        {{ label }}
    </label>
    <select
        :id="name"
        :name="name"
        :type="type"
        :value="inputValue"
        :class="{ 'has-error': !!errorMessage }"
        class="field"
        @input="handleChange"
        @blur="handleBlur"
    >
        <slot />
    </select>

    <p
        v-show="errorMessage"
        class="text-red-700"
    >
        {{ errorMessage }}
    </p>
</template>

<script setup lang="ts">
import { useField } from 'vee-validate';

const props = defineProps<{
    name: string;
    type?: string;
    label?: string;
    placeholder?: string;
    autocomplete?: string;
    successMessage?: string;
}>();

const {
    value: inputValue,
    errorMessage,
    handleBlur,
    handleChange,
} = useField(() => props.name);
</script>
