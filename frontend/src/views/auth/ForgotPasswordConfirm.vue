<template>
    <div class="login-header">
        <h2>
            Change Password
        </h2>
    </div>


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

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';

import { Form as VeeForm } from 'vee-validate';
import * as Yup from 'yup';

import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';

const schema = Yup.object().shape({
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters')
        .matches(/^(?=.*[0-9])/, 'Password must Contain One Number Character')
        .matches(
            /^(?=.*[!@#$%^&*])/,
            'Password must Contain  One Special Case Character',
        ),
});

const route = useRoute();
const router = useRouter();
const { success, error } = useAlert();

async function onSubmit(values: Record<string, unknown>): Promise<void> {
    const password = values.password as string;
    try {
        await authClient.resetPassword({
            token: route.params.token as string,
            newPassword: password,
        });

        success(
            'Password has been resetted. You will be redirected to login page in 2 seconds.',
        );

        setTimeout(() => {
            router.push({ name: 'account.login' });
        }, 2000);
    } catch (err) {
        error(err instanceof Error ? err.message : String(err));
    }
}
</script>

<style scoped>
.login-resend-mail-container {
    text-align: unset;
}
</style>
