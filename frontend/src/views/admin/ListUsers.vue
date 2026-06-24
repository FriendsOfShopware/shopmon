<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.userManagement") }}</h1>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Filter bar -->
    <div class="flex flex-wrap items-center justify-between gap-3 max-sm:flex-col max-sm:w-full">
      <div class="flex flex-wrap items-center gap-2">
        <!-- Role filter -->
        <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
          <button
            v-for="f in roleFilters"
            :key="f.value"
            :class="[
              'rounded-md px-3 py-1 text-sm font-medium transition-colors',
              roleFilter === f.value
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
            @click="
              roleFilter = f.value;
              currentPage = 1;
              loadUsers();
            "
          >
            {{ f.label }}
          </button>
        </div>

        <!-- Status filter -->
        <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
          <button
            v-for="f in statusFilters"
            :key="f.value"
            :class="[
              'rounded-md px-3 py-1 text-sm font-medium transition-colors',
              statusFilter === f.value
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
            @click="
              statusFilter = f.value;
              currentPage = 1;
              loadUsers();
            "
          >
            {{ f.label }}
          </button>
        </div>

        <!-- Sort -->
        <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
          <button
            v-for="s in sortOptions"
            :key="s.value"
            :class="[
              'rounded-md px-3 py-1 text-sm font-medium transition-colors',
              sortBy === s.value
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
            @click="
              sortBy = s.value;
              currentPage = 1;
              loadUsers();
            "
          >
            {{ s.label }}
          </button>
        </div>
      </div>

      <div class="relative">
        <icon-fa6-solid:magnifying-glass
          class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          v-model="searchQuery"
          :placeholder="$t('admin.searchUsers')"
          class="h-8 w-full pl-8 text-sm sm:w-56"
          @input="debouncedSearch"
        />
      </div>
    </div>

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
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/helpers/api";
import { formatDate } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";

type User = components["schemas"]["AdminUser"];

type StatusFilter = "" | "active" | "banned" | "unverified";
type SortBy = "createdAt" | "name" | "email";

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const roleFilter = ref("");
const statusFilter = ref<StatusFilter>("");
const sortBy = ref<SortBy>("createdAt");
const currentPage = ref(1);
const pageSize = ref(20);
const totalUsers = ref(0);

const totalPages = computed(() => Math.ceil(totalUsers.value / pageSize.value));
const { t } = useI18n();

const roleFilters = [
  { label: t("admin.allRoles"), value: "" },
  { label: t("admin.roleAdmin"), value: "admin" },
  { label: t("admin.roleUser"), value: "user" },
];

const statusFilters: { label: string; value: StatusFilter }[] = [
  { label: t("admin.allStatuses"), value: "" },
  { label: t("admin.active"), value: "active" },
  { label: t("admin.banned"), value: "banned" },
  { label: t("admin.unverified"), value: "unverified" },
];

const sortOptions: { label: string; value: SortBy }[] = [
  { label: t("admin.sortByCreated"), value: "createdAt" },
  { label: t("admin.sortByName"), value: "name" },
  { label: t("admin.sortByEmail"), value: "email" },
];

async function loadUsers() {
  loading.value = true;
  error.value = "";

  try {
    const query: {
      limit: number;
      offset: number;
      search?: string;
      role?: "user" | "admin";
      status?: "active" | "banned" | "unverified";
      sortBy?: SortBy;
      sortDirection?: "asc" | "desc";
    } = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
    };

    if (searchQuery.value) {
      query.search = searchQuery.value;
    }
    if (roleFilter.value === "admin" || roleFilter.value === "user") {
      query.role = roleFilter.value;
    }
    if (statusFilter.value) {
      query.status = statusFilter.value;
    }
    if (sortBy.value !== "createdAt") {
      query.sortBy = sortBy.value;
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
