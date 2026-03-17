<template>
  <div v-if="isLoading" class="loading">Loading uptime data...</div>

  <template v-else-if="uptimeData">
    <div v-if="!uptimeData.enabled" class="uptime-toggle">
      <Alert type="info"> Uptime monitoring is not enabled for this shop. </Alert>
      <button class="btn btn-primary" :disabled="isToggling" @click="toggleUptime(true)">
        Enable Uptime Monitoring
      </button>
    </div>

    <template v-else>
      <div class="uptime-header">
        <div class="uptime-status-card panel" :class="statusClass">
          <div class="status-indicator">
            <span class="status-dot" />
            <span class="status-text">{{ statusLabel }}</span>
          </div>
          <div v-if="uptimeData.downSince" class="down-since">
            Down since {{ formatDateTime(uptimeData.downSince) }}
          </div>
          <div v-if="uptimeData.lastCheckedAt" class="last-checked">
            Last checked: {{ formatDateTime(uptimeData.lastCheckedAt) }}
          </div>
        </div>

        <div class="uptime-stats panel">
          <div class="stat">
            <span class="stat-value">{{ uptimeData.stats.uptimePercentage ?? "-" }}%</span>
            <span class="stat-label">Uptime (30d)</span>
          </div>
          <div class="stat">
            <span class="stat-value">{{ uptimeData.stats.avgTtfb ?? "-" }} ms</span>
            <span class="stat-label">Avg TTFB</span>
          </div>
          <div class="stat">
            <span class="stat-value">{{ uptimeData.stats.totalChecks }}</span>
            <span class="stat-label">Total Checks</span>
          </div>
        </div>
      </div>

      <div class="mb-1">
        <button class="btn btn-danger-outline" :disabled="isToggling" @click="toggleUptime(false)">
          Disable Uptime Monitoring
        </button>
      </div>

      <div v-if="uptimeData.dailyStats.length > 0" class="panel chart-panel">
        <canvas ref="dailyUptimeChartCanvas" width="800" height="300" />
      </div>

      <div v-if="uptimeData.dailyStats.length > 0" class="panel chart-panel">
        <canvas ref="dailyTtfbChartCanvas" width="800" height="300" />
      </div>

      <div v-if="uptimeData.checks.length > 0" class="panel chart-panel">
        <canvas ref="recentTtfbChartCanvas" width="800" height="400" />
      </div>

      <div class="panel panel-table">
        <data-table
          :columns="[
            { key: 'checkedAt', name: 'Checked At' },
            { key: 'isUp', name: 'Status' },
            { key: 'statusCode', name: 'HTTP Code' },
            { key: 'ttfb', name: 'TTFB' },
            { key: 'responseTime', name: 'Response Time' },
            { key: 'error', name: 'Error' },
          ]"
          :data="uptimeData.checks"
        >
          <template #cell-checkedAt="{ row }">
            {{ formatDateTime(row.checkedAt) }}
          </template>
          <template #cell-isUp="{ row }">
            <span :class="row.isUp ? 'badge-up' : 'badge-down'">
              {{ row.isUp ? "UP" : "DOWN" }}
            </span>
          </template>
          <template #cell-statusCode="{ row }">
            {{ row.statusCode ?? "-" }}
          </template>
          <template #cell-ttfb="{ row }">
            {{ row.ttfb ? `${row.ttfb} ms` : "-" }}
          </template>
          <template #cell-responseTime="{ row }">
            {{ row.responseTime ? `${row.responseTime} ms` : "-" }}
          </template>
          <template #cell-error="{ row }">
            {{ row.error ?? "-" }}
          </template>
        </data-table>
      </div>
    </template>
  </template>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from "vue";
import { Chart, registerables } from "chart.js";
import "chartjs-adapter-date-fns";
import { formatDateTime } from "@/helpers/formatter";
import { trpcClient } from "@/helpers/trpc";
import { useShopDetail } from "@/composables/useShopDetail";
import { useAlert } from "@/composables/useAlert";
import Alert from "@/components/layout/Alert.vue";

Chart.register(...registerables);

const { shop } = useShopDetail();
const { success, error: showError } = useAlert();

type UptimeData = Awaited<ReturnType<typeof trpcClient.organization.uptime.getData.query>>;

const uptimeData = ref<UptimeData | null>(null);
const isLoading = ref(true);
const isToggling = ref(false);
const recentTtfbChartCanvas = ref<HTMLCanvasElement | null>(null);
const dailyUptimeChartCanvas = ref<HTMLCanvasElement | null>(null);
const dailyTtfbChartCanvas = ref<HTMLCanvasElement | null>(null);
let recentChart: Chart | null = null;
let dailyUptimeChart: Chart | null = null;
let dailyTtfbChart: Chart | null = null;

const statusClass = ref("");
const statusLabel = ref("");

function updateStatusDisplay() {
  if (!uptimeData.value) return;
  const status = uptimeData.value.status;
  statusClass.value =
    status === "up" ? "status-up" : status === "down" ? "status-down" : "status-unknown";
  statusLabel.value = status === "up" ? "Operational" : status === "down" ? "Down" : "Unknown";
}

async function loadUptimeData() {
  if (!shop.value) return;

  try {
    isLoading.value = true;
    uptimeData.value = await trpcClient.organization.uptime.getData.query({
      shopId: shop.value.id,
    });
    updateStatusDisplay();
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e));
  } finally {
    isLoading.value = false;
  }
}

async function toggleUptime(enabled: boolean) {
  if (!shop.value) return;

  try {
    isToggling.value = true;
    await trpcClient.organization.uptime.updateSettings.mutate({
      shopId: shop.value.id,
      enabled,
    });
    success(enabled ? "Uptime monitoring enabled" : "Uptime monitoring disabled");
    await loadUptimeData();
  } catch (e) {
    showError(e instanceof Error ? e.message : String(e));
  } finally {
    isToggling.value = false;
  }
}

function createRecentTtfbChart() {
  if (!recentTtfbChartCanvas.value || !uptimeData.value?.checks.length) return;

  if (recentChart) {
    recentChart.destroy();
  }

  const ctx = recentTtfbChartCanvas.value.getContext("2d");
  if (!ctx) return;

  const sortedData = [...uptimeData.value.checks]
    .filter((c) => c.ttfb !== null)
    .sort((a, b) => new Date(a.checkedAt).getTime() - new Date(b.checkedAt).getTime());

  recentChart = new Chart(ctx, {
    type: "line",
    data: {
      datasets: [
        {
          label: "TTFB (ms)",
          data: sortedData.map((item) => ({
            x: new Date(item.checkedAt).getTime(),
            y: item.ttfb!,
          })),
          borderColor: "rgb(59, 130, 246)",
          backgroundColor: "rgba(59, 130, 246, 0.1)",
          fill: true,
          tension: 0.3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        title: { display: true, text: "TTFB — Recent Checks" },
        tooltip: {
          callbacks: {
            label: (context) => `TTFB: ${context.parsed.y}ms`,
            title: (items) =>
              items.length > 0 ? new Date(items[0].parsed.x).toLocaleString() : "",
          },
        },
      },
      scales: {
        x: {
          type: "time",
          time: { unit: "hour", displayFormats: { hour: "MMM dd HH:mm" } },
          title: { display: true, text: "Time" },
        },
        y: { beginAtZero: true, title: { display: true, text: "TTFB (ms)" } },
      },
    },
  });
}

function createDailyUptimeChart() {
  if (!dailyUptimeChartCanvas.value || !uptimeData.value?.dailyStats.length) return;

  if (dailyUptimeChart) {
    dailyUptimeChart.destroy();
  }

  const ctx = dailyUptimeChartCanvas.value.getContext("2d");
  if (!ctx) return;

  const data = uptimeData.value.dailyStats;

  dailyUptimeChart = new Chart(ctx, {
    type: "bar",
    data: {
      datasets: [
        {
          label: "Uptime %",
          data: data.map((d) => ({
            x: new Date(d.date).getTime(),
            y: d.totalChecks > 0 ? Number(((d.upChecks / d.totalChecks) * 100).toFixed(1)) : 0,
          })),
          backgroundColor: data.map((d) => {
            const pct = d.totalChecks > 0 ? (d.upChecks / d.totalChecks) * 100 : 0;
            return pct >= 99
              ? "rgba(34, 197, 94, 0.7)"
              : pct >= 95
                ? "rgba(234, 179, 8, 0.7)"
                : "rgba(239, 68, 68, 0.7)";
          }),
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        title: { display: true, text: "Daily Uptime — Last Year" },
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (context) => `Uptime: ${context.parsed.y}%`,
            title: (items) =>
              items.length > 0 ? new Date(items[0].parsed.x).toLocaleDateString() : "",
          },
        },
      },
      scales: {
        x: {
          type: "time",
          time: { unit: "month", displayFormats: { month: "MMM yyyy" } },
          title: { display: true, text: "Date" },
        },
        y: { min: 0, max: 100, title: { display: true, text: "Uptime %" } },
      },
    },
  });
}

function createDailyTtfbChart() {
  if (!dailyTtfbChartCanvas.value || !uptimeData.value?.dailyStats.length) return;

  if (dailyTtfbChart) {
    dailyTtfbChart.destroy();
  }

  const ctx = dailyTtfbChartCanvas.value.getContext("2d");
  if (!ctx) return;

  const data = uptimeData.value.dailyStats.filter((d) => d.avgTtfb !== null);

  dailyTtfbChart = new Chart(ctx, {
    type: "line",
    data: {
      datasets: [
        {
          label: "Avg TTFB",
          data: data.map((d) => ({ x: new Date(d.date).getTime(), y: d.avgTtfb! })),
          borderColor: "rgb(59, 130, 246)",
          backgroundColor: "rgba(59, 130, 246, 0.1)",
          fill: true,
          tension: 0.3,
        },
        {
          label: "Min TTFB",
          data: data.map((d) => ({ x: new Date(d.date).getTime(), y: d.minTtfb ?? 0 })),
          borderColor: "rgba(34, 197, 94, 0.6)",
          borderDash: [5, 5],
          pointRadius: 0,
        },
        {
          label: "Max TTFB",
          data: data.map((d) => ({ x: new Date(d.date).getTime(), y: d.maxTtfb ?? 0 })),
          borderColor: "rgba(239, 68, 68, 0.6)",
          borderDash: [5, 5],
          pointRadius: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        title: { display: true, text: "Daily TTFB — Last Year" },
        tooltip: {
          callbacks: {
            label: (context) => `${context.dataset.label}: ${context.parsed.y}ms`,
            title: (items) =>
              items.length > 0 ? new Date(items[0].parsed.x).toLocaleDateString() : "",
          },
        },
      },
      scales: {
        x: {
          type: "time",
          time: { unit: "month", displayFormats: { month: "MMM yyyy" } },
          title: { display: true, text: "Date" },
        },
        y: { beginAtZero: true, title: { display: true, text: "TTFB (ms)" } },
      },
    },
  });
}

function createAllCharts() {
  createDailyUptimeChart();
  createDailyTtfbChart();
  createRecentTtfbChart();
}

function destroyAllCharts() {
  if (recentChart) {
    recentChart.destroy();
    recentChart = null;
  }
  if (dailyUptimeChart) {
    dailyUptimeChart.destroy();
    dailyUptimeChart = null;
  }
  if (dailyTtfbChart) {
    dailyTtfbChart.destroy();
    dailyTtfbChart = null;
  }
}

onMounted(async () => {
  if (shop.value) {
    await loadUptimeData();
    await nextTick();
    createAllCharts();
  }
});

onUnmounted(() => {
  destroyAllCharts();
});

watch(
  () => shop.value?.id,
  async () => {
    await loadUptimeData();
    await nextTick();
    createAllCharts();
  },
);

watch(
  () => [uptimeData.value?.checks, uptimeData.value?.dailyStats],
  async () => {
    await nextTick();
    createAllCharts();
  },
  { deep: true },
);
</script>

<style scoped>
.uptime-header {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1rem;
}

@media (max-width: 768px) {
  .uptime-header {
    grid-template-columns: 1fr;
  }
}

.uptime-status-card {
  padding: 1.5rem;
}

.uptime-status-card.status-up {
  border-left: 4px solid #22c55e;
}

.uptime-status-card.status-down {
  border-left: 4px solid #ef4444;
}

.uptime-status-card.status-unknown {
  border-left: 4px solid #a3a3a3;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  display: inline-block;
}

.status-up .status-dot {
  background-color: #22c55e;
}

.status-down .status-dot {
  background-color: #ef4444;
  animation: pulse 2s infinite;
}

.status-unknown .status-dot {
  background-color: #a3a3a3;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-text {
  font-size: 1.25rem;
  font-weight: 600;
}

.down-since,
.last-checked {
  font-size: 0.875rem;
  color: var(--text-muted);
}

.uptime-stats {
  padding: 1.5rem;
  display: flex;
  justify-content: space-around;
  align-items: center;
}

.stat {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
}

.badge-up {
  background-color: #dcfce7;
  color: #166534;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge-down {
  background-color: #fecaca;
  color: #991b1b;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.uptime-toggle {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: flex-start;
}

.loading {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
}

.mb-1 {
  margin-bottom: 1rem;
}
</style>
