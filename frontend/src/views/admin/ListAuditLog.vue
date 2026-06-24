<template>
  <AdminListLayout :title="$t('admin.auditLog')" :active-count="activeCount">
    <template #filters>
      <AdminFilterSidebar v-model="filters" :groups="filterGroups" @change="onFilterChange" />
    </template>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingAuditLog") }}
    </div>

    <!-- Entries list -->
    <div v-else-if="entries.length > 0" class="space-y-2">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="flex flex-wrap items-center gap-3 rounded-xl border bg-card px-4 py-3"
      >
        <Badge
          :class="
            entry.action.startsWith('admin.')
              ? 'bg-destructive/10 text-destructive border-destructive/20 text-xs'
              : 'bg-warning/10 text-warning border-warning/20 text-xs'
          "
        >
          {{ actionLabel(entry.action) }}
        </Badge>

        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
            <RouterLink
              v-if="entry.actorUserId"
              :to="{ name: 'admin.users.detail', params: { id: entry.actorUserId } }"
              class="font-medium hover:text-primary"
            >
              {{ entry.actorName ?? entry.actorEmail }}
            </RouterLink>
            <span v-else class="font-medium text-muted-foreground">{{ $t("admin.system") }}</span>

            <template v-if="entry.targetUserId || entry.targetName">
              <icon-fa6-solid:arrow-right class="size-2.5 text-muted-foreground" />
              <RouterLink
                v-if="entry.targetUserId"
                :to="{ name: 'admin.users.detail', params: { id: entry.targetUserId } }"
                class="hover:text-primary"
              >
                {{ entry.targetName ?? entry.targetEmail }}
              </RouterLink>
              <span v-else>{{ entry.targetName ?? entry.targetEmail }}</span>
            </template>
          </div>
          <div
            v-if="entry.detail || entry.ipAddress"
            class="mt-0.5 flex flex-wrap items-center gap-x-3 text-xs text-muted-foreground"
          >
            <span v-if="entry.detail" class="truncate">{{ entry.detail }}</span>
            <span v-if="entry.ipAddress" class="shrink-0 tabular-nums">{{ entry.ipAddress }}</span>
          </div>
        </div>

        <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
          {{ formatDateTime(entry.createdAt) }}
        </span>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState v-else :icon="IconClock" :title="$t('admin.noAuditLogFound')" />

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-center gap-4">
      <Button
        size="sm"
        variant="outline"
        :disabled="currentPage === 1"
        @click="changePage(currentPage - 1)"
      >
        {{ $t("common.previous") }}
      </Button>
      <span class="text-sm text-muted-foreground tabular-nums">{{
        $t("common.pageOf", { current: currentPage, total: totalPages })
      }}</span>
      <Button
        size="sm"
        variant="outline"
        :disabled="currentPage === totalPages"
        @click="changePage(currentPage + 1)"
      >
        {{ $t("common.next") }}
      </Button>
    </div>
  </AdminListLayout>
</template>

<script setup lang="ts">
import AdminFilterSidebar, { type FilterGroup } from "@/components/admin/AdminFilterSidebar.vue";
import AdminListLayout from "@/components/admin/AdminListLayout.vue";
import EmptyState from "@/components/EmptyState.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { api } from "@/helpers/api";
import { formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";
import IconClock from "~icons/fa6-solid/clock-rotate-left";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink } from "vue-router";

type AuditLogEntry = components["schemas"]["AdminAuditLogEntry"];

const { t } = useI18n();

const entries = ref<AuditLogEntry[]>([]);
const loading = ref(true);
const error = ref("");
const currentPage = ref(1);
const pageSize = 50;
const total = ref(0);

const filters = ref<Record<string, string>>({
  action: "",
});

const totalPages = computed(() => Math.ceil(total.value / pageSize));

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "action",
    label: t("admin.filterAction"),
    defaultValue: "",
    options: [
      { label: t("admin.allActions"), value: "" },
      { label: t("admin.actionSetRole"), value: "admin.set_role" },
      { label: t("admin.actionBanUser"), value: "admin.ban_user" },
      { label: t("admin.actionUnbanUser"), value: "admin.unban_user" },
      { label: t("admin.actionImpersonate"), value: "admin.impersonate" },
      { label: t("admin.actionPasswordChange"), value: "user.password_change" },
      { label: t("admin.actionPasswordReset"), value: "user.password_reset" },
    ],
  },
]);

const activeCount = computed(() => {
  let count = 0;
  for (const g of filterGroups.value) {
    if (filters.value[g.key] !== g.defaultValue) count++;
  }
  return count;
});

function actionLabel(a: string): string {
  switch (a) {
    case "admin.set_role":
      return t("admin.actionSetRole");
    case "admin.ban_user":
      return t("admin.actionBanUser");
    case "admin.unban_user":
      return t("admin.actionUnbanUser");
    case "admin.impersonate":
      return t("admin.actionImpersonate");
    case "user.password_change":
      return t("admin.actionPasswordChange");
    case "user.password_reset":
      return t("admin.actionPasswordReset");
    default:
      return a;
  }
}

async function loadEntries() {
  loading.value = true;
  error.value = "";

  const query: { limit: number; offset: number; action?: string } = {
    limit: pageSize,
    offset: (currentPage.value - 1) * pageSize,
  };
  if (filters.value.action) {
    query.action = filters.value.action;
  }

  const { data, error: respError } = await api.GET("/admin/audit-log", {
    params: { query },
  });

  if (respError || !data) {
    error.value = t("admin.failedLoadAuditLog", {
      error: (respError as { message?: string } | undefined)?.message ?? "",
    });
    loading.value = false;
    return;
  }

  entries.value = data.entries;
  total.value = data.total;
  loading.value = false;
}

function onFilterChange() {
  currentPage.value = 1;
  loadEntries();
}

function changePage(page: number) {
  currentPage.value = page;
  loadEntries();
}

onMounted(() => {
  loadEntries();
});
</script>
