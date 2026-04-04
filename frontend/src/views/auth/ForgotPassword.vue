<template>
  <div class="my-8 text-center">
    <h2 class="mb-2 text-3xl font-bold leading-tight">{{ $t("auth.forgotPasswordTitle") }}</h2>
    <p class="text-left text-muted-foreground">{{ $t("auth.forgotPasswordDesc") }}</p>
  </div>

  <form
    class="flex w-full flex-col gap-6 text-center"
    @submit="onSubmit"
  >
    <FormField v-slot="{ componentField }" name="email">
      <FormItem>
        <FormControl>
          <Input
            type="email"
            :placeholder="$t('common.emailAddress')"
            v-bind="componentField"
          />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <Button type="submit" class="w-full" :disabled="isSubmitting">
      <icon-fa6-solid:envelope v-if="!isSubmitting" class="size-5" aria-hidden="true" />
      <icon-line-md:loading-twotone-loop v-else class="size-5" />
      {{ $t("auth.sendEmail") }}
    </Button>

    <div>
      <RouterLink
        to="login"
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

import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField, FormItem, FormControl, FormMessage } from "@/components/ui/form";

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
