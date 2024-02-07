<script setup lang="ts">
import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

import { ref } from 'vue';

import { useAuthStore } from '@/stores/auth.store';
import { useRouter } from 'vue-router';
import { useAlertStore } from '@/stores/alert.store';

const router = useRouter();
const authStore = useAuthStore();

const isAuthenticated = ref(false);

const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        email: z.string().email('Email is required'),
        password: z.string().min(1, 'Password is required'),
    })),
});

const disableRegistration = import.meta.env.VITE_DISABLE_REGISTRATION;

const onSubmit = handleSubmit(async (values) =>  {
    const { email, password } = values;
    try {
        await authStore.login(email, password);

        // redirect to previous url or default to home page
        router.push(authStore.returnUrl || '/');
    } catch (e: any) {
        const alertStore = useAlertStore();

        alertStore.error(e);
    }
});

async function webauthnLogin() {
    isAuthenticated.value = true;

    try {
        await authStore.loginWithPasskey();

        // redirect to previous url or default to home page
        router.push(authStore.returnUrl || '/');
    } catch (e: unknown) {
        const alertStore = useAlertStore();

        alertStore.error(e);

        isAuthenticated.value = false;
    }
}

</script>

<template>
    <div v-if="!authStore.isAuthenticated">
        <div>
            <h2 class="text-center text-3xl tracking-tight font-bold">
                Sign in to your account
            </h2>
            <p
                v-if="!disableRegistration"
                class="mt-2 text-center text-sm text-gray-600 dark:text-neutral-500"
            >
                New to Shopmon?
                {{ ' ' }}
                <router-link
                    to="register"
                    class="font-medium"
                >
                    Create an account
                </router-link>
            </p>
        </div>

        <form
            class="mt-8 space-y-6"
            @submit="onSubmit"
        >
            <div class="space-y-2">
                <div>
                    <label
                        for="email-address"
                        class="sr-only"
                    >Email address</label>
                    <input
                        id="email-address"
                        v-model="values.email"
                        name="email"
                        type="email"
                        autocomplete="email"
                        required
                        class="field"
                        placeholder="Email address"
                        :class="{ 'has-error': errors.email }"
                    >

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
                type="submit"
                class="group w-full btn btn-primary"
                :disabled="isSubmitting"
            >
                <span class="relative -ml-8 mr-3 opacity-40 group-hover:opacity-60 w-4 h-4">
                    <icon-fa6-solid:key
                        v-if="!isSubmitting"
                        class="relative -top-[3px] h-5 w-5"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="w-5 h-5"
                    />
                </span>
                Sign in
            </button>

            <div class="text-center">
                <router-link
                    to="/account/forgot-password"
                    class="font-medium"
                >
                    Forgot your password?
                </router-link>
            </div>
        </form>

        <div class="mt-6 space-y-6">
            <div class="text-divider text-gray-600">
                Or
            </div>

            <button
                type="button"
                class="group w-full btn btn-primary"
                :disabled="isAuthenticated"
                @click="webauthnLogin"
            >
                <span class="relative -ml-8 mr-3 opacity-40 group-hover:opacity-60 h-5 w-5">
                    <icon-material-symbols:passkey
                        v-if="!isAuthenticated"
                        class="relative -top-[5px] h-7 w-7"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="w-5 h-5"
                    />
                </span>
                Login using Passkey
            </button>
        </div>
    </div>
</template>
