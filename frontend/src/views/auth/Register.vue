<template>
  <div class="my-8 text-center">
    <h2 class="mb-2 text-3xl font-bold leading-tight">{{ $t("auth.createAccountTitle") }}</h2>
  </div>

  <form class="flex w-full flex-col gap-6 text-center" @submit="onSubmit">
    <div class="flex flex-col gap-2">
      <FormField v-slot="{ componentField }" name="displayName">
        <FormItem>
          <FormControl>
            <Input :placeholder="$t('auth.displayName')" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="email">
        <FormItem>
          <FormControl>
            <Input type="email" :placeholder="$t('common.emailAddress')" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>

      <FormField v-slot="{ componentField }" name="password">
        <FormItem>
          <FormControl>
            <PasswordInput :placeholder="$t('common.password')" v-bind="componentField" />
          </FormControl>
          <FormMessage />
        </FormItem>
      </FormField>
    </div>

    <Button type="submit" class="w-full" :disabled="isSubmitting">
      <icon-fa6-solid:user-plus v-if="!isSubmitting" class="size-5" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="size-5" />
      {{ $t("auth.registerButton") }}
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
import { useI18n } from "vue-i18n";
import { z } from "zod";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";

import { useRouter } from "vue-router";
import { useAlert } from "@/composables/useAlert";
import { api, setToken } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField, FormItem, FormControl, FormMessage } from "@/components/ui/form";
import PasswordInput from "@/components/PasswordInput.vue";

const { t } = useI18n();
const router = useRouter();
const alert = useAlert();
const schema = z.object({
  displayName: z.string().min(1, t("validation.displayNameRequired")),
  email: z.string().min(1, t("validation.emailRequired")),
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

const onSubmit = handleSubmit(async (values) => {
  const { data, error } = await api.POST("/auth/sign-up/email", {
    body: {
      email: values.email,
      password: values.password,
      name: values.displayName,
    },
  });

  if (error) {
    alert.error((error as unknown as { message?: string }).message ?? t("auth.failedRegister"));
    return;
  }

  if ((data as { token?: string })?.token) {
    setToken((data as { token: string }).token);
  }

  await router.push({ name: "account.login" });
  alert.success(t("auth.registrationSuccess"));
});
</script>
