<template>
    <div class="login-header">
        <h2>Sign in to your account</h2>
        <p>
            New to Shopmon?
            {{ ' ' }}
            <router-link :to="{ name: 'account.register' }">
                Create an account
            </router-link>
        </p>
    </div>

    <Alert type="info" :dismissible="false">
        <p>
            <strong>Database Migration Notice:</strong> Due to recent database updates, you may need to reset your password.
            If you're unable to log in, please use the
            <router-link :to="{ name: 'account.forgot.password' }">password reset</router-link>
            option.
        </p>
    </Alert>

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
            <router-link :to="{ name: 'account.forgot.password' }">
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

        <button
            type="button"
            class="btn btn-github btn-block"
            :disabled="isGithubLoading"
            @click="githubLogin"
        >
            <icon-mdi:github
                v-if="!isGithubLoading"
                class="icon"
                aria-hidden="true"
            />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            
            Continue with GitHub
        </button>
    </div>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm, configure } from 'vee-validate';
import { ref } from 'vue';
import * as Yup from 'yup';

import Alert from '@/components/layout/Alert.vue';
import { useAlert } from '@/composables/useAlert';
import { useReturnUrl } from '@/composables/useReturnUrl';
import { authClient } from '@/helpers/auth-client';
import { useRouter } from 'vue-router';

const router = useRouter();
const { returnUrl, clearReturnUrl } = useReturnUrl();

const isAuthenticated = ref(false);
const isGithubLoading = ref(false);
const alert = useAlert();

configure({
    validateOnBlur: false,
});

const schema = Yup.object().shape({
    email: Yup.string().required('Email is required').email(),
    password: Yup.string().required('Password is required'),
});

async function onSubmit(values: Record<string, unknown>) {
    const email = values.email as string;
    const password = values.password as string;
    const resp = await authClient.signIn.email({
        email,
        password,
    });

    if (resp.error) {
        alert.error(resp.error.message ?? 'Failed to sign in');
        return;
    }

    // redirect to previous url or default to home page
    const redirectUrl = returnUrl.value ?? '/';
    clearReturnUrl();
    router.push(redirectUrl);
}

async function webauthnLogin() {
    isAuthenticated.value = true;

    try {
        await authClient.signIn.passkey();

        // redirect to previous url or default to home page
        const redirectUrl = returnUrl.value ?? '/';
        clearReturnUrl();
        router.push(redirectUrl);
    } catch (e: unknown) {
        const { error } = useAlert();

        error(e instanceof Error ? e.message : String(e));

        isAuthenticated.value = false;
    }
}

async function githubLogin() {
    isGithubLoading.value = true;

    try {
        const redirectUrl = returnUrl.value ?? '/';
        await authClient.signIn.social({
            provider: 'github',
            callbackURL: redirectUrl,
        });
    } catch (e: unknown) {
        const { error } = useAlert();

        error(e instanceof Error ? e.message : String(e));

        isGithubLoading.value = false;
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

.btn-github {
    border-color: #24292e;
    color: #fff;
    background-color: #24292e;

    &:hover {
        background-color: #1a1e22;
        border-color: #1a1e22;
        color: #fff;
    }

    .icon {
        width: 1.25rem;
        height: 1.25rem;
    }
}

.alert {
    margin-bottom: 1.5rem;
}
</style>
