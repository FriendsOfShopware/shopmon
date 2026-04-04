<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">
        {{ shop ? $t('shop.editShop', { name: shop.name }) : $t('nav.editShop') }}
      </h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button variant="outline" as-child>
        <RouterLink :to="{ name: 'account.shop.list' }">
          <icon-fa6-solid:arrow-left class="size-4" aria-hidden="true" />
          {{ $t("shop.backToShops") }}
        </RouterLink>
      </Button>

      <Button v-if="shop" variant="outline" as-child>
        <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop.id } }">
          <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
          {{ $t("environment.addEnvironment") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div class="space-y-6">
    <Card v-if="isPageLoading">
      <CardContent class="flex items-center gap-2 py-6">
        <icon-line-md:loading-twotone-loop class="size-5" />
        {{ $t("shop.loadingShop") }}
      </CardContent>
    </Card>

    <Card v-else-if="!shop">
      <CardContent class="py-6">
        <p>{{ $t("shop.shopNotFound") }}</p>
      </CardContent>
    </Card>

    <template v-else>
      <Card>
        <CardHeader>
          <CardTitle>{{ $t('shop.shopInfo') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <form @submit="onSubmitShop">
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

              <FormField v-slot="{ componentField }" name="description">
                <FormItem>
                  <FormLabel>{{ $t('common.description') }}</FormLabel>
                  <FormControl>
                    <Textarea
                      v-bind="componentField"
                      :placeholder="$t('shop.optionalDescription')"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="gitUrl">
                <FormItem>
                  <FormLabel>{{ $t('shop.gitRepoUrl') }}</FormLabel>
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

            <div class="flex justify-end pt-6">
              <Button
                :disabled="isShopSubmitting || isSavingShop"
                type="submit"
              >
                <icon-fa6-solid:floppy-disk
                  v-if="!(isShopSubmitting || isSavingShop)"
                  class="size-4"
                  aria-hidden="true"
                />
                <icon-line-md:loading-twotone-loop v-else class="size-4" />
                {{ $t("common.save") }}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card id="api-keys">
        <CardHeader>
          <CardTitle>{{ $t('shop.apiKeys') }}</CardTitle>
          <CardAction>
            <Button type="button" @click="openAddKeyModal">
              <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
              {{ $t("shop.createApiKey") }}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <Alert class="mb-4">
            <AlertTitle>{{ $t("shop.apiKeys") }}</AlertTitle>
            <AlertDescription>
              {{ $t("shop.apiKeyInfo") }}
            </AlertDescription>
          </Alert>

          <div v-if="isApiKeysLoading" class="py-12 text-center text-muted-foreground">
            <icon-line-md:loading-twotone-loop class="size-6 mx-auto mb-2" />
            {{ $t("shop.loadingApiKeys") }}
          </div>

          <div v-else-if="apiKeys.length === 0" class="py-16 text-center">
            <icon-fa6-solid:key class="size-12 text-muted-foreground mx-auto mb-4" aria-hidden="true" />
            <p>{{ $t("shop.noApiKeys") }}</p>
            <p class="mt-2 text-muted-foreground text-sm">
              {{ $t("shop.noApiKeysHint") }}
            </p>
          </div>

          <div v-else class="space-y-4 pt-4">
            <div
              v-for="apiKey in apiKeys"
              :key="apiKey.id"
              class="flex justify-between items-center p-4 border rounded-lg"
            >
              <div>
                <h4 class="text-lg font-medium mb-1">{{ apiKey.name }}</h4>
                <p class="text-sm text-muted-foreground mb-1">
                  {{ $t("shop.createdDate", { date: formatDate(apiKey.createdAt) }) }}
                </p>
                <p v-if="apiKey.lastUsedAt" class="text-sm text-muted-foreground mb-1">
                  {{ $t("shop.lastUsedDate", { date: formatDate(apiKey.lastUsedAt) }) }}
                </p>
                <div class="flex gap-2 mt-2">
                  <Badge v-for="scope in apiKey.scopes" :key="scope">
                    {{ getScopeLabel(scope) }}
                  </Badge>
                </div>
              </div>
              <div class="flex gap-2">
                <Button
                  type="button"
                  variant="destructive"
                  @click="confirmDeleteKey(apiKey)"
                >
                  <icon-fa6-solid:trash class="size-4" aria-hidden="true" />
                  {{ $t("common.delete") }}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card
        v-if="isPackagesConfigured"
        id="packages-tokens"
      >
        <CardHeader>
          <CardTitle>{{ $t('packages.title') }}</CardTitle>
          <CardDescription>{{ $t('packages.description') }}</CardDescription>
          <CardAction>
            <Button
              type="button"
              @click="showAddPackagesTokenModal = true"
            >
              <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
              {{ $t("packages.addToken") }}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div v-if="isPackagesTokensLoading" class="py-12 text-center text-muted-foreground">
            <icon-line-md:loading-twotone-loop class="size-6 mx-auto mb-2" />
            {{ $t("packages.loading") }}
          </div>

          <div v-else-if="packagesTokens.length === 0" class="py-16 text-center">
            <icon-fa6-solid:cube class="size-12 text-muted-foreground mx-auto mb-4" aria-hidden="true" />
            <p>{{ $t("packages.noTokens") }}</p>
            <p class="mt-2 text-muted-foreground text-sm">{{ $t("packages.noTokensHint") }}</p>
          </div>

          <template v-else>
            <div class="space-y-4">
              <div
                v-for="pt in packagesTokens"
                :key="pt.id"
                class="flex justify-between items-center p-4 border rounded-lg"
              >
                <div>
                  <h4 class="text-lg font-medium mb-1">Token #{{ pt.id }}</h4>
                  <p v-if="pt.lastSyncedAt" class="text-sm text-muted-foreground">
                    {{
                      $t("packages.lastSynced", {
                        time: timeAgo(new Date(pt.lastSyncedAt).getTime() / 1000),
                      })
                    }}
                  </p>
                  <p v-else class="text-sm text-muted-foreground">{{ $t("packages.notSyncedYet") }}</p>
                </div>
                <div class="flex gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    :disabled="isSyncingPackagesToken === pt.id"
                    @click="syncPackagesToken(pt)"
                  >
                    <icon-fa6-solid:arrows-rotate
                      v-if="isSyncingPackagesToken !== pt.id"
                      class="size-4"
                      aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop v-else class="size-4" />
                    {{ $t("packages.sync") }}
                  </Button>
                  <Button
                    type="button"
                    variant="destructive"
                    @click="confirmDeletePackagesToken(pt)"
                  >
                    <icon-fa6-solid:trash class="size-4" aria-hidden="true" />
                    {{ $t("common.delete") }}
                  </Button>
                </div>
              </div>
            </div>

            <div v-if="packagesComposerUrl" class="mt-6 pt-6 border-t">
              <h4 class="text-base font-medium mb-2">{{ $t("packages.composerSetup") }}</h4>
              <Alert variant="destructive" class="mb-4">
                <AlertDescription>
                  {{ $t("packages.composerWarning") }}
                </AlertDescription>
              </Alert>
              <p class="text-sm text-muted-foreground mb-2">
                {{ $t("packages.composerRepoHint") }}
              </p>
              <div class="relative my-2 bg-muted border rounded-lg overflow-hidden">
                <pre class="p-4 overflow-x-auto"><code class="font-mono text-sm leading-relaxed">{
    "repositories": [
        {
            "type": "composer",
            "url": "{{ packagesComposerUrl }}"
        }
    ]
}</code></pre>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  class="absolute top-2 right-2"
                  @click="copyComposerRepository"
                >
                  <icon-fa6-solid:copy class="size-3" aria-hidden="true" />
                  {{ $t("common.copy") }}
                </Button>
              </div>
              <p class="text-sm text-muted-foreground mb-2">{{ $t("packages.composerAuthHint") }}</p>
              <div class="relative my-2 bg-muted border rounded-lg overflow-hidden">
                <pre class="p-4 overflow-x-auto"><code class="font-mono text-sm leading-relaxed">composer config --auth bearer.{{ packagesComposerHost }} &lt;your-token&gt;</code></pre>
              </div>
            </div>
          </template>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{{ $t('shop.dangerZone') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-muted-foreground mb-4">{{ $t("shop.deleteShopWarning") }}</p>

          <p v-if="!canDeleteShop" class="text-sm text-destructive mb-4">
            {{ $t("shop.deleteShopEnvironmentsWarning", { count: environmentsInShopCount }) }}
          </p>

          <Button
            type="button"
            variant="destructive"
            :disabled="!canDeleteShop || isDeletingShop"
            @click="showDeleteShopModal = true"
          >
            <icon-fa6-solid:trash class="size-4" aria-hidden="true" />
            {{ $t("shop.deleteShop") }}
          </Button>
        </CardContent>
      </Card>
    </template>

    <!-- Add API Key Modal -->
    <modal :show="showAddKeyModal" close-x-mark @close="closeAddKeyModal">
      <template #title> {{ $t("shop.createApiKeyTitle") }} </template>

      <template #content>
        <form id="apiKeyForm" class="space-y-6" @submit="onSubmitApiKey">
          <div class="space-y-4">
            <FormField v-slot="{ componentField }" name="apiKeyName">
              <FormItem>
                <FormLabel>{{ $t('common.name') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    :placeholder="$t('shop.apiKeyPlaceholder')"
                  />
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">{{ $t("packages.apiKeyHelp") }}</p>
              </FormItem>
            </FormField>

            <div>
              <label class="text-sm font-medium">{{ $t("shop.scopes") }}</label>
              <p class="text-xs text-muted-foreground mb-2">{{ $t("shop.scopesHelp") }}</p>
              <div class="flex flex-col gap-3 mt-2">
                <label
                  v-for="scope in availableScopes"
                  :key="scope.value"
                  class="flex items-start gap-3 p-3 border rounded-lg cursor-pointer hover:bg-muted transition-colors"
                >
                  <input
                    type="checkbox"
                    :value="scope.value"
                    class="mt-1"
                    @change="toggleScope(scope.value)"
                    :checked="selectedScopes.includes(scope.value)"
                  />
                  <div class="flex flex-col gap-1">
                    <span class="font-medium">{{ scope.label }}</span>
                    <span class="text-sm text-muted-foreground">{{ scope.description }}</span>
                  </div>
                </label>
              </div>
              <p v-if="scopeError" class="text-sm text-destructive mt-1">{{ scopeError }}</p>
            </div>
          </div>
        </form>
      </template>

      <template #footer>
        <Button variant="outline" type="button" @click="closeAddKeyModal">
          {{ $t("common.cancel") }}
        </Button>
        <Button
          type="submit"
          form="apiKeyForm"
          :disabled="isCreatingApiKey"
        >
          <icon-fa6-solid:key v-if="!isCreatingApiKey" class="size-4" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-4" />
          {{ $t("shop.createApiKey") }}
        </Button>
      </template>
    </modal>

    <!-- Show Token Modal -->
    <modal :show="showTokenModal" @close="closeTokenModal">
      <template #title> {{ $t("shop.apiKeyCreatedTitle") }} </template>

      <template #content>
        <Alert variant="destructive" class="mb-4">
          <AlertTitle>{{ $t("shop.apiKeyCopyWarning") }}</AlertTitle>
          <AlertDescription>
            {{ $t("shop.apiKeyCopyDesc") }}
          </AlertDescription>
        </Alert>

        <div class="flex gap-2 mt-4 p-4 bg-muted rounded-lg">
          <code class="flex-1 p-3 bg-background border rounded text-sm font-mono break-all">{{ newToken }}</code>
          <Button type="button" variant="outline" @click="copyToken">
            <icon-fa6-solid:copy class="size-4" aria-hidden="true" />
            {{ $t("common.copy") }}
          </Button>
        </div>
      </template>

      <template #footer>
        <Button type="button" @click="closeTokenModal">
          {{ $t("common.done") }}
        </Button>
      </template>
    </modal>

    <!-- Delete API Key Modal -->
    <delete-confirmation-modal
      :show="showDeleteApiKeyModal"
      :title="$t('shop.deleteApiKeyTitle')"
      :entity-name="`the API key '${deletingApiKey?.name}'`"
      :custom-consequence="$t('shop.deleteApiKeyWarning')"
      :reversed-buttons="true"
      :is-loading="isDeletingApiKey"
      :confirm-button-text="$t('shop.deleteApiKeyConfirm')"
      @close="showDeleteApiKeyModal = false"
      @confirm="deleteApiKey"
    />

    <!-- Delete Project Modal -->
    <delete-confirmation-modal
      :show="showDeleteShopModal"
      :title="$t('shop.deleteShopTitle')"
      :entity-name="shop?.name || 'this shop'"
      @close="showDeleteShopModal = false"
      @confirm="deleteShop"
    />

    <!-- Add Packages Token Modal -->
    <modal
      :show="showAddPackagesTokenModal"
      close-x-mark
      @close="showAddPackagesTokenModal = false"
    >
      <template #title> {{ $t("packages.addTokenTitle") }} </template>

      <template #content>
        <form id="packagesTokenForm" class="space-y-4" @submit="onSubmitPackagesToken">
          <FormField v-slot="{ componentField }" name="packagesToken">
            <FormItem>
              <FormLabel>{{ $t('packages.tokenLabel') }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  :placeholder="$t('packages.tokenPlaceholder')"
                />
              </FormControl>
              <FormMessage />
              <p class="text-xs text-muted-foreground">
                {{ $t("packages.tokenHelp") }}
              </p>
            </FormItem>
          </FormField>
        </form>
      </template>

      <template #footer>
        <Button variant="outline" type="button" @click="showAddPackagesTokenModal = false">
          {{ $t("common.cancel") }}
        </Button>
        <Button
          type="submit"
          form="packagesTokenForm"
          :disabled="isCreatingPackagesToken"
        >
          <icon-fa6-solid:plus v-if="!isCreatingPackagesToken" class="size-4" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-4" />
          {{ $t("packages.addToken") }}
        </Button>
      </template>
    </modal>

    <!-- Delete Packages Token Modal -->
    <delete-confirmation-modal
      :show="showDeletePackagesTokenModal"
      :title="$t('packages.deleteTokenTitle')"
      :entity-name="`Token #${deletingPackagesToken?.id}`"
      :custom-consequence="$t('packages.deleteTokenWarning')"
      :reversed-buttons="true"
      :is-loading="isDeletingPackagesToken"
      :confirm-button-text="$t('packages.deleteTokenConfirm')"
      @close="showDeletePackagesTokenModal = false"
      @confirm="deletePackagesToken"
    />
  </div>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { fetchAccountEnvironments } from "@/composables/useAccountEnvironments";
import { formatDate, timeAgo } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

type Shop = components["schemas"]["AccountShop"];
type ApiKey = components["schemas"]["ApiKey"];
type AvailableScope = components["schemas"]["ApiKeyScope"];
type PackagesToken = components["schemas"]["PackagesToken"];

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const alert = useAlert();

const shopId = Number(route.params.shopId);

const shop = ref<Shop | null>(null);
const environmentsInShopCount = ref(0);
const apiKeys = ref<ApiKey[]>([]);
const availableScopes = ref<AvailableScope[]>([]);

const isPageLoading = ref(true);
const isSavingShop = ref(false);
const isApiKeysLoading = ref(false);
const isCreatingApiKey = ref(false);
const isDeletingApiKey = ref(false);
const isDeletingShop = ref(false);

const showAddKeyModal = ref(false);
const showTokenModal = ref(false);
const showDeleteApiKeyModal = ref(false);
const showDeleteShopModal = ref(false);

const deletingApiKey = ref<ApiKey | null>(null);
const newToken = ref("");

// API Key form state
const selectedScopes = ref<string[]>([]);
const scopeError = ref("");

// Packages tokens
const isPackagesConfigured = ref(false);
const packagesComposerUrl = ref<string | null>(null);
const packagesTokens = ref<PackagesToken[]>([]);
const isPackagesTokensLoading = ref(false);
const isCreatingPackagesToken = ref(false);
const isDeletingPackagesToken = ref(false);
const isSyncingPackagesToken = ref<number | null>(null);
const showAddPackagesTokenModal = ref(false);
const showDeletePackagesTokenModal = ref(false);
const deletingPackagesToken = ref<PackagesToken | null>(null);

const canDeleteShop = computed(() => environmentsInShopCount.value === 0);

const packagesComposerHost = computed(() => {
  if (!packagesComposerUrl.value) return "";
  try {
    return new URL(packagesComposerUrl.value).host;
  } catch {
    return packagesComposerUrl.value;
  }
});

// Shop edit form
const shopValidationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.shopNameRequired")),
    description: z.string().optional(),
    gitUrl: z.string().url(t("validation.urlInvalid")).optional().or(z.literal("")),
  }),
);

const { handleSubmit: handleShopSubmit, isSubmitting: isShopSubmitting, setValues: setShopValues } = useForm({
  validationSchema: shopValidationSchema,
  initialValues: {
    name: "",
    description: "",
    gitUrl: "",
  },
});

watch(shop, (s) => {
  if (s) {
    setShopValues({
      name: s.name ?? "",
      description: s.description ?? "",
      gitUrl: s.gitUrl ?? "",
    });
  }
});

// API Key form
const apiKeyValidationSchema = toTypedSchema(
  z.object({
    apiKeyName: z.string().min(1, t("validation.nameRequired")).max(100, t("validation.nameMaxLength")),
  }),
);

const { handleSubmit: handleApiKeySubmit, resetForm: resetApiKeyForm } = useForm({
  validationSchema: apiKeyValidationSchema,
  initialValues: {
    apiKeyName: "",
  },
});

// Packages token form
const packagesTokenValidationSchema = toTypedSchema(
  z.object({
    packagesToken: z.string().min(1, t("validation.tokenRequired")),
  }),
);

const { handleSubmit: handlePackagesTokenSubmit, resetForm: resetPackagesTokenForm } = useForm({
  validationSchema: packagesTokenValidationSchema,
  initialValues: {
    packagesToken: "",
  },
});

function toggleScope(value: string) {
  const idx = selectedScopes.value.indexOf(value);
  if (idx >= 0) {
    selectedScopes.value.splice(idx, 1);
  } else {
    selectedScopes.value.push(value);
  }
  scopeError.value = "";
}

async function loadShopSummary() {
  const [shopsRes, environmentsData] = await Promise.all([
    api.GET("/account/shops"),
    fetchAccountEnvironments(),
  ]);

  const shopsData = shopsRes.data ?? [];

  shop.value = shopsData.find((currentShop) => currentShop.id === shopId) ?? null;
  environmentsInShopCount.value = environmentsData.filter((env) => env.shopId === shopId).length;
}

async function loadApiKeys() {
  if (!shop.value) return;

  isApiKeysLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/shops/{shopId}/api-keys", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
    });
    apiKeys.value = data ?? [];
  } catch (error) {
    alert.error(
      `${t("shop.failedLoadApiKeys")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isApiKeysLoading.value = false;
  }
}

async function loadAvailableScopes() {
  try {
    const { data } = await api.GET("/api-key-scopes");
    availableScopes.value = data ?? [];
  } catch (error) {
    alert.error(
      `${t("shop.failedLoadScopes")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  }
}

async function scrollToHashSection() {
  if (route.hash !== "#api-keys") return;
  await nextTick();
  document.getElementById("api-keys")?.scrollIntoView({ behavior: "smooth" });
}

async function loadPageData() {
  isPageLoading.value = true;

  try {
    await loadShopSummary();

    if (!shop.value) {
      alert.error(t("shop.shopNotFound"));
      return;
    }

    await Promise.all([loadApiKeys(), loadAvailableScopes(), checkPackagesConfigured()]);
    await loadPackagesTokens();
    await scrollToHashSection();
  } catch (error) {
    alert.error(`${t("shop.failedLoadShop")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isPageLoading.value = false;
  }
}

function getScopeLabel(scope: string): string {
  const found = availableScopes.value.find((availableScope) => availableScope.value === scope);
  return found?.label ?? scope;
}

const onSubmitShop = handleShopSubmit(async (values) => {
  if (!shop.value) return;

  isSavingShop.value = true;
  try {
    await api.PATCH("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: {
        name: values.name,
        description: values.description ?? "",
        gitUrl: values.gitUrl || undefined,
      },
    });

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

function openAddKeyModal() {
  selectedScopes.value = [];
  scopeError.value = "";
  resetApiKeyForm({ values: { apiKeyName: "" } });
  showAddKeyModal.value = true;
}

function closeAddKeyModal() {
  showAddKeyModal.value = false;
}

function closeTokenModal() {
  showTokenModal.value = false;
  newToken.value = "";
}

const onSubmitApiKey = handleApiKeySubmit(async (values) => {
  if (!shop.value) return;

  if (selectedScopes.value.length === 0) {
    scopeError.value = t("validation.required", { field: "Scope" });
    return;
  }

  isCreatingApiKey.value = true;
  try {
    const { data: result } = await api.POST("/organizations/{orgId}/shops/{shopId}/api-keys", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: {
        name: values.apiKeyName,
        scopes: selectedScopes.value,
      },
    });

    newToken.value = result?.token ?? "";
    closeAddKeyModal();
    showTokenModal.value = true;

    await loadApiKeys();
    alert.success(t("shop.apiKeyCreated"));
  } catch (error) {
    alert.error(
      `${t("shop.failedCreateApiKey")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isCreatingApiKey.value = false;
  }
});

function copyToken() {
  if (!navigator.clipboard) {
    alert.error(t("shop.clipboardUnavailable"));
    return;
  }

  navigator.clipboard.writeText(newToken.value);
  alert.success(t("shop.apiKeyCopied"));
}

function copyComposerRepository() {
  if (!navigator.clipboard || !packagesComposerUrl.value) return;

  const json = JSON.stringify(
    { repositories: [{ type: "composer", url: packagesComposerUrl.value }] },
    null,
    4,
  );
  navigator.clipboard.writeText(json);
  alert.success(t("packages.composerCopied"));
}

function confirmDeleteKey(apiKey: ApiKey) {
  deletingApiKey.value = apiKey;
  showDeleteApiKeyModal.value = true;
}

async function deleteApiKey() {
  if (!deletingApiKey.value || !shop.value) return;

  isDeletingApiKey.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/shops/{shopId}/api-keys/{keyId}", {
      params: {
        path: {
          orgId: shop.value.organizationId,
          shopId: shop.value.id,
          keyId: deletingApiKey.value.id,
        },
      },
    });

    showDeleteApiKeyModal.value = false;
    deletingApiKey.value = null;
    await loadApiKeys();
    alert.success(t("shop.apiKeyDeleted"));
  } catch (error) {
    alert.error(
      `${t("shop.failedDeleteApiKey")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isDeletingApiKey.value = false;
  }
}

async function checkPackagesConfigured() {
  try {
    const { data: config } = await api.GET("/packages-token/configuration");
    isPackagesConfigured.value = config?.configured ?? false;
    packagesComposerUrl.value = config?.composerUrl ?? null;
  } catch {
    isPackagesConfigured.value = false;
  }
}

async function loadPackagesTokens() {
  if (!shop.value || !isPackagesConfigured.value) return;

  isPackagesTokensLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/shops/{shopId}/packages-tokens", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
    });
    packagesTokens.value = data ?? [];
  } catch (error) {
    alert.error(`${t("packages.failedLoad")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isPackagesTokensLoading.value = false;
  }
}

const onSubmitPackagesToken = handlePackagesTokenSubmit(async (values) => {
  if (!shop.value) return;

  isCreatingPackagesToken.value = true;
  try {
    await api.POST("/organizations/{orgId}/shops/{shopId}/packages-tokens", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: { token: values.packagesToken },
    });

    showAddPackagesTokenModal.value = false;
    await loadPackagesTokens();
    alert.success(t("packages.tokenAdded"));
  } catch (error) {
    alert.error(`${t("packages.failedAdd")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isCreatingPackagesToken.value = false;
  }
});

function confirmDeletePackagesToken(pt: PackagesToken) {
  deletingPackagesToken.value = pt;
  showDeletePackagesTokenModal.value = true;
}

async function deletePackagesToken() {
  if (!deletingPackagesToken.value || !shop.value) return;

  isDeletingPackagesToken.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}", {
      params: {
        path: {
          orgId: shop.value.organizationId,
          shopId: shop.value.id,
          tokenId: deletingPackagesToken.value.id,
        },
      },
    });

    showDeletePackagesTokenModal.value = false;
    deletingPackagesToken.value = null;
    await loadPackagesTokens();
    alert.success(t("packages.tokenDeleted"));
  } catch (error) {
    alert.error(
      `${t("packages.failedDelete")}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isDeletingPackagesToken.value = false;
  }
}

async function syncPackagesToken(pt: PackagesToken) {
  if (!shop.value) return;

  isSyncingPackagesToken.value = pt.id;
  try {
    await api.POST("/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}/sync", {
      params: {
        path: { orgId: shop.value.organizationId, shopId: shop.value.id, tokenId: pt.id },
      },
    });

    await loadPackagesTokens();
    alert.success(t("packages.syncTriggered"));
  } catch (error) {
    alert.error(`${t("packages.failedSync")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isSyncingPackagesToken.value = null;
  }
}

async function deleteShop() {
  if (!shop.value) return;

  isDeletingShop.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
    });

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
