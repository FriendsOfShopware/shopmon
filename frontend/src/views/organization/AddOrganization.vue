<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ $t('organization.newOrganization') }}</h1>
    </div>
    <div class="flex gap-2 items-start" />
  </header>

  <div v-if="session?.user">
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.orgInfo') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="space-y-4">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ $t('common.name') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" autocomplete="name" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <div class="flex justify-end pt-6">
            <Button :disabled="isSubmitting" type="submit">
              <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="size-4" aria-hidden="true" />
              <icon-line-md:loading-twotone-loop v-else class="size-4" />
              {{ $t("common.save") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession, fetchSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

const { t } = useI18n();
const { session } = useSession();

const { error } = useAlert();
const router = useRouter();

const schema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.orgNameRequired")),
  }),
);

const { handleSubmit, isSubmitting } = useForm({
  validationSchema: schema,
});

const onSubmit = handleSubmit(async (values) => {
  try {
    const { error: respError } = await api.POST("/auth/organizations", {
      body: {
        name: values.name,
      },
    });

    if (respError) {
      error((respError as { message?: string }).message ?? "Failed to create organization");
      return;
    }
    await fetchSession();
    await router.push({ name: "account.dashboard" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
});
</script>
