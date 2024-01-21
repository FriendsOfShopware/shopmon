<template>
    <div>
        <div class="relative">
            <label
                for="password"
                class="sr-only"
            >Password</label>
            <input
                id="password"
                :value="modelValue"
                :name="name"
                :type="passwordType"
                autocomplete="current-password"
                required
                class="field field-password"
                placeholder="Password"
                :class="{ 'has-error': error }"
                @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
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
            {{ error }}
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

defineProps<{ 
    modelValue: string | undefined,
    name: string, error: string | undefined
}>();

defineEmits<{
    'update:modelValue': [value: string]
}>();

const passwordType = ref('password');
</script>
