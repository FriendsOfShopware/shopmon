<template>
  <Card class="border-0 shadow-none sm:border sm:shadow-sm">
    <CardHeader class="space-y-1 px-6 pt-6 pb-2 text-center">
      <CardTitle class="text-2xl font-semibold tracking-tight">
        {{ $t("auth.confirmingAccount") }}
      </CardTitle>
    </CardHeader>

    <CardContent class="px-6 pb-6">
      <Alert v-if="isLoading">
        <icon-fa6-solid:circle-info class="size-4" />
        <AlertDescription>{{ $t("common.loading") }}</AlertDescription>
      </Alert>

      <template v-else>
        <Alert
          v-if="confirmSuccess"
          class="border-green-500/30 bg-green-500/10 text-green-600 dark:text-green-400"
        >
          <icon-fa6-solid:circle-check class="size-4" />
          <AlertDescription>
            {{ $t("auth.emailConfirmed") }}
          </AlertDescription>
        </Alert>

        <Alert v-else variant="destructive">
          <icon-fa6-solid:circle-xmark class="size-4" />
          <AlertDescription>{{ $t("auth.tokenExpired") }}</AlertDescription>
        </Alert>
      </template>

      <p class="mt-4 text-center text-sm text-muted-foreground">
        <RouterLink
          :to="{ name: 'account.login' }"
          class="font-medium text-primary underline-offset-4 hover:underline"
        >
          {{ $t("auth.backToSignIn", "Back to sign in") }}
        </RouterLink>
      </p>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useI18n } from "vue-i18n";
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const { t } = useI18n();
const route = useRoute();
const alert = useAlert();

const isLoading = ref(true);
const confirmSuccess = ref(false);

onMounted(async () => {
  const resp = await api.GET("/auth/verify-email", {
    params: { query: { token: route.params.token as string } },
  });

  if (resp.error) {
    confirmSuccess.value = false;
    isLoading.value = false;

    alert.error((resp.error as { message?: string }).message ?? t("auth.failedVerifyEmail"));

    return;
  }

  confirmSuccess.value = true;
  isLoading.value = false;
});
</script>
