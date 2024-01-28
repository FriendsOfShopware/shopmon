<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        email: z.string().min(1, 'Email is required'),
    })),
});

const onSubmit = handleSubmit(async (values) => {
    const usersStore = useAuthStore();
    const alertStore = useAlertStore();

    try {
        await usersStore.resetPassword(values.email);
        alertStore.success('Password reset email sent');
    } catch (error: any) {
        alertStore.error(error.message);
    }
});
</script>

<template>
    <div>
        <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
            Forgot password
        </h4>
        <p>We will send you a confirmation email. Click on the link in it to change your password.</p>
    </div>

    <form
        class="space-y-4"
        @submit="onSubmit"
    >
        <div>
            <input
                v-model="values.email"
                name="email"
                placeholder="Email address"
                type="text"
                class="field"
                :class="{ 'is-invalid': errors.email }"
            >
            <div class="text-red-700">
                {{ errors.email }}
            </div>
        </div>

        <button
            class="btn btn-primary w-full group"
            :disabled="isSubmitting"
            type="submit"
        >
            <span class="relative -ml-7 mr-2 opacity-40 group-hover:opacity-60">
                <icon-fa6-solid:envelope
                    v-if="!isSubmitting"
                    class="h-5 w-5"
                    aria-hidden="true"
                />
                <icon-line-md:loading-twotone-loop
                    v-else
                    class="w-5 h-5"
                />
            </span>
            Send email
        </button>

        <router-link
            to="login"
            class="inline-block center text-center w-full"
        >
            Cancel
        </router-link>
    </form>
</template>
