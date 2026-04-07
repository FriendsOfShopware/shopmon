<template>
  <div v-if="shops">
    <!-- Page header -->
    <div
      class="mb-8 flex items-end justify-between max-sm:flex-col max-sm:items-start max-sm:gap-4"
    >
      <div>
        <h1 class="text-2xl font-bold tracking-tight sm:text-3xl">{{ $t("dashboard.title") }}</h1>
        <p class="mt-1 text-muted-foreground">{{ $t("dashboard.myEnvironments") }}</p>
      </div>
      <div v-if="shops.length > 0" class="hidden items-center gap-6 sm:flex">
        <div class="text-right">
          <div class="text-2xl font-bold tabular-nums">{{ shops.length }}</div>
          <div class="text-xs text-muted-foreground">{{ $t("dashboard.shops") }}</div>
        </div>
        <Separator orientation="vertical" class="h-10" />
        <div class="text-right">
          <div class="text-2xl font-bold tabular-nums text-success">{{ greenCount }}</div>
          <div class="text-xs text-muted-foreground">{{ $t("dashboard.healthy") }}</div>
        </div>
        <div class="text-right">
          <div
            class="text-2xl font-bold tabular-nums"
            :class="warnCount > 0 ? 'text-warning' : 'text-muted-foreground'"
          >
            {{ warnCount }}
          </div>
          <div class="text-xs text-muted-foreground">{{ $t("dashboard.warnings") }}</div>
        </div>
        <div class="text-right">
          <div
            class="text-2xl font-bold tabular-nums"
            :class="errorCount > 0 ? 'text-destructive' : 'text-muted-foreground'"
          >
            {{ errorCount }}
          </div>
          <div class="text-xs text-muted-foreground">{{ $t("dashboard.errors") }}</div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState
      v-if="shops.length === 0"
      :icon="IconFolder"
      :title="$t('shop.noShops')"
      :description="$t('shop.getStarted')"
    >
      <Button as-child>
        <RouterLink :to="{ name: 'account.shops.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("shop.addShop") }}
        </RouterLink>
      </Button>
    </EmptyState>

    <template v-else>
      <!-- Alerts row -->
      <div
        v-if="outdatedExtensionCount > 0 || errorCount > 0"
        class="mb-6 grid gap-3 sm:grid-cols-2"
      >
        <!-- Outdated extensions alert -->
        <RouterLink
          v-if="outdatedExtensionCount > 0"
          :to="{ name: 'account.extension.list' }"
          class="flex items-center gap-3 rounded-xl border border-warning/20 bg-warning/5 p-4 transition-colors hover:bg-warning/10"
        >
          <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-warning/10">
            <icon-fa6-solid:arrow-up class="size-4 text-warning" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm font-semibold">
              {{ $t("dashboard.extensionUpdatesAvailable", { count: outdatedExtensionCount }) }}
            </div>
            <div class="text-xs text-muted-foreground">{{ $t("dashboard.acrossYourShops") }}</div>
          </div>
          <icon-fa6-solid:chevron-right class="size-3 shrink-0 text-muted-foreground" />
        </RouterLink>

        <!-- Error shops alert -->
        <div
          v-if="errorCount > 0"
          class="flex items-center gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4"
        >
          <div
            class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-destructive/10"
          >
            <icon-fa6-solid:circle-xmark class="size-4 text-destructive" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm font-semibold">
              {{ $t("dashboard.shopsNeedAttention", { count: errorCount }) }}
            </div>
            <div class="text-xs text-muted-foreground">
              {{ $t("dashboard.healthChecksFailing") }}
            </div>
          </div>
        </div>
      </div>

      <!-- Shops grid -->
      <section class="mb-8">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          <RouterLink
            v-for="shop in shops"
            :key="shop.id"
            :to="shopLink(shop)"
            class="group relative flex items-start gap-3 rounded-xl border bg-card p-4 shadow-sm transition-all duration-200 hover:border-primary/30 hover:shadow-md"
          >
            <div
              class="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted"
            >
              <img
                v-if="defaultEnv(shop)?.favicon"
                :src="defaultEnv(shop)!.favicon!"
                alt=""
                class="size-5 rounded"
              />
              <icon-fa6-solid:folder v-else class="size-4 text-muted-foreground/50" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span
                  class="truncate font-semibold leading-tight transition-colors group-hover:text-primary"
                  >{{ shop.name }}</span
                >
                <StatusIcon v-if="defaultEnv(shop)" :status="defaultEnv(shop)!.status" />
              </div>
              <div class="mt-1.5 flex items-center gap-2">
                <Badge v-if="defaultEnv(shop)" variant="secondary" class="font-mono text-xs">
                  {{ defaultEnv(shop)!.shopwareVersion }}
                </Badge>
                <span v-if="envCount(shop) > 1" class="text-xs text-muted-foreground">
                  {{ $t("shop.envCount", { count: envCount(shop) }) }}
                </span>
              </div>
            </div>
          </RouterLink>
        </div>
      </section>

      <!-- Shopware version overview + recent changes -->
      <div class="mb-8 grid gap-6 lg:grid-cols-3">
        <!-- Version distribution -->
        <CardSection :icon="IconCodeBranch" :title="$t('dashboard.shopwareVersions')">
          <div class="space-y-2">
            <div
              v-for="version in versionDistribution"
              :key="version.version"
              class="flex items-center gap-3"
            >
              <Badge variant="secondary" class="min-w-[5rem] justify-center font-mono text-xs">{{
                version.version
              }}</Badge>
              <div class="flex-1">
                <div class="h-2 overflow-hidden rounded-full bg-muted">
                  <div
                    class="h-full rounded-full bg-primary transition-all"
                    :style="{ width: `${(version.count / environments.length) * 100}%` }"
                  />
                </div>
              </div>
              <span class="w-6 text-right text-xs tabular-nums text-muted-foreground">{{
                version.count
              }}</span>
            </div>
          </div>
        </CardSection>

        <!-- Recent changes -->
        <CardSection
          :icon="IconClockRotateLeft"
          :title="$t('dashboard.lastChanges')"
          class="lg:col-span-2"
        >
          <div
            v-if="changelogs.length === 0"
            class="flex flex-col items-center gap-2 py-8 text-center text-muted-foreground"
          >
            <icon-fa6-solid:clock-rotate-left class="size-8 opacity-30" />
            <p class="text-sm">{{ $t("dashboard.noRecentChanges") }}</p>
          </div>
          <div v-else class="space-y-1.5">
            <RouterLink
              v-for="log in changelogs.slice(0, 8)"
              :key="log.id"
              :to="{
                name: 'account.environments.detail',
                params: { environmentId: log.environmentId },
              }"
              class="flex items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-accent"
            >
              <span class="shrink-0 text-xs tabular-nums text-muted-foreground">{{
                formatDate(log.date)
              }}</span>
              <Separator orientation="vertical" class="h-4" />
              <span class="min-w-0 truncate text-sm font-medium">{{ log.environmentName }}</span>
              <span class="min-w-0 flex-1 truncate text-xs text-muted-foreground">{{
                sumChanges(log)
              }}</span>
            </RouterLink>
          </div>
        </CardSection>
      </div>
    </template>
  </div>

  <!-- Loading skeleton -->
  <div v-else class="space-y-8">
    <div>
      <Skeleton class="mb-2 h-9 w-48" />
      <Skeleton class="h-5 w-64" />
    </div>
    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Skeleton v-for="i in 4" :key="i" class="h-24 rounded-xl" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { sumChanges } from "@/helpers/changelog";
import { formatDate } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import {
  useAccountEnvironments,
  fetchAccountEnvironments,
} from "@/composables/useAccountEnvironments";
import { useSession } from "@/composables/useSession";

import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import StatusIcon from "@/components/StatusIcon.vue";
import EmptyState from "@/components/EmptyState.vue";
import CardSection from "@/components/CardSection.vue";
import IconFolder from "~icons/fa6-solid/folder";
import IconCodeBranch from "~icons/fa6-solid/code-branch";
import IconClockRotateLeft from "~icons/fa6-solid/clock-rotate-left";

type AccountShop = components["schemas"]["AccountShop"];
type AccountEnvironment = components["schemas"]["AccountEnvironment"];

const { t } = useI18n();
const { activeOrganizationId } = useSession();

const changelogs = ref<components["schemas"]["AccountChangelog"][]>([]);
const extensions = ref<components["schemas"]["AccountExtension"][]>([]);
const shops = ref<AccountShop[] | null>(null);

function loadDashboardData() {
  api.GET("/account/changelogs").then(({ data }) => {
    if (data) changelogs.value = data;
  });
  api.GET("/account/extensions").then(({ data }) => {
    if (data) extensions.value = data;
  });
  api.GET("/account/shops").then(({ data }) => {
    shops.value = data ?? [];
  });
}

loadDashboardData();

const { environments } = useAccountEnvironments();

watch(activeOrganizationId, () => {
  fetchAccountEnvironments();
  loadDashboardData();
});

function defaultEnv(shop: AccountShop): AccountEnvironment | undefined {
  return environments.value.find((e) => e.id === shop.defaultEnvironmentId);
}

function envCount(shop: AccountShop): number {
  return environments.value.filter((e) => e.shopId === shop.id).length;
}

function shopLink(shop: AccountShop) {
  return {
    name: "account.environments.detail",
    params: { environmentId: shop.defaultEnvironmentId },
  };
}

// Status counts based on default environments
const greenCount = computed(
  () => (shops.value ?? []).filter((s) => defaultEnv(s)?.status === "green").length,
);
const warnCount = computed(
  () => (shops.value ?? []).filter((s) => defaultEnv(s)?.status === "yellow").length,
);
const errorCount = computed(
  () => (shops.value ?? []).filter((s) => defaultEnv(s)?.status === "red").length,
);

// Outdated extensions count
const outdatedExtensionCount = computed(
  () =>
    extensions.value.filter((e) => e.installed && e.latestVersion && e.version !== e.latestVersion)
      .length,
);

// Shopware version distribution (still per-environment for accuracy)
const versionDistribution = computed(() => {
  const counts = new Map<string, number>();
  for (const env of environments.value ?? []) {
    counts.set(env.shopwareVersion, (counts.get(env.shopwareVersion) ?? 0) + 1);
  }
  return Array.from(counts.entries())
    .map(([version, count]) => ({ version, count }))
    .sort((a, b) => b.count - a.count);
});
</script>
