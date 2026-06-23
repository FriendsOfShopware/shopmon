<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h1 class="min-w-0 text-2xl font-bold tracking-tight break-words">
        {{ shop ? $t("shop.editShop", { name: shop.name }) : $t("nav.editShop") }}
      </h1>
      <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
        <Button variant="outline" size="sm" class="w-full sm:w-auto" as-child>
          <RouterLink :to="{ name: 'account.shop.list' }">
            <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
            {{ $t("shop.backToShops") }}
          </RouterLink>
        </Button>
        <Button v-if="shop" size="sm" class="w-full sm:w-auto" as-child>
          <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop.id } }">
            <icon-fa6-solid:plus class="mr-1.5 size-3" />
            {{ $t("environment.addEnvironment") }}
          </RouterLink>
        </Button>
      </div>
    </div>

    <!-- Loading -->
    <Card v-if="isPageLoading">
      <CardContent class="flex items-center gap-2 py-8">
        <icon-line-md:loading-twotone-loop class="size-5" />
        {{ $t("shop.loadingShop") }}
      </CardContent>
    </Card>

    <!-- Not found -->
    <Card v-else-if="!shop">
      <CardContent class="py-8 text-center text-muted-foreground">
        {{ $t("shop.shopNotFound") }}
      </CardContent>
    </Card>

    <template v-else>
      <!-- Shop info form -->
      <Card>
        <CardHeader class="px-4 pb-3 sm:px-6">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:folder class="size-4 text-muted-foreground" />
            {{ $t("shop.shopInfo") }}
          </CardTitle>
        </CardHeader>
        <CardContent class="px-4 sm:px-6">
          <form @submit="onSubmitShop" class="space-y-4">
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

            <div class="flex justify-end">
              <Button :disabled="isShopSubmitting || isSavingShop" type="submit">
                <icon-fa6-solid:floppy-disk
                  v-if="!(isShopSubmitting || isSavingShop)"
                  class="mr-1.5 size-3.5"
                />
                <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
                {{ $t("common.save") }}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <!-- Environments -->
      <Card v-if="shopEnvironments.length > 0">
        <CardHeader class="px-4 pb-3 sm:px-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <CardTitle class="flex items-center gap-2 text-base">
              <icon-fa6-solid:earth-americas class="size-4 text-muted-foreground" />
              {{ $t("environment.title") }}
            </CardTitle>
            <Button size="sm" class="w-full sm:w-auto" as-child>
              <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop!.id } }">
                <icon-fa6-solid:plus class="mr-1.5 size-3" />
                {{ $t("environment.addEnvironment") }}
              </RouterLink>
            </Button>
          </div>
        </CardHeader>
        <CardContent class="space-y-2 px-4 sm:px-6">
          <div
            v-for="env in shopEnvironments"
            :key="env.id"
            :class="[
              'flex flex-col gap-3 rounded-lg border p-3 transition-colors sm:flex-row sm:items-center sm:px-4',
              env.id === shop!.defaultEnvironmentId
                ? 'border-primary/30 bg-primary/5'
                : 'hover:bg-muted/50',
            ]"
          >
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div
                class="flex size-9 shrink-0 items-center justify-center rounded-lg border bg-muted"
              >
                <img v-if="env.favicon" :src="env.favicon" alt="" class="size-5 rounded" />
                <icon-fa6-solid:earth-americas v-else class="size-3.5 text-muted-foreground/50" />
              </div>

              <div class="min-w-0 flex-1">
                <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                  <RouterLink
                    :to="{ name: 'account.environments.detail', params: { environmentId: env.id } }"
                    class="min-w-0 max-w-full truncate font-medium transition-colors hover:text-primary"
                  >
                    {{ env.name }}
                  </RouterLink>
                  <StatusIcon :status="env.status" />
                  <Badge
                    v-if="env.id === shop!.defaultEnvironmentId"
                    variant="default"
                    class="text-[10px]"
                  >
                    {{ $t("common.default") }}
                  </Badge>
                </div>
                <div class="mt-0.5 flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                  <Badge variant="secondary" class="shrink-0 font-mono text-[10px]">{{
                    env.shopwareVersion
                  }}</Badge>
                  <span v-if="env.url" class="min-w-0 truncate">{{ env.url }}</span>
                </div>
              </div>
            </div>

            <div
              class="flex w-full shrink-0 items-center justify-end gap-1 border-t pt-2 sm:w-auto sm:border-t-0 sm:pt-0"
            >
              <Button
                v-if="env.id !== shop!.defaultEnvironmentId"
                variant="ghost"
                size="sm"
                class="text-xs text-muted-foreground"
                :disabled="isSettingDefaultEnv === env.id"
                @click="setDefaultEnvironment(env.id)"
              >
                <icon-fa6-solid:star v-if="isSettingDefaultEnv !== env.id" class="mr-1 size-3" />
                <icon-line-md:loading-twotone-loop v-else class="mr-1 size-3" />
                {{ $t("shop.setDefault") }}
              </Button>
              <Button
                as-child
                variant="ghost"
                size="icon"
                class="size-8"
                :title="$t('common.edit')"
              >
                <RouterLink
                  :to="{
                    name: 'account.environments.edit',
                    params: { environmentId: env.id },
                  }"
                >
                  <icon-fa6-solid:pencil class="size-3.5" />
                </RouterLink>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- API Keys -->
      <ApiKeysCard :org-id="shop.organizationId" :shop-id="shop.id" />

      <!-- Packages Tokens -->
      <PackagesTokensCard
        v-if="instanceConfig?.packageMirrorEnabled !== false"
        :org-id="shop.organizationId"
        :shop-id="shop.id"
      />

      <!-- Danger zone -->
      <Card class="border-destructive/30">
        <CardHeader class="px-4 pb-3 sm:px-6">
          <CardTitle class="flex items-center gap-2 text-base text-destructive">
            <icon-fa6-solid:triangle-exclamation class="size-4" />
            {{ $t("shop.dangerZone") }}
          </CardTitle>
        </CardHeader>
        <CardContent class="px-4 sm:px-6">
          <p class="mb-3 text-sm text-muted-foreground">{{ $t("shop.deleteShopWarning") }}</p>
          <p v-if="!canDeleteShop" class="mb-3 text-sm text-destructive">
            {{ $t("shop.deleteShopEnvironmentsWarning", { count: environmentsInShopCount }) }}
          </p>
          <Button
            variant="destructive"
            :disabled="!canDeleteShop || isDeletingShop"
            @click="showDeleteShopModal = true"
          >
            <icon-fa6-solid:trash class="mr-1.5 size-3.5" />
            {{ $t("shop.deleteShop") }}
          </Button>
        </CardContent>
      </Card>
    </template>

    <!-- Delete Shop -->
    <DeleteConfirmationModal
      :show="showDeleteShopModal"
      :title="$t('shop.deleteShopTitle')"
      :entity-name="shop?.name || $t('shop.thisShop')"
      @close="showDeleteShopModal = false"
      @confirm="deleteShop"
    />
  </div>
</template>

<script setup lang="ts">
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import ApiKeysCard from "@/components/shop/ApiKeysCard.vue";
import PackagesTokensCard from "@/components/shop/PackagesTokensCard.vue";
import { useAlert } from "@/composables/useAlert";
import { useInstanceConfig } from "@/composables/useInstanceConfig";
import { fetchAccountEnvironments } from "@/composables/useAccountEnvironments";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import StatusIcon from "@/components/StatusIcon.vue";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRoute, useRouter } from "vue-router";

type Shop = components["schemas"]["AccountShop"];
type AccountEnvironment = components["schemas"]["AccountEnvironment"];

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const alert = useAlert();
const { config: instanceConfig } = useInstanceConfig();

const shopId = Number(route.params.shopId);

const shop = ref<Shop | null>(null);
const shopEnvironments = ref<AccountEnvironment[]>([]);
const environmentsInShopCount = ref(0);
const isSettingDefaultEnv = ref<number | null>(null);

const isPageLoading = ref(true);
const isSavingShop = ref(false);
const isDeletingShop = ref(false);

const showDeleteShopModal = ref(false);

const canDeleteShop = computed(() => environmentsInShopCount.value === 0);

const shopValidationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.shopNameRequired")),
    description: z.string().optional(),
    gitUrl: z.string().url(t("validation.urlInvalid")).optional().or(z.literal("")),
  }),
);

const {
  handleSubmit: handleShopSubmit,
  isSubmitting: isShopSubmitting,
  setValues: setShopValues,
} = useForm({
  validationSchema: shopValidationSchema,
  initialValues: { name: "", description: "", gitUrl: "" },
});

watch(shop, (s) => {
  if (s) {
    setShopValues({ name: s.name ?? "", description: s.description ?? "", gitUrl: s.gitUrl ?? "" });
  }
});

async function loadShopSummary() {
  const [shopsRes, environmentsData] = await Promise.all([
    api.GET("/account/shops"),
    fetchAccountEnvironments(),
  ]);
  const shopsData = shopsRes.data ?? [];
  shop.value = shopsData.find((currentShop) => currentShop.id === shopId) ?? null;
  const envs = environmentsData.filter((env) => env.shopId === shopId);
  shopEnvironments.value = envs;
  environmentsInShopCount.value = envs.length;
}

async function setDefaultEnvironment(envId: number) {
  if (!shop.value) return;
  isSettingDefaultEnv.value = envId;
  try {
    const { error } = await api.PATCH("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: { defaultEnvironmentId: envId },
    });
    if (error) {
      alert.error((error as { message?: string }).message ?? t("shop.defaultEnvSet"));
      return;
    }
    await loadShopSummary();
    alert.success(t("shop.defaultEnvSet"));
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  } finally {
    isSettingDefaultEnv.value = null;
  }
}

async function loadPageData() {
  isPageLoading.value = true;
  try {
    await loadShopSummary();
    if (!shop.value) {
      alert.error(t("shop.shopNotFound"));
      return;
    }
  } catch (error) {
    alert.error(`${t("shop.failedLoadShop")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isPageLoading.value = false;
  }
}

const onSubmitShop = handleShopSubmit(async (values) => {
  if (!shop.value) return;
  isSavingShop.value = true;
  try {
    const { error } = await api.PATCH("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: {
        name: values.name,
        description: values.description ?? "",
        gitUrl: values.gitUrl || undefined,
      },
    });
    if (error) {
      alert.error(
        `${t("shop.failedUpdateShop")}: ${(error as { message?: string }).message ?? ""}`,
      );
      return;
    }
    await loadShopSummary();
    alert.success(t("shop.shopUpdated"));
  } catch (error) {
    alert.error(
      `${t("shop.failedUpdateShop")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isSavingShop.value = false;
  }
});

async function deleteShop() {
  if (!shop.value) return;
  isDeletingShop.value = true;
  try {
    const { error } = await api.DELETE("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
    });
    if (error) {
      alert.error(
        `${t("shop.failedDeleteShop")}: ${(error as { message?: string }).message ?? ""}`,
      );
      return;
    }
    alert.success(t("shop.shopDeleted"));
    router.push({ name: "account.shop.list" });
  } catch (error) {
    alert.error(
      `${t("shop.failedDeleteShop")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isDeletingShop.value = false;
    showDeleteShopModal.value = false;
  }
}

onMounted(() => {
  if (Number.isNaN(shopId)) {
    isPageLoading.value = false;
    alert.error(t("shop.invalidShop"));
    return;
  }
  loadPageData();
});
</script>
