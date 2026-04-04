<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.orgManagement") }}</h1>
  </div>

  <Card>
    <CardContent>
      <Alert v-if="error" variant="destructive">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <div class="flex flex-wrap items-center gap-4 mb-8">
        <div class="flex-1 min-w-[250px]">
          <Input
            v-model="searchQuery"
            :placeholder="$t('admin.searchOrgs')"
            @input="debouncedSearch"
          />
        </div>

        <div>
          <select
            v-model="sortBy"
            class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
            @change="loadOrganizations"
          >
            <option value="createdAt">{{ $t("admin.sortByCreated") }}</option>
            <option value="name">{{ $t("admin.sortByName") }}</option>
            <option value="shopCount">{{ $t("admin.sortByShops") }}</option>
            <option value="memberCount">{{ $t("admin.sortByMembers") }}</option>
          </select>
        </div>

        <div>
          <select
            v-model="sortDirection"
            class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
            @change="loadOrganizations"
          >
            <option value="desc">{{ $t("common.descending") }}</option>
            <option value="asc">{{ $t("common.ascending") }}</option>
          </select>
        </div>
      </div>

      <DataTable
        v-if="!loading && organizations.length > 0"
        :columns="tableColumns"
        :data="organizations"
      >
        <template #cell-logo="{ row }">
          <div class="flex items-center justify-center size-10">
            <img
              v-if="row.logo"
              :src="row.logo"
              :alt="row.name"
              class="size-8 rounded object-cover"
            />
            <div
              v-else
              class="flex items-center justify-center size-8 rounded bg-muted text-muted-foreground"
            >
              <icon-fa6-solid:building class="size-4" />
            </div>
          </div>
        </template>

        <template #cell-name="{ row }">
          <router-link
            :to="{ name: 'account.organizations.detail', params: { organizationId: row.id } }"
            class="text-primary hover:underline"
          >
            {{ row.name }}
          </router-link>
        </template>

        <template #cell-environmentCount="{ row }">
          <Badge class="bg-sky-500 text-white border-transparent">
            {{ $t("admin.nShops", { count: row.environmentCount }) }}
          </Badge>
        </template>

        <template #cell-memberCount="{ row }">
          <Badge variant="secondary">
            {{ $t("admin.nMembers", { count: row.memberCount }) }}
          </Badge>
        </template>

        <template #cell-createdAt="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </DataTable>

      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
      >
        <icon-line-md:loading-twotone-loop class="size-6 animate-spin" />
        {{ $t("admin.loadingOrgs") }}
      </div>

      <div v-if="totalPages > 1" class="flex items-center justify-center gap-4 mt-8">
        <Button
          size="sm"
          variant="outline"
          :disabled="currentPage === 1"
          @click="changePage(currentPage - 1)"
        >
          {{ $t("common.previous") }}
        </Button>
        <span class="text-sm text-muted-foreground">{{
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
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import DataTable from "@/components/layout/DataTable.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { formatDate } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";
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

const { t } = useI18n();

const tableColumns = computed(() => [
  { key: "logo" as const, name: t("admin.logo") },
  { key: "name" as const, name: t("common.name"), sortable: true, searchable: true },
  { key: "environmentCount" as const, name: t("common.environments"), sortable: true },
  { key: "memberCount" as const, name: t("common.members"), sortable: true },
  { key: "createdAt" as const, name: t("admin.created"), sortable: true },
]);

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
