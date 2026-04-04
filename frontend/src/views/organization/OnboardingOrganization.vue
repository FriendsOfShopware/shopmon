<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ $t("onboarding.welcome") }}</h1>
    </div>
    <div class="flex gap-2 items-start" />
  </header>

  <div class="w-full">
    <Card>
      <CardContent class="pt-6">
        <div class="text-center pb-8 mb-8 border-b">
          <icon-fa6-solid:building class="size-12 text-primary mx-auto mb-4" />
          <h2 class="text-2xl font-semibold mb-2">{{ $t("onboarding.title") }}</h2>
          <p class="text-muted-foreground max-w-md mx-auto leading-relaxed">
            {{ $t("onboarding.description") }}
          </p>
        </div>

        <div class="flex flex-col gap-6">
          <div class="flex gap-4 items-start">
            <div
              class="shrink-0 size-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center"
            >
              <icon-fa6-solid:plus class="size-4" />
            </div>
            <div class="flex-1">
              <h3 class="text-lg font-semibold mb-1">{{ $t("onboarding.createTitle") }}</h3>
              <p class="text-muted-foreground text-sm leading-relaxed mb-0">
                {{ $t("onboarding.createDescription") }}
              </p>

              <form class="mt-4 flex gap-3 items-end" @submit="onSubmit">
                <div class="flex-1">
                  <FormField v-slot="{ componentField }" name="name">
                    <FormItem>
                      <FormLabel>{{ $t("common.name") }}</FormLabel>
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
                </div>

                <Button :disabled="isSubmitting" type="submit" class="shrink-0">
                  <icon-fa6-solid:floppy-disk
                    v-if="!isSubmitting"
                    class="size-4"
                    aria-hidden="true"
                  />
                  <icon-line-md:loading-twotone-loop v-else class="size-4" />
                  {{ $t("onboarding.createButton") }}
                </Button>
              </form>
            </div>
          </div>

          <div
            class="flex items-center gap-4 text-muted-foreground text-sm uppercase tracking-wide"
          >
            <div class="flex-1 h-px bg-border" />
            <span>{{ $t("common.or") }}</span>
            <div class="flex-1 h-px bg-border" />
          </div>

          <div class="flex gap-4 items-start">
            <div
              class="shrink-0 size-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center"
            >
              <icon-fa6-solid:envelope class="size-4" />
            </div>
            <div class="flex-1">
              <h3 class="text-lg font-semibold mb-1">{{ $t("onboarding.inviteTitle") }}</h3>
              <p class="text-muted-foreground text-sm leading-relaxed mb-0">
                {{ $t("onboarding.inviteDescription") }}
              </p>
            </div>
          </div>
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
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
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

const { handleSubmit, isSubmitting } = useForm({
  validationSchema,
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
