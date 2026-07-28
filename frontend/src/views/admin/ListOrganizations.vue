<template>
  <AdminListLayout :title="$t('admin.orgManagement')" :active-count="activeCount">
    <template #filters>
      <AdminFilterSidebar
        v-model="filters"
        v-model:search="searchQuery"
        :groups="filterGroups"
        searchable
        :search-placeholder="$t('admin.searchOrgs')"
        @update:search="debouncedSearch"
        @change="onFilterChange"
      />
    </template>

    <Alert v-if="error" variant="destructive" class="mb-4">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <div v-if="!loading && !error" class="mb-3 text-sm text-muted-foreground">
      {{ $t("admin.resultCount", { count: totalOrganizations }) }}
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingOrgs") }}
    </div>

    <!-- Org list -->
    <div v-else-if="organizations.length > 0" class="space-y-2">
      <RouterLink
        v-for="org in organizations"
        :key="org.id"
        :to="{ name: 'admin.organizations.detail', params: { id: org.id } }"
        class="group flex items-center gap-4 rounded-xl border bg-card px-4 py-3.5 transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-sm"
      >
        <div
          class="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted transition-transform group-hover:scale-105"
        >
          <img
            v-if="org.logo"
            :src="org.logo"
            :alt="org.name"
            class="size-7 rounded object-cover"
          />
          <icon-fa6-solid:building
            v-else
            class="size-4 text-muted-foreground group-hover:text-primary transition-colors"
          />
        </div>

        <div class="min-w-0 flex-1">
          <div class="truncate font-semibold transition-colors group-hover:text-primary">
            {{ org.name }}
          </div>
          <div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
            <span class="flex items-center gap-1 font-medium">
              <icon-fa6-solid:earth-americas class="size-3 text-emerald-500" />
              {{ org.environmentCount }}
            </span>
            <span class="flex items-center gap-1 font-medium">
              <icon-fa6-solid:users class="size-3 text-indigo-500" />
              {{ org.memberCount }}
            </span>
            <span class="hidden tabular-nums sm:inline">{{ formatDate(org.createdAt) }}</span>
          </div>
        </div>

        <icon-fa6-solid:chevron-right
          class="size-3.5 shrink-0 text-muted-foreground opacity-30 transition-all duration-150 group-hover:opacity-100 group-hover:translate-x-0.5 group-hover:text-primary"
        />
      </RouterLink>
    </div>

    <!-- Empty -->
    <div
      v-else
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:building class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">{{ $t("admin.noOrganizationsFound") }}</h3>
      <p v-if="searchQuery" class="text-sm text-muted-foreground">
        {{ $t("admin.noOrganizationsMatch", { query: searchQuery }) }}
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
import { Button } from "@/components/ui/button";
import {
  adminGetOrganizations,
  type AdminGetOrganizationsData,
  type AccountOrganization as Organization,
} from "@/api/generated";
import { formatDate } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";
import { useAdminListQuery } from "@/composables/useAdminListQuery";

const {
  search: searchQuery,
  page: currentPage,
  filters,
  syncToUrl,
} = useAdminListQuery({
  search: "",
  page: 1,
  filters: {
    sortBy: "createdAt",
  },
});

const organizations = ref<Organization[]>([]);
const loading = ref(true);
const error = ref("");
const sortDirection = ref<"asc" | "desc">("desc");
const pageSize = ref(20);
const totalOrganizations = ref(0);

const totalPages = computed(() => Math.ceil(totalOrganizations.value / pageSize.value));
const { t } = useI18n();

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "sortBy",
    label: t("admin.filterSortBy"),
    defaultValue: "createdAt",
    options: [
      { label: t("admin.sortByCreated"), value: "createdAt" },
      { label: t("admin.sortByName"), value: "name" },
      { label: t("admin.sortByMembers"), value: "memberCount" },
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

async function loadOrganizations() {
  loading.value = true;
  error.value = "";

  try {
    const sortBy = filters.value.sortBy as "createdAt" | "name" | "shopCount" | "memberCount";

    const query: AdminGetOrganizationsData["query"] = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      sortBy,
      sortDirection: sortDirection.value,
    };

    if (searchQuery.value) {
      query.searchField = "name";
      query.searchOperator = "contains";
      query.searchValue = searchQuery.value;
    }

    const { data: response } = await adminGetOrganizations({ query });

    if (response) {
      organizations.value = response.organizations;
      totalOrganizations.value = response.total;
    }
  } catch (err) {
    error.value = `Failed to load organizations: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

function onFilterChange() {
  currentPage.value = 1;
  syncToUrl();
  loadOrganizations();
}

function changePage(page: number) {
  currentPage.value = page;
  syncToUrl();
  loadOrganizations();
}

let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    syncToUrl();
    loadOrganizations();
  }, 300);
}

onMounted(() => {
  loadOrganizations();
});
</script>
