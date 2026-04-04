<template>
  <HeaderContainer :title="$t('admin.userManagement')" />

  <Panel>
    <Alert v-if="error" type="danger">
      {{ error }}
    </Alert>

    <div class="users-filter">
      <div class="search-container">
        <BaseInput
          v-model="searchQuery"
          :placeholder="$t('admin.searchUsers')"
          @input="debouncedSearch"
        />
      </div>

      <div class="filter-container">
        <BaseSelect v-model="roleFilter" @change="loadUsers">
          <option value="">{{ $t("admin.allRoles") }}</option>
          <option value="admin">{{ $t("admin.roleAdmin") }}</option>
          <option value="user">{{ $t("admin.roleUser") }}</option>
        </BaseSelect>
      </div>
    </div>

    <DataTable v-if="!loading && users.length > 0" :columns="tableColumns" :data="users">
      <template #cell-role="{ row }">
        <span class="badge" :class="`badge-${row.role || 'user'}`">
          {{ row.role || "user" }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <span v-if="row.banned" class="badge badge-danger"> {{ $t("admin.banned") }} </span>

        <span v-else-if="!row.emailVerified" class="badge badge-warning">
          {{ $t("admin.unverified") }}
        </span>

        <span v-else class="badge badge-success"> {{ $t("admin.active") }} </span>
      </template>

      <template #cell-createdAt="{ row }">
        {{ formatUserDate(row.createdAt) }}
      </template>

      <template #cell-actions="{ row }">
        <div class="actions-group">
          <UiButton
            v-if="row.id != sessionData?.user?.id"
            size="sm"
            variant="primary"
            @click="impersonateUser(row.id)"
          >
            <icon-fa6-solid:user-secret class="icon" />
            {{ $t("admin.impersonate") }}
          </UiButton>

          <UiButton
            v-if="!row.banned && row.id != sessionData?.user?.id"
            size="sm"
            variant="destructive"
            @click="banUser(row.id)"
          >
            {{ $t("admin.ban") }}
          </UiButton>

          <UiButton v-else-if="row.banned" size="sm" @click="unbanUser(row.id)">
            {{ $t("admin.unban") }}
          </UiButton>
        </div>
      </template>
    </DataTable>

    <div v-if="loading" class="loading-container">
      <icon-line-md:loading-twotone-loop class="loading-icon" />
      {{ $t("admin.loadingUsers") }}
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <UiButton size="sm" :disabled="currentPage === 1" @click="changePage(currentPage - 1)">
        {{ $t("common.previous") }}
      </UiButton>
      <span class="page-info">{{
        $t("common.pageOf", { current: currentPage, total: totalPages })
      }}</span>
      <UiButton
        size="sm"
        :disabled="currentPage === totalPages"
        @click="changePage(currentPage + 1)"
      >
        {{ $t("common.next") }}
      </UiButton>
    </div>
  </Panel>
</template>

<script setup lang="ts">
import Alert from "@/components/layout/Alert.vue";
import DataTable from "@/components/layout/DataTable.vue";
import HeaderContainer from "@/components/layout/HeaderContainer.vue";
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

<style scoped>
.users-filter {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  align-items: center;
}

.search-container {
  flex: 1;
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

.badge-admin {
  background-color: #7c3aed;
  color: white;
}

.badge-user {
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
