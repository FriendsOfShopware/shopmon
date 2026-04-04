<template>
  <HeaderContainer :title="$t('admin.shopManagement')" />

  <Panel>
    <Alert v-if="error" type="danger">
      {{ error }}
    </Alert>

    <div class="shops-filter">
      <div class="search-container">
        <BaseInput
          v-model="searchQuery"
          :placeholder="$t('admin.searchShops')"
          @input="debouncedSearch"
        />
      </div>

      <div class="filter-container">
        <BaseSelect v-model="sortBy" @change="loadShops">
          <option value="createdAt">{{ $t("admin.sortByCreated") }}</option>
          <option value="name">{{ $t("admin.sortByName") }}</option>
          <option value="url">{{ $t("admin.sortByUrl") }}</option>
          <option value="status">{{ $t("admin.sortByStatus") }}</option>
          <option value="shopwareVersion">{{ $t("admin.sortByShopwareVersion") }}</option>
          <option value="lastScrapedAt">{{ $t("admin.sortByLastScraped") }}</option>
          <option value="organizationName">{{ $t("admin.sortByOrg") }}</option>
        </BaseSelect>
      </div>

      <div class="filter-container">
        <BaseSelect v-model="sortDirection" @change="loadShops">
          <option value="desc">{{ $t("common.descending") }}</option>
          <option value="asc">{{ $t("common.ascending") }}</option>
        </BaseSelect>
      </div>
    </div>

    <DataTable v-if="!loading && shops.length > 0" :columns="tableColumns" :data="shops">
      <template #cell-name="{ row }">
        <a :href="row.url" target="_blank" class="shop-link">
          {{ row.name }}
        </a>
      </template>

      <template #cell-url="{ row }">
        <a :href="row.url" target="_blank" class="shop-link">
          {{ row.url }}
        </a>
      </template>

      <template #cell-status="{ row }">
        <span class="badge" :class="`badge-${row.status}`">
          {{ row.status }}
        </span>
      </template>

      <template #cell-shopwareVersion="{ row }">
        <code class="version">{{ row.shopwareVersion }}</code>
      </template>

      <template #cell-lastScrapedAt="{ row }">
        {{ row.lastScrapedAt ? formatDate(row.lastScrapedAt) : "-" }}
      </template>

      <template #cell-organizationName="{ row }">
        <router-link
          :to="{
            name: 'account.organizations.detail',
            params: { organizationId: row.organizationId },
          }"
        >
          {{ row.organizationName }}
        </router-link>
      </template>
    </DataTable>

    <div v-if="loading" class="loading-container">
      <icon-line-md:loading-twotone-loop class="loading-icon" />
      {{ $t("admin.loadingShops") }}
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button class="btn btn-sm" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">
        {{ $t("common.previous") }}
      </button>
      <span class="page-info">{{
        $t("common.pageOf", { current: currentPage, total: totalPages })
      }}</span>
      <button
        class="btn btn-sm"
        :disabled="currentPage === totalPages"
        @click="changePage(currentPage + 1)"
      >
        {{ $t("common.next") }}
      </button>
    </div>
  </Panel>
</template>

<script setup lang="ts">
import Alert from "@/components/layout/Alert.vue";
import DataTable from "@/components/layout/DataTable.vue";
import HeaderContainer from "@/components/layout/HeaderContainer.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { formatDate } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";

type Shop = components["schemas"]["AccountShop"];

const shops = ref<Shop[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const sortBy = ref<
  "createdAt" | "name" | "url" | "status" | "shopwareVersion" | "lastScrapedAt" | "organizationName"
>("createdAt");
const sortDirection = ref<"asc" | "desc">("desc");
const currentPage = ref(1);
const pageSize = ref(20);
const totalShops = ref(0);

const totalPages = computed(() => Math.ceil(totalShops.value / pageSize.value));

const { t } = useI18n();

const tableColumns = computed(() => [
  { key: "name" as const, name: t("common.name"), sortable: true, searchable: true },
  { key: "url" as const, name: t("common.url"), sortable: true, searchable: true },
  { key: "status" as const, name: t("common.status"), sortable: true },
  { key: "shopwareVersion" as const, name: t("admin.shopwareVersion"), sortable: true },
  { key: "lastScrapedAt" as const, name: t("admin.lastScraped"), sortable: true },
  { key: "organizationName" as const, name: t("settings.organization"), sortable: true },
]);

async function loadShops() {
  loading.value = true;
  error.value = "";

  try {
    const query: {
      limit: number;
      offset: number;
      sortBy:
        | "createdAt"
        | "name"
        | "url"
        | "status"
        | "shopwareVersion"
        | "lastScrapedAt"
        | "organizationName";
      sortDirection: "asc" | "desc";
      searchField?: "name" | "url";
      searchOperator?: "contains";
      searchValue?: string;
    } = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      sortBy: sortBy.value,
      sortDirection: sortDirection.value,
    };

    if (searchQuery.value) {
      // Search by name if it contains letters, otherwise by URL
      const isNameSearch = /[a-zA-Z]/.test(searchQuery.value);
      query.searchField = isNameSearch ? "name" : "url";
      query.searchOperator = "contains";
      query.searchValue = searchQuery.value;
    }

    const { data: response } = await api.GET("/admin/shops", {
      params: { query },
    });

    if (response) {
      shops.value = response.shops;
      totalShops.value = response.total;
    }
  } catch (err) {
    error.value = `Failed to load shops: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

function changePage(page: number) {
  currentPage.value = page;
  loadShops();
}

let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    loadShops();
  }, 300);
}

onMounted(() => {
  loadShops();
});
</script>

<style scoped>
.shops-filter {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  align-items: center;
  flex-wrap: wrap;
}

.search-container {
  flex: 1;
  min-width: 250px;
}

.search-input {
  width: 100%;
  max-width: 400px;
}

.filter-container select {
  min-width: 150px;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: var(--text-color-muted);
  gap: 0.5rem;
}

.loading-icon {
  width: 24px;
  height: 24px;
}

.shop-link {
  color: var(--text-color-primary);
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }
}

.badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.badge-green {
  background-color: #10b981;
  color: white;
}

.badge-yellow {
  background-color: #f59e0b;
  color: white;
}

.badge-red {
  background-color: #ef4444;
  color: white;
}

.version {
  background-color: var(--bg-color-muted);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-family: monospace;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-top: 2rem;
}

.page-info {
  color: var(--text-color-muted);
}
</style>
