<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">
        {{ shop ? $t("shop.editShop", { name: shop.name }) : $t("nav.editShop") }}
      </h1>
      <div class="flex gap-2">
        <Button variant="outline" size="sm" as-child>
          <RouterLink :to="{ name: 'account.shop.list' }">
            <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
            {{ $t("shop.backToShops") }}
          </RouterLink>
        </Button>
        <Button v-if="shop" size="sm" as-child>
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
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:folder class="size-4 text-muted-foreground" />
            {{ $t("shop.shopInfo") }}
          </CardTitle>
        </CardHeader>
        <CardContent>
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
        <CardHeader class="pb-3">
          <div class="flex items-center justify-between">
            <CardTitle class="flex items-center gap-2 text-base">
              <icon-fa6-solid:earth-americas class="size-4 text-muted-foreground" />
              {{ $t("environment.title") }}
            </CardTitle>
            <Button size="sm" as-child>
              <RouterLink :to="{ name: 'account.environments.new', query: { shopId: shop!.id } }">
                <icon-fa6-solid:plus class="mr-1.5 size-3" />
                {{ $t("environment.addEnvironment") }}
              </RouterLink>
            </Button>
          </div>
        </CardHeader>
        <CardContent class="space-y-2">
          <div
            v-for="env in shopEnvironments"
            :key="env.id"
            :class="[
              'flex items-center gap-3 rounded-xl border px-4 py-3 transition-colors',
              env.id === shop!.defaultEnvironmentId
                ? 'border-primary/30 bg-primary/5'
                : 'hover:bg-muted/50',
            ]"
          >
            <div
              class="flex size-9 shrink-0 items-center justify-center rounded-lg border bg-muted"
            >
              <img v-if="env.favicon" :src="env.favicon" alt="" class="size-5 rounded" />
              <icon-fa6-solid:earth-americas v-else class="size-3.5 text-muted-foreground/50" />
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <RouterLink
                  :to="{ name: 'account.environments.detail', params: { environmentId: env.id } }"
                  class="truncate font-medium hover:text-primary transition-colors"
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
              <div class="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
                <Badge variant="secondary" class="font-mono text-[10px]">{{
                  env.shopwareVersion
                }}</Badge>
                <span v-if="env.url" class="truncate">{{ env.url }}</span>
              </div>
            </div>

            <div class="flex shrink-0 items-center gap-1">
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
      <Card id="api-keys">
        <CardHeader class="pb-3">
          <div class="flex items-center justify-between">
            <CardTitle class="flex items-center gap-2 text-base">
              <icon-fa6-solid:key class="size-4 text-muted-foreground" />
              {{ $t("shop.apiKeys") }}
            </CardTitle>
            <Button size="sm" @click="openAddKeyModal">
              <icon-fa6-solid:plus class="mr-1.5 size-3" />
              {{ $t("shop.createApiKey") }}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Alert class="mb-4">
            <AlertDescription>{{ $t("shop.apiKeyInfo") }}</AlertDescription>
          </Alert>

          <div
            v-if="isApiKeysLoading"
            class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
          >
            <icon-line-md:loading-twotone-loop class="size-5" />
            {{ $t("shop.loadingApiKeys") }}
          </div>

          <div
            v-else-if="apiKeys.length === 0"
            class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-12 text-center"
          >
            <icon-fa6-solid:key class="size-8 text-muted-foreground" />
            <p class="font-medium">{{ $t("shop.noApiKeys") }}</p>
            <p class="text-sm text-muted-foreground">{{ $t("shop.noApiKeysHint") }}</p>
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="apiKey in apiKeys"
              :key="apiKey.id"
              class="flex items-center justify-between rounded-xl border px-4 py-3"
            >
              <div class="min-w-0">
                <div class="font-medium">{{ apiKey.name }}</div>
                <div
                  class="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground"
                >
                  <span>{{ $t("shop.createdDate", { date: formatDate(apiKey.createdAt) }) }}</span>
                  <span v-if="apiKey.lastUsedAt">{{
                    $t("shop.lastUsedDate", { date: formatDate(apiKey.lastUsedAt) })
                  }}</span>
                </div>
                <div class="mt-2 flex flex-wrap gap-1">
                  <Badge
                    v-for="scope in apiKey.scopes"
                    :key="scope"
                    variant="secondary"
                    class="text-xs"
                  >
                    {{ getScopeLabel(scope) }}
                  </Badge>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                class="size-8 shrink-0 text-muted-foreground hover:text-destructive"
                @click="confirmDeleteKey(apiKey)"
              >
                <icon-fa6-solid:trash class="size-3.5" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Packages Tokens -->
      <Card
        v-if="instanceConfig?.packageMirrorEnabled !== false && isPackagesConfigured"
        id="packages-tokens"
      >
        <CardHeader class="pb-3">
          <div class="flex items-center justify-between">
            <div>
              <CardTitle class="flex items-center gap-2 text-base">
                <icon-fa6-solid:cube class="size-4 text-muted-foreground" />
                {{ $t("packages.title") }}
              </CardTitle>
              <CardDescription class="mt-1">{{ $t("packages.description") }}</CardDescription>
            </div>
            <Button size="sm" @click="showAddPackagesTokenModal = true">
              <icon-fa6-solid:plus class="mr-1.5 size-3" />
              {{ $t("packages.addToken") }}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div
            v-if="isPackagesTokensLoading"
            class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
          >
            <icon-line-md:loading-twotone-loop class="size-5" />
            {{ $t("packages.loading") }}
          </div>

          <div
            v-else-if="packagesTokens.length === 0"
            class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-12 text-center"
          >
            <icon-fa6-solid:cube class="size-8 text-muted-foreground" />
            <p class="font-medium">{{ $t("packages.noTokens") }}</p>
            <p class="text-sm text-muted-foreground">{{ $t("packages.noTokensHint") }}</p>
          </div>

          <template v-else>
            <div class="space-y-2">
              <div
                v-for="pt in packagesTokens"
                :key="pt.id"
                class="flex items-center justify-between rounded-xl border px-4 py-3"
              >
                <div class="min-w-0">
                  <div class="font-medium">Token #{{ pt.id }}</div>
                  <div class="mt-0.5 text-xs text-muted-foreground">
                    <template v-if="pt.lastSyncedAt">{{
                      $t("packages.lastSynced", {
                        time: timeAgo(new Date(pt.lastSyncedAt).getTime() / 1000),
                      })
                    }}</template>
                    <template v-else>{{ $t("packages.notSyncedYet") }}</template>
                  </div>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="size-8"
                    :disabled="isSyncingPackagesToken === pt.id"
                    :title="$t('packages.sync')"
                    @click="syncPackagesToken(pt)"
                  >
                    <icon-fa6-solid:arrows-rotate
                      v-if="isSyncingPackagesToken !== pt.id"
                      class="size-3.5"
                    />
                    <icon-line-md:loading-twotone-loop v-else class="size-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="size-8 text-muted-foreground hover:text-destructive"
                    @click="confirmDeletePackagesToken(pt)"
                  >
                    <icon-fa6-solid:trash class="size-3.5" />
                  </Button>
                </div>
              </div>
            </div>

            <!-- Composer setup -->
            <div v-if="packagesComposerUrl" class="mt-6 border-t pt-6">
              <h4 class="mb-3 text-sm font-semibold">{{ $t("packages.composerSetup") }}</h4>
              <Alert variant="destructive" class="mb-3">
                <AlertDescription>{{ $t("packages.composerWarning") }}</AlertDescription>
              </Alert>
              <p class="mb-2 text-xs text-muted-foreground">
                {{ $t("packages.composerRepoHint") }}
              </p>
              <div class="relative overflow-hidden rounded-md border bg-muted">
                <pre class="overflow-x-auto p-4"><code class="font-mono text-sm leading-relaxed">{
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
                  class="absolute right-2 top-2"
                  @click="copyComposerRepository"
                >
                  <icon-fa6-solid:copy class="mr-1 size-3" />
                  {{ $t("common.copy") }}
                </Button>
              </div>
              <p class="mt-3 mb-2 text-xs text-muted-foreground">
                {{ $t("packages.composerAuthHint") }}
              </p>
              <div class="overflow-hidden rounded-md border bg-muted">
                <pre
                  class="overflow-x-auto p-4"
                ><code class="font-mono text-sm">composer config --auth bearer.{{ packagesComposerHost }} &lt;your-token&gt;</code></pre>
              </div>
            </div>
          </template>
        </CardContent>
      </Card>

      <!-- Danger zone -->
      <Card class="border-destructive/30">
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base text-destructive">
            <icon-fa6-solid:triangle-exclamation class="size-4" />
            {{ $t("shop.dangerZone") }}
          </CardTitle>
        </CardHeader>
        <CardContent>
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

    <!-- ═══ Modals ═══ -->

    <!-- Add API Key -->
    <Dialog :open="showAddKeyModal" @update:open="(v: boolean) => !v && closeAddKeyModal()">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("shop.createApiKeyTitle") }}</DialogTitle>
        </DialogHeader>
        <form id="apiKeyForm" class="space-y-4" @submit="onSubmitApiKey">
          <FormField v-slot="{ componentField }" name="apiKeyName">
            <FormItem>
              <FormLabel>{{ $t("common.name") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" :placeholder="$t('shop.apiKeyPlaceholder')" />
              </FormControl>
              <FormMessage />
              <p class="text-xs text-muted-foreground">{{ $t("packages.apiKeyHelp") }}</p>
            </FormItem>
          </FormField>

          <div>
            <label class="text-sm font-medium">{{ $t("shop.scopes") }}</label>
            <p class="mb-2 text-xs text-muted-foreground">{{ $t("shop.scopesHelp") }}</p>
            <div class="mt-2 flex flex-col gap-2">
              <label
                v-for="scope in availableScopes"
                :key="scope.value"
                class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-muted"
              >
                <input
                  type="checkbox"
                  :value="scope.value"
                  class="mt-1"
                  :checked="selectedScopes.includes(scope.value)"
                  @change="toggleScope(scope.value)"
                />
                <div>
                  <span class="text-sm font-medium">{{ scope.label }}</span>
                  <p class="text-xs text-muted-foreground">{{ scope.description }}</p>
                </div>
              </label>
            </div>
            <p v-if="scopeError" class="mt-1 text-sm text-destructive">{{ scopeError }}</p>
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="closeAddKeyModal">{{ $t("common.cancel") }}</Button>
          <Button type="submit" form="apiKeyForm" :disabled="isCreatingApiKey">
            <icon-fa6-solid:key v-if="!isCreatingApiKey" class="mr-1.5 size-3.5" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
            {{ $t("shop.createApiKey") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Show created token -->
    <Dialog :open="showTokenModal" @update:open="(v: boolean) => !v && closeTokenModal()">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("shop.apiKeyCreatedTitle") }}</DialogTitle>
        </DialogHeader>
        <Alert variant="destructive">
          <AlertTitle>{{ $t("shop.apiKeyCopyWarning") }}</AlertTitle>
          <AlertDescription>{{ $t("shop.apiKeyCopyDesc") }}</AlertDescription>
        </Alert>
        <div class="flex gap-2 rounded-lg border bg-muted p-3">
          <code class="flex-1 break-all rounded border bg-background p-2 font-mono text-sm">{{
            newToken
          }}</code>
          <Button variant="outline" size="sm" @click="copyToken">
            <icon-fa6-solid:copy class="mr-1 size-3" />
            {{ $t("common.copy") }}
          </Button>
        </div>
        <DialogFooter>
          <Button @click="closeTokenModal">{{ $t("common.done") }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete API Key -->
    <DeleteConfirmationModal
      :show="showDeleteApiKeyModal"
      :title="$t('shop.deleteApiKeyTitle')"
      :entity-name="`the API key '${deletingApiKey?.name}'`"
      :custom-consequence="$t('shop.deleteApiKeyWarning')"
      :is-loading="isDeletingApiKey"
      :confirm-button-text="$t('shop.deleteApiKeyConfirm')"
      @close="showDeleteApiKeyModal = false"
      @confirm="deleteApiKey"
    />

    <!-- Delete Shop -->
    <DeleteConfirmationModal
      :show="showDeleteShopModal"
      :title="$t('shop.deleteShopTitle')"
      :entity-name="shop?.name || 'this shop'"
      @close="showDeleteShopModal = false"
      @confirm="deleteShop"
    />

    <!-- Add Packages Token -->
    <Dialog
      :open="showAddPackagesTokenModal"
      @update:open="(v: boolean) => !v && (showAddPackagesTokenModal = false)"
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("packages.addTokenTitle") }}</DialogTitle>
        </DialogHeader>
        <form id="packagesTokenForm" class="space-y-4" @submit="onSubmitPackagesToken">
          <FormField v-slot="{ componentField }" name="packagesToken">
            <FormItem>
              <FormLabel>{{ $t("packages.tokenLabel") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" :placeholder="$t('packages.tokenPlaceholder')" />
              </FormControl>
              <FormMessage />
              <p class="text-xs text-muted-foreground">{{ $t("packages.tokenHelp") }}</p>
            </FormItem>
          </FormField>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="showAddPackagesTokenModal = false">{{
            $t("common.cancel")
          }}</Button>
          <Button type="submit" form="packagesTokenForm" :disabled="isCreatingPackagesToken">
            <icon-fa6-solid:plus v-if="!isCreatingPackagesToken" class="mr-1.5 size-3.5" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
            {{ $t("packages.addToken") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Packages Token -->
    <DeleteConfirmationModal
      :show="showDeletePackagesTokenModal"
      :title="$t('packages.deleteTokenTitle')"
      :entity-name="`Token #${deletingPackagesToken?.id}`"
      :custom-consequence="$t('packages.deleteTokenWarning')"
      :is-loading="isDeletingPackagesToken"
      :confirm-button-text="$t('packages.deleteTokenConfirm')"
      @close="showDeletePackagesTokenModal = false"
      @confirm="deletePackagesToken"
    />
  </div>
</template>

<script setup lang="ts">
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { useInstanceConfig } from "@/composables/useInstanceConfig";
import { fetchAccountEnvironments } from "@/composables/useAccountEnvironments";
import { formatDate, timeAgo } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import StatusIcon from "@/components/StatusIcon.vue";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRoute, useRouter } from "vue-router";

type Shop = components["schemas"]["AccountShop"];
type ApiKey = components["schemas"]["ApiKey"];
type AvailableScope = components["schemas"]["ApiKeyScope"];
type PackagesToken = components["schemas"]["PackagesToken"];
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

const selectedScopes = ref<string[]>([]);
const scopeError = ref("");

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

const apiKeyValidationSchema = toTypedSchema(
  z.object({
    apiKeyName: z
      .string()
      .min(1, t("validation.nameRequired"))
      .max(100, t("validation.nameMaxLength")),
  }),
);

const { handleSubmit: handleApiKeySubmit, resetForm: resetApiKeyForm } = useForm({
  validationSchema: apiKeyValidationSchema,
  initialValues: { apiKeyName: "" },
});

const packagesTokenValidationSchema = toTypedSchema(
  z.object({
    packagesToken: z.string().min(1, t("validation.tokenRequired")),
  }),
);

const { handleSubmit: handlePackagesTokenSubmit, resetForm: resetPackagesTokenForm } = useForm({
  validationSchema: packagesTokenValidationSchema,
  initialValues: { packagesToken: "" },
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
  const envs = environmentsData.filter((env) => env.shopId === shopId);
  shopEnvironments.value = envs;
  environmentsInShopCount.value = envs.length;
}

async function setDefaultEnvironment(envId: number) {
  if (!shop.value) return;
  isSettingDefaultEnv.value = envId;
  try {
    await api.PATCH("/organizations/{orgId}/shops/{shopId}", {
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id } },
      body: { defaultEnvironmentId: envId },
    });
    await loadShopSummary();
    alert.success(t("shop.defaultEnvSet"));
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  } finally {
    isSettingDefaultEnv.value = null;
  }
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
  return availableScopes.value.find((s) => s.value === scope)?.label ?? scope;
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
      body: { name: values.apiKeyName, scopes: selectedScopes.value },
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
      params: { path: { orgId: shop.value.organizationId, shopId: shop.value.id, tokenId: pt.id } },
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
