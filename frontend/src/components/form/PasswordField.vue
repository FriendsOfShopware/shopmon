<template>
    <div>
        <label :for="name" :class="{ 'sr-only': !label }">{{ passwordLabel }}</label>
        <div class="password-wrapper">
            <Field 
                :id="name" 
                :name="name" 
                :type="passwordType"
                class="field field-password" 
                :placeholder="placeholder" 
                :class="{ 'has-error': error }"
             />

            <div class="password-toggle">
                <icon-fa6-solid:eye v-if="passwordType == 'password'" class="password-toggle-icon"
                    @click="passwordType = 'text'" />
                <icon-fa6-solid:eye-slash v-else class="password-toggle-icon" @click="passwordType = 'password'" />
            </div>
        </div>
        <div class="error-message">{{ error }}</div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Field } from 'vee-validate';

const props = defineProps<{
    label?: string;
    placeholder?: string;
    name: string;
    error: string | undefined;
}>();

const passwordType = ref('password');
const passwordLabel = props.label || 'Password';
</script>

<style>
.password-wrapper {
    position: relative;
}

.password-field {
    padding: 8px 12px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 16px;
    width: 100%;
    box-sizing: border-box;
}

.password-field.has-error {
    border-color: #dc3545;
}

.password-toggle {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    padding-right: 8px;
    cursor: pointer;
    z-index: 10;
}

.password-toggle-icon {
    width: 18px;
}

.error-message {
    color: #dc3545;
}
</style>
