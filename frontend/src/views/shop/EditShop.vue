<template>
  <header-container :title="shop ? $t('shop.editShop', { name: shop.name }) : $t('nav.editShop')">
    <div class="header-actions">
      <UiButton :to="{ name: 'account.shop.list' }" type="button">
        <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
        {{ $t("shop.backToShops") }}
      </UiButton>

      <UiButton
        v-if="shop"
        :to="{ name: 'account.environments.new', query: { shopId: shop.id } }"
        type="button"
      >
        <icon-fa6-solid:plus class="icon" aria-hidden="true" />
        {{ $t("environment.addEnvironment") }}
      </UiButton>
    </div>
  </header-container>

  <main-container>
    <Panel v-if="isPageLoading" class="state-panel">
      <icon-line-md:loading-twotone-loop class="icon" />
      {{ $t("shop.loadingShop") }}
    </Panel>

    <Panel v-else-if="!shop" class="state-panel">
      <p>{{ $t("shop.shopNotFound") }}</p>
    </Panel>

    <template v-else>
      <Panel>
        <vee-form
          :key="`shop-${shop.id}`"
          v-slot="{ errors, isSubmitting }"
          :validation-schema="shopSchema"
          :initial-values="shopFormInitialValues"
          @submit="onSubmitShop"
        >
          <form-group :title="$t('shop.shopInfo')">
            <InputField name="name" :label="$t('common.name')" :error="errors.name" />

            <TextareaField
              name="description"
              :label="$t('common.description')"
              :placeholder="$t('shop.optionalDescription')"
              :error="errors.description"
            />

            <InputField
              name="gitUrl"
              :label="$t('shop.gitRepoUrl')"
              type="url"
              placeholder="https://github.com/org/repo"
              :error="errors.gitUrl"
            />
          </form-group>

          <div class="form-submit">
            <UiButton
              :disabled="isSubmitting || isSavingShop"
              type="submit"
              variant="primary"
            >
              <icon-fa6-solid:floppy-disk
                v-if="!(isSubmitting || isSavingShop)"
                class="icon"
                aria-hidden="true"
              />
              <icon-line-md:loading-twotone-loop v-else class="icon" />
              {{ $t("common.save") }}
            </UiButton>
          </div>
        </vee-form>
      </Panel>

      <Panel id="api-keys" :title="$t('shop.apiKeys')">
        <template #action>
          <UiButton type="button" variant="primary" @click="openAddKeyModal">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            {{ $t("shop.createApiKey") }}
          </UiButton>
        </template>

        <Banner variant="default">
          <p>
            <strong>{{ $t("shop.apiKeys") }}</strong>
          </p>
          <p>
            {{ $t("shop.apiKeyInfo") }}
          </p>
        </Banner>

        <div v-if="isApiKeysLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          {{ $t("shop.loadingApiKeys") }}
        </div>

        <div v-else-if="apiKeys.length === 0" class="api-keys-empty">
          <icon-fa6-solid:key class="icon icon-large" aria-hidden="true" />
          <p>{{ $t("shop.noApiKeys") }}</p>
          <p class="text-muted">
            {{ $t("shop.noApiKeysHint") }}
          </p>
        </div>

        <div v-else class="api-keys-list">
          <div v-for="apiKey in apiKeys" :key="apiKey.id" class="api-key-item">
            <div class="api-key-info">
              <h4>{{ apiKey.name }}</h4>
              <p class="text-muted">
                {{ $t("shop.createdDate", { date: formatDate(apiKey.createdAt) }) }}
              </p>
              <p v-if="apiKey.lastUsedAt" class="text-muted">
                {{ $t("shop.lastUsedDate", { date: formatDate(apiKey.lastUsedAt) }) }}
              </p>
              <div class="api-key-scopes">
                <span v-for="scope in apiKey.scopes" :key="scope" class="badge badge-primary">
                  {{ getScopeLabel(scope) }}
                </span>
              </div>
            </div>
            <div class="api-key-actions">
              <UiButton
                type="button"
                variant="destructive"
                @click="confirmDeleteKey(apiKey)"
              >
                <icon-fa6-solid:trash class="icon" aria-hidden="true" />
                {{ $t("common.delete") }}
              </UiButton>
            </div>
          </div>
        </div>
      </Panel>

      <Panel
        v-if="isPackagesConfigured"
        id="packages-tokens"
        :title="$t('packages.title')"
        :description="$t('packages.description')"
      >
        <template #action>
          <UiButton
            type="button"
            variant="primary"
            @click="showAddPackagesTokenModal = true"
          >
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            {{ $t("packages.addToken") }}
          </UiButton>
        </template>

        <div v-if="isPackagesTokensLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          {{ $t("packages.loading") }}
        </div>

        <div v-else-if="packagesTokens.length === 0" class="api-keys-empty">
          <icon-fa6-solid:cube class="icon icon-large" aria-hidden="true" />
          <p>{{ $t("packages.noTokens") }}</p>
          <p class="text-muted">{{ $t("packages.noTokensHint") }}</p>
        </div>

        <template v-else>
          <div class="api-keys-list">
            <div v-for="pt in packagesTokens" :key="pt.id" class="api-key-item">
              <div class="api-key-info">
                <h4>Token #{{ pt.id }}</h4>
                <p v-if="pt.lastSyncedAt" class="text-muted">
                  {{
                    $t("packages.lastSynced", {
                      time: timeAgo(new Date(pt.lastSyncedAt).getTime() / 1000),
                    })
                  }}
                </p>
                <p v-else class="text-muted">{{ $t("packages.notSyncedYet") }}</p>
              </div>
              <div class="api-key-actions">
                <UiButton
                  type="button"
                  :disabled="isSyncingPackagesToken === pt.id"
                  @click="syncPackagesToken(pt)"
                >
                  <icon-fa6-solid:arrows-rotate
                    v-if="isSyncingPackagesToken !== pt.id"
                    class="icon"
                    aria-hidden="true"
                  />
                  <icon-line-md:loading-twotone-loop v-else class="icon" />
                  {{ $t("packages.sync") }}
                </UiButton>
                <UiButton
                  type="button"
                  variant="destructive"
                  @click="confirmDeletePackagesToken(pt)"
                >
                  <icon-fa6-solid:trash class="icon" aria-hidden="true" />
                  {{ $t("common.delete") }}
                </UiButton>
              </div>
            </div>
          </div>

          <div v-if="packagesComposerUrl" class="composer-setup">
            <h4>{{ $t("packages.composerSetup") }}</h4>
            <Banner variant="alert">
              <p>
                {{ $t("packages.composerWarning") }}
              </p>
            </Banner>
            <p class="text-muted">
              {{ $t("packages.composerRepoHint") }}
            </p>
            <div class="code-block">
              <pre><code>{
    "repositories": [
        {
            "type": "composer",
            "url": "{{ packagesComposerUrl }}"
        }
    ]
}</code></pre>
              <UiButton
                type="button"
                size="sm"
                @click="copyComposerRepository"
              >
                <icon-fa6-solid:copy class="icon" aria-hidden="true" />
                {{ $t("common.copy") }}
              </UiButton>
            </div>
            <p class="text-muted">{{ $t("packages.composerAuthHint") }}</p>
            <div class="code-block">
              <pre><code>composer config --auth bearer.{{ packagesComposerHost }} &lt;your-token&gt;</code></pre>
            </div>
          </div>
        </template>
      </Panel>

      <Panel :title="$t('shop.dangerZone')">
        <p>{{ $t("shop.deleteShopWarning") }}</p>

        <p v-if="!canDeleteShop" class="delete-shop-warning">
          {{ $t("shop.deleteShopEnvironmentsWarning", { count: environmentsInShopCount }) }}
        </p>

        <UiButton
          type="button"
          variant="destructive"
          :disabled="!canDeleteShop || isDeletingShop"
          @click="showDeleteShopModal = true"
        >
          <icon-fa6-solid:trash class="icon" aria-hidden="true" />
          {{ $t("shop.deleteShop") }}
        </UiButton>
      </Panel>
    </template>

    <!-- Add API Key Modal -->
    <modal :show="showAddKeyModal" close-x-mark @close="closeAddKeyModal">
      <template #title> {{ $t("shop.createApiKeyTitle") }} </template>

      <template #content>
        <vee-form
          id="apiKeyForm"
          v-slot="{ errors }"
          :validation-schema="apiKeySchema"
          :initial-values="{ name: '', scopes: [] }"
          class="api-key-form"
          @submit="onSubmitApiKey"
        >
          <div class="form-group">
            <InputField
              name="name"
              :label="$t('common.name')"
              :placeholder="$t('shop.apiKeyPlaceholder')"
              :error="errors.name"
            />
            <p class="field-help">{{ $t("packages.apiKeyHelp") }}</p>
          </div>

          <div class="form-group">
            <label>{{ $t("shop.scopes") }}</label>
            <p class="field-help">{{ $t("shop.scopesHelp") }}</p>
            <div class="scopes-list">
              <label v-for="scope in availableScopes" :key="scope.value" class="scope-checkbox">
                <field type="checkbox" name="scopes" :value="scope.value" />
                <div class="scope-info">
                  <span class="scope-label">{{ scope.label }}</span>
                  <span class="scope-description">{{ scope.description }}</span>
                </div>
              </label>
            </div>
            <div class="field-error-message">{{ errors.scopes }}</div>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <UiButton type="button" @click="closeAddKeyModal">
          {{ $t("common.cancel") }}
        </UiButton>
        <UiButton
          type="submit"
          variant="primary"
          form="apiKeyForm"
          :disabled="isCreatingApiKey"
        >
          <icon-fa6-solid:key v-if="!isCreatingApiKey" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t("shop.createApiKey") }}
        </UiButton>
      </template>
    </modal>

    <!-- Show Token Modal -->
    <modal :show="showTokenModal" @close="closeTokenModal">
      <template #title> {{ $t("shop.apiKeyCreatedTitle") }} </template>

      <template #content>
        <Banner variant="alert">
          <p>
            <strong>{{ $t("shop.apiKeyCopyWarning") }}</strong>
          </p>
          <p>
            {{ $t("shop.apiKeyCopyDesc") }}
          </p>
        </Banner>

        <div class="token-display">
          <code class="token-value">{{ newToken }}</code>
          <UiButton type="button" @click="copyToken">
            <icon-fa6-solid:copy class="icon" aria-hidden="true" />
            {{ $t("common.copy") }}
          </UiButton>
        </div>
      </template>

      <template #footer>
        <UiButton type="button" variant="primary" @click="closeTokenModal">
          {{ $t("common.done") }}
        </UiButton>
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
        <vee-form
          id="packagesTokenForm"
          v-slot="{ errors }"
          :validation-schema="packagesTokenSchema"
          :initial-values="{ token: '' }"
          class="api-key-form"
          @submit="onSubmitPackagesToken"
        >
          <div class="form-group">
            <InputField
              name="token"
              :label="$t('packages.tokenLabel')"
              :placeholder="$t('packages.tokenPlaceholder')"
              :error="errors.token"
            />
            <p class="field-help">
              {{ $t("packages.tokenHelp") }}
            </p>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <UiButton type="button" @click="showAddPackagesTokenModal = false">
          {{ $t("common.cancel") }}
        </UiButton>
        <UiButton
          type="submit"
          variant="primary"
          form="packagesTokenForm"
          :disabled="isCreatingPackagesToken"
        >
          <icon-fa6-solid:plus v-if="!isCreatingPackagesToken" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t("packages.addToken") }}
        </UiButton>
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
  </main-container>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { fetchAccountEnvironments } from "@/composables/useAccountEnvironments";
import { formatDate, timeAgo } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Field, Form as VeeForm } from "vee-validate";
import { computed, nextTick, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

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

const shopFormInitialValues = computed(() => ({
  name: shop.value?.name ?? "",
  description: shop.value?.description ?? "",
  gitUrl: shop.value?.gitUrl ?? "",
}));

const shopSchema = Yup.object().shape({
  name: Yup.string().required(t("validation.shopNameRequired")),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url(t("validation.urlInvalid")).optional(),
});

const apiKeySchema = Yup.object().shape({
  name: Yup.string()
    .required(t("validation.nameRequired"))
    .min(1, t("validation.nameRequired"))
    .max(100, t("validation.nameMaxLength")),
  scopes: Yup.array()
    .of(Yup.string())
    .min(1, t("validation.required", { field: "Scope" }))
    .required(t("validation.required", { field: "Scope" })),
});

const packagesTokenSchema = Yup.object().shape({
  token: Yup.string().required(t("validation.tokenRequired")).min(1, t("validation.tokenRequired")),
});

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

async function onSubmitShop(values: Record<string, unknown>) {
  if (!shop.value) return;

  const typedValues = values as Yup.InferType<typeof shopSchema>;

  isSavingShop.value = true;
  try {
    await api.PATCH("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: {
        name: typedValues.name,
        description: typedValues.description ?? "",
        gitUrl: typedValues.gitUrl || undefined,
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
}

function openAddKeyModal() {
  showAddKeyModal.value = true;
}

function closeAddKeyModal() {
  showAddKeyModal.value = false;
}

function closeTokenModal() {
  showTokenModal.value = false;
  newToken.value = "";
}

async function onSubmitApiKey(values: Record<string, unknown>) {
  if (!shop.value) return;

  const typedValues = values as Yup.InferType<typeof apiKeySchema>;

  isCreatingApiKey.value = true;
  try {
    const { data: result } = await api.POST("/organizations/{orgId}/shops/{shopId}/api-keys", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: {
        name: typedValues.name,
        scopes: typedValues.scopes as string[],
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
}

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

async function onSubmitPackagesToken(values: Record<string, unknown>) {
  if (!shop.value) return;

  const typedValues = values as Yup.InferType<typeof packagesTokenSchema>;

  isCreatingPackagesToken.value = true;
  try {
    await api.POST("/organizations/{orgId}/shops/{shopId}/packages-tokens", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: { token: typedValues.token },
    });

    showAddPackagesTokenModal.value = false;
    await loadPackagesTokens();
    alert.success(t("packages.tokenAdded"));
  } catch (error) {
    alert.error(`${t("packages.failedAdd")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isCreatingPackagesToken.value = false;
  }
}

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

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
}

.state-panel {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.delete-shop-warning {
  color: var(--color-danger);
}

.api-keys-loading {
  padding: 3rem;
  text-align: center;
  color: var(--text-color-muted);
}

.api-keys-empty {
  padding: 4rem 2rem;
  text-align: center;

  .icon-large {
    font-size: 3rem;
    color: var(--text-color-muted);
    margin-bottom: 1rem;
  }

  p {
    margin: 0;

    &.text-muted {
      margin-top: 0.5rem;
      color: var(--text-color-muted);
      font-size: 0.875rem;
    }
  }
}

.api-keys-list {
  padding: 1rem 0 0;
}

.api-key-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border: 1px solid var(--panel-border-color);
  border-radius: 0.5rem;
  margin-bottom: 1rem;

  &:last-child {
    margin-bottom: 0;
  }
}

.api-key-info {
  h4 {
    margin: 0 0 0.25rem 0;
    font-size: 1.125rem;
    font-weight: 500;
  }

  p {
    margin: 0 0 0.25rem 0;
    font-size: 0.875rem;
  }
}

.api-key-scopes {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.api-key-actions {
  display: flex;
  gap: 0.5rem;
}

.api-key-form {
  .form-group {
    margin-bottom: 1.5rem;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

.field-help {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--text-color-muted);
}

.scopes-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.scope-checkbox {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--panel-border-color);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background-color 0.2s;

  &:hover {
    background-color: var(--panel-background-color);
  }

  input[type="checkbox"] {
    margin-top: 0.25rem;
  }
}

.scope-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.scope-label {
  font-weight: 500;
}

.scope-description {
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.token-display {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
  padding: 1rem;
  background-color: var(--panel-background-color);
  border-radius: 0.5rem;
}

.token-value {
  flex: 1;
  padding: 0.75rem;
  background-color: var(--input-background-color);
  border: 1px solid var(--input-border-color);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
  word-break: break-all;
}

.composer-setup {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--panel-border-color);

  h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    font-weight: 500;
  }

  .text-muted {
    font-size: 0.875rem;
    color: var(--text-color-muted);
    margin: 0.5rem 0;

    code {
      padding: 0.125rem 0.375rem;
      background-color: var(--input-background-color);
      border-radius: 0.25rem;
      font-size: 0.8125rem;
    }
  }
}

.code-block {
  position: relative;
  margin: 0.5rem 0 1rem;
  background-color: var(--input-background-color);
  border: 1px solid var(--input-border-color);
  border-radius: 0.5rem;
  overflow: hidden;

  pre {
    margin: 0;
    padding: 1rem;
    overflow-x: auto;

    code {
      font-family: monospace;
      font-size: 0.8125rem;
      line-height: 1.5;
    }
  }

  .ui-button--sm {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }
}
</style>
