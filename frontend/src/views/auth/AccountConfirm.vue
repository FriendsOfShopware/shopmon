<template>
  <div class="login-header">
    <h2>{{ $t("auth.confirmingAccount") }}</h2>
  </div>

  <Banner v-if="isLoading" variant="default"> {{ $t("common.loading") }} </Banner>

  <template v-else>
    <Banner v-if="success" variant="success">
      {{ $t("auth.emailConfirmed") }}
      <router-link :to="{ name: 'account.login' }"> {{ $t("nav.login") }} </router-link>
    </Banner>

    <Banner v-else variant="error"> {{ $t("auth.tokenExpired") }} </Banner>
  </template>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useI18n } from "vue-i18n";
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const alert = useAlert();

const isLoading = ref(true);
const success = ref(false);

onMounted(async () => {
  const resp = await api.GET("/auth/verify-email", {
    params: { query: { token: route.params.token as string } },
  });

  if (resp.error) {
    success.value = false;
    isLoading.value = false;

    alert.error((resp.error as { message?: string }).message ?? t("auth.failedVerifyEmail"));

    return;
  }

  success.value = true;
  isLoading.value = false;
});
</script>
