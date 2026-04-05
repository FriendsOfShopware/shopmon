<template>
  <div class="mx-auto max-w-lg py-12">
    <!-- Welcome header -->
    <div class="mb-10 text-center">
      <div class="mx-auto mb-4 flex size-16 items-center justify-center rounded-2xl bg-primary/10">
        <icon-fa6-solid:rocket class="size-7 text-primary" />
      </div>
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("onboarding.welcome") }}</h1>
      <p class="mx-auto mt-2 max-w-sm text-sm leading-relaxed text-muted-foreground">
        {{ $t("onboarding.description") }}
      </p>
    </div>

    <!-- Create org -->
    <Card class="mb-4">
      <CardContent class="p-6">
        <div class="mb-4 flex items-center gap-3">
          <div class="flex size-9 items-center justify-center rounded-lg bg-primary/10">
            <icon-fa6-solid:plus class="size-3.5 text-primary" />
          </div>
          <div>
            <h2 class="font-semibold">{{ $t("onboarding.createTitle") }}</h2>
            <p class="text-xs text-muted-foreground">{{ $t("onboarding.createDescription") }}</p>
          </div>
        </div>

        <form class="flex gap-2" @submit="onSubmit">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem class="flex-1">
              <FormControl>
                <Input
                  v-bind="componentField"
                  :placeholder="$t('onboarding.orgNamePlaceholder')"
                  autocomplete="organization"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <Button :disabled="isSubmitting" type="submit" class="shrink-0">
            <icon-fa6-solid:arrow-right v-if="!isSubmitting" class="mr-1.5 size-3" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3" />
            {{ $t("onboarding.createButton") }}
          </Button>
        </form>
      </CardContent>
    </Card>

    <!-- Divider -->
    <div class="flex items-center gap-3 py-2 text-xs uppercase tracking-wide text-muted-foreground">
      <div class="h-px flex-1 bg-border" />
      <span>{{ $t("common.or") }}</span>
      <div class="h-px flex-1 bg-border" />
    </div>

    <!-- Join existing -->
    <Card class="mt-4">
      <CardContent class="flex items-center gap-3 p-6">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
          <icon-fa6-solid:envelope class="size-3.5 text-muted-foreground" />
        </div>
        <div>
          <h2 class="font-semibold">{{ $t("onboarding.inviteTitle") }}</h2>
          <p class="mt-0.5 text-xs leading-relaxed text-muted-foreground">
            {{ $t("onboarding.inviteDescription") }}
          </p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { fetchSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();

const validationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.orgNameRequired")),
  }),
);

const { handleSubmit, isSubmitting } = useForm({ validationSchema });

const onSubmit = handleSubmit(async (values) => {
  try {
    const { error: respError } = await api.POST("/auth/organizations", {
      body: { name: values.name },
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
