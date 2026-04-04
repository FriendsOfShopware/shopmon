<template>
  <div class="login-header">
    <h2>{{ $t("auth.signIn") }}</h2>
    <p>
      {{ $t("auth.newToShopmon") }}
      {{ " " }}
      <router-link :to="{ name: 'account.register' }"> {{ $t("auth.createAccount") }} </router-link>
    </p>
  </div>

  <vee-form
    v-slot="{ errors, isSubmitting }"
    class="login-form-container"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div class="login-form-group">
      <InputField
        name="email"
        type="email"
        autocomplete="email"
        :placeholder="$t('common.emailAddress')"
        required
        :error="errors.email"
      />

      <PasswordField
        name="password"
        :placeholder="$t('common.password')"
        :error="errors.password"
      />
    </div>

    <button type="submit" class="btn btn-primary btn-block" :disabled="isSubmitting">
      <icon-fa6-solid:right-to-bracket v-if="!isSubmitting" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t("auth.signInButton") }}
    </button>

    <div>
      <router-link :to="{ name: 'account.forgot.password' }">
        {{ $t("auth.forgotPassword") }}
      </router-link>
    </div>
  </vee-form>

  <div class="passkey-container">
    <div class="text-divider">{{ $t("common.or") }}</div>

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
      {{ $t("auth.loginPasskey") }}
    </button>

    <button
      type="button"
      class="btn btn-github btn-block"
      :disabled="isGithubLoading"
      @click="githubLogin"
    >
      <icon-mdi:github v-if="!isGithubLoading" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t("auth.continueGithub") }}
    </button>

    <div class="sso-section">
      <div class="text-divider">{{ $t("auth.enterpriseSSO") }}</div>

      <BaseInput
        v-model="ssoEmail"
        type="email"
        :placeholder="$t('auth.enterWorkEmail')"
        @keyup.enter="ssoLogin"
      >
        <template #append>
          <button
            type="button"
            class="btn btn-secondary icon-only"
            :disabled="isSSOLoading || !ssoEmail"
            @click="ssoLogin"
          >
            <icon-fa6-solid:arrow-right v-if="!isSSOLoading" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
          </button>
        </template>
      </BaseInput>
      <p class="sso-help">{{ $t("auth.ssoHelp") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm, configure } from "vee-validate";
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { useReturnUrl } from "@/composables/useReturnUrl";
import { fetchSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";
import { useRouter } from "vue-router";
import { startAuthentication } from "@simplewebauthn/browser";

const { t } = useI18n();
const router = useRouter();
const { returnUrl, clearReturnUrl } = useReturnUrl();

const isAuthenticated = ref(false);
const isGithubLoading = ref(false);
const isSSOLoading = ref(false);
const ssoEmail = ref("");
const alert = useAlert();

configure({
  validateOnBlur: false,
});

const schema = Yup.object().shape({
  email: Yup.string().required(t("validation.emailRequired")).email(),
  password: Yup.string().required(t("validation.passwordRequired")),
});

async function goToDashboard() {
  await fetchSession();
  const redirectUrl = returnUrl.value ?? "/app/dashboard";
  clearReturnUrl();
  await router.push(redirectUrl);
}

async function onSubmit(values: Record<string, unknown>) {
  const email = values.email as string;
  const password = values.password as string;
  const { data, error } = await api.POST("/auth/sign-in/email", {
    body: { email, password },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? t("auth.failedSignIn"));
    return;
  }

  if ((data as { token?: string })?.token) {
    setToken((data as { token: string }).token);
  }

  await goToDashboard();
}

async function webauthnLogin() {
  isAuthenticated.value = true;

  try {
    // Get login options from server
    const { data: optionsData, error: optionsError } = await api.POST(
      "/auth/passkey/login-options",
    );

    if (optionsError || !optionsData) {
      alert.error(t("auth.failedSignInPasskey"));
      isAuthenticated.value = false;
      return;
    }

    const { options, challengeKey } = optionsData as {
      options: { publicKey: Parameters<typeof startAuthentication>[0]["optionsJSON"] };
      challengeKey: string;
    };

    // Run WebAuthn browser API
    const assertion = await startAuthentication({ optionsJSON: options.publicKey });

    // Send assertion to server
    const { data: loginData, error } = await api.POST("/auth/passkey/login", {
      body: { challengeKey, ...assertion } as never,
    });

    if (error) {
      alert.error((error as { message?: string }).message ?? "Failed to sign in with Passkey");
      isAuthenticated.value = false;
      return;
    }

    if ((loginData as unknown as { token?: string })?.token) {
      setToken((loginData as unknown as { token: string }).token);
    }

    await goToDashboard();
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : String(e));
    isAuthenticated.value = false;
  }
}

async function githubLogin() {
  isGithubLoading.value = true;

  try {
    const redirectUrl = returnUrl.value ?? "/";
    const { data, error } = await api.POST("/auth/sign-in/social", {
      body: { provider: "github", callbackURL: redirectUrl },
    });

    if (error || !data?.url) {
      alert.error(t("auth.failedSignInGithub"));
      isGithubLoading.value = false;
      return;
    }

    // Redirect to the OAuth provider
    window.location.href = data.url;
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : String(e));
    isGithubLoading.value = false;
  }
}

async function ssoLogin() {
  if (!ssoEmail.value) {
    alert.error(t("auth.pleaseEnterWorkEmail"));
    return;
  }

  isSSOLoading.value = true;

  try {
    const redirectUrl = returnUrl.value ?? "/";
    const { data, error } = await api.POST("/auth/sign-in/sso", {
      body: {
        email: ssoEmail.value,
        callbackURL: `${window.location.origin}${redirectUrl}`,
      },
    });

    if (error || !data?.url) {
      if (error && (error as unknown as { status?: number }).status === 404) {
        alert.error(t("auth.noSSOProvider"));
      } else {
        alert.error((error as unknown as { message?: string })?.message ?? t("auth.ssoLoginFailed"));
      }
    } else {
      window.location.href = data.url;
    }
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : t("auth.ssoLoginFailed"));
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
