<template>
    <main-container></main-container>
    <div class="users-container">
        <HeaderContainer title="User Management" />

        <div class="users-content">
            <Alert v-if="error" type="danger">
                {{ error }}
            </Alert>

            <div class="users-header">
                <div class="search-container">
                    <input
                        v-model="searchQuery"
                        type="text"
                        placeholder="Search users by email..."
                        class="field search-input"
                        @input="debouncedSearch"
                    />
                </div>
                <div class="filter-container">
                    <select v-model="roleFilter" class="field" @change="loadUsers">
                        <option value="">All Roles</option>
                        <option value="admin">Admin</option>
                        <option value="user">User</option>
                    </select>
                </div>
            </div>

            <DataTable 
                v-if="!loading && users.length > 0"
                :columns="tableColumns"
                :data="users"
            >
                <template #cell-role="{ row }">
                    <span class="badge" :class="`badge-${row.role || 'user'}`">
                        {{ row.role || 'user' }}
                    </span>
                </template>
                
                <template #cell-status="{ row }">
                    <span v-if="row.banned" class="badge badge-danger">
                        Banned
                    </span>
                    <span v-else-if="!row.emailVerified" class="badge badge-warning">
                        Unverified
                    </span>
                    <span v-else class="badge badge-success">
                        Active
                    </span>
                </template>
                
                <template #cell-createdAt="{ row }">
                    {{ formatUserDate(row.createdAt) }}
                </template>
                
                <template #cell-actions="{ row }">
                    <div class="action-buttons">
                        <button
                            v-if="row.id !== session?.data?.user?.id"
                            class="btn btn-sm btn-primary"
                            @click="impersonateUser(row.id)"
                        >
                            <icon-fa6-solid:user-secret class="btn-icon" />
                            Impersonate
                        </button>
                        
                        <button
                            v-if="!row.banned && row.id !== session?.data?.user?.id"
                            class="btn btn-sm btn-danger-outline"
                            @click="banUser(row.id)"
                        >
                            Ban
                        </button>
                        <button
                            v-else-if="row.banned"
                            class="btn btn-sm"
                            @click="unbanUser(row.id)"
                        >
                            Unban
                        </button>
                    </div>
                </template>
            </DataTable>

            <div v-if="loading" class="loading-container">
                <icon-line-md:loading-twotone-loop class="loading-icon" />
                Loading users...
            </div>

            <div v-if="totalPages > 1" class="pagination">
                <button
                    class="btn btn-sm"
                    :disabled="currentPage === 1"
                    @click="changePage(currentPage - 1)"
                >
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
        </div>
    </div>
</template>

<script setup lang="ts">
import Alert from '@/components/layout/Alert.vue';
import DataTable from '@/components/layout/DataTable.vue';
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import { authClient } from '@/helpers/auth-client';
import { formatDate } from '@/helpers/formatter';
import { computed, onMounted, ref } from 'vue';

type User = Awaited<
    ReturnType<typeof authClient.admin.listUsers>
>['data']['users'][number];

const users = ref<User[]>([]);
const loading = ref(true);
const error = ref('');
const searchQuery = ref('');
const roleFilter = ref('');
const currentPage = ref(1);
const pageSize = ref(20);
const totalUsers = ref(0);

const session = authClient.useSession();

const totalPages = computed(() => Math.ceil(totalUsers.value / pageSize.value));

const tableColumns = [
    { key: 'name', name: 'Name', sortable: true, searchable: true },
    { key: 'email', name: 'Email', sortable: true, searchable: true },
    { key: 'role', name: 'Role', sortable: true },
    { key: 'status', name: 'Status' },
    { key: 'createdAt', name: 'Created', sortable: true },
];

async function loadUsers() {
    loading.value = true;
    error.value = '';

    try {
        const query: { [key: string]: unknown } = {
            limit: pageSize.value,
            offset: (currentPage.value - 1) * pageSize.value,
            sortBy: 'createdAt',
            sortDirection: 'desc' as const,
        };

        if (searchQuery.value) {
            query.searchField = 'email';
            query.searchOperator = 'contains';
            query.searchValue = searchQuery.value;
        }

        if (roleFilter.value) {
            query.filterField = 'role';
            query.filterOperator = 'eq';
            query.filterValue = roleFilter.value;
        }

        const response = await authClient.admin.listUsers({ query });

        if (!response.error) {
            users.value = response.data?.users || [];
            totalUsers.value = response.data?.total || 0;
        } else {
            error.value = response.error.message || 'Failed to load users';
        }
    } catch (err) {
        error.value = 'Failed to load users';
        console.error('Error loading users:', err);
    } finally {
        loading.value = false;
    }
}

async function impersonateUser(userId: string) {
    try {
        const response = await authClient.admin.impersonateUser({ userId });

        if (!response.error) {
            // Reload the page to apply the new session
            window.location.href = '/';
        } else {
            error.value =
                response.error.message || 'Failed to impersonate user';
        }
    } catch (err) {
        error.value = 'Failed to impersonate user';
        console.error('Error impersonating user:', err);
    }
}

async function banUser(userId: string) {
    const reason = prompt('Ban reason:');
    if (!reason) return;

    try {
        const response = await authClient.admin.banUser({
            userId,
            banReason: reason,
        });

        if (!response.error) {
            await loadUsers();
        } else {
            error.value = response.error.message || 'Failed to ban user';
        }
    } catch (err) {
        error.value = 'Failed to ban user';
        console.error('Error banning user:', err);
    }
}

async function unbanUser(userId: string) {
    try {
        const response = await authClient.admin.unbanUser({ userId });

        if (!response.error) {
            await loadUsers();
        } else {
            error.value = response.error.message || 'Failed to unban user';
        }
    } catch (err) {
        error.value = 'Failed to unban user';
        console.error('Error unbanning user:', err);
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
.users-container {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: white;
}

.users-content {
    flex: 1;
    padding: 2rem;
    overflow-y: auto;
}

.users-header {
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
    color: var(--text-muted);
    gap: 0.5rem;
}

.loading-icon {
    width: 24px;
    height: 24px;
}

.action-buttons {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.btn-sm {
    padding: 0.25rem 0.75rem;
    font-size: 0.875rem;
}

.btn-primary {
    background-color: #3b82f6;
    color: white;
    border: 1px solid #3b82f6;
}

.btn-primary:hover {
    background-color: #2563eb;
    border-color: #2563eb;
}

.btn-icon {
    width: 0.875rem;
    height: 0.875rem;
    margin-right: 0.25rem;
    display: inline-block;
    vertical-align: middle;
}

.badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
}

.badge-admin {
    background-color: #7c3aed;
    color: white;
}

.badge-user {
    background-color: #6b7280;
    color: white;
}

.badge-success {
    background-color: #10b981;
    color: white;
}

.badge-warning {
    background-color: #f59e0b;
    color: white;
}

.badge-danger {
    background-color: #ef4444;
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
    color: var(--text-muted);
}
</style>
