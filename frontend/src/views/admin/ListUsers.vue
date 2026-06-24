<template>
  <AdminListLayout :title="$t('admin.userManagement')" :active-count="activeCount">
    <template #filters>
      <AdminFilterSidebar
        v-model="filters"
        v-model:search="searchQuery"
        :groups="filterGroups"
        searchable
        :search-placeholder="$t('admin.searchUsers')"
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
      {{ $t("admin.loadingUsers") }}
    </div>

    <!-- User list -->
    <div v-else-if="users.length > 0" class="space-y-2">
      <RouterLink
        v-for="user in users"
        :key="user.id"
        :to="{ name: 'admin.users.detail', params: { id: user.id } }"
        class="group flex items-center gap-4 rounded-xl border bg-card px-4 py-3 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
      >
        <!-- Avatar -->
        <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10">
          <icon-fa6-solid:user class="size-3.5 text-primary" />
        </div>

        <!-- Info -->
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="truncate font-medium transition-colors group-hover:text-primary">{{
              user.name
            }}</span>
            <Badge
              v-if="user.role === 'admin'"
              class="bg-primary/10 text-primary border-primary/20 text-[10px]"
              >{{ $t("admin.roleAdmin") }}</Badge
            >
          </div>
          <div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
            <span class="truncate">{{ user.email }}</span>
            <span class="hidden shrink-0 tabular-nums sm:inline">{{
              formatDate(user.createdAt)
            }}</span>
          </div>
        </div>

        <!-- Status -->
        <div class="hidden shrink-0 sm:block">
          <Badge
            v-if="user.banned"
            class="bg-destructive/10 text-destructive border-destructive/20 text-xs"
          >
            {{ $t("admin.banned") }}
          </Badge>
          <Badge
            v-else-if="!user.emailVerified"
            class="bg-warning/10 text-warning border-warning/20 text-xs"
          >
            {{ $t("admin.unverified") }}
          </Badge>
          <Badge v-else class="bg-success/10 text-success border-success/20 text-xs">
            {{ $t("admin.active") }}
          </Badge>
        </div>

        <!-- Chevron -->
        <icon-fa6-solid:chevron-right
          class="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
        />
      </RouterLink>
    </div>

    <!-- Empty state -->
    <div
      v-else
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:users class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">{{ $t("admin.noUsersFound") }}</h3>
      <p v-if="searchQuery" class="text-sm text-muted-foreground">
        {{ $t("admin.noUsersMatch", { query: searchQuery }) }}
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { api } from "@/helpers/api";
import { formatDate } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";

type User = components["schemas"]["AdminUser"];

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const currentPage = ref(1);
const pageSize = ref(20);
const totalUsers = ref(0);

const filters = ref<Record<string, string>>({
  role: "",
  status: "",
  sortBy: "createdAt",
});

const totalPages = computed(() => Math.ceil(totalUsers.value / pageSize.value));
const { t } = useI18n();

const filterGroups = computed<FilterGroup[]>(() => [
  {
    key: "role",
    label: t("admin.filterRole"),
    defaultValue: "",
    options: [
      { label: t("admin.allRoles"), value: "" },
      { label: t("admin.roleAdmin"), value: "admin" },
      { label: t("admin.roleUser"), value: "user" },
    ],
  },
  {
    key: "status",
    label: t("admin.filterStatus"),
    defaultValue: "",
    options: [
      { label: t("admin.allStatuses"), value: "" },
      { label: t("admin.active"), value: "active" },
      { label: t("admin.banned"), value: "banned" },
      { label: t("admin.unverified"), value: "unverified" },
    ],
  },
  {
    key: "sortBy",
    label: t("admin.filterSortBy"),
    defaultValue: "createdAt",
    options: [
      { label: t("admin.sortByCreated"), value: "createdAt" },
      { label: t("admin.sortByName"), value: "name" },
      { label: t("admin.sortByEmail"), value: "email" },
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

async function loadUsers() {
  loading.value = true;
  error.value = "";

  try {
    const role = filters.value.role;
    const status = filters.value.status;
    const sortBy = filters.value.sortBy;

    const query: {
      limit: number;
      offset: number;
      search?: string;
      role?: "user" | "admin";
      status?: "active" | "banned" | "unverified";
      sortBy?: "createdAt" | "name" | "email";
      sortDirection?: "asc" | "desc";
    } = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
    };

    if (searchQuery.value) {
      query.search = searchQuery.value;
    }
    if (role === "admin" || role === "user") {
      query.role = role;
    }
    if (status === "active" || status === "banned" || status === "unverified") {
      query.status = status;
    }
    if (sortBy === "name" || sortBy === "email") {
      query.sortBy = sortBy;
      query.sortDirection = "asc";
    }

    const { data, error: respError } = await api.GET("/auth/admin/users", {
      params: { query },
    });

    if (!respError && data) {
      users.value = data.users ?? [];
      totalUsers.value = data.total ?? users.value.length;
    } else {
      error.value =
        (respError as unknown as { message?: string })?.message ?? "Failed to load users";
    }
  } catch (err) {
    error.value = `Failed to load users: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

function onFilterChange() {
  currentPage.value = 1;
  loadUsers();
}

function changePage(page: number) {
  currentPage.value = page;
  loadUsers();
}

let searchTimeout: ReturnType<typeof setTimeout>;
function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    currentPage.value = 1;
    loadUsers();
  }, 300);
}

onMounted(() => {
  loadUsers();
});
</script>
