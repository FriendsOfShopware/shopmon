<template>
  <Card v-if="isPackagesConfigured" id="packages-tokens">
    <CardHeader class="px-4 pb-3 sm:px-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="min-w-0">
          <CardTitle as="h2" class="flex items-center gap-2 text-base">
            <icon-fa6-solid:cube class="size-4 text-muted-foreground" />
            {{ $t("packages.title") }}
          </CardTitle>
          <CardDescription class="mt-1">{{ $t("packages.description") }}</CardDescription>
        </div>
        <Button size="sm" class="w-full sm:w-auto" @click="showAddPackagesTokenModal = true">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("packages.addToken") }}
        </Button>
      </div>
    </CardHeader>
    <CardContent class="px-4 sm:px-6">
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
            class="flex items-center justify-between gap-3 rounded-lg border px-4 py-3"
          >
            <div class="min-w-0">
              <div class="font-medium">{{ $t("packages.tokenNumber", { id: pt.id }) }}</div>
              <div class="mt-0.5 text-xs text-muted-foreground">
                <template v-if="pt.lastSyncedAt">{{
                  $t("packages.lastSynced", {
                    time: timeAgo(pt.lastSyncedAt),
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
            ><code class="font-mono text-sm">composer config --auth bearer.{{ packagesComposerHost }} &lt;{{ $t("packages.yourToken") }}&gt;</code></pre>
          </div>
        </div>
      </template>
    </CardContent>
  </Card>

  <!-- Add Packages Token -->
  <AddPackagesTokenModal
    :open="showAddPackagesTokenModal"
    :org-id="orgId"
    :shop-id="shopId"
    @update:open="showAddPackagesTokenModal = $event"
    @created="loadPackagesTokens"
  />

  <!-- Delete Packages Token -->
  <DeleteConfirmationModal
    :show="showDeletePackagesTokenModal"
    :title="$t('packages.deleteTokenTitle')"
    :entity-name="$t('packages.tokenNumber', { id: deletingPackagesToken?.id })"
    :custom-consequence="$t('packages.deleteTokenWarning')"
    :is-loading="isDeletingPackagesToken"
    :confirm-button-text="$t('packages.deleteTokenConfirm')"
    @close="showDeletePackagesTokenModal = false"
    @confirm="deletePackagesToken"
  />
</template>

<script setup lang="ts">
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import AddPackagesTokenModal from "@/components/shop/AddPackagesTokenModal.vue";
import { useAlert } from "@/composables/useAlert";
import { timeAgo } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

type PackagesToken = components["schemas"]["PackagesToken"];

const props = defineProps<{
  orgId: string;
  shopId: number;
}>();

const { t } = useI18n();
const alert = useAlert();

const isPackagesConfigured = ref(false);
const packagesComposerUrl = ref<string | null>(null);
const packagesTokens = ref<PackagesToken[]>([]);
const isPackagesTokensLoading = ref(false);
const isDeletingPackagesToken = ref(false);
const isSyncingPackagesToken = ref<number | null>(null);
const showAddPackagesTokenModal = ref(false);
const showDeletePackagesTokenModal = ref(false);
const deletingPackagesToken = ref<PackagesToken | null>(null);

const packagesComposerHost = computed(() => {
  if (!packagesComposerUrl.value) return "";
  try {
    return new URL(packagesComposerUrl.value).host;
  } catch {
    return packagesComposerUrl.value;
  }
});

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
  if (!isPackagesConfigured.value) return;
  isPackagesTokensLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/shops/{shopId}/packages-tokens", {
      params: { path: { orgId: props.orgId, shopId: props.shopId } },
    });
    packagesTokens.value = data ?? [];
  } catch (error) {
    alert.error(`${t("packages.failedLoad")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isPackagesTokensLoading.value = false;
  }
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

function confirmDeletePackagesToken(pt: PackagesToken) {
  deletingPackagesToken.value = pt;
  showDeletePackagesTokenModal.value = true;
}

async function deletePackagesToken() {
  if (!deletingPackagesToken.value) return;
  isDeletingPackagesToken.value = true;
  try {
    const { error } = await api.DELETE(
      "/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}",
      {
        params: {
          path: {
            orgId: props.orgId,
            shopId: props.shopId,
            tokenId: deletingPackagesToken.value.id,
          },
        },
      },
    );
    if (error) {
      alert.error(
        `${t("packages.failedDelete")}: ${(error as { message?: string }).message ?? ""}`,
      );
      return;
    }
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
  isSyncingPackagesToken.value = pt.id;
  try {
    const { error } = await api.POST(
      "/organizations/{orgId}/shops/{shopId}/packages-tokens/{tokenId}/sync",
      {
        params: {
          path: { orgId: props.orgId, shopId: props.shopId, tokenId: pt.id },
        },
      },
    );
    if (error) {
      alert.error(`${t("packages.failedSync")}: ${(error as { message?: string }).message ?? ""}`);
      return;
    }
    await loadPackagesTokens();
    alert.success(t("packages.syncTriggered"));
  } catch (error) {
    alert.error(`${t("packages.failedSync")}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isSyncingPackagesToken.value = null;
  }
}

onMounted(async () => {
  await checkPackagesConfigured();
  await loadPackagesTokens();
});
</script>
