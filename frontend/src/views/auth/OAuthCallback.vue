<template>
  <div class="flex items-center justify-center py-12">
    <icon-line-md:loading-twotone-loop class="size-8" />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { useReturnUrl } from "@/composables/useReturnUrl";
import { fetchSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";

const router = useRouter();
const { returnUrl, clearReturnUrl } = useReturnUrl();

const params = new URLSearchParams(window.location.search);
const code = params.get("code");
const state = params.get("state");
const savedState = sessionStorage.getItem("oauth_state");
sessionStorage.removeItem("oauth_state");

async function handleCallback() {
  if (!code || !state || savedState !== state) {
    router.replace({ name: "account.login" });
    return;
  }

  try {
    const { data: callbackData } = await api.GET("/auth/callback/github", {
      params: { query: { code, state } },
    });

    if (!callbackData?.code) {
      router.replace({ name: "account.login" });
      return;
    }

    const { data } = await api.POST("/auth/exchange-code", {
      body: { code: callbackData.code },
    });

    if (data?.token) {
      setToken(data.token);
      await fetchSession();
      const redirectUrl = returnUrl.value ?? "/app/dashboard";
      clearReturnUrl();
      router.replace(redirectUrl);
    } else {
      router.replace({ name: "account.login" });
    }
  } catch {
    router.replace({ name: "account.login" });
  }
}

handleCallback();
</script>
