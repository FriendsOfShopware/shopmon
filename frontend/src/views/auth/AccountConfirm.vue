<template>
  <div class="login-header">
    <h2>Confirming your Account Registration</h2>
  </div>

  <Alert v-if="isLoading" type="info"> Loading... </Alert>

  <template v-else>
    <Alert v-if="success" type="success">
      Your email address has been confirmed.
      <router-link :to="{ name: 'account.login' }"> Login </router-link>
    </Alert>

    <Alert v-else type="error"> The given token has been expired </Alert>
  </template>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

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

    alert.error((resp.error as { message?: string }).message ?? "Failed to verify email");

    return;
  }

  success.value = true;
  isLoading.value = false;
});
</script>
