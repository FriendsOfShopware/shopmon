<template>
  <div class="login-header">
    <h2>{{ $t("auth.createAccountTitle") }}</h2>
  </div>

  <vee-form
    v-slot="{ errors, isSubmitting }"
    class="login-form-container"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div class="login-form-group">
      <InputField name="displayName" :placeholder="$t('auth.displayName')" :error="errors.displayName" />
      <InputField name="email" :placeholder="$t('common.emailAddress')" :error="errors.email" />

      <PasswordField
        name="password"
        :placeholder="$t('common.password')"
        :error="errors.password"
      />
    </div>

    <button class="btn btn-primary btn-block" :disabled="isSubmitting" type="submit">
      <icon-fa6-solid:user-plus v-if="!isSubmitting" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t("auth.registerButton") }}
    </button>

    <div>
      <router-link :to="{ name: 'account.login' }"> {{ $t("common.cancel") }} </router-link>
    </div>
  </vee-form>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { router } from "@/router";

import { useAlert } from "@/composables/useAlert";
import { api, setToken } from "@/helpers/api";

const { t } = useI18n();
const alert = useAlert();

const schema = Yup.object().shape({
  displayName: Yup.string().required(t("validation.displayNameRequired")),
  email: Yup.string().required(t("validation.emailRequired")),
  password: Yup.string()
    .required(t("validation.passwordRequired"))
    .min(8, t("validation.passwordMinLength"))
    .matches(/^(?=.*[0-9])/, t("validation.passwordNumber"))
    .matches(/^(?=.*[!@#$%^&*])/, t("validation.passwordSpecial")),
});

async function onSubmit(values: Record<string, unknown>) {
  const { data, error } = await api.POST("/auth/sign-up/email", {
    body: {
      email: values.email as string,
      password: values.password as string,
      name: values.displayName as string,
    },
  });

  if (error) {
    alert.error((error as unknown as { message?: string }).message ?? t("auth.failedRegister"));
    return;
  }

  if ((data as { token?: string })?.token) {
    setToken((data as { token: string }).token);
  }

  await router.push({ name: "account.login" });
  alert.success(t("auth.registrationSuccess"));
}
</script>
