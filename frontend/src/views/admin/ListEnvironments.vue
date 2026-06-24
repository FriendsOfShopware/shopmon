<template>
  <AdminListLayout :title="$t('admin.environmentManagement')" :active-count="activeCount">
    <template #filters>
      <AdminFilterSidebar
        v-model="filters"
        v-model:search="searchQuery"
        :groups="filterGroups"
        searchable
        :search-placeholder="$t('admin.searchEnvironments')"
        @update:search="debouncedSearch"
        @change="onFilterChange"
      />
    </template>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingEnvironments") }}
    </div>

    <!-- Environment list -->
    <div v-else-if="environments.length > 0" class="space-y-2">
      <div
        v-for="env in environments"
        :key="env.id"
        class="flex items-center gap-4 rounded-xl border bg-card px-4 py-3"
      >
        <StatusIcon :status="env.status" class="shrink-0" />

        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <RouterLink
              :to="{ name: 'admin.environments.detail', params: { id: env.id } }"
              class="truncate font-medium hover:text-primary transition-colors"
            >
              {{ env.name }}
            </RouterLink>
            <Badge variant="secondary" class="font-mono text-[10px]">{{
              env.shopwareVersion
            }}</Badge>
          </div>
          <div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
            <RouterLink
              :to="{ name: 'admin.organizations.detail', params: { id: env.organizationId } }"
              class="hover:text-primary transition-colors"
            >
              {{ env.organizationName }}
            </RouterLink>
            <span v-if="env.shopName" class="flex items-center gap-1">
              <icon-fa6-solid:store class="size-2.5" />
              {{ env.shopName }}
            </span>
            <span v-if="env.lastScrapedAt" class="hidden tabular-nums sm:inline">{{
              formatDate(env.lastScrapedAt)
            }}</span>
          </div>
        </div>

        <a
          :href="env.url"
          target="_blank"
          class="shrink-0 text-muted-foreground hover:text-foreground"
        >
          <icon-fa6-solid:arrow-up-right-from-square class="size-3" />
        </a>
      </div>
    </div>

    <!-- Empty -->
    <div
      v-else
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:earth-americas class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">{{ $t("admin.noEnvironmentsFound") }}</h3>
      <p v-if="searchQuery" class="text-sm text-muted-foreground">
        {{ $t("admin.noEnvironmentsMatch", { query: searchQuery }) }}
      </p>
    </div>

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
      <span class="text-sm tabular-nums text-muted-foreground">{{
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import StatusIcon from "@/components/StatusIcon.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { formatDate } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";

type Environment = components["schemas"]["AccountEnvironment"];
type SortBy =
  | "createdAt"
  | "name"
  | "url"
  | "status"
  | "shopwareVersion"
  | "lastScrapedAt"
  | "organizationName";

const environments = ref<Environment[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const sortDirection = ref<"asc" | "desc">("desc");
const currentPage = ref(1);
const pageSize = ref(20);
const totalEnvironments = ref(0);

const filters = ref<Record<string, string>>({
  sortBy: "createdAt",
});

const totalPages = computed(() => Math.ceil(totalEnvironments.value / pageSize.value));
const { t } = useI18n();

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "sortBy",
    label: t("admin.filterSortBy"),
    defaultValue: "createdAt",
    options: [
      { label: t("admin.sortByCreated"), value: "createdAt" },
      { label: t("admin.sortByName"), value: "name" },
      { label: t("admin.sortByStatus"), value: "status" },
      { label: t("admin.sortByOrg"), value: "organizationName" },
    ],
  },
]);

const activeCount = computed(() => {
  let count = 0;
  for (const g of filterGroups.value) {
    if (filters.value[g.key] !== g.defaultValue) count++;
  }
  if (searchQuery.value) count++;
  return count;
});

async function loadEnvironments() {
  loading.value = true;
  error.value = "";

  try {
    const query: {
      limit: number;
      offset: number;
      sortBy: SortBy;
      sortDirection: "asc" | "desc";
      searchField?: "name" | "url";
      searchOperator?: "contains";
      searchValue?: string;
    } = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      sortBy: filters.value.sortBy as SortBy,
      sortDirection: sortDirection.value,
    };

    if (searchQuery.value) {
      const isNameSearch = /[a-zA-Z]/.test(searchQuery.value);
      query.searchField = isNameSearch ? "name" : "url";
      query.searchOperator = "contains";
      query.searchValue = searchQuery.value;
    }

    const { data: response } = await api.GET("/admin/environments", { params: { query } });

    if (response) {
      environments.value = response.environments;
      totalEnvironments.value = response.total;
    }
  } catch (err) {
    error.value = `Failed to load environments: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

function onFilterChange() {
  currentPage.value = 1;
  loadEnvironments();
}

function changePage(page: number) {
  currentPage.value = page;
  loadEnvironments();
}

let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    loadEnvironments();
  }, 300);
}

onMounted(() => {
  loadEnvironments();
});
</script>
