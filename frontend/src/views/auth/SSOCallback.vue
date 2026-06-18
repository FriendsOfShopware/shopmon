<template>
  <div class="flex items-center justify-center py-12">
    <icon-line-md:loading-twotone-loop class="size-8" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import { useAlert } from "@/composables/useAlert";
import { useReturnUrl } from "@/composables/useReturnUrl";
import { fetchSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const alert = useAlert();
const { returnUrl, clearReturnUrl } = useReturnUrl();

const params = new URLSearchParams(window.location.search);
const code = params.get("code");
const state = params.get("state");
const savedState = sessionStorage.getItem("oauth_state");
sessionStorage.removeItem("oauth_state");

async function handleCallback() {
  const providerId = route.params.providerId as string;

  if (!code || !state || !providerId || savedState !== state) {
    router.replace({ name: "account.login" });
    return;
  }

  try {
    const { data: callbackData, error: callbackError } = await api.GET(
      "/auth/sso/callback/{providerId}",
      {
        params: { path: { providerId }, query: { code, state } },
      },
    );

    if (callbackError || !callbackData?.code) {
      alert.error(
        (callbackError as unknown as { message?: string })?.message ?? t("auth.failedSignIn"),
      );
      router.replace({ name: "account.login" });
      return;
    }

    const { data, error } = await api.POST("/auth/exchange-code", {
      body: { code: callbackData.code },
    });

    if (error || !data?.token) {
      alert.error((error as unknown as { message?: string })?.message ?? t("auth.failedSignIn"));
      router.replace({ name: "account.login" });
      return;
    }

    setToken(data.token);
    await fetchSession();
    const redirectUrl = returnUrl.value ?? "/app/dashboard";
    clearReturnUrl();
    router.replace(redirectUrl);
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : t("auth.failedSignIn"));
    router.replace({ name: "account.login" });
  }
}

handleCallback();
</script>
