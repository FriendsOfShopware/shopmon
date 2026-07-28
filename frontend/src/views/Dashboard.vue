<template>
  <div class="space-y-6">
    <PageHeader :title="$t('dashboard.title')" />
    <!-- KPI Header -->
    <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <button
        type="button"
        @click="toggleStatusFilter('green')"
        class="group flex flex-col justify-between rounded-xl border p-4 text-left transition-all hover:border-emerald-500/50 hover:shadow-sm"
        :class="statusFilter === 'green' ? 'border-emerald-500 bg-emerald-500/10' : 'bg-card'"
      >
        <div class="flex items-center justify-between text-muted-foreground">
          <span class="text-xs font-semibold uppercase tracking-wider">{{
            $t("dashboard.statHealthy")
          }}</span>
          <StatusIcon status="green" class="size-4 text-emerald-500" />
        </div>
        <div class="mt-3 flex items-baseline justify-between">
          <span class="text-2xl font-bold tracking-tight">{{ healthyCount }}</span>
          <span class="text-xs text-muted-foreground" v-if="totalEnvironmentsCount > 0"
            >{{ Math.round((healthyCount / totalEnvironmentsCount) * 100) }}%</span
          >
        </div>
      </button>

      <button
        type="button"
        @click="toggleStatusFilter('yellow')"
        class="group flex flex-col justify-between rounded-xl border p-4 text-left transition-all hover:border-amber-500/50 hover:shadow-sm"
        :class="statusFilter === 'yellow' ? 'border-amber-500 bg-amber-500/10' : 'bg-card'"
      >
        <div class="flex items-center justify-between text-muted-foreground">
          <span class="text-xs font-semibold uppercase tracking-wider">{{
            $t("dashboard.statWarning")
          }}</span>
          <StatusIcon status="yellow" class="size-4 text-amber-500" />
        </div>
        <div class="mt-3 flex items-baseline justify-between">
          <span class="text-2xl font-bold tracking-tight">{{ warningCount }}</span>
          <span class="text-xs text-muted-foreground" v-if="warningCount > 0">{{
            $t("dashboard.statAttention")
          }}</span>
        </div>
      </button>

      <button
        type="button"
        @click="toggleStatusFilter('red')"
        class="group flex flex-col justify-between rounded-xl border p-4 text-left transition-all hover:border-rose-500/50 hover:shadow-sm"
        :class="statusFilter === 'red' ? 'border-rose-500 bg-rose-500/10' : 'bg-card'"
      >
        <div class="flex items-center justify-between text-muted-foreground">
          <span class="text-xs font-semibold uppercase tracking-wider">{{
            $t("dashboard.statCritical")
          }}</span>
          <StatusIcon status="red" class="size-4 text-rose-500" />
        </div>
        <div class="mt-3 flex items-baseline justify-between">
          <span
            class="text-2xl font-bold tracking-tight"
            :class="{ 'text-rose-500': redCount > 0 }"
            >{{ redCount }}</span
          >
          <span class="text-xs text-rose-500 font-medium" v-if="redCount > 0">{{
            $t("dashboard.statActionRequired")
          }}</span>
        </div>
      </button>

      <RouterLink
        :to="{ name: 'account.extension.list', query: { hasUpdate: 'true' } }"
        class="group flex flex-col justify-between rounded-xl border bg-card p-4 text-left transition-all hover:border-primary/50 hover:shadow-sm"
      >
        <div class="flex items-center justify-between text-muted-foreground">
          <span class="text-xs font-semibold uppercase tracking-wider">{{
            $t("dashboard.statUpdates")
          }}</span>
          <icon-fa6-solid:arrows-rotate class="size-4 text-primary" />
        </div>
        <div class="mt-3 flex items-baseline justify-between">
          <span class="text-2xl font-bold tracking-tight">{{ pendingUpdatesCount }}</span>
          <span class="text-xs text-muted-foreground">{{ $t("dashboard.statExtensions") }}</span>
        </div>
      </RouterLink>
    </div>

    <!-- Active Filter Banner -->
    <div
      v-if="statusFilter"
      class="flex items-center justify-between rounded-lg border bg-muted/40 px-4 py-2 text-sm"
    >
      <span class="text-muted-foreground">
        <i18n-t keypath="dashboard.filterActive" tag="span">
          <template #status>
            <span class="font-medium text-foreground capitalize">{{
              statusFilter === "green"
                ? $t("dashboard.statHealthy")
                : statusFilter === "yellow"
                  ? $t("dashboard.statWarning")
                  : $t("dashboard.statCritical")
            }}</span>
          </template>
        </i18n-t>
      </span>
      <Button variant="ghost" size="sm" class="h-7 text-xs" @click="statusFilter = null">
        {{ $t("dashboard.filterClear") }}
      </Button>
    </div>

    <!-- Layout Container: Left (Shops) / Right (Insights) -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Left 2 Cols: Shops Grid -->
      <div class="space-y-6 lg:col-span-2">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-bold tracking-tight">{{ $t("dashboard.shopsHeading") }}</h2>
          <Button variant="outline" size="sm" as-child>
            <RouterLink :to="{ name: 'account.shops.new' }">
              <icon-fa6-solid:plus class="mr-1.5 size-3.5" />
              {{ $t("dashboard.addShop") }}
            </RouterLink>
          </Button>
        </div>

        <!-- Skeleton loader -->
        <div v-if="shops === null" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div v-for="i in 4" :key="i" class="rounded-xl border bg-card p-5 space-y-3">
            <div class="flex items-center justify-between">
              <Skeleton class="h-5 w-32" />
              <Skeleton class="h-4 w-12" />
            </div>
            <Skeleton class="h-4 w-full" />
            <Skeleton class="h-4 w-2/3" />
          </div>
        </div>

        <!-- Empty state -->
        <EmptyState
          v-else-if="shops.length === 0"
          :title="$t('dashboard.noShopsTitle')"
          :description="$t('dashboard.noShopsDesc')"
          :action-label="$t('dashboard.createShop')"
          :action-to="{ name: 'account.shops.new' }"
          :icon="IconStore"
        />

        <!-- Filtered Empty State -->
        <EmptyState
          v-else-if="filteredShops.length === 0"
          :title="$t('dashboard.noMatchingShopsTitle')"
          :description="$t('dashboard.noMatchingShopsDesc')"
          :action-label="$t('dashboard.filterClear')"
          @action="statusFilter = null"
          :icon="IconFilter"
        />

        <!-- Shop Cards Grid -->
        <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <CardSection
            v-for="shop in filteredShops"
            :key="shop.id"
            :title="shop.name"
            :icon="IconFolder"
            :to="{ name: 'account.shops.edit', params: { shopId: shop.id } }"
          >
            <template #actions>
              <Badge variant="outline" class="text-xs font-medium">
                {{ $t("dashboard.envCount", { count: getShopEnvironments(shop.id).length }) }}
              </Badge>
            </template>

            <div class="space-y-2">
              <div
                v-if="getShopEnvironments(shop.id).length === 0"
                class="text-xs text-muted-foreground py-2 italic"
              >
                {{ $t("dashboard.noEnvironmentsInShop") }}
              </div>
              <div
                v-for="env in getShopEnvironments(shop.id)"
                :key="env.id"
                class="group flex items-center justify-between rounded-lg border bg-muted/20 p-2.5 transition-colors hover:bg-accent"
              >
                <RouterLink
                  :to="{ name: 'account.environments.detail', params: { environmentId: env.id } }"
                  class="flex items-center gap-2.5 min-w-0 flex-1"
                >
                  <StatusIcon :status="env.status" class="size-3.5 shrink-0" />
                  <span
                    class="truncate text-sm font-medium group-hover:text-primary transition-colors"
                  >
                    {{ env.name }}
                  </span>
                </RouterLink>
                <span class="text-xs text-muted-foreground shrink-0 font-mono ml-2">
                  v{{ env.shopwareVersion || "?" }}
                </span>
              </div>
            </div>
          </CardSection>
        </div>
      </div>

      <!-- Right Col: Activity & Updates -->
      <div class="space-y-6">
        <!-- Extension Updates Card -->
        <CardSection
          :title="$t('dashboard.extensionUpdates')"
          :icon="IconCodeBranch"
          :to="{ name: 'account.extension.list', query: { hasUpdate: 'true' } }"
          :link-text="$t('dashboard.viewAll')"
        >
          <div
            v-if="extensionsNeedingUpdate.length === 0"
            class="text-sm text-muted-foreground py-4 text-center"
          >
            {{ $t("dashboard.allExtensionsUpToDate") }}
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="ext in extensionsNeedingUpdate.slice(0, 5)"
              :key="ext.name"
              class="flex items-center justify-between border-b pb-2.5 last:border-0 last:pb-0"
            >
              <div class="min-w-0 pr-2">
                <RouterLink
                  :to="{ name: 'account.extension.detail', params: { name: ext.name } }"
                  class="truncate text-sm font-medium hover:underline block"
                >
                  {{ ext.label || ext.name }}
                </RouterLink>
                <div class="text-xs text-muted-foreground truncate">
                  {{ $t("dashboard.affectedEnvs", { count: updateCount(ext) }) }}
                </div>
              </div>
              <Badge variant="secondary" class="shrink-0 text-xs">
                {{ $t("dashboard.updateAvailable") }}
              </Badge>
            </div>
          </div>
        </CardSection>

        <!-- Recent Changes Card -->
        <CardSection :title="$t('dashboard.recentChanges')" :icon="IconClockRotateLeft">
          <div
            v-if="changelogs.length === 0"
            class="text-sm text-muted-foreground py-4 text-center"
          >
            {{ $t("dashboard.noRecentChanges") }}
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="cl in changelogs.slice(0, 5)"
              :key="cl.id"
              class="flex items-start justify-between border-b pb-2.5 last:border-0 last:pb-0 gap-2"
            >
              <div class="min-w-0 flex-1">
                <div class="text-sm font-medium truncate">{{ cl.environmentName }}</div>
                <div class="text-xs text-muted-foreground truncate">
                  {{ sumChanges(cl) }}
                </div>
              </div>
              <button
                type="button"
                @click="openEnvironmentChangelog(cl)"
                class="shrink-0 text-xs text-primary hover:underline font-medium"
              >
                {{ $t("dashboard.details") }}
              </button>
            </div>
          </div>
        </CardSection>
      </div>
    </div>

    <!-- Environment Changelog Modal -->
    <ShopChangelog
      :show="viewEnvironmentChangelogDialog"
      :changelog="dialogEnvironmentChangelog"
      @close="closeEnvironmentChangelog"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink } from "vue-router";
import { useEnvironmentChangelogModal } from "@/composables/useEnvironmentChangelogModal";
import ShopChangelog from "@/components/modal/ShopChangelog.vue";
import { sumChanges } from "@/helpers/changelog";
import { hasUpdate, updateCount } from "@/composables/useAccountExtensions";
import {
  getAccountChangelogs,
  getAccountExtensions,
  getAccountShops,
  type AccountChangelog,
  type AccountEnvironment,
  type AccountExtension,
  type AccountShop,
} from "@/api/generated";
import {
  useAccountEnvironments,
  fetchAccountEnvironments,
} from "@/composables/useAccountEnvironments";
import { useSession } from "@/composables/useSession";

import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import StatusIcon from "@/components/StatusIcon.vue";
import EmptyState from "@/components/EmptyState.vue";
import CardSection from "@/components/CardSection.vue";
import PageHeader from "@/components/PageHeader.vue";
import IconFolder from "~icons/fa6-solid/folder";
import IconCodeBranch from "~icons/fa6-solid/code-branch";
import IconClockRotateLeft from "~icons/fa6-solid/clock-rotate-left";
import IconStore from "~icons/fa6-solid/store";
import IconFilter from "~icons/fa6-solid/filter";

const { t } = useI18n();
const { activeOrganizationId } = useSession();

const changelogs = ref<AccountChangelog[]>([]);
const extensions = ref<AccountExtension[]>([]);
const shops = ref<AccountShop[] | null>(null);

// Bumped on every load so a slow response from a previous organization
// cannot overwrite shops/changelogs/extensions after the user switched.
let loadGeneration = 0;

function loadDashboardData() {
  const generation = ++loadGeneration;
  getAccountChangelogs().then(({ data }) => {
    if (generation !== loadGeneration) return;
    if (data) changelogs.value = data;
  });
  getAccountExtensions().then(({ data }) => {
    if (generation !== loadGeneration) return;
    if (data) extensions.value = data;
  });
  getAccountShops().then(({ data }) => {
    if (generation !== loadGeneration) return;
    shops.value = data ?? [];
  });
}

loadDashboardData();

const { environments } = useAccountEnvironments();

// Header stats double as filters for the shops grid
const statusFilter = ref<"green" | "yellow" | "red" | null>(null);

watch(activeOrganizationId, () => {
  // Drop status filter so a selection from the previous org (e.g. "Warnings")
  // cannot hide every shop in the newly loaded organization.
  statusFilter.value = null;
  fetchAccountEnvironments();
  loadDashboardData();
});

const totalEnvironmentsCount = computed(() => environments.value.length);
const healthyCount = computed(() => environments.value.filter((e) => e.status === "green").length);
const warningCount = computed(() => environments.value.filter((e) => e.status === "yellow").length);
const redCount = computed(() => environments.value.filter((e) => e.status === "red").length);

const extensionsNeedingUpdate = computed(() => extensions.value.filter(hasUpdate));
const pendingUpdatesCount = computed(() => extensionsNeedingUpdate.value.length);

function toggleStatusFilter(status: "green" | "yellow" | "red") {
  statusFilter.value = statusFilter.value === status ? null : status;
}

const getShopEnvironments = (shopId: number): AccountEnvironment[] => {
  return environments.value.filter((e) => e.shopId === shopId);
};

const filteredShops = computed(() => {
  if (!shops.value) return [];
  if (!statusFilter.value) return shops.value;

  return shops.value.filter((shop) => {
    const shopEnvs = getShopEnvironments(shop.id);
    return shopEnvs.some((e) => e.status === statusFilter.value);
  });
});

const {
  viewEnvironmentChangelogDialog,
  dialogEnvironmentChangelog,
  openEnvironmentChangelog,
  closeEnvironmentChangelog,
} = useEnvironmentChangelogModal();
</script>
