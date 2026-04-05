<template>
  <Card class="border-0 shadow-none sm:border sm:shadow-sm">
    <CardHeader class="space-y-1 px-6 pt-6 pb-2 text-center">
      <CardTitle class="text-2xl font-semibold tracking-tight">
        {{ $t("auth.forgotPasswordTitle") }}
      </CardTitle>
      <CardDescription>
        {{ $t("auth.forgotPasswordDesc") }}
      </CardDescription>
    </CardHeader>

    <CardContent class="px-6 pb-6">
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

        <Button type="submit" class="mt-2 w-full" :disabled="isSubmitting">
          <icon-line-md:loading-twotone-loop v-if="isSubmitting" class="size-5" />
          {{ $t("auth.sendEmail") }}
        </Button>
      </form>

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
import { useI18n } from "vue-i18n";
import { z } from "zod";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";

import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { FormField, FormItem, FormControl, FormLabel, FormMessage } from "@/components/ui/form";

const { t } = useI18n();

const schema = z.object({
  email: z.string().min(1, t("validation.emailRequired")),
});

const { handleSubmit, isSubmitting } = useForm({
  validationSchema: toTypedSchema(schema),
});

const onSubmit = handleSubmit(async (values) => {
  const { success, error } = useAlert();

  try {
    await api.POST("/auth/forget-password", {
      body: { email: values.email },
    });
    success(t("auth.resetEmailSent"));
  } catch (e) {
    error(e instanceof Error ? e.message : t("auth.failedSendReset"));
  }
});
</script>
