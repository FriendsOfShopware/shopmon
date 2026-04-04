<template>
  <HeaderContainer :title="$t('admin.dashboard')" />

  <Panel>
    <Banner v-if="error" variant="error">
      {{ error }}
    </Banner>

    <div v-if="loading" class="loading-container">
      <icon-line-md:loading-twotone-loop class="loading-icon" />
      {{ $t("admin.loadingStats") }}
    </div>

    <div v-if="!loading && stats" class="stats-grid">
      <div class="stat-card">
        <div class="stat-header">
          <h3 class="stat-title">{{ $t("admin.totalUsers") }}</h3>
          <icon-fa6-solid:users class="stat-icon" />
        </div>
        <div class="stat-value">{{ stats.totalUsers }}</div>
        <p class="stat-description">{{ $t("admin.totalUsersDesc") }}</p>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <h3 class="stat-title">{{ $t("admin.totalOrgs") }}</h3>
          <icon-fa6-solid:building class="stat-icon" />
        </div>
        <div class="stat-value">{{ stats.totalOrganizations }}</div>
        <p class="stat-description">{{ $t("admin.totalOrgsDesc") }}</p>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <h3 class="stat-title">{{ $t("admin.totalEnvironments") }}</h3>
          <icon-fa6-solid:store class="stat-icon" />
        </div>
        <div class="stat-value">{{ stats.totalEnvironments }}</div>
        <p class="stat-description">{{ $t("admin.totalEnvironmentsDesc") }}</p>
      </div>

      <div class="stat-card status-breakdown">
        <div class="stat-header">
          <h3 class="stat-title">{{ $t("admin.environmentStatus") }}</h3>
          <icon-fa6-solid:chart-bar class="stat-icon" />
        </div>
        <div class="status-list">
          <div
            v-for="(count, status) in stats.environmentsByStatus"
            :key="status"
            class="status-item"
          >
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
      <UiButton to="/admin/organizations" variant="primary">
        {{ $t("admin.manageOrgs") }}
      </UiButton>
      <UiButton to="/admin/environments" variant="primary">
        {{ $t("admin.manageEnvironments") }}
      </UiButton>
    </div>
  </Panel>

  <!-- Growth Charts -->
  <div v-if="growthData" class="charts-grid">
    <Panel title="User Growth">
      <div class="chart-container">
        <canvas ref="userChartCanvas" />
      </div>
    </Panel>

    <Panel title="Shop Growth">
      <div class="chart-container">
        <canvas ref="shopChartCanvas" />
      </div>
    </Panel>
  </div>

  <!-- Shopware Version Distribution -->
  <Panel v-if="versionData && versionData.length > 0" title="Shopware Version Distribution">
    <div class="chart-container">
      <canvas ref="versionChartCanvas" />
    </div>
  </Panel>

  <!-- Recent Activity -->
  <div v-if="activity" class="charts-grid">
    <Panel title="Recent Signups">
      <div v-if="activity.recentUsers.length === 0" class="empty-state">No recent signups</div>
      <div v-else class="activity-list">
        <div v-for="user in activity.recentUsers" :key="user.id" class="activity-item">
          <div class="activity-info">
            <span class="activity-name">{{ user.displayName }}</span>
            <span class="activity-detail">{{ user.email }}</span>
          </div>
          <span class="activity-time">{{ formatDateTime(user.createdAt) }}</span>
        </div>
      </div>
    </Panel>

    <Panel title="Recent Environments">
      <div v-if="activity.recentEnvironments.length === 0" class="empty-state">
        No recent environments
      </div>
      <div v-else class="activity-list">
        <div v-for="env in activity.recentEnvironments" :key="env.id" class="activity-item">
          <div class="activity-info">
            <span class="activity-name">{{ env.name }}</span>
            <span class="activity-detail">{{ env.organizationName }}</span>
          </div>
          <span class="activity-time">{{
            env.lastScrapedAt ? formatDateTime(env.lastScrapedAt) : ""
          }}</span>
        </div>
      </div>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import Banner from "@/components/layout/Banner.vue";
import HeaderContainer from "@/components/layout/HeaderContainer.vue";
import Panel from "@/components/layout/Panel.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { formatDateTime } from "@/helpers/formatter";
import { onMounted, onUnmounted, ref, nextTick, watch } from "vue";
import { Chart, registerables } from "chart.js";

Chart.register(...registerables);

type Stats = components["schemas"]["AdminStats"];
type GrowthData = components["schemas"]["AdminGrowth"];
type Activity = components["schemas"]["AdminRecentActivity"];
type VersionData = components["schemas"]["ShopwareVersionCount"][];

const stats = ref<Stats | null>(null);
const growthData = ref<GrowthData | null>(null);
const activity = ref<Activity | null>(null);
const versionData = ref<VersionData | null>(null);
const loading = ref(true);
const error = ref("");

const userChartCanvas = ref<HTMLCanvasElement | null>(null);
const shopChartCanvas = ref<HTMLCanvasElement | null>(null);
const versionChartCanvas = ref<HTMLCanvasElement | null>(null);
let userChartInstance: Chart | null = null;
let shopChartInstance: Chart | null = null;
let versionChartInstance: Chart | null = null;

function createGrowthChart(
  canvas: HTMLCanvasElement,
  data: { month: string; count: number }[],
  label: string,
  color: string,
) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;

  return new Chart(ctx, {
    type: "line",
    data: {
      labels: data.map((d) => d.month),
      datasets: [
        {
          label,
          data: data.map((d) => d.count),
          borderColor: color,
          backgroundColor: color + "20",
          fill: true,
          tension: 0.3,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
      },
      scales: {
        x: {
          title: { display: true, text: "Month" },
        },
        y: {
          beginAtZero: true,
          title: { display: true, text: "Total Count" },
          ticks: { precision: 0 },
        },
      },
    },
  });
}

function createVersionChart(canvas: HTMLCanvasElement, data: { version: string; count: number }[]) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;

  const colors = [
    "#6366f1",
    "#10b981",
    "#f59e0b",
    "#ef4444",
    "#8b5cf6",
    "#06b6d4",
    "#ec4899",
    "#84cc16",
    "#f97316",
    "#14b8a6",
  ];

  return new Chart(ctx, {
    type: "pie",
    data: {
      labels: data.map((d) => d.version),
      datasets: [
        {
          data: data.map((d) => d.count),
          backgroundColor: data.map((_, i) => colors[i % colors.length] + "cc"),
          borderColor: data.map((_, i) => colors[i % colors.length]),
          borderWidth: 1,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: "right",
        },
      },
    },
  });
}

function renderCharts() {
  if (userChartInstance) userChartInstance.destroy();
  if (shopChartInstance) shopChartInstance.destroy();
  if (versionChartInstance) versionChartInstance.destroy();

  if (growthData.value) {
    if (userChartCanvas.value) {
      userChartInstance = createGrowthChart(
        userChartCanvas.value,
        growthData.value.users,
        "Users",
        "#6366f1",
      );
    }
    if (shopChartCanvas.value) {
      shopChartInstance = createGrowthChart(
        shopChartCanvas.value,
        growthData.value.environments,
        "Environments",
        "#10b981",
      );
    }
  }

  if (versionData.value && versionChartCanvas.value) {
    versionChartInstance = createVersionChart(versionChartCanvas.value, versionData.value);
  }
}

async function loadStats() {
  loading.value = true;
  error.value = "";

  try {
    const [statsRes, growthRes, activityRes, versionRes] = await Promise.all([
      api.GET("/admin/stats"),
      api.GET("/admin/growth"),
      api.GET("/admin/recent-activity"),
      api.GET("/admin/shopware-versions"),
    ]);
    stats.value = statsRes.data ?? null;
    growthData.value = growthRes.data ?? null;
    activity.value = activityRes.data ?? null;
    versionData.value = versionRes.data ?? null;
  } catch (err) {
    error.value = `Failed to load dashboard stats: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await loadStats();
  await nextTick();
  renderCharts();
});

watch([growthData, versionData], async () => {
  await nextTick();
  renderCharts();
});

onUnmounted(() => {
  if (userChartInstance) userChartInstance.destroy();
  if (shopChartInstance) shopChartInstance.destroy();
  if (versionChartInstance) versionChartInstance.destroy();
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

.action-links .ui-button {
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  font-weight: 600;
  text-decoration: none;
  display: inline-block;
  transition: all 0.2s ease;
}

.action-links .ui-button--primary {
  background-color: var(--primary-color);
  color: white;
}

.action-links .ui-button--primary:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 2rem;
}

.chart-container {
  height: 300px;
  position: relative;
}

.activity-list {
  display: flex;
  flex-direction: column;
}

.activity-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border-color);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.activity-name {
  font-weight: 600;
  color: var(--text-color-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.activity-detail {
  font-size: 0.8rem;
  color: var(--text-color-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.activity-time {
  font-size: 0.8rem;
  color: var(--text-color-muted);
  white-space: nowrap;
  margin-left: 1rem;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--text-color-muted);
}
</style>
