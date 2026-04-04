<template>
  <header-container :title="$t('environment.newEnvironment')" />
  <main-container>
    <Card v-if="shops.length === 0 && shopsLoaded">
      <CardContent>
        <div class="text-center pt-8 pb-4 px-4">
          <div class="text-4xl text-primary opacity-70 mb-4">
            <icon-fa6-solid:folder-tree aria-hidden="true" />
          </div>

          <h3 class="text-xl font-semibold mb-2">Create a shop first</h3>

          <p class="text-muted-foreground mb-6">
            Shops group related environments together. For example:
          </p>

          <div class="inline-flex flex-col gap-2 bg-muted border rounded-lg px-6 py-4 mb-8 text-left">
            <div class="flex items-center gap-2 font-medium">
              <icon-fa6-solid:folder class="text-xs text-primary opacity-80" aria-hidden="true" />
              <span>Toy Shop X</span>
            </div>
            <div class="flex flex-col gap-1.5 ml-6 pl-3 border-l-2">
              <div class="flex items-center gap-2 text-sm text-muted-foreground">
                <icon-fa6-solid:shop class="text-xs text-primary opacity-80" aria-hidden="true" />
                Production
              </div>
              <div class="flex items-center gap-2 text-sm text-muted-foreground">
                <icon-fa6-solid:shop class="text-xs text-primary opacity-80" aria-hidden="true" />
                Staging
              </div>
            </div>
          </div>

          <div class="mt-8">
            <Button as-child>
              <RouterLink :to="{ name: 'account.shops.new', query: { redirect: $route.fullPath } }">
                <icon-fa6-solid:plus class="size-3.5" aria-hidden="true" />
                Create Shop
              </RouterLink>
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card v-else>
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
                      <Input v-bind="componentField" />
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
              <h3 class="text-lg font-semibold mb-1">{{ $t('environment.bypassAuthHeader') }}</h3>
              <p class="text-sm text-muted-foreground mb-4">{{ $t("environment.bypassAuthHeaderDesc") }}</p>

              <div class="flex items-center gap-2">
                <code class="text-xs break-all">{{ shopToken }}</code>
                <Button type="button" size="icon-sm" variant="outline" @click="copyToken">
                  <icon-fa6-solid:copy />
                </Button>
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
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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

function generateShopToken(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32));
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

const shopToken = generateShopToken();

async function copyToken() {
  await navigator.clipboard.writeText(shopToken);
  success(t("environment.tokenCopied"));
}

const shops = ref<components["schemas"]["AccountShop"][]>([]);
const shopsLoaded = ref(false);
const selectedShopId = ref<number>(route.query.shopId ? Number(route.query.shopId) : 0);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

api.GET("/account/shops").then(({ data }) => {
  if (data) {
    shops.value = data;
    if (!selectedShopId.value && data.length > 0) {
      selectedShopId.value = data[0].id;
    }
  }
  shopsLoaded.value = true;
});

const isValidUrl = (url: string) => {
  try {
    const parsedUrl = new URL(url);
    return !!parsedUrl;
  } catch (_e) {
    return false;
  }
};

const schema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.environmentNameRequired")),
    shopUrl: z.string().min(1, t("validation.environmentUrlRequired")).refine((val) => isValidUrl(val), t("validation.environmentUrlInvalid")),
    shopId: z.string().min(1, t("validation.shopRequired")),
    clientId: z.string().min(1, t("validation.clientIdRequired")),
    clientSecret: z.string().min(1, t("validation.clientSecretRequired")),
  })
);

const { handleSubmit, isSubmitting, setFieldValue } = useForm({
  validationSchema: schema,
  initialValues: {
    shopId: selectedShopId.value ? String(selectedShopId.value) : "",
  },
});

const onSubmit = handleSubmit(async (values) => {
  try {
    const { error: apiError } = await api.POST("/environments", {
      body: {
        name: values.name,
        shopUrl: values.shopUrl.replace(/\/+$/, ""),
        clientId: values.clientId,
        clientSecret: values.clientSecret,
        shopId: Number(values.shopId),
        shopToken,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

    router.push({ name: "account.shop.list" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
});

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
