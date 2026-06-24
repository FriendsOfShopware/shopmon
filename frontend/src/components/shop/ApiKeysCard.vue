<template>
  <Card id="api-keys">
    <CardHeader class="px-4 pb-3 sm:px-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <CardTitle as="h2" class="flex items-center gap-2 text-base">
          <icon-fa6-solid:key class="size-4 text-muted-foreground" />
          {{ $t("shop.apiKeys") }}
        </CardTitle>
        <Button size="sm" class="w-full sm:w-auto" @click="showAddKeyModal = true">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("shop.createApiKey") }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="px-4 sm:px-6">
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
          class="flex items-center justify-between gap-3 rounded-lg border px-4 py-3"
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

  <!-- Add API Key -->
  <AddApiKeyModal
    :open="showAddKeyModal"
    :org-id="orgId"
    :shop-id="shopId"
    :available-scopes="availableScopes"
    @update:open="showAddKeyModal = $event"
    @created="onApiKeyCreated"
  />

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
    :entity-name="$t('shop.apiKeyEntity', { name: deletingApiKey?.name })"
    :custom-consequence="$t('shop.deleteApiKeyWarning')"
    :is-loading="isDeletingApiKey"
    :confirm-button-text="$t('shop.deleteApiKeyConfirm')"
    @close="showDeleteApiKeyModal = false"
    @confirm="deleteApiKey"
  />
</template>

<script setup lang="ts">
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import AddApiKeyModal from "@/components/shop/AddApiKeyModal.vue";
import { useAlert } from "@/composables/useAlert";
import { formatDate } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { nextTick, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

type ApiKey = components["schemas"]["ApiKey"];
type AvailableScope = components["schemas"]["ApiKeyScope"];

const props = defineProps<{
  orgId: string;
  shopId: number;
}>();

const { t } = useI18n();
const route = useRoute();
const alert = useAlert();

const apiKeys = ref<ApiKey[]>([]);
const availableScopes = ref<AvailableScope[]>([]);
const isApiKeysLoading = ref(false);
const isDeletingApiKey = ref(false);

const showAddKeyModal = ref(false);
const showTokenModal = ref(false);
const showDeleteApiKeyModal = ref(false);
const deletingApiKey = ref<ApiKey | null>(null);
const newToken = ref("");

async function loadApiKeys() {
  isApiKeysLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/shops/{shopId}/api-keys", {
      params: { path: { orgId: props.orgId, shopId: props.shopId } },
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

function getScopeLabel(scope: string): string {
  return availableScopes.value.find((s) => s.value === scope)?.label ?? scope;
}

function closeTokenModal() {
  showTokenModal.value = false;
  newToken.value = "";
}

async function onApiKeyCreated(token: string) {
  newToken.value = token;
  showTokenModal.value = true;
  await loadApiKeys();
}

function copyToken() {
  if (!navigator.clipboard) {
    alert.error(t("shop.clipboardUnavailable"));
    return;
  }
  navigator.clipboard.writeText(newToken.value);
  alert.success(t("shop.apiKeyCopied"));
}

function confirmDeleteKey(apiKey: ApiKey) {
  deletingApiKey.value = apiKey;
  showDeleteApiKeyModal.value = true;
}

async function deleteApiKey() {
  if (!deletingApiKey.value) return;
  isDeletingApiKey.value = true;
  try {
    const { error } = await api.DELETE("/organizations/{orgId}/shops/{shopId}/api-keys/{keyId}", {
      params: {
        path: {
          orgId: props.orgId,
          shopId: props.shopId,
          keyId: deletingApiKey.value.id,
        },
      },
    });
    if (error) {
      alert.error(
        `${t("shop.failedDeleteApiKey")}: ${(error as { message?: string }).message ?? ""}`,
      );
      return;
    }
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

onMounted(async () => {
  await Promise.all([loadApiKeys(), loadAvailableScopes()]);
  if (route.hash === "#api-keys") {
    await nextTick();
    document.getElementById("api-keys")?.scrollIntoView({ behavior: "smooth" });
  }
});
</script>
