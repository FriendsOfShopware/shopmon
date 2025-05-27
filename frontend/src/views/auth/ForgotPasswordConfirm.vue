<template>
    <div class="login-header">
        <h2>
            Change Password
        </h2>
    </div>

    <Alert type="info" v-if="isLoading">
        Loading...
    </Alert>

    <template v-else>
        <vee-form
            v-if="tokenFound"
            v-slot="{ errors, isSubmitting }"
            class="login-form-container"
            :validation-schema="schema"
            @submit="onSubmit"
        >
            <password-field
                name="password"
                placeholder="Password"
                :error="errors.password"
            />

            <button
                class="btn btn-primary btn-block"
                :disabled="isSubmitting"
                type="submit"
            >
                <icon-fa6-solid:key
                    v-if="isSubmitting"
                    class="icon"
                    aria-hidden="true"
                />
                <icon-line-md:loading-twotone-loop
                    v-else
                    class="icon"
                />
                Change Password
            </button>

            <div>
                <router-link to="login">
                    Cancel
                </router-link>
            </div>            
        </vee-form>

        <div class="login-form-container login-resend-mail-container" v-else>
            <Alert type="error">
                <strong>Invalid Token</strong>
                <p>It looks like your Token is expired.</p>
            </Alert>

            <button
                type="button"
                class="btn btn-primary btn-block"
                @click="goToResend"
            >
                <icon-fa6-solid:envelope
                    class="icon"
                    aria-hidden="true"
                />
                Resend email
            </button>
        </div>
    </template>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Form as VeeForm } from 'vee-validate';
import * as Yup from 'yup';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';

const schema = Yup.object().shape({
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters')
        .matches(/^(?=.*[0-9])/, 'Password must Contain One Number Character')
        .matches(/^(?=.*[!@#\$%\^&\*])/, 'Password must Contain  One Special Case Character'),
});

const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();

async function onSubmit(values: any): Promise<void> {
    try {
        await authStore.confirmResetPassword(route.params.token as string, values.password);
        alertStore.success('Password has been resetted. You will be redirected to login page in 2 seconds.');

        setTimeout(() => {
            router.push('/account/login');
        }, 2000);
    } catch (error: Error) {
        alertStore.error(error.message);
    }
}

const isLoading = ref(true);
const tokenFound = ref(false);

onMounted(async() => {
    tokenFound.value = await authStore.resetAvailable(route.params.token as string);
    isLoading.value = false;
});

function goToResend() {
    router.push('/account/forgot-password');
}
</script>

<style scoped>
.login-resend-mail-container {
    text-align: unset;
}
</style>
