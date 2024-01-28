<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import PasswordField from '@/components/layout/PasswordField.vue';

import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';

import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        password: z.string().min(8, 'Password must be at least 8 characters'),

        //.matches(/^(?=.*[0-9])/, 'Password must Contain One Number Character')
        //.matches(/^(?=.*[!@#\$%\^&\*])/, 'Password must Contain  One Special Case Character'),
    })),
});

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();

const onSubmit = handleSubmit(async (values) => {
    try {
        await authStore.confirmResetPassword(route.params.token as string, values.password);
        alertStore.success('Password has been reset. You will be redirected to login page in 2 seconds.');

        setTimeout(() => {
            router.push('/account/login');
        }, 2000);
    } catch (error: any) {
        alertStore.error(error.message);
    }
});

const isLoading = ref(true);
const tokenFound = ref(false);

onMounted(async () => {
    tokenFound.value = await authStore.resetAvailable(route.params.token as string);
    isLoading.value = false;
});

function goToResend() {
    router.push('/account/forgot-password');
}
</script>

<template>
    <div>
        <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
            Change Password
        </h4>
    </div>
    <div
        v-if="isLoading"
        class="rounded-md bg-blue-50 p-4 border border-sky-200 dark:bg-gray-900 dark:border-sky-400"
    >
        <div class="flex">
            <icon-fa6-solid:circle-info
                class="h-5 w-5 text-sky-500"
                aria-hidden="true"
            />
            <div class="ml-3 flex-1 md:flex md:justify-between ">
                <p class="text-sky-900 dark:text-sky-600">
                    Loading...
                </p>
            </div>
        </div>
    </div>

    <div v-else>
        <form
            v-if="tokenFound"
            class="space-y-4"
            @submit="onSubmit"
        >
            <password-field
                v-model="values.password"
                name="password"
                :error="errors.password"
            />

            <button
                class="w-full group btn btn-primary"
                :disabled="isSubmitting"
                type="submit"
            >
                <span class="relative -ml-7 mr-2 opacity-40 group-hover:opacity-60">
                    <icon-fa6-solid:key
                        v-if="!isSubmitting"
                        class="w-5 h-5"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="w-5 h-5"
                    />
                </span>
                Change Password
            </button>

            <router-link
                to="login"
                class="inline-block mt-2 center text-center w-full"
            >
                Cancel
            </router-link>
        </form>

        <div v-else>
            <div
                class="rounded-md bg-red-50 p-4 border border-red-200
                dark:bg-red-900 dark:bg-opacity-30 dark:border-red-400"
            >
                <div class="flex">
                    <div class="flex-shrink-0">
                        <icon-fa6-solid:circle-xmark
                            class="h-5 w-5 text-red-600 dark:text-red-400"
                            aria-hidden="true"
                        />
                    </div>
                    <div class="ml-3">
                        <h3 class="font-medium text-red-900 dark:text-red-500">
                            Invalid Token
                        </h3>
                        <div class="mt-1 text-red-800 dark:text-red-400">
                            <p>It looks like your Token is expired.</p>
                        </div>
                    </div>
                </div>
            </div>

            <button
                type="button"
                class="mt-4 btn btn-primary w-full"
                @click="goToResend"
            >
                <span class="relative -ml-7 mr-2 opacity-40 group-hover:opacity-60">
                    <icon-fa6-solid:envelope
                        class="h-5 w-5"
                        aria-hidden="true"
                    />
                </span>
                Resend email
            </button>
        </div>
    </div>
</template>
