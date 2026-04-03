<template>
  <div class="login-header">
    <h2>{{ $t("auth.changePassword") }}</h2>
  </div>

  <vee-form
    v-slot="{ errors, isSubmitting }"
    class="login-form-container"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <password-field name="password" :placeholder="$t('common.password')" :error="errors.password" />

    <button class="btn btn-primary btn-block" :disabled="isSubmitting" type="submit">
      <icon-fa6-solid:key v-if="isSubmitting" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t("auth.changePasswordButton") }}
    </button>

    <div>
      <router-link :to="{ name: 'account.login' }"> {{ $t("common.cancel") }} </router-link>
    </div>
  </vee-form>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";

import { Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";

const { t } = useI18n();

const schema = Yup.object().shape({
  password: Yup.string()
    .required(t("validation.passwordRequired"))
    .min(8, t("validation.passwordMinLength"))
    .matches(/^(?=.*[0-9])/, t("validation.passwordNumber"))
    .matches(/^(?=.*[!@#$%^&*])/, t("validation.passwordSpecial")),
});

const route = useRoute();
const router = useRouter();
const { success, error } = useAlert();

async function onSubmit(values: Record<string, unknown>): Promise<void> {
  const password = values.password as string;

  const resp = await authClient.resetPassword({
    token: route.params.token as string,
    newPassword: password,
  });

  if (resp.error) {
    error(resp.error.message ?? t("auth.failedResetPassword"));
    return;
  }

  success(t("auth.passwordResetSuccess"));

  setTimeout(() => {
    router.push({ name: "account.login" });
  }, 2000);
}
</script>

<style scoped>
.login-resend-mail-container {
  text-align: unset;
}
</style>
