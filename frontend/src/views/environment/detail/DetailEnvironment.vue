<template>
  <div v-if="environment" class="space-y-6">

    <!-- ═══════ TOP: Status bar + key metrics ═══════ -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card class="relative overflow-hidden">
        <CardContent class="flex items-center gap-3 p-4">
          <div :class="['flex size-10 shrink-0 items-center justify-center rounded-xl', statusBg]">
            <StatusIcon :status="environment.status" />
          </div>
          <div class="min-w-0">
            <div class="text-xs font-medium text-muted-foreground">Status</div>
            <div class="truncate font-semibold capitalize">{{ environment.status === 'green' ? 'Healthy' : environment.status === 'yellow' ? 'Warning' : 'Error' }}</div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10">
            <icon-fa6-solid:code-branch class="size-4 text-primary" />
          </div>
          <div class="min-w-0">
            <div class="text-xs font-medium text-muted-foreground">{{ $t("shopDetail.shopwareVersion") }}</div>
            <div class="flex items-center gap-2">
              <span class="font-mono font-semibold">{{ environment.shopwareVersion }}</span>
              <a
                v-if="latestShopwareVersion && latestShopwareVersion !== environment.shopwareVersion"
                :href="'https://github.com/shopware/platform/releases/tag/v' + latestShopwareVersion"
                target="_blank"
              >
                <Badge variant="secondary" class="font-mono text-[10px]">{{ latestShopwareVersion }}</Badge>
              </a>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10">
            <icon-fa6-solid:puzzle-piece class="size-4 text-primary" />
          </div>
          <div class="min-w-0">
            <div class="text-xs font-medium text-muted-foreground">{{ $t("shopDetail.extensions") }}</div>
            <div class="font-semibold tabular-nums">
              {{ environment.extensions?.length ?? 0 }}
              <span v-if="outdatedExtensions.length > 0" class="text-sm font-normal text-warning">({{ outdatedExtensions.length }} outdated)</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10">
            <icon-fa6-solid:clock class="size-4 text-primary" />
          </div>
          <div class="min-w-0">
            <div class="text-xs font-medium text-muted-foreground">{{ $t("shopDetail.lastCheckedAt") }}</div>
            <div class="truncate text-sm font-semibold tabular-nums">{{ formatDateTime(environment.lastScrapedAt) }}</div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- ═══════ MIDDLE: Bento grid — checks + extensions + tasks ═══════ -->
    <div class="grid gap-4 lg:grid-cols-2">

      <!-- Security checks -->
      <Card>
        <CardHeader class="flex-row items-center justify-between pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:shield-halved class="size-4 text-muted-foreground" />
            {{ $t("shopDetail.securityChecks") }}
          </CardTitle>
          <Button as-child variant="ghost" size="sm" class="h-7 text-xs">
            <RouterLink :to="checksRoute">
              {{ $t("shopDetail.viewAllChecks") }}
              <icon-fa6-solid:arrow-right class="ml-1 size-2.5" />
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent class="pt-0">
          <div v-if="sortedCriticalChecks.length === 0" class="flex items-center gap-2 rounded-lg bg-success/10 px-3 py-2.5 text-sm text-success">
            <icon-fa6-solid:circle-check class="size-4 shrink-0" />
            {{ $t("shopDetail.allChecksPassed") }}
          </div>
          <div v-else class="space-y-1.5">
            <div v-for="check in sortedCriticalChecks" :key="check.id" class="flex items-start gap-2.5 rounded-lg border px-3 py-2">
              <StatusIcon :status="check.level" class="mt-0.5 shrink-0" />
              <div class="min-w-0 flex-1 text-sm">
                <span>{{ check.message }}</span>
                <a v-if="check.link" :href="check.link" target="_blank" class="ml-1 inline-flex items-center gap-0.5 text-primary hover:underline">
                  <icon-fa6-solid:arrow-up-right-from-square class="size-2.5" />
                </a>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Outdated extensions -->
      <Card>
        <CardHeader class="flex-row items-center justify-between pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:plug class="size-4 text-muted-foreground" />
            {{ $t("shopDetail.extensions") }}
          </CardTitle>
          <Button as-child variant="ghost" size="sm" class="h-7 text-xs">
            <RouterLink :to="extensionsRoute">
              {{ $t("shopDetail.viewAllExtensions") }}
              <icon-fa6-solid:arrow-right class="ml-1 size-2.5" />
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent class="pt-0">
          <div v-if="outdatedExtensions.length === 0" class="flex items-center gap-2 rounded-lg bg-success/10 px-3 py-2.5 text-sm text-success">
            <icon-fa6-solid:circle-check class="size-4 shrink-0" />
            {{ $t("shopDetail.allExtensionsUpToDate") }}
          </div>
          <div v-else class="space-y-1.5">
            <div v-for="ext in outdatedExtensions" :key="ext.name" class="flex items-center gap-2.5 rounded-lg border px-3 py-2">
              <icon-fa6-solid:arrow-up class="size-3.5 shrink-0 text-info" />
              <div class="min-w-0 flex-1 text-sm">
                <span class="font-medium">{{ ext.label }}</span>
                <button class="ml-1 text-primary hover:underline" @click="openExtensionChangelog(ext)">
                  {{ ext.version }} → {{ ext.latestVersion }}
                </button>
              </div>
              <a v-if="ext.storeLink" :href="ext.storeLink" target="_blank" class="shrink-0 text-muted-foreground hover:text-foreground">
                <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
              </a>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Overdue tasks -->
      <Card>
        <CardHeader class="flex-row items-center justify-between pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:list-check class="size-4 text-muted-foreground" />
            {{ $t("shopDetail.scheduledTasks") }}
          </CardTitle>
          <Button as-child variant="ghost" size="sm" class="h-7 text-xs">
            <RouterLink :to="tasksRoute">
              {{ $t("shopDetail.viewAllScheduledTasks") }}
              <icon-fa6-solid:arrow-right class="ml-1 size-2.5" />
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent class="pt-0">
          <div v-if="overdueTasks.length === 0" class="flex items-center gap-2 rounded-lg bg-success/10 px-3 py-2.5 text-sm text-success">
            <icon-fa6-solid:circle-check class="size-4 shrink-0" />
            {{ $t("shopDetail.noOverdueTasks") }}
          </div>
          <div v-else class="space-y-1.5">
            <div v-for="task in overdueTasks" :key="task.id" class="flex items-start gap-2.5 rounded-lg border px-3 py-2">
              <icon-fa6-solid:clock class="mt-0.5 size-3.5 shrink-0 text-warning" />
              <div class="min-w-0 text-sm">
                <div class="font-medium">{{ task.name }}</div>
                <div class="text-xs text-muted-foreground">{{ getOverdueTime(task.nextExecutionTime ?? "") }} overdue</div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Recent changes -->
      <Card>
        <CardHeader class="flex-row items-center justify-between pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-fa6-solid:clock-rotate-left class="size-4 text-muted-foreground" />
            {{ $t("shopDetail.recentChanges") }}
          </CardTitle>
          <Button as-child variant="ghost" size="sm" class="h-7 text-xs">
            <RouterLink :to="changelogRoute">
              {{ $t("shopDetail.viewAllChanges") }}
              <icon-fa6-solid:arrow-right class="ml-1 size-2.5" />
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent class="pt-0">
          <div v-if="recentChangelogs.length === 0" class="flex items-center gap-2 rounded-lg bg-muted px-3 py-2.5 text-sm text-muted-foreground">
            <icon-fa6-solid:circle-info class="size-4 shrink-0" />
            {{ $t("shopDetail.noRecentChanges") }}
          </div>
          <div v-else class="space-y-1.5">
            <div
              v-for="changelog in recentChangelogs"
              :key="changelog.id"
              class="flex cursor-pointer items-center gap-3 rounded-lg border px-3 py-2 transition-colors hover:bg-accent"
              @click="openEnvironmentChangelog(changelog)"
            >
              <span class="shrink-0 text-xs tabular-nums text-muted-foreground">{{ formatDate(changelog.date) }}</span>
              <Separator orientation="vertical" class="h-4" />
              <span class="min-w-0 flex-1 truncate text-sm">{{ sumChanges(changelog) }}</span>
              <icon-fa6-solid:chevron-right class="size-2.5 shrink-0 text-muted-foreground" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- ═══════ BOTTOM: Environment info ═══════ -->
    <Card>
      <CardHeader class="flex-row items-center justify-between pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:circle-info class="size-4 text-muted-foreground" />
          {{ $t("shopDetail.shopInfo") }}
        </CardTitle>
        <Button
          v-if="latestShopwareVersion && latestShopwareVersion !== environment.shopwareVersion"
          variant="outline"
          size="sm"
          class="h-7 text-xs"
          @click="openUpdateWizard"
        >
          <icon-fa6-solid:rotate class="mr-1 size-2.5" />
          {{ $t("shopDetail.compatibilityCheck") }}
        </Button>
      </CardHeader>
      <CardContent class="pt-0">
        <div class="grid gap-x-8 gap-y-4 sm:grid-cols-2 lg:grid-cols-3">
          <div v-for="item in infoItems" :key="item.label">
            <dt class="text-xs font-medium text-muted-foreground">{{ item.label }}</dt>
            <dd class="mt-0.5">
              <component :is="item.component" v-if="item.component" v-bind="item.componentProps" />
              <span v-else class="text-sm">{{ item.value }}</span>
            </dd>
          </div>

          <div class="sm:col-span-2 lg:col-span-3">
            <dt class="text-xs font-medium text-muted-foreground">
              Bypass Authentication Header
              <span title="If your website is protected by authentication, configure the header 'shopmon-shop-token' with this value to be excluded">
                <icon-fa6-solid:circle-info class="ml-0.5 inline size-3 text-info" />
              </span>
            </dt>
            <dd class="mt-1 flex items-center gap-2">
              <code class="rounded bg-muted px-2 py-1 font-mono text-xs break-all">{{ environment.environmentToken }}</code>
              <Button type="button" variant="ghost" size="icon" class="size-7 shrink-0" title="Copy token" @click="copyToken">
                <icon-fa6-solid:copy class="size-3" />
              </Button>
            </dd>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Modals -->
    <EnvironmentChangelog
      :show="viewEnvironmentChangelogDialog"
      :changelog="dialogEnvironmentChangelog"
      @close="closeEnvironmentChangelog"
    />
    <ExtensionChangelog
      :show="viewExtensionChangelogDialog"
      :extension="dialogExtension"
      @close="closeExtensionChangelog"
    />
    <ShopwareUpdateWizard
      :show="viewUpdateWizardDialog"
      :shopware-versions="shopwareVersions"
      :loading="loadingUpdateWizard"
      :extensions="dialogUpdateWizard"
      @close="viewUpdateWizardDialog = false"
      @version-selected="loadUpdateWizard"
    />
  </div>
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from "@/helpers/formatter";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useEnvironmentChangelogModal } from "@/composables/useEnvironmentChangelogModal";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import EnvironmentChangelog from "@/components/modal/ShopChangelog.vue";
import { ref, computed } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useAlert } from "@/composables/useAlert";
import { sumChanges } from "@/helpers/changelog";
import { useI18n } from "vue-i18n";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import StatusIcon from "@/components/StatusIcon.vue";

type Extension = components["schemas"]["EnvironmentExtension"];
type ExtensionWithCompatibility = Extension & {
  compatibility?: { name: string; label: string; type: string };
};

const route = useRoute();
const { t } = useI18n();
const { error, success } = useAlert();
const { environment, shopwareVersions, latestShopwareVersion } = useEnvironmentDetail();

// Route helpers
const routeParams = computed(() => ({
  organizationId: route.params.organizationId,
  environmentId: route.params.environmentId,
}));
const checksRoute = computed(() => ({ name: "account.environments.detail.checks", params: routeParams.value }));
const extensionsRoute = computed(() => ({ name: "account.environments.detail.extensions", params: routeParams.value }));
const tasksRoute = computed(() => ({ name: "account.environments.detail.tasks", params: routeParams.value }));
const changelogRoute = computed(() => ({ name: "account.environments.detail.changelog", params: routeParams.value }));

// Status styling
const statusBg = computed(() => {
  switch (environment.value?.status) {
    case "green": return "bg-success/15";
    case "yellow": return "bg-warning/15";
    case "red": return "bg-destructive/15";
    default: return "bg-muted";
  }
});

// Info grid items
const infoItems = computed(() => {
  if (!environment.value) return [];
  return [
    { label: t("shopDetail.shopwareVersion"), value: environment.value.shopwareVersion },
    { label: t("shopDetail.lastCheckedAt"), value: formatDateTime(environment.value.lastScrapedAt) },
    {
      label: t("shopDetail.lastShopUpdate"),
      value: environment.value.lastChangelog?.date
        ? formatDate(environment.value.lastChangelog.date)
        : t("common.never"),
    },
    {
      label: t("shopDetail.lastDeployment"),
      value: environment.value.deploymentsCount > 0
        ? `${environment.value.deploymentsCount} deployment${environment.value.deploymentsCount !== 1 ? "s" : ""}`
        : t("common.never"),
    },
    { label: t("shopDetail.environment"), value: environment.value.cache?.environment ?? "-" },
    { label: t("shopDetail.cacheAdapter"), value: environment.value.cache?.cacheAdapter ?? "-" },
    {
      label: t("shopDetail.httpCache"),
      value: environment.value.cache?.httpCache ? t("common.enabled") : t("common.disabled"),
    },
    { label: "Organization", value: environment.value.organizationName },
    { label: "Project", value: environment.value.shopName ?? "-" },
  ];
});

async function copyToken() {
  if (environment.value?.environmentToken) {
    await navigator.clipboard.writeText(environment.value.environmentToken);
    success("Token copied to clipboard");
  }
}

// Changelogs
const { viewExtensionChangelogDialog, dialogExtension, openExtensionChangelog, closeExtensionChangelog } = useExtensionChangelogModal();
const { viewEnvironmentChangelogDialog, dialogEnvironmentChangelog, openEnvironmentChangelog, closeEnvironmentChangelog } = useEnvironmentChangelogModal();

// Computed data
const sortedCriticalChecks = computed(() => {
  if (!environment.value?.checks) return [];
  return environment.value.checks
    .filter((c) => c.level !== "green" && !environment.value?.ignores?.includes(c.id))
    .sort((a, b) => {
      if (a.level === "red" && b.level !== "red") return -1;
      if (a.level !== "red" && b.level === "red") return 1;
      return 0;
    })
    .slice(0, 5);
});

const outdatedExtensions = computed(() => {
  if (!environment.value?.extensions) return [];
  return environment.value.extensions
    .filter((ext) => ext.installed && ext.latestVersion && ext.version !== ext.latestVersion)
    .slice(0, 5);
});

const overdueTasks = computed(() => {
  if (!environment.value?.scheduledTasks) return [];
  const fifteenMinutesAgo = new Date(Date.now() - 15 * 60 * 1000);
  return environment.value.scheduledTasks
    .filter(
      (task) =>
        task.overdue && task.status !== "inactive" && task.nextExecutionTime && new Date(task.nextExecutionTime) < fifteenMinutesAgo,
    )
    .slice(0, 5);
});

const recentChangelogs = computed(() => {
  if (!environment.value?.changelogs) return [];
  return environment.value.changelogs.slice(0, 5);
});

function getOverdueTime(nextExecutionTime: string): string {
  const diffMs = Date.now() - new Date(nextExecutionTime).getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  if (diffMinutes < 60) return `${diffMinutes}m`;
  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) return `${diffHours}h`;
  return `${Math.floor(diffHours / 24)}d`;
}

// Update wizard
const viewUpdateWizardDialog = ref(false);
const loadingUpdateWizard = ref(false);
const dialogUpdateWizard = ref<ExtensionWithCompatibility[] | null>(null);

function openUpdateWizard() {
  dialogUpdateWizard.value = null;
  viewUpdateWizardDialog.value = true;
}

async function loadUpdateWizard(version: string) {
  if (!environment.value?.extensions) return;
  loadingUpdateWizard.value = true;

  try {
    const { data: pluginCompatibility } = await api.POST("/info/extension-compatibility", {
      body: {
        currentVersion: environment.value.shopwareVersion,
        futureVersion: version,
        extensions: environment.value.extensions.map((e) => ({ name: e.name, version: e.version })),
      },
    });

    if (!pluginCompatibility) return;

    const extensions = JSON.parse(JSON.stringify(environment.value.extensions)) as ExtensionWithCompatibility[];
    for (const ext of extensions) {
      const compat = pluginCompatibility.find((p) => p.name === ext.name);
      ext.compatibility = compat ? (compat.status as ExtensionWithCompatibility["compatibility"]) : undefined;
    }

    dialogUpdateWizard.value = extensions.sort((a, b) => {
      if (a.active !== b.active) return a.active ? -1 : 1;
      return a.label.localeCompare(b.label);
    });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  } finally {
    loadingUpdateWizard.value = false;
  }
}
</script>
