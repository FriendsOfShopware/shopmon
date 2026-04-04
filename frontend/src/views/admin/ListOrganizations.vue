<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.orgManagement") }}</h1>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Filter bar -->
    <div class="flex flex-wrap items-center justify-between gap-3 max-sm:flex-col max-sm:w-full">
      <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
        <button
          v-for="s in sortOptions"
          :key="s.value"
          :class="[
            'rounded-md px-3 py-1 text-sm font-medium transition-colors',
            sortBy === s.value ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="sortBy = s.value; loadOrganizations()"
        >
          {{ s.label }}
        </button>
      </div>

      <div class="relative">
        <icon-fa6-solid:magnifying-glass class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          :placeholder="$t('admin.searchOrgs')"
          class="h-8 w-full pl-8 text-sm sm:w-56"
          @input="debouncedSearch"
        />
      </div>
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
        :to="{ name: 'account.organizations.detail', params: { organizationId: org.id } }"
        class="group flex items-center gap-4 rounded-xl border bg-card px-4 py-3 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
      >
        <div class="flex size-10 shrink-0 items-center justify-center rounded-lg border bg-muted">
          <img v-if="org.logo" :src="org.logo" :alt="org.name" class="size-7 rounded object-cover" />
          <icon-fa6-solid:building v-else class="size-4 text-muted-foreground" />
        </div>

        <div class="min-w-0 flex-1">
          <div class="truncate font-medium transition-colors group-hover:text-primary">{{ org.name }}</div>
          <div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
            <span class="flex items-center gap-1">
              <icon-fa6-solid:earth-americas class="size-2.5" />
              {{ org.environmentCount }}
            </span>
            <span class="flex items-center gap-1">
              <icon-fa6-solid:users class="size-2.5" />
              {{ org.memberCount }}
            </span>
            <span class="hidden tabular-nums sm:inline">{{ formatDate(org.createdAt) }}</span>
          </div>
        </div>

        <icon-fa6-solid:chevron-right class="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
      </RouterLink>
    </div>

    <!-- Empty -->
    <div v-else class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center">
      <icon-fa6-solid:building class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">No organizations found</h3>
      <p v-if="searchQuery" class="text-sm text-muted-foreground">No organizations match "{{ searchQuery }}"</p>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-center gap-4">
      <Button size="sm" variant="outline" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">
        {{ $t("common.previous") }}
      </Button>
      <span class="text-sm tabular-nums text-muted-foreground">{{ $t("common.pageOf", { current: currentPage, total: totalPages }) }}</span>
      <Button size="sm" variant="outline" :disabled="currentPage === totalPages" @click="changePage(currentPage + 1)">
        {{ $t("common.next") }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

const sortOptions = [
  { label: t("admin.sortByCreated"), value: "createdAt" as const },
  { label: t("admin.sortByName"), value: "name" as const },
  { label: t("admin.sortByMembers"), value: "memberCount" as const },
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

    const { data: response } = await api.GET("/admin/organizations", { params: { query } });

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
  searchTimeout = setTimeout(() => { currentPage.value = 1; loadOrganizations(); }, 300);
}

onMounted(() => { loadOrganizations(); });
</script>
