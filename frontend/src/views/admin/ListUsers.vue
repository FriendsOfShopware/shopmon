<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t('admin.userManagement') }}</h1>
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
            :placeholder="$t('admin.searchUsers')"
            @input="debouncedSearch"
          />
        </div>

        <div>
          <select
            v-model="roleFilter"
            class="h-9 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
            @change="loadUsers"
          >
            <option value="">{{ $t("admin.allRoles") }}</option>
            <option value="admin">{{ $t("admin.roleAdmin") }}</option>
            <option value="user">{{ $t("admin.roleUser") }}</option>
          </select>
        </div>
      </div>

      <DataTable v-if="!loading && users.length > 0" :columns="tableColumns" :data="users">
        <template #cell-role="{ row }">
          <Badge
            :class="{
              'bg-violet-600 text-white border-transparent': row.role === 'admin',
              'bg-gray-500 text-white border-transparent': row.role !== 'admin',
            }"
          >
            {{ row.role || "user" }}
          </Badge>
        </template>

        <template #cell-status="{ row }">
          <Badge
            v-if="row.banned"
            class="bg-red-500 text-white border-transparent"
          >
            {{ $t("admin.banned") }}
          </Badge>

          <Badge
            v-else-if="!row.emailVerified"
            class="bg-amber-500 text-white border-transparent"
          >
            {{ $t("admin.unverified") }}
          </Badge>

          <Badge
            v-else
            class="bg-emerald-500 text-white border-transparent"
          >
            {{ $t("admin.active") }}
          </Badge>
        </template>

        <template #cell-createdAt="{ row }">
          {{ formatUserDate(row.createdAt) }}
        </template>

        <template #cell-actions="{ row }">
          <div class="flex flex-wrap items-center justify-end gap-2">
            <Button
              v-if="row.id != sessionData?.user?.id"
              size="sm"
              @click="impersonateUser(row.id)"
            >
              <icon-fa6-solid:user-secret class="size-3.5" />
              {{ $t("admin.impersonate") }}
            </Button>

            <Button
              v-if="!row.banned && row.id != sessionData?.user?.id"
              size="sm"
              variant="destructive"
              @click="banUser(row.id)"
            >
              {{ $t("admin.ban") }}
            </Button>

            <Button v-else-if="row.banned" size="sm" variant="outline" @click="unbanUser(row.id)">
              {{ $t("admin.unban") }}
            </Button>
          </div>
        </template>
      </DataTable>

      <div v-if="loading" class="flex items-center justify-center gap-2 py-12 text-muted-foreground">
        <icon-line-md:loading-twotone-loop class="size-6 animate-spin" />
        {{ $t("admin.loadingUsers") }}
      </div>

      <div v-if="totalPages > 1" class="flex items-center justify-center gap-4 mt-8">
        <Button size="sm" variant="outline" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">
          {{ $t("common.previous") }}
        </Button>
        <span class="text-muted-foreground text-sm">{{
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
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import { formatDate } from "@/helpers/formatter";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";

interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  banned: boolean;
  emailVerified: boolean;
  createdAt: string;
  status?: string;
}

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref("");
const searchQuery = ref("");
const roleFilter = ref("");
const currentPage = ref(1);
const pageSize = ref(20);
const totalUsers = ref(0);

const { session: sessionData } = useSession();

const totalPages = computed(() => Math.ceil(totalUsers.value / pageSize.value));

const { t } = useI18n();

const tableColumns = computed(() => [
  { key: "name" as const, name: t("common.name"), sortable: true, searchable: true },
  { key: "email" as const, name: t("common.email"), sortable: true, searchable: true },
  { key: "role" as const, name: t("common.role"), sortable: true },
  { key: "status" as const, name: t("common.status") },
  { key: "createdAt" as const, name: t("admin.created"), sortable: true },
]);

async function loadUsers() {
  loading.value = true;
  error.value = "";

  try {
    const { data, error: respError } = await api.GET("/auth/admin/users");

    if (!respError && data) {
      // The admin endpoint may return data in various shapes depending on server impl
      const userData = data as unknown as { users?: User[]; total?: number };
      users.value = userData.users ?? (Array.isArray(data) ? (data as User[]) : []);
      totalUsers.value = userData.total ?? users.value.length;
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

async function impersonateUser(userId: string) {
  try {
    const { error: respError } = await api.POST("/auth/admin/users/{userId}/impersonate", {
      params: { path: { userId } },
    });

    if (!respError) {
      // Reload the page to apply the new session
      window.location.href = "/";
    } else {
      error.value = (respError as { message?: string }).message ?? "Failed to impersonate user";
    }
  } catch (err) {
    error.value = `Failed to impersonate user: ${err instanceof Error ? err.message : String(err)}`;
  }
}

async function banUser(userId: string) {
  const reason = window.prompt("Ban reason:");
  if (!reason) return;

  try {
    const { error: respError } = await api.POST("/auth/admin/users/{userId}/ban", {
      params: { path: { userId } },
      body: { banReason: reason },
    });

    if (!respError) {
      await loadUsers();
    } else {
      error.value = (respError as { message?: string }).message ?? "Failed to ban user";
    }
  } catch (err) {
    error.value = `Failed to ban user: ${err instanceof Error ? err.message : String(err)}`;
  }
}

async function unbanUser(userId: string) {
  try {
    const { error: respError } = await api.POST("/auth/admin/users/{userId}/unban", {
      params: { path: { userId } },
    });

    if (!respError) {
      await loadUsers();
    } else {
      error.value = (respError as { message?: string }).message ?? "Failed to unban user";
    }
  } catch (err) {
    error.value = `Failed to unban user: ${err instanceof Error ? err.message : String(err)}`;
  }
}

function changePage(page: number) {
  currentPage.value = page;
  loadUsers();
}

function formatUserDate(date: string | Date) {
  return formatDate(date);
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
