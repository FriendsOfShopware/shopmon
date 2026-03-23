<template>
  <HeaderContainer title="Organization Management" />

  <Panel>
    <Alert v-if="error" type="danger">
      {{ error }}
    </Alert>

    <div class="organizations-filter">
      <div class="search-container">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search organizations by name..."
          class="field search-input"
          @input="debouncedSearch"
        />
      </div>

      <div class="filter-container">
        <select v-model="sortBy" class="field" @change="loadOrganizations">
          <option value="createdAt">Sort by Created</option>
          <option value="name">Sort by Name</option>
          <option value="shopCount">Sort by Shops</option>
          <option value="memberCount">Sort by Members</option>
        </select>
      </div>

      <div class="filter-container">
        <select v-model="sortDirection" class="field" @change="loadOrganizations">
          <option value="desc">Descending</option>
          <option value="asc">Ascending</option>
        </select>
      </div>
    </div>

    <DataTable
      v-if="!loading && organizations.length > 0"
      :columns="tableColumns"
      :data="organizations"
    >
      <template #cell-logo="{ row }">
        <div class="logo-container">
          <img v-if="row.logo" :src="row.logo" :alt="row.name" class="organization-logo" />
          <div v-else class="logo-placeholder">
            <icon-fa6-solid:building class="icon" />
          </div>
        </div>
      </template>

      <template #cell-name="{ row }">
        <router-link
          :to="{ name: 'account.organizations.detail', params: { organizationId: row.id } }"
        >
          {{ row.name }}
        </router-link>
      </template>

      <template #cell-shopCount="{ row }">
        <span class="badge badge-info"> {{ row.shopCount }} shops </span>
      </template>

      <template #cell-memberCount="{ row }">
        <span class="badge badge-secondary"> {{ row.memberCount }} members </span>
      </template>

      <template #cell-createdAt="{ row }">
        {{ formatDate(row.createdAt) }}
      </template>
    </DataTable>

    <div v-if="loading" class="loading-container">
      <icon-line-md:loading-twotone-loop class="loading-icon" />
      Loading organizations...
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button class="btn btn-sm" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">
        Previous
      </button>
      <span class="page-info">Page {{ currentPage }} of {{ totalPages }}</span>
      <button
        class="btn btn-sm"
        :disabled="currentPage === totalPages"
        @click="changePage(currentPage + 1)"
      >
        Next
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
import { computed, onMounted, ref } from "vue";

type Organization = components["schemas"]["AccountOrganization"];

const organizations = ref<Organization[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const sortBy = ref<"createdAt" | "name" | "shopCount" | "memberCount">("createdAt");
const sortDirection = ref<"asc" | "desc">("desc");
const currentPage = ref(1);
const pageSize = ref(20);
const totalOrganizations = ref(0);

const totalPages = computed(() => Math.ceil(totalOrganizations.value / pageSize.value));

const tableColumns = [
  { key: "logo" as const, name: "Logo" },
  { key: "name" as const, name: "Name", sortable: true, searchable: true },
  { key: "shopCount" as const, name: "Shops", sortable: true },
  { key: "memberCount" as const, name: "Members", sortable: true },
  { key: "createdAt" as const, name: "Created", sortable: true },
];

async function loadOrganizations() {
  loading.value = true;
  error.value = "";

  try {
    const query: {
      limit: number;
      offset: number;
      sortBy: "createdAt" | "name" | "shopCount" | "memberCount";
      sortDirection: "asc" | "desc";
      searchField?: "name";
      searchOperator?: "contains";
      searchValue?: string;
    } = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      sortBy: sortBy.value,
      sortDirection: sortDirection.value,
    };

    if (searchQuery.value) {
      query.searchField = "name";
      query.searchOperator = "contains";
      query.searchValue = searchQuery.value;
    }

    const { data: response } = await api.GET("/admin/organizations", {
      params: { query },
    });

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

function changePage(page: number) {
  currentPage.value = page;
  loadOrganizations();
}

let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    loadOrganizations();
  }, 300);
}

onMounted(() => {
  loadOrganizations();
});
</script>

<style scoped>
.organizations-filter {
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

.actions-group {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.logo-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
}

.organization-logo {
  width: 32px;
  height: 32px;
  object-fit: cover;
  border-radius: 4px;
}

.logo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background-color: var(--bg-color-muted);
  border-radius: 4px;
  color: var(--text-color-muted);
}

.logo-placeholder .icon {
  width: 16px;
  height: 16px;
}

.badge-info {
  background-color: #0ea5e9;
  color: white;
}

.badge-secondary {
  background-color: #6b7280;
  color: white;
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
