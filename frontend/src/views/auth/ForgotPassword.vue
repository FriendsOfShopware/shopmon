<template>
  <div class="login-header">
    <h2>{{ $t("auth.forgotPasswordTitle") }}</h2>
    <p>{{ $t("auth.forgotPasswordDesc") }}</p>
  </div>

  <vee-form
    v-slot="{ errors, isSubmitting }"
    class="login-form-container"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <InputField name="email" :placeholder="$t('common.emailAddress')" :error="errors.email" />

    <button class="btn btn-primary btn-block" :disabled="isSubmitting" type="submit">
      <icon-fa6-solid:envelope v-if="!isSubmitting" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t("auth.sendEmail") }}
    </button>

    <div>
      <router-link to="login"> {{ $t("common.cancel") }} </router-link>
    </div>
  </vee-form>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";

const { t } = useI18n();

const schema = Yup.object().shape({
  email: Yup.string().required(t("validation.emailRequired")),
});

async function onSubmit(values: Record<string, unknown>): Promise<void> {
  const { success, error } = useAlert();

  try {
    await api.POST("/auth/forget-password", {
      body: { email: values.email as string },
    });
    success(t("auth.resetEmailSent"));
  } catch (e) {
    error(e instanceof Error ? e.message : t("auth.failedSendReset"));
  }
}
</script>

<style scoped>
.login-header {
  p {
    text-align: left;
  }
}
</style>
