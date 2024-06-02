<template>
    <div class="login-header">
        <h2>
            Forgot password
        </h2>
        <p>We will send you a confirmation email. Click on the link in it to change your password.</p>
    </div>

    <vee-form
        v-slot="{ errors, isSubmitting }"
        class="login-form-container"
        :validation-schema="schema"
        @submit="onSubmit"
    >
        <div>
            <field
                name="email"
                placeholder="Email address"
                type="text"
                class="field"
                :class="{ 'has-error': errors.email }"
            />
            <div class="field-error-message">
                {{ errors.email }}
            </div>
        </div>

        <button
            class="btn btn-primary btn-block"
            :disabled="isSubmitting"
            type="submit"
        >
            <icon-fa6-solid:envelope
                v-if="!isSubmitting"
                class="icon"
                aria-hidden="true"
            />
            <icon-line-md:loading-twotone-loop
                v-else
                class="icon"
            />
            Send email
        </button>

        <div>
            <router-link to="login">
                Cancel
            </router-link>
        </div>
        
    </vee-form>
</template>

<script setup lang="ts">
import { Form as VeeForm, Field } from 'vee-validate';
import * as Yup from 'yup';

import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
    email: Yup.string().required('Email is required'),
});

async function onSubmit(values): Promise<void> {
    const usersStore = useAuthStore();
    const alertStore = useAlertStore();

    try {
        await usersStore.resetPassword(values.email);
        alertStore.success('Password reset email sent');
    } catch (error: Error) {
        alertStore.error(error.message);
    }
}
</script>

<style scoped>
.login-header {
    p {
        text-align: left;
    }
}
</style>
