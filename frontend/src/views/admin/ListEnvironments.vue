<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.environmentManagement") }}</h1>
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
            :placeholder="$t('admin.searchEnvironments')"
            @input="debouncedSearch"
          />
        </div>

        <div>
          <select
            v-model="sortBy"
            class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
            @change="loadEnvironments"
          >
            <option value="createdAt">{{ $t("admin.sortByCreated") }}</option>
            <option value="name">{{ $t("admin.sortByName") }}</option>
            <option value="url">{{ $t("admin.sortByUrl") }}</option>
            <option value="status">{{ $t("admin.sortByStatus") }}</option>
            <option value="shopwareVersion">{{ $t("admin.sortByShopwareVersion") }}</option>
            <option value="lastScrapedAt">{{ $t("admin.sortByLastScraped") }}</option>
            <option value="organizationName">{{ $t("admin.sortByOrg") }}</option>
          </select>
        </div>

        <div>
          <select
            v-model="sortDirection"
            class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
            @change="loadEnvironments"
          >
            <option value="desc">{{ $t("common.descending") }}</option>
            <option value="asc">{{ $t("common.ascending") }}</option>
          </select>
        </div>
      </div>

      <DataTable
        v-if="!loading && environments.length > 0"
        :columns="tableColumns"
        :data="environments"
      >
        <template #cell-name="{ row }">
          <a :href="row.url" target="_blank" class="hover:underline">
            {{ row.name }}
          </a>
        </template>

        <template #cell-url="{ row }">
          <a :href="row.url" target="_blank" class="hover:underline">
            {{ row.url }}
          </a>
        </template>

        <template #cell-status="{ row }">
          <Badge
            :class="{
              'bg-emerald-500 text-white border-transparent': row.status === 'green',
              'bg-amber-500 text-white border-transparent': row.status === 'yellow',
              'bg-red-500 text-white border-transparent': row.status === 'red',
            }"
          >
            {{ row.status }}
          </Badge>
        </template>

        <template #cell-shopwareVersion="{ row }">
          <code class="rounded bg-muted px-1.5 py-0.5 text-sm font-mono">{{
            row.shopwareVersion
          }}</code>
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
            class="text-primary hover:underline"
          >
            {{ row.organizationName }}
          </router-link>
        </template>
      </DataTable>

      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
      >
        <icon-line-md:loading-twotone-loop class="size-6 animate-spin" />
        {{ $t("admin.loadingEnvironments") }}
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

type Environment = components["schemas"]["AccountEnvironment"];

const environments = ref<Environment[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const sortBy = ref<
  "createdAt" | "name" | "url" | "status" | "shopwareVersion" | "lastScrapedAt" | "organizationName"
>("createdAt");
const sortDirection = ref<"asc" | "desc">("desc");
const currentPage = ref(1);
const pageSize = ref(20);
const totalEnvironments = ref(0);

const totalPages = computed(() => Math.ceil(totalEnvironments.value / pageSize.value));

const { t } = useI18n();

const tableColumns = computed(() => [
  { key: "name" as const, name: t("common.name"), sortable: true, searchable: true },
  { key: "url" as const, name: t("common.url"), sortable: true, searchable: true },
  { key: "status" as const, name: t("common.status"), sortable: true },
  { key: "shopwareVersion" as const, name: t("admin.shopwareVersion"), sortable: true },
  { key: "lastScrapedAt" as const, name: t("admin.lastScraped"), sortable: true },
  { key: "organizationName" as const, name: t("settings.organization"), sortable: true },
]);

async function loadEnvironments() {
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

    const { data: response } = await api.GET("/admin/environments", {
      params: { query },
    });

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
