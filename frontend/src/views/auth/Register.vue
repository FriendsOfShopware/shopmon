<template>
    <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
        Create account
    </h4>
    <vee-form
        v-slot="{ errors, isSubmitting }"
        class="space-y-6"
        :validation-schema="schema"
        @submit="onSubmit"
    >
        <div class="space-y-2">
            <div>
                <field
                    name="displayName"
                    placeholder="Display Name"
                    type="text"
                    class="field"
                    :class="{ 'has-error': errors.displayName }"
                />
                <div class="text-red-700">
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
                <div class="text-red-700">
                    {{ errors.email }}
                </div>
            </div>

            <PasswordField
                name="password"
                :error="errors.password"
            />
        </div>

        <button
            class="btn btn-primary w-full group"
            :disabled="isSubmitting"
            type="submit"
        >
            <span class="relative -ml-7 mr-2 opacity-40 group-hover:opacity-60">
                <icon-fa6-solid:user-plus
                    v-if="!isSubmitting"
                    class="h-5 w-5"
                    aria-hidden="true"
                />
                <icon-line-md:loading-twotone-loop
                    v-else
                    class="w-5 h-5"
                />
            </span>
            Register
        </button>

        <router-link
            to="login"
            class="inline-block mt-2 center text-center w-full"
        >
            Cancel
        </router-link>
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
