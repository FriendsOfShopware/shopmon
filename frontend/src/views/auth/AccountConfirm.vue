<template>
  <div class="login-header">
    <h2>{{ $t('auth.confirmingAccount') }}</h2>
  </div>

  <Alert v-if="isLoading" type="info"> {{ $t('common.loading') }} </Alert>

  <template v-else>
    <Alert v-if="success" type="success">
      {{ $t('auth.emailConfirmed') }}
      <router-link :to="{ name: 'account.login' }"> {{ $t('nav.login') }} </router-link>
    </Alert>

    <Alert v-else type="error"> {{ $t('auth.tokenExpired') }} </Alert>
  </template>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";
import { useI18n } from "vue-i18n";
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const alert = useAlert();

const isLoading = ref(true);
const success = ref(false);

onMounted(async () => {
  const resp = await authClient.verifyEmail({
    query: { token: route.params.token as string },
  });

  if (resp.error) {
    success.value = false;
    isLoading.value = false;

    alert.error(resp.error.message ?? t('auth.failedVerifyEmail'));

    return;
  }

  success.value = true;
  isLoading.value = false;
});
</script>
