<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("shop.newShop") }}</h1>

    <Card v-if="!isLoadingOrgs">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:folder class="size-4 text-muted-foreground" />
          {{ $t("shop.shopInfo") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit" class="space-y-6">
          <div class="space-y-4">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ $t("common.name") }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="description">
              <FormItem>
                <FormLabel>{{ $t("common.description") }}</FormLabel>
                <FormControl>
                  <Textarea v-bind="componentField" :placeholder="$t('shop.optionalDescription')" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="gitUrl">
              <FormItem>
                <FormLabel>{{ $t("shop.gitRepoUrl") }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://github.com/org/repo"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <div class="flex justify-end">
            <Button :disabled="isSubmitting" type="submit">
              <icon-fa6-solid:floppy-disk
                v-if="!isSubmitting"
                class="mr-1.5 size-3.5"
                aria-hidden="true"
              />
              <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
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
import { api } from "@/helpers/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();
const route = useRoute();

const isLoadingOrgs = ref(false);

const validationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.shopNameRequired")),
    description: z.string().optional(),
    gitUrl: z.string().url(t("validation.urlInvalid")).optional().or(z.literal("")),
  }),
);

const { handleSubmit, isSubmitting } = useForm({
  validationSchema,
  initialValues: {
    name: "",
    description: "",
    gitUrl: "",
  },
});

const onSubmit = handleSubmit(async (values) => {
  try {
    const { error: apiError } = await api.POST("/organization/shops", {
      body: {
        name: values.name,
        description: values.description ?? undefined,
        gitUrl: values.gitUrl || undefined,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

    if (route.query.redirect) {
      router.push(route.query.redirect as string);
    } else {
      router.push({ name: "account.shop.list" });
    }
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
});

onMounted(() => {
  // no-op, organization comes from session
});
</script>
