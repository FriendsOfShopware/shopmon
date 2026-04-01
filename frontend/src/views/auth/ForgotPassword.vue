<template>
  <div class="login-header">
    <h2>{{ $t('auth.forgotPasswordTitle') }}</h2>
    <p>{{ $t('auth.forgotPasswordDesc') }}</p>
  </div>

  <vee-form
    v-slot="{ errors, isSubmitting }"
    class="login-form-container"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div>
      <field
        name="email"
        :placeholder="$t('common.emailAddress')"
        type="text"
        class="field"
        :class="{ 'has-error': errors.email }"
      />
      <div class="field-error-message">
        {{ errors.email }}
      </div>
    </div>

    <button class="btn btn-primary btn-block" :disabled="isSubmitting" type="submit">
      <icon-fa6-solid:envelope v-if="!isSubmitting" class="icon" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="icon" />
      {{ $t('auth.sendEmail') }}
    </button>

    <div>
      <router-link to="login"> {{ $t('common.cancel') }} </router-link>
    </div>
  </vee-form>
</template>

<script setup lang="ts">
import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";

const { t } = useI18n();

const schema = Yup.object().shape({
  email: Yup.string().required(t('validation.emailRequired')),
});

async function onSubmit(values: { email: string }): Promise<void> {
  const { success, error } = useAlert();

  const resp = await authClient.forgetPassword({ email: values.email });

  if (resp.error) {
    error(resp.error.message ?? t('auth.failedSendReset'));
    return;
  }

  success(t('auth.resetEmailSent'));
}
</script>

<style scoped>
.login-header {
  p {
    text-align: left;
  }
}
</style>
