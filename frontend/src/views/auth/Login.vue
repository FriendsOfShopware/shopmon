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
      <div>
        <label for="email-address" class="sr-only">{{ $t("common.emailAddress") }}</label>
        <field
          id="email-address"
          name="email"
          type="email"
          autocomplete="email"
          required=""
          class="field"
          :placeholder="$t('common.emailAddress')"
          :class="{ 'has-error': errors.email }"
        />

        <div class="field-error-message">{{ errors.email }}</div>
      </div>

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

      <div class="sso-input-group">
        <input
          v-model="ssoEmail"
          type="email"
          class="field"
          :placeholder="$t('auth.enterWorkEmail')"
          @keyup.enter="ssoLogin"
        />
        <button
          type="button"
          class="btn btn-secondary icon-only"
          :disabled="isSSOLoading || !ssoEmail"
          @click="ssoLogin"
        >
          <icon-fa6-solid:arrow-right v-if="!isSSOLoading" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
        </button>
      </div>
      <p class="sso-help">{{ $t("auth.ssoHelp") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm, configure } from "vee-validate";
import { ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { useReturnUrl } from "@/composables/useReturnUrl";
import { authClient } from "@/helpers/auth-client";
import { useRouter } from "vue-router";

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

function goToDashboard() {
  const sess = authClient.useSession();

  const stop = watch(
    sess,
    async (user) => {
      if (!user.data?.user) {
        return;
      }

      stop();
      const redirectUrl = returnUrl.value ?? "/app/dashboard";
      clearReturnUrl();
      await router.push(redirectUrl);
    },
    { immediate: true },
  );
}

async function onSubmit(values: Record<string, unknown>) {
  const email = values.email as string;
  const password = values.password as string;
  const resp = await authClient.signIn.email({
    email,
    password,
  });

  if (resp.error) {
    alert.error(resp.error.message ?? t("auth.failedSignIn"));
    return;
  }

  goToDashboard();
}

async function webauthnLogin() {
  isAuthenticated.value = true;

  try {
    const resp = await authClient.signIn.passkey();

    if (resp?.error) {
      const { error } = useAlert();
      error(resp.error.message ?? t("auth.failedSignInPasskey"));
      isAuthenticated.value = false;
      return;
    }

    goToDashboard();
  } catch (e: unknown) {
    const { error } = useAlert();

    error(e instanceof Error ? e.message : String(e));

    isAuthenticated.value = false;
  }
}

async function githubLogin() {
  isGithubLoading.value = true;

  try {
    const redirectUrl = returnUrl.value ?? "/";
    const resp = await authClient.signIn.social({
      provider: "github",
      callbackURL: redirectUrl,
    });
    if (resp.error) {
      const { error } = useAlert();
      error(resp.error.message ?? t("auth.failedSignInGithub"));
      isGithubLoading.value = false;
      return;
    }

    goToDashboard();
  } catch (e: unknown) {
    const { error } = useAlert();

    error(e instanceof Error ? e.message : String(e));

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
    const result = await authClient.signIn.sso({
      email: ssoEmail.value,
      callbackURL: `${window.location.origin}${redirectUrl}`,
    });

    if (result.error) {
      if (result.error.code === "SSO_PROVIDER_NOT_FOUND") {
        alert.error(t("auth.noSSOProvider"));
      } else {
        alert.error(result.error.message ?? t("auth.ssoLoginFailed"));
      }
    } else {
      goToDashboard();
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
