<template>
  <div class="space-y-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ t("admin.loadingEnvironment") }}
    </div>

    <!-- Error -->
    <Alert v-else-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Not found -->
    <EmptyState
      v-else-if="!environment"
      :icon="IconEarthAmericas"
      :title="t('admin.environmentNotFound')"
    />

    <!-- Detail -->
    <template v-else>
      <div class="space-y-4">
        <!-- Breadcrumb nav -->
        <nav class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <RouterLink
            :to="{ name: 'admin.dashboard' }"
            class="hover:text-primary transition-colors"
          >
            {{ t("admin.dashboard") }}
          </RouterLink>
          <icon-fa6-solid:chevron-right class="size-2.5 opacity-50" />
          <RouterLink
            :to="{ name: 'admin.environments' }"
            class="hover:text-primary transition-colors"
          >
            {{ t("common.environments") }}
          </RouterLink>
          <icon-fa6-solid:chevron-right class="size-2.5 opacity-50" />
          <span class="font-medium text-foreground">{{ environment.name }}</span>
        </nav>

        <PageHeader :title="environment.name">
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              @click="router.push({ name: 'admin.environments' })"
            >
              <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
              {{ t("admin.back") }}
            </Button>
            <Button size="sm" as-child class="gap-1.5 shadow-xs">
              <a :href="environment.url" target="_blank" rel="noopener noreferrer">
                <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
                {{ t("admin.openEnvironment") }}
              </a>
            </Button>
          </div>
        </PageHeader>
      </div>

      <!-- Stat cards -->
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          :icon="IconHeartPulse"
          :value="friendlyStatusLabel"
          :label="t('common.status')"
          :color="statusColor"
        />
        <StatCard
          :icon="IconTag"
          :value="environment.shopwareVersion"
          :label="t('admin.shopwareVersion')"
          color="primary"
        />
        <StatCard
          :icon="IconPuzzlePiece"
          :value="environment.extensions.length"
          :label="t('admin.extensions')"
          color="primary"
        />
        <StatCard
          :icon="IconClock"
          :value="environment.lastScrapedAt ? formatDateTime(environment.lastScrapedAt) : '—'"
          :label="t('admin.lastScraped')"
          color="muted"
        />
      </div>

      <!-- Scrape error -->
      <Alert
        v-if="environment.lastScrapedError"
        variant="destructive"
        class="border-destructive/30 bg-destructive/10"
      >
        <AlertTitle class="font-semibold">{{ t("admin.lastScrapedError") }}</AlertTitle>
        <AlertDescription>{{ environment.lastScrapedError }}</AlertDescription>
      </Alert>

      <!-- Connection issue warning -->
      <Alert
        v-if="environment.connectionIssueCount > 0"
        variant="destructive"
        class="border-amber-500/30 bg-amber-500/10 text-amber-900 dark:text-amber-200"
      >
        <icon-fa6-solid:triangle-exclamation class="size-4 text-amber-600 dark:text-amber-400" />
        <AlertTitle class="font-semibold text-amber-700 dark:text-amber-300"
          >Unstable Connection</AlertTitle
        >
        <AlertDescription class="text-amber-800 dark:text-amber-300">
          This environment has recorded {{ environment.connectionIssueCount }} consecutive
          connection issue(s) during background monitoring.
        </AlertDescription>
      </Alert>

      <!-- Org / shop info -->
      <div
        class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground rounded-lg border bg-card/50 px-3.5 py-2"
      >
        <icon-fa6-solid:building class="size-3.5 text-primary" />
        <RouterLink
          :to="{ name: 'admin.organizations.detail', params: { id: environment.organizationId } }"
          class="font-semibold text-foreground hover:text-primary transition-colors"
        >
          {{ environment.organizationName }}
        </RouterLink>
        <span class="text-muted-foreground/40">/</span>
        <span class="flex items-center gap-1.5 font-medium text-foreground">
          <icon-fa6-solid:store class="size-3 text-muted-foreground" />
          {{ environment.shopName }}
        </span>
      </div>

      <!-- Bento grid -->
      <div class="grid gap-4 lg:grid-cols-2">
        <!-- Health checks -->
        <CardSection :icon="IconShieldHalved" :title="t('admin.checks')">
          <EmptyState
            v-if="environment.checks.length === 0"
            :icon="IconCircleCheck"
            :title="t('admin.noChecks')"
            size="sm"
          />
          <div v-else class="space-y-2 max-h-96 overflow-y-auto pr-1">
            <div
              v-for="check in environment.checks"
              :key="check.id"
              class="group flex items-start gap-3 rounded-xl border bg-card px-4 py-3 transition-all hover:border-primary/30"
            >
              <Badge
                v-if="check.level === 'green'"
                class="inline-flex items-center gap-1 shrink-0 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 text-xs px-2 py-0.5 font-medium"
              >
                <span class="size-1.5 rounded-full bg-emerald-500" />
                Pass
              </Badge>
              <Badge
                v-else-if="check.level === 'yellow'"
                class="inline-flex items-center gap-1 shrink-0 bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 text-xs px-2 py-0.5 font-medium"
              >
                <span class="size-1.5 rounded-full bg-amber-500" />
                Warning
              </Badge>
              <Badge
                v-else
                class="inline-flex items-center gap-1 shrink-0 bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20 text-xs px-2 py-0.5 font-medium"
              >
                <span class="size-1.5 rounded-full bg-rose-500 animate-pulse" />
                Error
              </Badge>
              <div class="min-w-0 flex-1">
                <div class="text-sm font-medium leading-snug">{{ check.message }}</div>
                <div class="mt-0.5 text-xs text-muted-foreground font-mono text-[11px]">
                  {{ check.source }}
                </div>
              </div>
              <a
                v-if="check.link"
                :href="check.link"
                target="_blank"
                rel="noopener noreferrer"
                class="shrink-0 p-1 text-muted-foreground hover:text-primary transition-colors"
              >
                <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
              </a>
            </div>
          </div>
        </CardSection>

        <!-- Extensions -->
        <CardSection :icon="IconPuzzlePiece" :title="t('admin.extensions')">
          <EmptyState
            v-if="environment.extensions.length === 0"
            :icon="IconPuzzlePiece"
            :title="t('admin.noExtensions')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="ext in environment.extensions"
              :key="ext.name"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">{{ ext.label }}</div>
                <div class="mt-0.5 font-mono text-xs text-muted-foreground tabular-nums">
                  <template v-if="isOutdated(ext)">
                    {{ ext.version }} → {{ ext.latestVersion }}
                  </template>
                  <template v-else>{{ ext.version }}</template>
                </div>
              </div>
              <Badge
                v-if="isOutdated(ext)"
                class="shrink-0 border-warning/30 bg-warning/10 text-warning"
                >{{ t("admin.outdated") }}</Badge
              >
              <a
                v-if="ext.storeLink"
                :href="ext.storeLink"
                target="_blank"
                rel="noopener noreferrer"
                class="shrink-0 text-muted-foreground hover:text-foreground"
              >
                <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
              </a>
            </div>
          </div>
        </CardSection>

        <!-- Scheduled tasks -->
        <CardSection :icon="IconListCheck" :title="t('admin.scheduledTasks')">
          <EmptyState
            v-if="environment.scheduledTasks.length === 0"
            :icon="IconListCheck"
            :title="t('admin.noTasks')"
            size="sm"
          />
          <div v-else class="space-y-2">
            <div
              v-for="task in environment.scheduledTasks"
              :key="task.id"
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium">{{ task.name }}</div>
                <div class="mt-0.5 text-xs text-muted-foreground tabular-nums">
                  {{ task.nextExecutionTime ? formatDateTime(task.nextExecutionTime) : "—" }}
                </div>
              </div>
              <Badge variant="secondary" class="shrink-0 capitalize">{{ task.status }}</Badge>
              <Badge v-if="task.overdue" variant="destructive" class="shrink-0">{{
                t("admin.overdue")
              }}</Badge>
            </div>
          </div>
        </CardSection>

        <!-- Last deployment -->
        <CardSection :icon="IconRocket" :title="t('admin.lastDeployment')">
          <EmptyState
            v-if="!environment.lastDeployment"
            :icon="IconRocket"
            :title="t('admin.noDeployment')"
            size="sm"
          />
          <div v-else class="space-y-3 rounded-xl border bg-card px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <span class="truncate text-sm font-medium">{{
                environment.lastDeployment.name
              }}</span>
              <Badge
                :variant="environment.lastDeployment.returnCode === 0 ? undefined : 'destructive'"
                :class="
                  environment.lastDeployment.returnCode === 0
                    ? 'shrink-0 border-success/30 bg-success/10 text-success'
                    : 'shrink-0'
                "
              >
                {{ t("admin.returnCode") }}: {{ environment.lastDeployment.returnCode }}
              </Badge>
            </div>
            <code
              class="block rounded bg-muted px-2 py-1.5 font-mono text-xs break-all text-muted-foreground"
            >
              {{ environment.lastDeployment.command }}
            </code>
            <div class="text-xs text-muted-foreground tabular-nums">
              {{ formatDateTime(environment.lastDeployment.createdAt) }}
            </div>
          </div>
        </CardSection>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import {
  adminGetEnvironmentDetail,
  type AdminEnvironmentDetail,
  type AdminEnvironmentExtension,
} from "@/api/generated";
import { formatDateTime } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import PageHeader from "@/components/PageHeader.vue";
import StatCard from "@/components/StatCard.vue";
import CardSection from "@/components/CardSection.vue";
import EmptyState from "@/components/EmptyState.vue";

import IconHeartPulse from "~icons/fa6-solid/heart-pulse";
import IconTag from "~icons/fa6-solid/tag";
import IconPuzzlePiece from "~icons/fa6-solid/puzzle-piece";
import IconClock from "~icons/fa6-solid/clock";
import IconShieldHalved from "~icons/fa6-solid/shield-halved";
import IconListCheck from "~icons/fa6-solid/list-check";
import IconRocket from "~icons/fa6-solid/rocket";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconEarthAmericas from "~icons/fa6-solid/earth-americas";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const environment = ref<AdminEnvironmentDetail | null>(null);
const loading = ref(true);
const error = ref("");

const statusColor = computed(() => {
  switch (environment.value?.status) {
    case "green":
      return "success" as const;
    case "yellow":
      return "warning" as const;
    default:
      return "destructive" as const;
  }
});

const friendlyStatusLabel = computed(() => {
  switch (environment.value?.status) {
    case "green":
      return t("admin.statusGreen");
    case "yellow":
      return t("admin.statusYellow");
    case "red":
      return t("admin.statusRed");
    default:
      return environment.value?.status ?? "";
  }
});

function checkLevelClass(level: string): string {
  switch (level) {
    case "error":
      return "border-transparent bg-destructive text-white";
    case "warning":
      return "border-warning/30 bg-warning/10 text-warning";
    default:
      return "border-transparent bg-secondary text-secondary-foreground";
  }
}

function isOutdated(ext: AdminEnvironmentExtension): boolean {
  return Boolean(ext.latestVersion && ext.latestVersion !== ext.version);
}

async function loadEnvironment(id: number) {
  loading.value = true;
  error.value = "";
  environment.value = null;

  const { data, error: respError } = await adminGetEnvironmentDetail({
    path: { envId: id },
  });

  if (respError) {
    error.value = (respError as { message?: string }).message ?? t("admin.environmentNotFound");
  } else {
    environment.value = data;
  }

  loading.value = false;
}

watch(
  () => route.params.id,
  (id) => {
    loadEnvironment(Number(id));
  },
  { immediate: true },
);
</script>
