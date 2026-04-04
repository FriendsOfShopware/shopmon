<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("environment.newEnvironment") }}</h1>

    <!-- No shops yet — onboarding -->
    <Card v-if="shops.length === 0 && shopsLoaded">
      <CardContent class="flex flex-col items-center px-6 py-12 text-center">
        <div class="mb-4 flex size-14 items-center justify-center rounded-2xl bg-primary/10">
          <icon-fa6-solid:folder-tree class="size-6 text-primary" />
        </div>

        <h3 class="mb-2 text-xl font-semibold">Create a shop first</h3>
        <p class="mb-6 max-w-sm text-muted-foreground">
          Shops group related environments together. For example:
        </p>

        <div
          class="mb-8 inline-flex flex-col gap-2 rounded-lg border bg-muted/50 px-6 py-4 text-left"
        >
          <div class="flex items-center gap-2 font-medium">
            <icon-fa6-solid:folder class="size-3 text-primary" />
            <span>Toy Shop X</span>
          </div>
          <div class="ml-6 flex flex-col gap-1.5 border-l-2 pl-3">
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <icon-fa6-solid:shop class="size-3 text-primary/60" />
              Production
            </div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <icon-fa6-solid:shop class="size-3 text-primary/60" />
              Staging
            </div>
          </div>
        </div>

        <Button as-child>
          <RouterLink :to="{ name: 'account.shops.new', query: { redirect: $route.fullPath } }">
            <icon-fa6-solid:plus class="mr-1.5 size-3.5" />
            Create Shop
          </RouterLink>
        </Button>
      </CardContent>
    </Card>

    <!-- Add environment form -->
    <template v-else-if="shopsLoaded">
      <Card>
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:gear class="size-4 text-muted-foreground" />
            {{ $t("environment.environmentInfo") }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form @submit="onSubmit" class="space-y-6">
            <div class="grid gap-4 sm:grid-cols-2">
              <FormField v-slot="{ componentField }" name="name">
                <FormItem>
                  <FormLabel>{{ $t("common.name") }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="shopId">
                <FormItem>
                  <FormLabel>{{ $t("environment.shop") }}</FormLabel>
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
                  <FormLabel>{{ $t("common.url") }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" autocomplete="url" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <Separator />

            <!-- Bypass auth header -->
            <div>
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-semibold">{{ $t("environment.bypassAuthHeader") }}</h3>
                  <p class="mt-0.5 text-xs text-muted-foreground">
                    {{ $t("environment.bypassAuthHeaderDesc") }}
                  </p>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-7 shrink-0"
                  title="Copy token"
                  @click="copyToken"
                >
                  <icon-fa6-solid:copy class="size-3" />
                </Button>
              </div>
              <code class="mt-2 block rounded bg-muted px-3 py-2 font-mono text-xs break-all">{{
                shopToken
              }}</code>
            </div>

            <Separator />

            <!-- Integration -->
            <div>
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-semibold">{{ $t("environment.integration") }}</h3>
                  <p class="mt-0.5 text-xs text-muted-foreground">
                    <i18n-t keypath="environment.integrationDesc" tag="span">
                      <template #pluginLink>
                        <a
                          href="https://github.com/FriendsOfShopware/FroshShopmon"
                          target="_blank"
                          class="text-primary hover:underline"
                          >{{ $t("environment.pluginName") }}</a
                        >
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
                    <FormLabel>{{ $t("environment.clientId") }}</FormLabel>
                    <FormControl>
                      <Input v-bind="componentField" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                </FormField>

                <FormField v-slot="{ componentField }" name="clientSecret">
                  <FormItem>
                    <FormLabel>{{ $t("environment.clientSecret") }}</FormLabel>
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

      <!-- Plugin Connection Modal -->
      <PluginConnectionModal
        v-model:base64="pluginBase64"
        :show="showPluginModal"
        :error="pluginError"
        @close="closePluginModal"
        @import="processPluginData"
      />
    </template>
  </div>
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import PluginConnectionModal from "@/components/modal/PluginConnectionModal.vue";

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
    return !!new URL(url);
  } catch {
    return false;
  }
};

const schema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.environmentNameRequired")),
    shopUrl: z
      .string()
      .min(1, t("validation.environmentUrlRequired"))
      .refine((val) => isValidUrl(val), t("validation.environmentUrlInvalid")),
    shopId: z.string().min(1, t("validation.shopRequired")),
    clientId: z.string().min(1, t("validation.clientIdRequired")),
    clientSecret: z.string().min(1, t("validation.clientSecretRequired")),
  }),
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
