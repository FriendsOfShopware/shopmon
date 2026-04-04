<template>
  <div v-if="environment" class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ $t('environment.editEnvironment', { name: environment.name }) }}</h1>
      <Button as-child variant="outline">
        <RouterLink
          :to="{
            name: 'account.environments.detail',
            params: {
              organizationId: route.params.organizationId as string,
              environmentId: environment.id,
            },
          }"
        >
          {{ $t("common.cancel") }}
        </RouterLink>
      </Button>
    </div>

    <!-- Environment settings form -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:gear class="size-4 text-muted-foreground" />
          {{ $t('environment.environmentInfo') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit" class="space-y-6">
          <div class="grid gap-4 sm:grid-cols-2">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ $t('common.name') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" autocomplete="name" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="shopId">
              <FormItem>
                <FormLabel>{{ $t('environment.shop') }}</FormLabel>
                <Select v-bind="componentField">
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue :placeholder="$t('environment.shop')" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem v-for="shop in shops" :key="shop.id" :value="String(shop.id)">
                      {{ shop.organizationName + " / " + shop.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="shopUrl" class="sm:col-span-2">
              <FormItem>
                <FormLabel>{{ $t('common.url') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" autocomplete="url" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <Separator />

          <!-- Integration section -->
          <div>
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-sm font-semibold">{{ $t('environment.integration') }}</h3>
                <p class="mt-0.5 text-xs text-muted-foreground">
                  <i18n-t keypath="environment.integrationDesc" tag="span">
                    <template #pluginLink>
                      <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank" class="text-primary hover:underline">{{
                        $t("environment.pluginName")
                      }}</a>
                    </template>
                  </i18n-t>
                </p>
              </div>
              <Button type="button" variant="outline" size="sm" @click="openPluginModal">
                <icon-fa6-solid:plug class="mr-1.5 size-3" />
                {{ $t("environment.connectPlugin") }}
              </Button>
            </div>

            <div class="mt-4 grid gap-4 sm:grid-cols-2">
              <FormField v-slot="{ componentField }" name="clientId">
                <FormItem>
                  <FormLabel>{{ $t('environment.clientId') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="clientSecret">
                <FormItem>
                  <FormLabel>{{ $t('environment.clientSecret') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>
          </div>

          <div class="flex justify-end">
            <Button :disabled="isSubmitting" type="submit">
              <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="mr-1.5 size-3.5" aria-hidden="true" />
              <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
              {{ $t("common.save") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Sitespeed settings -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:rocket class="size-4 text-muted-foreground" />
          {{ $t('environment.sitespeed') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("environment.sitespeedDesc") }}</p>

        <form @submit.prevent="onSitespeedSubmit" class="space-y-4">
          <label class="flex cursor-pointer items-center gap-3">
            <Switch :checked="sitespeedEnabled" @update:checked="sitespeedEnabled = $event" />
            <span class="text-sm font-medium">{{ $t("environment.sitespeedEnabled") }}</span>
          </label>

          <div v-if="sitespeedEnabled" class="space-y-3 rounded-lg border bg-muted/30 p-4">
            <label class="text-sm font-medium">{{ $t("environment.sitespeedUrls") }}</label>

            <div v-for="(url, index) in sitespeedUrls" :key="index" class="flex items-center gap-2">
              <Input
                :model-value="sitespeedUrls[index]"
                type="url"
                placeholder="https://example.com"
                class="flex-1"
                @update:model-value="sitespeedUrls[index] = ($event as string)"
              />
              <Button type="button" variant="ghost" size="icon" class="size-8 shrink-0 text-muted-foreground hover:text-destructive" @click="removeSitespeedUrl(index)">
                <icon-fa6-solid:xmark class="size-3.5" />
              </Button>
            </div>

            <Button
              v-if="sitespeedUrls.length < 5"
              type="button"
              variant="outline"
              size="sm"
              @click="addSitespeedUrl"
            >
              <icon-fa6-solid:plus class="mr-1.5 size-3" />
              {{ $t("environment.newUrl") }}
            </Button>

            <p class="text-xs text-muted-foreground">{{ $t("environment.sitespeedScheduleDesc") }}</p>
          </div>

          <div class="flex justify-end">
            <Button
              :disabled="isSitespeedSubmitting || !isSitespeedFormValid"
              type="submit"
            >
              <icon-fa6-solid:floppy-disk v-if="!isSitespeedSubmitting" class="mr-1.5 size-3.5" aria-hidden="true" />
              <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
              {{ $t("environment.saveSitespeedSettings") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Danger zone -->
    <Card class="border-destructive/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base text-destructive">
          <icon-fa6-solid:triangle-exclamation class="size-4" />
          {{ $t('environment.deleteEnvironmentTitle', { name: environment.name }) }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("environment.deleteEnvironmentWarning") }}</p>
        <Button
          type="button"
          variant="destructive"
          @click="showEnvironmentDeletionModal = true"
        >
          <icon-fa6-solid:trash class="mr-1.5 size-3.5" />
          {{ $t("environment.deleteEnvironment") }}
        </Button>
      </CardContent>
    </Card>

    <!-- Modals -->
    <DeleteConfirmationModal
      :show="showEnvironmentDeletionModal"
      :title="$t('environment.deleteEnvironment')"
      :entity-name="environment?.name || 'this environment'"
      @close="showEnvironmentDeletionModal = false"
      @confirm="deleteEnvironment"
    />

    <PluginConnectionModal
      v-model:base64="pluginBase64"
      :show="showPluginModal"
      :error="pluginError"
      @close="closePluginModal"
      @import="processPluginData"
    />
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useForm } from "vee-validate";
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import PluginConnectionModal from "@/components/modal/PluginConnectionModal.vue";

const { t } = useI18n();
const { error, success } = useAlert();
const router = useRouter();
const route = useRoute();
const environment = ref<components["schemas"]["EnvironmentDetail"] | null>(null);
const isLoading = ref(false);
const shops = ref<components["schemas"]["AccountShop"][]>([]);
const selectedShopId = ref<number>(0);

const environmentId = Number.parseInt(route.params.environmentId as string, 10);

api.GET("/account/shops").then(({ data }) => {
  if (data) shops.value = data;
});

async function loadEnvironment() {
  isLoading.value = true;
  const { data } = await api.GET("/environments/{environmentId}", {
    params: { path: { environmentId } },
  });
  environment.value = data ?? null;

  if (environment.value?.organizationId) {
    const { data: shopsData } = await api.GET("/account/shops");
    if (shopsData) shops.value = shopsData;
    selectedShopId.value = environment.value.shopId ?? 0;
  }

  if (environment.value) {
    sitespeedEnabled.value = environment.value.sitespeedEnabled ?? false;
    sitespeedUrls.value = environment.value.sitespeedUrls
      ? [...environment.value.sitespeedUrls]
      : [];

    setFieldValue("name", environment.value.name);
    setFieldValue("shopUrl", environment.value.url);
    setFieldValue("shopId", String(environment.value.shopId ?? ""));
  }

  isLoading.value = false;
}

loadEnvironment();

const showEnvironmentDeletionModal = ref(false);
const sitespeedUrls = ref<string[]>([]);
const sitespeedEnabled = ref(false);
const isSitespeedSubmitting = ref(false);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

const isSitespeedFormValid = computed(() => {
  if (!sitespeedEnabled.value) return true;
  return sitespeedUrls.value.some((url) => url.trim() !== "");
});

const schema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.environmentNameRequired")),
    shopUrl: z.string().min(1, t("validation.environmentUrlRequired")).url(t("validation.environmentUrlInvalid")),
    shopId: z.string().min(1, t("validation.shopRequired")),
    clientId: z.string().optional(),
    clientSecret: z.string().optional(),
  })
);

const { handleSubmit, isSubmitting, setFieldValue } = useForm({
  validationSchema: schema,
});

const onSubmit = handleSubmit(async (values) => {
  if (environment.value) {
    try {
      let shopUrl = values.shopUrl;
      if (shopUrl) {
        shopUrl = shopUrl.replace(/\/+$/, "");
      }
      await api.PATCH("/environments/{environmentId}", {
        params: { path: { environmentId: environment.value.id } },
        body: {
          name: values.name,
          shopUrl,
          clientId: values.clientId,
          clientSecret: values.clientSecret,
          shopId: Number(values.shopId),
        },
      });

      router.push({
        name: "account.environments.detail",
        params: {
          organizationId: route.params.organizationId as string,
          environmentId: environment.value.id,
        },
      });
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    }
  }
});

async function deleteEnvironment() {
  if (environment.value) {
    try {
      await api.DELETE("/environments/{environmentId}", {
        params: { path: { environmentId: environment.value.id } },
      });
      router.push({ name: "account.shop.list" });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
}

function addSitespeedUrl() {
  sitespeedUrls.value.push("");
}

function removeSitespeedUrl(index: number) {
  sitespeedUrls.value.splice(index, 1);
}

async function onSitespeedSubmit() {
  if (environment.value) {
    try {
      isSitespeedSubmitting.value = true;

      if (sitespeedEnabled.value && sitespeedUrls.value.length === 0) {
        error(t("environment.sitespeedUrlRequired"));
        return;
      }

      await api.PUT("/environments/{environmentId}/sitespeed-settings", {
        params: { path: { environmentId: environment.value.id } },
        body: {
          enabled: sitespeedEnabled.value,
          urls: sitespeedUrls.value,
        },
      });

      await loadEnvironment();
      success(t("environment.sitespeedSaved"));
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isSitespeedSubmitting.value = false;
    }
  }
}

function openPluginModal() {
  showPluginModal.value = true;
  pluginBase64.value = "";
  pluginError.value = "";
}

const closePluginModal = () => {
  showPluginModal.value = false;
  pluginBase64.value = "";
  pluginError.value = "";
};

function processPluginData() {
  try {
    pluginError.value = "";

    if (!pluginBase64.value.trim()) {
      pluginError.value = t("environment.base64Error");
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = t("environment.base64InvalidData");
      return;
    }

    setFieldValue("shopUrl", data.url);
    setFieldValue("clientId", data.clientId);
    setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("environment.base64InvalidFormat");
  }
}
</script>
