<template>
  <Card class="border-0 shadow-none sm:border sm:shadow-sm">
    <CardHeader class="space-y-1 px-6 pt-6 pb-2 text-center">
      <CardTitle class="text-2xl font-semibold tracking-tight">
        {{ $t("auth.signIn") }}
      </CardTitle>
      <CardDescription>
        {{ $t("auth.newToShopmon") }}
        {{ " " }}
        <RouterLink
          :to="{ name: 'account.register' }"
          class="font-medium text-primary underline-offset-4 hover:underline"
        >
          {{ $t("auth.createAccount") }}
        </RouterLink>
      </CardDescription>
    </CardHeader>

    <CardContent class="px-6 pb-6">
      <!-- Social & Passkey login buttons -->
      <div class="flex flex-col gap-3">
        <Button
          type="button"
          variant="outline"
          class="w-full"
          :disabled="isGithubLoading"
          @click="githubLogin"
        >
          <icon-mdi:github v-if="!isGithubLoading" class="size-5" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-5" />
          {{ $t("auth.continueGithub") }}
        </Button>

        <Button
          type="button"
          variant="outline"
          class="w-full"
          :disabled="isAuthenticated"
          @click="webauthnLogin"
        >
          <icon-material-symbols:passkey
            v-if="!isAuthenticated"
            class="size-5"
            aria-hidden="true"
          />
          <icon-line-md:loading-twotone-loop v-else class="size-5" />
          {{ $t("auth.loginPasskey") }}
        </Button>
      </div>

      <!-- Divider -->
      <div class="relative my-6">
        <Separator />
        <span
          class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-card px-3 text-xs text-muted-foreground"
        >
          {{ $t("common.or") }}
        </span>
      </div>

      <!-- Email / Password form -->
      <form class="flex flex-col gap-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="email">
          <FormItem>
            <FormLabel>{{ $t("common.emailAddress") }}</FormLabel>
            <FormControl>
              <Input type="email" autocomplete="email" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="password">
          <FormItem>
            <div class="flex items-center justify-between">
              <FormLabel>{{ $t("common.password") }}</FormLabel>
              <RouterLink
                :to="{ name: 'account.forgot.password' }"
                class="text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
              >
                {{ $t("auth.forgotPassword") }}
              </RouterLink>
            </div>
            <FormControl>
              <PasswordInput v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <Button type="submit" class="mt-2 w-full" :disabled="isSubmitting">
          <icon-line-md:loading-twotone-loop v-if="isSubmitting" class="size-5" />
          {{ $t("auth.signInButton") }}
        </Button>
      </form>

      <!-- Enterprise SSO -->
      <div class="mt-6 rounded-lg border border-dashed p-4">
        <p class="mb-3 text-center text-sm font-medium text-muted-foreground">
          {{ $t("auth.enterpriseSSO") }}
        </p>
        <div class="flex gap-2">
          <Input
            v-model="ssoEmail"
            type="email"
            class="flex-1"
            :placeholder="$t('auth.enterWorkEmail')"
            @keyup.enter="ssoLogin"
          />
          <Button type="button" size="icon" :disabled="isSSOLoading || !ssoEmail" @click="ssoLogin">
            <icon-fa6-solid:arrow-right v-if="!isSSOLoading" class="size-4" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="size-4" />
          </Button>
        </div>
        <p class="mt-2 text-center text-xs text-muted-foreground">{{ $t("auth.ssoHelp") }}</p>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { z } from "zod";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { startAuthentication } from "@simplewebauthn/browser";

import { useAlert } from "@/composables/useAlert";
import { useReturnUrl } from "@/composables/useReturnUrl";
import { fetchSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { FormField, FormItem, FormControl, FormLabel, FormMessage } from "@/components/ui/form";
import PasswordInput from "@/components/PasswordInput.vue";

const { t } = useI18n();
const router = useRouter();
const { returnUrl, clearReturnUrl } = useReturnUrl();

const isAuthenticated = ref(false);
const isGithubLoading = ref(false);
const isSSOLoading = ref(false);
const ssoEmail = ref("");
const alert = useAlert();

const schema = z.object({
  email: z.string().min(1, t("validation.emailRequired")).email(),
  password: z.string().min(1, t("validation.passwordRequired")),
});

const { handleSubmit, isSubmitting } = useForm({
  validationSchema: toTypedSchema(schema),
});

async function goToDashboard() {
  await fetchSession();
  const redirectUrl = returnUrl.value ?? "/app/dashboard";
  clearReturnUrl();
  await router.push(redirectUrl);
}

const onSubmit = handleSubmit(async (values) => {
  const { data, error } = await api.POST("/auth/sign-in/email", {
    body: { email: values.email, password: values.password },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? t("auth.failedSignIn"));
    return;
  }

  if ((data as { token?: string })?.token) {
    setToken((data as { token: string }).token);
  }

  await goToDashboard();
});

async function webauthnLogin() {
  isAuthenticated.value = true;

  try {
    // Get login options from server
    const { data: optionsData, error: optionsError } = await api.POST(
      "/auth/passkey/login-options",
    );

    if (optionsError || !optionsData) {
      alert.error(t("auth.failedSignInPasskey"));
      isAuthenticated.value = false;
      return;
    }

    const { options, challengeKey } = optionsData as {
      options: { publicKey: Parameters<typeof startAuthentication>[0]["optionsJSON"] };
      challengeKey: string;
    };

    // Run WebAuthn browser API
    const assertion = await startAuthentication({ optionsJSON: options.publicKey });

    // Send assertion to server
    const { data: loginData, error } = await api.POST("/auth/passkey/login", {
      body: { challengeKey, ...assertion } as never,
    });

    if (error) {
      alert.error((error as { message?: string }).message ?? "Failed to sign in with Passkey");
      isAuthenticated.value = false;
      return;
    }

    if ((loginData as unknown as { token?: string })?.token) {
      setToken((loginData as unknown as { token: string }).token);
    }

    await goToDashboard();
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : String(e));
    isAuthenticated.value = false;
  }
}

async function githubLogin() {
  isGithubLoading.value = true;

  try {
    const redirectUrl = returnUrl.value ?? "/";
    const { data, error } = await api.POST("/auth/sign-in/social", {
      body: { provider: "github", callbackURL: `${window.location.origin}${redirectUrl}` },
    });

    if (error || !data?.url) {
      alert.error(t("auth.failedSignInGithub"));
      isGithubLoading.value = false;
      return;
    }

    // Store state for client-side verification on callback
    const oauthUrl = new URL(data.url);
    const state = oauthUrl.searchParams.get("state");
    if (state) {
      sessionStorage.setItem("oauth_state", state);
    }

    // Redirect to the OAuth provider
    window.location.href = data.url;
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : String(e));
    isGithubLoading.value = false;
  }
}

async function ssoLogin() {
  if (!ssoEmail.value) {
    alert.error(t("auth.pleaseEnterWorkEmail"));
    return;
  }

  isSSOLoading.value = true;

  try {
    const redirectUrl = returnUrl.value ?? "/";
    const { data, error } = await api.POST("/auth/sign-in/sso", {
      body: {
        email: ssoEmail.value,
        callbackURL: `${window.location.origin}${redirectUrl}`,
      },
    });

    if (error || !data?.url) {
      if (error && (error as unknown as { status?: number }).status === 404) {
        alert.error(t("auth.noSSOProvider"));
      } else {
        alert.error(
          (error as unknown as { message?: string })?.message ?? t("auth.ssoLoginFailed"),
        );
      }
    } else {
      // Store state for client-side verification on callback
      const oauthUrl = new URL(data.url);
      const state = oauthUrl.searchParams.get("state");
      if (state) {
        sessionStorage.setItem("oauth_state", state);
      }

      window.location.href = data.url;
    }
  } catch (e: unknown) {
    alert.error(e instanceof Error ? e.message : t("auth.ssoLoginFailed"));
  } finally {
    isSSOLoading.value = false;
  }
}
</script>
