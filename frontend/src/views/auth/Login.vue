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
        
        <div class="sso-section">
            <div class="text-divider">Enterprise SSO</div>
            
            <div class="sso-input-group">
                <input
                    v-model="ssoEmail"
                    type="email"
                    class="field"
                    placeholder="Enter your work email"
                    @keyup.enter="ssoLogin"
                />
                <button
                    type="button"
                    class="btn btn-secondary"
                    :disabled="isSSOLoading || !ssoEmail"
                    @click="ssoLogin"
                >
                    <icon-fa6-solid:arrow-right
                        v-if="!isSSOLoading"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop v-else class="icon" />
                </button>
            </div>
            <p class="sso-help">Use your company email to sign in with SSO</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm, configure } from 'vee-validate';
import { ref, watch } from 'vue';
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
const isSSOLoading = ref(false);
const ssoEmail = ref('');
const alert = useAlert();

configure({
    validateOnBlur: false,
});

const schema = Yup.object().shape({
    email: Yup.string().required('Email is required').email(),
    password: Yup.string().required('Password is required'),
});

function goToHome() {
    const sess = authClient.useSession()

    watch(sess, (user) => {
        if (user.data?.user) {
            const redirectUrl = returnUrl.value ?? '/';
            clearReturnUrl();
            router.push(redirectUrl);
        }
    });
}

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

    goToHome();
}

async function webauthnLogin() {
    isAuthenticated.value = true;

    try {
        const resp = await authClient.signIn.passkey();

        if (resp?.error) {
            const { error } = useAlert();
            error(resp.error.message ?? 'Failed to sign in with Passkey');
            isAuthenticated.value = false;
            return;
        }

        goToHome();
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
        const resp = await authClient.signIn.social({
            provider: 'github',
            callbackURL: redirectUrl,
        });
        if (resp.error) {
            const { error } = useAlert();
            error(resp.error.message ?? 'Failed to sign in with GitHub');
            isGithubLoading.value = false;
            return;
        }

        goToHome();
    } catch (e: unknown) {
        const { error } = useAlert();

        error(e instanceof Error ? e.message : String(e));

        isGithubLoading.value = false;
    }
}

async function ssoLogin() {
    if (!ssoEmail.value) {
        alert.error('Please enter your work email');
        return;
    }

    isSSOLoading.value = true;

    try {
        const redirectUrl = returnUrl.value ?? '/';
        const result = await authClient.signIn.sso({
            email: ssoEmail.value,
            callbackURL: `${window.location.origin}${redirectUrl}`,
        });

        if (result.error) {
            if (result.error.code === 'SSO_PROVIDER_NOT_FOUND') {
                alert.error('No SSO provider found for this email domain');
            } else {
                alert.error(result.error.message ?? 'SSO login failed');
            }
        } else {
            goToHome();
        }
    } catch (e: unknown) {
        alert.error(e instanceof Error ? e.message : 'SSO login failed');
    } finally {
        isSSOLoading.value = false;
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

.sso-section {
    margin-top: 1rem;
}

.sso-input-group {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;

    input {
        flex: 1;
    }

    button {
        padding-left: 1rem;
        padding-right: 1rem;
    }
}

.sso-help {
    margin-top: 0.5rem;
    font-size: 0.875rem;
    color: var(--text-color-muted);
    text-align: center;
}
</style>
