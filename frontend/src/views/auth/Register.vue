<template>
    <div class="login-header">
        <h2>Create account</h2>
    </div>
    
    <vee-form
        v-slot="{ errors, isSubmitting }"
        class="login-form-container"
        :validation-schema="schema"
        @submit="onSubmit"
    >
        <div class="login-form-group">
            <div>
                <field
                    name="displayName"
                    placeholder="Display Name"
                    type="text"
                    class="field"
                    :class="{ 'has-error': errors.displayName }"
                />
                <div class="field-error-message">
                    {{ errors.displayName }}
                </div>
            </div>

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

            <PasswordField
                name="password"
                placeholder="Password"
                :error="errors.password"
            />
        </div>

        <button
            class="btn btn-primary btn-block"
            :disabled="isSubmitting"
            type="submit"
        >
            <icon-fa6-solid:user-plus
                v-if="!isSubmitting"
                class="icon"
                aria-hidden="true"
            />
            <icon-line-md:loading-twotone-loop
                v-else
                class="icon"
            />
            Register
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

import { router } from '@/router';

import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
    displayName: Yup.string().required('Display Name is required'),
    email: Yup.string().required('Email is required'),
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters')
        .matches(/^(?=.*[0-9])/, 'Password must Contain One Number Character')
        .matches(/^(?=.*[!@#\$%\^&\*])/, 'Password must Contain  One Special Case Character'),
});

async function onSubmit(values: any) {
    const authStore = useAuthStore();
    const alertStore = useAlertStore();
    try {
        await authStore.register(values);
        await router.push('/account/login');
        alertStore.success('Registration successful. Please check your mailbox and confirm your email address.');
    } catch (error: any) {
        alertStore.error(error);
    }
}
</script>
