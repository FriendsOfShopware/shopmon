<template>
  <header-container
    v-if="environment"
    :title="$t('environment.editEnvironment', { name: environment.name })"
  >
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
  </header-container>

  <main-container v-if="environment">
    <Card>
      <CardContent>
        <form @submit="onSubmit">
          <div class="space-y-6">
            <div>
              <h3 class="text-lg font-semibold mb-4">{{ $t('environment.environmentInfo') }}</h3>
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

                <FormField v-slot="{ componentField }" name="shopUrl">
                  <FormItem>
                    <FormLabel>{{ $t('common.url') }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" autocomplete="url" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>
            </div>

            <div>
              <h3 class="text-lg font-semibold mb-1">{{ $t('environment.integration') }}</h3>
              <p class="text-sm text-muted-foreground mb-4">
                <i18n-t keypath="environment.integrationDesc" tag="span">
                  <template #pluginLink>
                    <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank" class="text-primary hover:underline">{{
                      $t("environment.pluginName")
                    }}</a>
                  </template>
                </i18n-t>
                <a
                  href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions"
                  class="text-primary hover:underline ml-1"
                >
                  {{ $t("environment.permissions") }}
                </a>
              </p>

              <div class="space-y-4">
                <Button type="button" variant="outline" @click="openPluginModal">
                  {{ $t("environment.connectPlugin") }}
                </Button>

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

            <div class="flex justify-end pt-4">
              <Button :disabled="isSubmitting" type="submit">
                <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="size-3.5" aria-hidden="true" />
                <icon-line-md:loading-twotone-loop v-else class="size-3.5" />
                {{ $t("common.save") }}
              </Button>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>

    <Card v-if="environment">
      <CardHeader>
        <CardTitle>{{ $t('environment.sitespeed') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="space-y-2 mb-4">
          <p class="text-sm text-muted-foreground">{{ $t("environment.sitespeedDesc") }}</p>
          <p class="text-sm text-muted-foreground">{{ $t("environment.sitespeedActivateDesc") }}</p>
          <p class="text-sm text-muted-foreground">{{ $t("environment.sitespeedScheduleDesc") }}</p>
        </div>

        <form @submit.prevent="onSitespeedSubmit">
          <div class="mb-4">
            <label for="sitespeedEnabled" class="flex items-center gap-2 text-sm font-medium">
              <input id="sitespeedEnabled" v-model="sitespeedEnabled" type="checkbox" class="rounded" />
              {{ $t("environment.sitespeedEnabled") }}
            </label>
          </div>

          <div v-if="sitespeedEnabled" class="space-y-2">
            <label class="text-sm font-medium">{{ $t("environment.sitespeedUrls") }}</label>

            <div class="space-y-2 mt-2">
              <div v-for="(url, index) in sitespeedUrls" :key="index" class="flex items-center gap-2">
                <Input
                  :model-value="sitespeedUrls[index]"
                  type="url"
                  placeholder="https://example.com"
                  class="flex-1"
                  @update:model-value="sitespeedUrls[index] = ($event as string)"
                />
                <Button type="button" variant="outline" size="icon-sm" @click="removeSitespeedUrl(index)">
                  <icon-fa6-solid:xmark />
                </Button>
              </div>

              <Button
                v-if="sitespeedUrls.length < 5"
                type="button"
                variant="outline"
                @click="addSitespeedUrl"
              >
                <icon-fa6-solid:plus class="size-3.5" />
                {{ $t("environment.newUrl") }}
              </Button>
            </div>
          </div>

          <div class="flex justify-end pt-4">
            <Button
              :disabled="isSitespeedSubmitting || !isSitespeedFormValid"
              type="submit"
            >
              <icon-fa6-solid:floppy-disk
                v-if="!isSitespeedSubmitting"
                class="size-3.5"
                aria-hidden="true"
              />
              <icon-line-md:loading-twotone-loop v-else class="size-3.5" />
              {{ $t("environment.saveSitespeedSettings") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t('environment.deleteEnvironmentTitle', { name: environment.name }) }}</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-sm text-muted-foreground mb-4">{{ $t("environment.deleteEnvironmentWarning") }}</p>

        <Button
          type="button"
          variant="destructive"
          @click="showEnvironmentDeletionModal = true"
        >
          <icon-fa6-solid:trash class="size-3.5" />
          {{ $t("environment.deleteEnvironment") }}
        </Button>
      </CardContent>
    </Card>

    <delete-confirmation-modal
      :show="showEnvironmentDeletionModal"
      :title="$t('environment.deleteEnvironment')"
      :entity-name="environment?.name || 'this environment'"
      @close="showEnvironmentDeletionModal = false"
      @confirm="deleteEnvironment"
    />

    <!-- Plugin Connection Modal -->
    <plugin-connection-modal
      v-model:base64="pluginBase64"
      :show="showPluginModal"
      :error="pluginError"
      @close="closePluginModal"
      @import="processPluginData"
    />
  </main-container>
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

  // Load shops for the organization
  if (environment.value?.organizationId) {
    const { data: shopsData } = await api.GET("/account/shops");
    if (shopsData) shops.value = shopsData;
    selectedShopId.value = environment.value.shopId ?? 0;
  }

  // Initialize sitespeed settings
  if (environment.value) {
    sitespeedEnabled.value = environment.value.sitespeedEnabled ?? false;
    sitespeedUrls.value = environment.value.sitespeedUrls
      ? [...environment.value.sitespeedUrls]
      : [];

    // Set form values once environment loads
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
  if (!sitespeedEnabled.value) {
    return true;
  }
  const hasNonEmptyUrl = sitespeedUrls.value.some((url) => url.trim() !== "");
  return hasNonEmptyUrl;
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
