<template>
  <AdminListLayout :title="$t('admin.advisories')" :active-count="activeFilterCount">
    <template #filters>
      <div class="flex h-full flex-col">
        <div class="px-4 pb-4">
          <Label class="mb-1.5 block text-xs font-medium text-muted-foreground">
            {{ $t("admin.filterPackage") }}
          </Label>
          <Select v-model="packageFilter" @update:model-value="onFilterChange">
            <SelectTrigger class="h-9 w-full text-sm">
              <SelectValue :placeholder="$t('advisories.allPackages')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{{ $t("advisories.allPackages") }}</SelectItem>
              <SelectItem v-for="pkg in packages" :key="pkg" :value="pkg">{{ pkg }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Separator />
        <AdminFilterSidebar
          v-model="filters"
          v-model:search="searchQuery"
          :groups="filterGroups"
          searchable
          :search-placeholder="$t('advisories.searchPlaceholder')"
          class="min-h-0 flex-1"
          @update:search="debouncedSearch"
          @change="onFilterChange"
        />
      </div>
    </template>

    <div class="mb-4 flex flex-wrap items-center gap-2">
      <Button size="sm" :disabled="syncing" @click="syncNow">
        <icon-line-md:loading-twotone-loop v-if="syncing" class="mr-1.5 size-3.5" />
        <icon-fa6-solid:rotate v-else class="mr-1.5 size-3.5" />
        {{ $t("admin.syncAdvisories") }}
      </Button>
      <span v-if="syncMessage" class="text-sm text-muted-foreground">{{ syncMessage }}</span>
    </div>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <div v-if="loading" class="space-y-2" aria-hidden="true">
      <div v-for="i in 8" :key="i" class="flex items-center gap-4 rounded-xl border bg-card p-4">
        <Skeleton class="size-9 shrink-0 rounded-lg" />
        <div class="flex-1 space-y-2">
          <Skeleton class="h-4 w-2/3" />
          <Skeleton class="h-3 w-1/3" />
        </div>
        <Skeleton class="hidden h-5 w-16 rounded-full sm:block" />
      </div>
    </div>

    <div v-else-if="advisories.length > 0" class="space-y-2">
      <RouterLink
        v-for="adv in advisories"
        :key="adv.advisoryId"
        :to="{ name: 'admin.advisories.detail', params: { id: adv.advisoryId } }"
        class="group flex items-center gap-3 rounded-xl border bg-card p-4 transition-all hover:border-primary/30 hover:shadow-sm sm:gap-4"
      >
        <div
          class="flex size-9 shrink-0 items-center justify-center rounded-lg"
          :class="severityTileClass(adv.effectiveSeverity)"
        >
          <icon-fa6-solid:shield-halved class="size-4" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
            <span class="font-medium transition-colors group-hover:text-primary">{{
              adv.title
            }}</span>
            <Badge v-if="!adv.isVisible" variant="outline" class="text-[10px]">{{
              $t("admin.hidden")
            }}</Badge>
          </div>
          <div
            class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground"
          >
            <span class="flex flex-wrap gap-1">
              <Badge
                v-for="pkg in adv.packages"
                :key="pkg.packageName"
                variant="secondary"
                class="font-mono text-[10px]"
                >{{ pkg.packageName }}</Badge
              >
            </span>
            <span v-if="adv.cve" class="font-mono">{{ adv.cve }}</span>
            <span v-if="adv.reportedAt" class="inline-flex items-center gap-1">
              <icon-fa6-regular:calendar class="size-3" />
              {{ formatDate(adv.reportedAt) }}
            </span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-3">
          <SeverityBadge :severity="adv.effectiveSeverity" class="text-[10px]" />
          <icon-fa6-solid:chevron-right
            class="hidden size-3 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 sm:block"
          />
        </div>
      </RouterLink>
    </div>

    <EmptyState
      v-else
      :icon="IconShield"
      :title="$t('advisories.empty')"
      :description="$t('advisories.emptyHint')"
    />

    <div v-if="totalPages > 1" class="mt-4 flex items-center justify-center gap-4">
      <Button size="sm" variant="outline" :disabled="page <= 1" @click="changePage(page - 1)">
        {{ $t("common.previous") }}
      </Button>
      <span class="text-sm tabular-nums text-muted-foreground">{{
        $t("common.pageOf", { current: page, total: totalPages })
      }}</span>
      <Button
        size="sm"
        variant="outline"
        :disabled="page >= totalPages"
        @click="changePage(page + 1)"
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
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  type AdminAdvisory,
  type AdminListAdvisoriesData,
  adminListAdvisories,
  adminSyncAdvisories,
  listAdvisoryPackages,
} from "@/api/generated";
import { useAdminListQuery } from "@/composables/useAdminListQuery";
import { severityTileClass } from "@/helpers/advisory";
import { formatDate } from "@/helpers/formatter";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import IconShield from "~icons/fa6-solid/shield-halved";

const { t } = useI18n();
const advisories = ref<AdminAdvisory[]>([]);
const packages = ref<string[]>([]);
const total = ref(0);
const loading = ref(true);
const error = ref("");
const perPage = 25;
const syncing = ref(false);
const syncMessage = ref("");

const severities = ["critical", "high", "medium", "low", "none"] as const;

// Shared with the other admin lists so search, filters, and pagination survive
// a reload and can be reproduced from a shared URL. revision fires when the
// route query changes from outside (browser back/forward), which local refs
// would miss.
const {
  search: searchQuery,
  page,
  filters,
  revision,
  syncToUrl,
} = useAdminListQuery({
  search: "",
  page: 1,
  filters: {
    severity: "",
    visible: "",
    package: "all",
  },
});

const packageFilter = computed({
  get: () => filters.value.package ?? "all",
  set: (value: string) => {
    filters.value = { ...filters.value, package: value };
  },
});

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "severity",
    label: t("admin.filterSeverity"),
    defaultValue: "",
    options: [
      { label: t("advisories.allSeverities"), value: "" },
      ...severities.map((s) => ({ label: t(`advisories.severity.${s}`), value: s })),
    ],
  },
  {
    key: "visible",
    label: t("admin.visibility"),
    defaultValue: "",
    options: [
      { label: t("admin.allVisibility"), value: "" },
      { label: t("admin.visibleOnly"), value: "visible" },
      { label: t("admin.hiddenOnly"), value: "hidden" },
    ],
  },
]);

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / perPage)));

// Badge on the mobile filter trigger: number of active filters, not results.
const activeFilterCount = computed(
  () =>
    (searchQuery.value.trim() !== "" ? 1 : 0) +
    (packageFilter.value !== "all" ? 1 : 0) +
    (filters.value.severity !== "" ? 1 : 0) +
    (filters.value.visible !== "" ? 1 : 0),
);

function onFilterChange() {
  clearTimeout(searchTimeout);
  page.value = 1;
  syncToUrl();
  reload();
}

function changePage(next: number) {
  page.value = next;
  syncToUrl();
  reload();
}

let searchTimeout: ReturnType<typeof setTimeout> | undefined;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    page.value = 1;
    syncToUrl();
    reload();
  }, 300);
}

// Filter changes, pagination, debounced search, and browser history can each
// start a reload while an earlier one is still in flight. Responses are not
// guaranteed to arrive in request order, so every result is checked against the
// latest request id and a superseded one is discarded — otherwise a slow older
// response would overwrite newer results, leaving the table disagreeing with
// the URL and the controls.
let requestId = 0;

async function reload() {
  const currentRequest = ++requestId;
  loading.value = true;
  error.value = "";
  try {
    const query: AdminListAdvisoriesData["query"] = {
      limit: perPage,
      offset: (page.value - 1) * perPage,
    };
    if (searchQuery.value.trim()) query.q = searchQuery.value.trim();
    if (packageFilter.value !== "all") query.package = packageFilter.value;
    const severity = filters.value.severity;
    if ((severities as readonly string[]).includes(severity)) {
      query.severity = severity as (typeof severities)[number];
    }
    if (filters.value.visible === "visible") query.visible = true;
    if (filters.value.visible === "hidden") query.visible = false;

    const { data, error: apiError } = await adminListAdvisories({ query });
    if (currentRequest !== requestId) return;
    if (apiError || !data) {
      error.value = t("advisories.loadFailed");
      return;
    }
    advisories.value = data.advisories;
    total.value = data.total;
  } finally {
    // Only the newest request owns the spinner: an older one clearing it would
    // signal "done" while the current request is still running.
    if (currentRequest === requestId) {
      loading.value = false;
    }
  }
}

async function syncNow() {
  syncing.value = true;
  syncMessage.value = "";
  try {
    const { error: apiError } = await adminSyncAdvisories();
    if (apiError) {
      syncMessage.value = t("admin.syncFailed");
      return;
    }
    syncMessage.value = t("admin.syncEnqueued");
  } finally {
    syncing.value = false;
  }
}

// Browser back/forward changes the query without remounting, so the list must
// follow the URL rather than keep whatever it last rendered.
watch(revision, () => {
  reload();
});

onMounted(async () => {
  const { data } = await listAdvisoryPackages();
  packages.value = data ?? [];
  await reload();
});
</script>
