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
                <router-link :to="{ name: 'account.login' }">
                    Cancel
                </router-link>
            </div>            
        </vee-form>
    </template>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Form as VeeForm } from 'vee-validate';
import * as Yup from 'yup';

import { authClient } from '@/helpers/auth-client';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters')
        .matches(/^(?=.*[0-9])/, 'Password must Contain One Number Character')
        .matches(
            /^(?=.*[!@#\$%\^&\*])/,
            'Password must Contain  One Special Case Character',
        ),
});

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();

async function onSubmit(values: { password: string }): Promise<void> {
    try {
        await authClient.resetPassword({
            token: route.params.token as string,
            newPassword: values.password,
        });

        alertStore.success(
            'Password has been resetted. You will be redirected to login page in 2 seconds.',
        );

        setTimeout(() => {
            router.push({ name: 'account.login' });
        }, 2000);
    } catch (error) {
        alertStore.error(
            error instanceof Error ? error.message : String(error),
        );
    }
}

const isLoading = ref(true);

function goToResend() {
    router.push({ name: 'account.forgot.password' });
}
</script>

<style scoped>
.login-resend-mail-container {
    text-align: unset;
}
</style>
