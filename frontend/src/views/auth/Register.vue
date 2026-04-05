<template>
  <Card class="border-0 shadow-none sm:border sm:shadow-sm">
    <CardHeader class="space-y-1 px-6 pt-6 pb-2 text-center">
      <CardTitle class="text-2xl font-semibold tracking-tight">
        {{ $t("auth.createAccountTitle") }}
      </CardTitle>
      <CardDescription>
        {{ $t("auth.alreadyHaveAccount", "Already have an account?") }}
        {{ " " }}
        <RouterLink
          :to="{ name: 'account.login' }"
          class="font-medium text-primary underline-offset-4 hover:underline"
        >
          {{ $t("auth.signIn") }}
        </RouterLink>
      </CardDescription>
    </CardHeader>

    <CardContent class="px-6 pb-6">
      <form class="flex flex-col gap-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="displayName">
          <FormItem>
            <FormLabel>{{ $t("auth.displayName") }}</FormLabel>
            <FormControl>
              <Input v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="email">
          <FormItem>
            <FormLabel>{{ $t("common.emailAddress") }}</FormLabel>
            <FormControl>
              <Input type="email" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="password">
          <FormItem>
            <FormLabel>{{ $t("common.password") }}</FormLabel>
            <FormControl>
              <PasswordInput v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <Button type="submit" class="mt-2 w-full" :disabled="isSubmitting">
          <icon-line-md:loading-twotone-loop v-if="isSubmitting" class="size-5" />
          {{ $t("auth.registerButton") }}
        </Button>
      </form>
    </CardContent>
  </Card>
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { FormField, FormItem, FormControl, FormLabel, FormMessage } from "@/components/ui/form";
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
