<template>
  <div v-if="environments">
    <!-- Page header with welcome + quick stats -->
    <div
      class="mb-8 flex items-end justify-between max-sm:flex-col max-sm:items-start max-sm:gap-4"
    >
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold tracking-tight">{{ $t("dashboard.title") }}</h1>
        <p class="mt-1 text-muted-foreground">{{ $t("dashboard.myEnvironments") }}</p>
      </div>
      <div v-if="environments.length > 0" class="hidden items-center gap-6 sm:flex">
        <div class="text-right">
          <div class="text-2xl font-bold tabular-nums">{{ environments.length }}</div>
          <div class="text-xs text-muted-foreground">Environments</div>
        </div>
        <Separator orientation="vertical" class="h-10" />
        <div class="text-right">
          <div class="text-2xl font-bold tabular-nums text-success">{{ greenCount }}</div>
          <div class="text-xs text-muted-foreground">Healthy</div>
        </div>
        <div class="text-right">
          <div
            class="text-2xl font-bold tabular-nums"
            :class="warnCount > 0 ? 'text-warning' : 'text-muted-foreground'"
          >
            {{ warnCount }}
          </div>
          <div class="text-xs text-muted-foreground">Warnings</div>
        </div>
        <div class="text-right">
          <div
            class="text-2xl font-bold tabular-nums"
            :class="errorCount > 0 ? 'text-destructive' : 'text-muted-foreground'"
          >
            {{ errorCount }}
          </div>
          <div class="text-xs text-muted-foreground">Errors</div>
        </div>
      </div>
    </div>

    <!-- Environments grid -->
    <section class="mb-8">
      <div
        v-if="environments.length === 0"
        class="flex w-full flex-col items-center gap-6 rounded-xl border border-dashed bg-card px-10 py-16 text-center"
      >
        <FolderPlus class="size-12 text-muted-foreground" />
        <h2 class="text-2xl font-semibold">No environments yet</h2>
        <p class="max-w-sm text-muted-foreground">
          Add your first Shopware environment to start monitoring.
        </p>
        <Button as-child>
          <RouterLink :to="{ name: 'account.environments.new' }">
            <Plus class="mr-1 size-4" />
            Add Environment
          </RouterLink>
        </Button>
      </div>

      <div v-else class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <RouterLink
          v-for="env in environments"
          :key="env.id"
          :to="{
            name: 'account.environments.detail',
            params: {
              organizationId: env.organizationId,
              environmentId: env.id,
            },
          }"
          class="group relative flex items-start gap-3 rounded-xl border bg-card p-4 shadow-sm transition-all duration-200 hover:border-primary/30 hover:shadow-md"
        >
          <!-- Favicon -->
          <div class="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted">
            <img
              v-if="env.favicon"
              :src="env.favicon"
              :alt="$t('dashboard.environmentLogo')"
              class="size-5 rounded"
            />
            <icon-fa6-solid:earth-americas v-else class="size-4 text-muted-foreground/50" />
          </div>

          <!-- Info -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span
                class="truncate font-semibold leading-tight group-hover:text-primary transition-colors"
                >{{ env.name }}</span
              >
              <StatusIcon :status="env.status" />
            </div>
            <div class="mt-0.5 truncate text-sm text-muted-foreground">
              {{ env.organizationName }}
            </div>
            <div class="mt-1.5 flex items-center gap-2">
              <Badge variant="secondary" class="font-mono text-xs">
                {{ env.shopwareVersion }}
              </Badge>
            </div>
          </div>
        </RouterLink>
      </div>
    </section>

    <!-- Changelog table -->
    <section v-if="changelogs.length > 0">
      <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold">
        <icon-fa6-solid:file-waveform class="size-4 text-muted-foreground" />
        {{ $t("dashboard.lastChanges") }}
      </h2>

      <Card class="overflow-hidden p-0">
        <DataTable
          :columns="[
            { key: 'environmentName', name: $t('dashboard.environment'), sortable: true },
            { key: 'extensions', name: $t('dashboard.changes'), sortable: true },
            { key: 'date', name: $t('common.date'), sortable: true, sortPath: 'date' },
          ]"
          :data="changelogs"
        >
          <template #cell-environmentName="{ row }">
            <RouterLink
              :to="{
                name: 'account.environments.detail',
                params: {
                  organizationId: row.environmentOrganizationId,
                  environmentId: row.environmentId,
                },
              }"
              class="font-medium text-primary hover:underline"
            >
              {{ row.environmentName }}
            </RouterLink>
          </template>

          <template #cell-extensions="{ row }">
            {{ sumChanges(row) }}
          </template>

          <template #cell-date="{ row }">
            <span class="tabular-nums text-muted-foreground">{{ formatDateTime(row.date) }}</span>
          </template>
        </DataTable>
      </Card>
    </section>
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
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import {
  useAccountEnvironments,
  fetchAccountEnvironments,
} from "@/composables/useAccountEnvironments";
import { useSession } from "@/composables/useSession";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import DataTable from "@/components/layout/DataTable.vue";
import StatusIcon from "@/components/StatusIcon.vue";
import { Button } from "@/components/ui/button";
import { FolderPlus, Plus } from "lucide-vue-next";

const { t } = useI18n();
const { activeOrganizationId } = useSession();

const changelogs = ref<components["schemas"]["AccountChangelog"][]>([]);

function loadDashboardData() {
  api.GET("/account/changelogs").then(({ data }) => {
    if (data) changelogs.value = data;
  });
}

loadDashboardData();

const { environments } = useAccountEnvironments();

watch(activeOrganizationId, () => {
  fetchAccountEnvironments();
  loadDashboardData();
});

const greenCount = computed(
  () => environments.value?.filter((e) => e.status === "green").length ?? 0,
);
const warnCount = computed(
  () => environments.value?.filter((e) => e.status === "yellow").length ?? 0,
);
const errorCount = computed(
  () => environments.value?.filter((e) => e.status === "red").length ?? 0,
);
</script>
