<template>
    <div class="login-header">
        <h2>Sign in to your account</h2>
        <p v-if="!disableRegistration">
            New to Shopmon?
            {{ ' ' }}
            <router-link to="register">
                Create an account
            </router-link>
        </p>
    </div>

    <vee-form
        v-slot="{ errors, isSubmitting }"
        class="login-form-container"
        :validation-schema="schema"
        @submit="onSubmit"
    >
        <div class="login-form-group">
            <div>
                <label for="email-address" class="sr-only">Email address</label>
                <field
                    id="email-address"
                    name="email"
                    type="email"
                    autocomplete="email"
                    required=""
                    class="field"
                    placeholder="Email address"
                    :class="{ 'has-error': errors.email }"
                />

                <div class="field-error-message">{{ errors.email }}</div>
            </div>

            <PasswordField
                name="password"
                placeholder="Password"
                :error="errors.password"
            />
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="isSubmitting">
            <icon-fa6-solid:key
                v-if="!isSubmitting"
                class="icon"
                aria-hidden="true"
            />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            Sign in
        </button>

        <div>
            <router-link to="/account/forgot-password">
                Forgot your password?
            </router-link>
        </div>
    </vee-form>

    <div class="passkey-container">
        <div class="text-divider">Or</div>

        <button
            type="button"
            class="btn btn-primary btn-block btn-passkey"
            :disabled="isAuthenticated"
            @click="webauthnLogin"
        >
            <icon-material-symbols:passkey
                v-if="!isAuthenticated"
                class="icon icon-passkey"
                aria-hidden="true"
            />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            
            Login using Passkey
        </button>
    </div>
</template>

<script setup lang="ts">
import { Form as VeeForm, Field, configure } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import { useAuthStore } from '@/stores/auth.store';
import { useRouter } from 'vue-router';
import { useAlertStore } from '@/stores/alert.store';

const router = useRouter();
const authStore = useAuthStore();

const isAuthenticated = ref(false);

configure({
    validateOnBlur: false,
});

const schema = Yup.object().shape({
    email: Yup.string().required('Email is required').email(),
    password: Yup.string().required('Password is required'),
});

const disableRegistration = import.meta.env.VITE_DISABLE_REGISTRATION;

async function onSubmit(values: any) {
    const { email, password } = values;
    try {
        await authStore.login(email, password);

        // redirect to previous url or default to home page
        router.push(authStore.returnUrl || '/');
    } catch (e: unknown) {
        const alertStore = useAlertStore();

        alertStore.error(e);
    }
}

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

<style scoped>
.passkey-container {
    margin-top: 2rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;

    .text-divider {
        color: #6b7280;
    }
}
</style>
