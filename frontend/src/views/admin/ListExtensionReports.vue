<template>
  <AdminListLayout :title="$t('admin.extensionReports')" :active-count="activeCount">
    <template #filters>
      <AdminFilterSidebar v-model="filters" :groups="filterGroups" @change="onFilterChange" />
    </template>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingExtensionReports") }}
    </div>

    <!-- Reports list -->
    <div v-else-if="reports.length > 0" class="space-y-2">
      <div v-for="report in reports" :key="report.id" class="rounded-xl border bg-card px-4 py-3">
        <div class="flex flex-wrap items-center gap-3">
          <Badge :class="categoryClass(report.category)">
            {{ $t(`reportExtension.categories.${report.category}`) }}
          </Badge>
          <Badge :class="statusClass(report.status)">
            {{ $t(`admin.reportStatus.${report.status}`) }}
          </Badge>

          <div class="min-w-0 flex-1">
            <a
              :href="`https://store.shopware.com/search?search=${report.extensionName}`"
              target="_blank"
              rel="noopener noreferrer"
              class="font-medium hover:text-primary"
            >
              {{ report.extensionLabel ?? report.extensionName }}
            </a>
            <span class="ml-2 text-xs text-muted-foreground">{{ report.extensionName }}</span>
          </div>

          <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
            {{ formatDateTime(report.createdAt) }}
          </span>
        </div>

        <p class="mt-2 whitespace-pre-wrap text-sm text-muted-foreground">{{ report.comment }}</p>

        <div class="mt-2 flex flex-wrap items-center justify-between gap-2">
          <div class="text-xs text-muted-foreground">
            <template v-if="report.reporterName || report.reporterEmail">
              {{ $t("admin.reportedBy") }}: {{ report.reporterName ?? report.reporterEmail }}
            </template>
            <template v-else>{{ $t("admin.reporterRemoved") }}</template>
            <template v-if="report.reviewerName">
              · {{ $t("admin.reviewedBy") }}: {{ report.reviewerName }}
            </template>
          </div>

          <div v-if="report.status === 'pending'" class="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              :disabled="acting === report.id"
              @click="moderate(report.id, 'reject')"
            >
              <icon-fa6-solid:xmark class="mr-1.5 size-3" />
              {{ $t("admin.reject") }}
            </Button>
            <Button
              size="sm"
              :disabled="acting === report.id"
              @click="moderate(report.id, 'approve')"
            >
              <icon-fa6-solid:check class="mr-1.5 size-3" />
              {{ $t("admin.approve") }}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState v-else :icon="IconFlag" :title="$t('admin.noExtensionReportsFound')" />

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
import { useAlert } from "@/composables/useAlert";
import type { components } from "@/types/api";
import IconFlag from "~icons/fa6-solid/flag";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

type ExtensionReport = components["schemas"]["AdminExtensionReport"];
type ReportStatus = "pending" | "approved" | "rejected";

const { t } = useI18n();
const { success, error: alertError } = useAlert();

const reports = ref<ExtensionReport[]>([]);
const loading = ref(true);
const error = ref("");
const acting = ref<number | null>(null);
const currentPage = ref(1);
const pageSize = 50;
const total = ref(0);

const filters = ref<Record<string, string>>({
  status: "pending",
});

const totalPages = computed(() => Math.ceil(total.value / pageSize));

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "status",
    label: t("admin.reportFilterStatus"),
    defaultValue: "pending",
    options: [
      { label: t("admin.reportStatus.pending"), value: "pending" },
      { label: t("admin.reportStatus.approved"), value: "approved" },
      { label: t("admin.reportStatus.rejected"), value: "rejected" },
      { label: t("common.all"), value: "" },
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

function categoryClass(category: string): string {
  if (category === "security") {
    return "bg-destructive/10 text-destructive border-destructive/20 text-xs";
  }
  return "bg-warning/10 text-warning border-warning/20 text-xs";
}

function statusClass(status: string): string {
  switch (status) {
    case "approved":
      return "bg-success/10 text-success border-success/20 text-xs";
    case "rejected":
      return "bg-muted text-muted-foreground text-xs";
    default:
      return "bg-primary/10 text-primary border-primary/20 text-xs";
  }
}

async function loadReports() {
  loading.value = true;
  error.value = "";

  const query: { limit: number; offset: number; status?: ReportStatus } = {
    limit: pageSize,
    offset: (currentPage.value - 1) * pageSize,
  };
  if (filters.value.status) {
    query.status = filters.value.status as ReportStatus;
  }

  const { data, error: respError } = await api.GET("/admin/extension-reports", {
    params: { query },
  });

  if (respError || !data) {
    error.value = t("admin.failedLoadExtensionReports", {
      error: (respError as { message?: string } | undefined)?.message ?? "",
    });
    loading.value = false;
    return;
  }

  reports.value = data.reports;
  total.value = data.total;
  loading.value = false;
}

async function moderate(reportId: number, action: "approve" | "reject") {
  acting.value = reportId;
  try {
    const path =
      action === "approve"
        ? "/admin/extension-reports/{reportId}/approve"
        : "/admin/extension-reports/{reportId}/reject";
    const { error: respError } = await api.POST(path, {
      params: { path: { reportId } },
    });
    if (respError) {
      throw new Error((respError as { message?: string })?.message ?? "request failed");
    }
    success(action === "approve" ? t("admin.reportApproved") : t("admin.reportRejected"));
    await loadReports();
  } catch (e) {
    alertError(e instanceof Error ? e.message : String(e));
  } finally {
    acting.value = null;
  }
}

function onFilterChange() {
  currentPage.value = 1;
  loadReports();
}

function changePage(page: number) {
  currentPage.value = page;
  loadReports();
}

onMounted(() => {
  loadReports();
});
</script>
