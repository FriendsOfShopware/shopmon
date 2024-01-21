<template>
    <div>
        <div class="relative">
            <label
                for="password"
                class="sr-only"
            >
                Password
            </label>
            <input
                id="password"
                :value="inputValue"
                :name="name"
                :type="passwordType"
                :autocomplete="autocomplete || 'current-password'"
                required
                class="field field-password"
                placeholder="Password"
                :class="{ 'has-error': !!errorMessage }"
                @input="handleChange"
                @blur="handleBlur"
            >
            <div class="absolute right-0 inset-y-0 flex items-center pr-3 cursor-pointer z-10">
                <icon-fa6-solid:eye
                    v-if="passwordType == 'password'"
                    class="w-[18px]"
                    @click="passwordType = 'text'"
                />
                <icon-fa6-solid:eye-slash
                    v-else
                    class="w-[18px]"
                    @click="passwordType = 'password'"
                />
            </div>
        </div>

        <div class="text-red-700">
            {{ errorMessage }}
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useField } from 'vee-validate';

const props = defineProps<{
    name: string;
    label?: string;
    successMessage?: string;
    placeholder?: string;
    autocomplete?: string;
}>();

const {
    value: inputValue,
    errorMessage,
    handleBlur,
    handleChange,
} = useField(() => props.name);

const passwordType = ref('password');
</script>
