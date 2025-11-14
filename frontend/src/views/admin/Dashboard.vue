<template>
    <HeaderContainer title="Admin Dashboard" />

    <div class="panel">
        <Alert v-if="error" type="danger">
            {{ error }}
        </Alert>

        <div v-if="loading" class="loading-container">
            <icon-line-md:loading-twotone-loop class="loading-icon" />
            Loading dashboard stats...
        </div>

        <div v-if="!loading && stats" class="stats-grid">
            <!-- Users Stat -->
            <div class="stat-card">
                <div class="stat-header">
                    <h3 class="stat-title">Total Users</h3>
                    <icon-fa6-solid:users class="stat-icon" />
                </div>
                <div class="stat-value">{{ stats.totalUsers }}</div>
                <p class="stat-description">Active users in the system</p>
            </div>

            <!-- Organizations Stat -->
            <div class="stat-card">
                <div class="stat-header">
                    <h3 class="stat-title">Total Organizations</h3>
                    <icon-fa6-solid:building class="stat-icon" />
                </div>
                <div class="stat-value">{{ stats.totalOrganizations }}</div>
                <p class="stat-description">Registered organizations</p>
            </div>

            <!-- Total Shops Stat -->
            <div class="stat-card">
                <div class="stat-header">
                    <h3 class="stat-title">Total Shops</h3>
                    <icon-fa6-solid:store class="stat-icon" />
                </div>
                <div class="stat-value">{{ stats.totalShops }}</div>
                <p class="stat-description">Monitored shops</p>
            </div>

            <!-- Shop Status Breakdown -->
            <div class="stat-card status-breakdown">
                <div class="stat-header">
                    <h3 class="stat-title">Shop Status</h3>
                    <icon-fa6-solid:chart-bar class="stat-icon" />
                </div>
                <div class="status-list">
                    <div v-for="(count, status) in stats.shopsByStatus" :key="status" class="status-item">
                        <div class="status-label-wrapper">
                            <span class="badge" :class="`badge-${status}`">{{ status }}</span>
                        </div>
                        <div class="status-count">{{ count }}</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Action Links -->
        <div v-if="!loading && stats" class="action-links">
            <router-link to="/admin/organizations" class="btn btn-primary">
                Manage Organizations
            </router-link>
            <router-link to="/admin/shops" class="btn btn-primary">
                Manage Shops
            </router-link>
        </div>
    </div>
</template>

<script setup lang="ts">
import Alert from '@/components/layout/Alert.vue';
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import { trpcClient } from '@/helpers/trpc';
import { onMounted, ref } from 'vue';

type Stats = Awaited<ReturnType<typeof trpcClient.admin.getStats.query>>;

const stats = ref<Stats | null>(null);
const loading = ref(true);
const error = ref('');

async function loadStats() {
    loading.value = true;
    error.value = '';

    try {
        const response = await trpcClient.admin.getStats.query();
        stats.value = response;
    } catch (err) {
        error.value = `Failed to load dashboard stats: ${err instanceof Error ? err.message : String(err)}`;
    } finally {
        loading.value = false;
    }
}

onMounted(() => {
    loadStats();
});
</script>

<style scoped>
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
    margin-bottom: 2rem;
}

.stat-card {
    background: var(--bg-color-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    transition: all 0.3s ease;
}

.stat-card:hover {
    border-color: var(--primary-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.stat-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-color-muted);
    margin: 0;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.stat-icon {
    width: 24px;
    height: 24px;
    color: var(--primary-color);
    opacity: 0.7;
}

.stat-value {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--text-color-primary);
    margin: 0.5rem 0;
}

.stat-description {
    font-size: 0.875rem;
    color: var(--text-color-muted);
    margin: 0;
}

.status-breakdown {
    grid-column: span 1;
}

.status-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background: var(--bg-color-primary);
    border-radius: 6px;
}

.status-label-wrapper {
    flex: 1;
}

.badge {
    display: inline-block;
    padding: 0.35rem 0.75rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 600;
    text-transform: capitalize;
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

.status-count {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-color-primary);
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
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.action-links {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
}

.btn {
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    font-weight: 600;
    text-decoration: none;
    display: inline-block;
    transition: all 0.2s ease;
    cursor: pointer;
    border: none;
}

.btn-primary {
    background-color: var(--primary-color);
    color: white;
}

.btn-primary:hover {
    opacity: 0.9;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
</style>
