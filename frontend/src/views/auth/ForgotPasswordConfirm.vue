<template>
  <div class="my-8 text-center">
    <h2 class="mb-2 text-3xl font-bold leading-tight">{{ $t("auth.changePassword") }}</h2>
  </div>

  <form class="flex w-full flex-col gap-6 text-center" @submit="onSubmit">
    <FormField v-slot="{ componentField }" name="password">
      <FormItem>
        <FormControl>
          <div class="relative">
            <Input
              :type="passwordType"
              :placeholder="$t('common.password')"
              v-bind="componentField"
            />
            <div
              class="absolute inset-y-0 right-0 z-10 flex cursor-pointer items-center pr-3 opacity-40 transition-opacity hover:opacity-100"
              @click="passwordType = passwordType === 'password' ? 'text' : 'password'"
            >
              <icon-fa6-solid:eye v-if="passwordType === 'password'" class="size-4" />
              <icon-fa6-solid:eye-slash v-else class="size-4" />
            </div>
          </div>
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <Button type="submit" class="w-full" :disabled="isSubmitting">
      <icon-fa6-solid:key v-if="!isSubmitting" class="size-5" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="size-5" />
      {{ $t("auth.changePasswordButton") }}
    </Button>

    <div>
      <RouterLink
        :to="{ name: 'account.login' }"
        class="text-sm text-muted-foreground underline-offset-4 hover:underline"
      >
        {{ $t("common.cancel") }}
      </RouterLink>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { z } from "zod";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";

import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField, FormItem, FormControl, FormMessage } from "@/components/ui/form";

const { t } = useI18n();
const passwordType = ref("password");

const schema = z.object({
  password: z
    .string()
    .min(1, t("validation.passwordRequired"))
    .min(8, t("validation.passwordMinLength"))
    .regex(/^(?=.*[0-9])/, t("validation.passwordNumber"))
    .regex(/^(?=.*[!@#$%^&*])/, t("validation.passwordSpecial")),
});

const { handleSubmit, isSubmitting } = useForm({
  validationSchema: toTypedSchema(schema),
});

const route = useRoute();
const router = useRouter();
const { success, error } = useAlert();

const onSubmit = handleSubmit(async (values) => {
  const resp = await api.POST("/auth/reset-password", {
    body: {
      token: route.params.token as string,
      newPassword: values.password,
    },
  });

  if (resp.error) {
    error((resp.error as { message?: string }).message ?? t("auth.failedResetPassword"));
    return;
  }

  success(t("auth.passwordResetSuccess"));

  setTimeout(() => {
    router.push({ name: "account.login" });
  }, 2000);
});
</script>
