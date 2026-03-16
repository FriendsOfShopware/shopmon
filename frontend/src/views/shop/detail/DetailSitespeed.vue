<template>
  <Alert v-if="shop && !shop.sitespeedEnabled" type="info">
    Sitespeed monitoring is not enabled for this shop. Please enable it in the shop settings to view
    performance metrics.
  </Alert>
  <div class="mb-1">
    <a
      v-if="showSitespeedData"
      class="btn btn-primary-outline"
      :href="shop.sitespeedReportUrl"
      target="_blank"
    >
      <i class="fa fa-chart-line" /> View Sitespeed Report
    </a>
  </div>

  <div v-if="showSitespeedData" class="panel chart-panel">
    <canvas ref="timeChartCanvas" width="800" height="400" />
  </div>

  <div v-if="showSitespeedData" class="panel chart-panel">
    <canvas ref="transferSizeChartCanvas" width="800" height="400" />
  </div>

  <div v-if="showSitespeedData" class="panel chart-panel">
    <canvas ref="clsChartCanvas" width="800" height="400" />
  </div>

  <div v-if="shop && shop.sitespeedEnabled" class="panel panel-table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'createdAt', name: 'Checked At' },
        { key: 'deployment', name: 'Deployment' },
        { key: 'ttfb', name: 'TTFB' },
        { key: 'fullyLoaded', name: 'Fully Loaded' },
        { key: 'largestContentfulPaint', name: 'Largest Contentful Paint' },
        { key: 'firstContentfulPaint', name: 'First Contentful Paint' },
        { key: 'cumulativeLayoutShift', name: 'Cumulative Layout Shift' },
        { key: 'transferSize', name: 'Transfer Size' },
      ]"
      :data="shop.sitespeed || []"
    >
      <template #cell-ttfb="{ row }">
        {{ row.ttfb ? `${row.ttfb} ms` : "-" }}
      </template>
      <template #cell-fullyLoaded="{ row }">
        {{ row.fullyLoaded ? `${row.fullyLoaded} ms` : "-" }}
      </template>
      <template #cell-largestContentfulPaint="{ row }">
        {{ row.largestContentfulPaint ? `${row.largestContentfulPaint} ms` : "-" }}
      </template>
      <template #cell-firstContentfulPaint="{ row }">
        {{ row.firstContentfulPaint ? `${row.firstContentfulPaint} ms` : "-" }}
      </template>
      <template #cell-cumulativeLayoutShift="{ row }">
        {{ row.cumulativeLayoutShift ? row.cumulativeLayoutShift : "-" }}
      </template>
      <template #cell-transferSize="{ row }">
        {{ row.transferSize ? formatBytes(row.transferSize) : "-" }}
      </template>
      <template #cell-createdAt="{ row }">
        {{ formatDateTime(row.createdAt) }}
      </template>
      <template #cell-deployment="{ row }">
        <router-link
          v-if="row.deployment"
          :to="{
            name: 'account.shops.detail.deployment',
            params: {
              slug: $route.params.slug,
              shopId: $route.params.shopId,
              deploymentId: row.deployment.id,
            },
          }"
        >
          {{ row.deployment.name || `Deployment #${row.deployment.id}` }}
        </router-link>
        <span v-else>-</span>
      </template>
    </data-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, Ref, computed } from "vue";
import { Chart, registerables } from "chart.js";
import annotationPlugin from "chartjs-plugin-annotation";
import "chartjs-adapter-date-fns";
import { formatDateTime } from "@/helpers/formatter";
import { useShopDetail } from "@/composables/useShopDetail";

const { shop } = useShopDetail();

Chart.register(annotationPlugin);

// Computed property to check if sitespeed data should be displayed
const showSitespeedData = computed(() => {
  return (
    shop.value &&
    shop.value.sitespeedEnabled &&
    shop.value.sitespeed &&
    shop.value.sitespeed.length > 0
  );
});

// Extract unique deployments from sitespeed data for chart annotations
const deploymentMarkers = computed(() => {
  if (!shop.value?.sitespeed) return [];

  const deploymentMap = new Map<number, { id: number; name: string; timestamp: number }>();

  shop.value.sitespeed.forEach((item) => {
    if (item.deployment?.id) {
      const timestamp = new Date(item.createdAt).getTime();
      if (!deploymentMap.has(item.deployment.id)) {
        deploymentMap.set(item.deployment.id, {
          id: item.deployment.id,
          name: item.deployment.name || `Deployment #${item.deployment.id}`,
          timestamp,
        });
      }
    }
  });

  return Array.from(deploymentMap.values()).sort((a, b) => a.timestamp - b.timestamp);
});

function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

Chart.register(...registerables);

const timeChartCanvas = ref<HTMLCanvasElement | null>(null);
const timeChartInstance = ref<Chart | null>(null);
const transferSizeChartCanvas = ref<HTMLCanvasElement | null>(null);
const transferSizeChartInstance = ref<Chart | null>(null);
const clsChartCanvas = ref<HTMLCanvasElement | null>(null);
const clsChartInstance = ref<Chart | null>(null);

const timeMetrics = [
  { key: "ttfb", label: "TTFB" },
  { key: "fullyLoaded", label: "Fully Loaded" },
  { key: "largestContentfulPaint", label: "LCP" },
  { key: "firstContentfulPaint", label: "FCP" },
];

interface SitespeedDataItem {
  createdAt: string;
  deployment?: {
    id: number;
    name: string;
  } | null;
  ttfb?: number;
  fullyLoaded?: number;
  largestContentfulPaint?: number;
  firstContentfulPaint?: number;
  cumulativeLayoutShift?: number;
  transferSize?: number;
}

interface ChartConfig {
  canvasRef: Ref<HTMLCanvasElement | null>;
  chartInstance: Ref<Chart | null>;
  title: string;
  yAxisLabel: string;
  datasets: Array<{
    label: string;
    dataKey?: string;

    valueFormatter?: (_item: SitespeedDataItem) => number;

    tooltipFormatter?: (_value: number) => string;
  }>;
}

const chartConfigs: ChartConfig[] = [
  {
    canvasRef: timeChartCanvas,
    chartInstance: timeChartInstance,
    title: "Performance Metrics Over Time",
    yAxisLabel: "Time (ms)",
    datasets: timeMetrics.map((metric) => ({
      label: metric.label,
      dataKey: metric.key,
      valueFormatter: (item) => item[metric.key as keyof typeof item] as number,
      tooltipFormatter: (value) => `${value}ms`,
    })),
  },
  {
    canvasRef: transferSizeChartCanvas,
    chartInstance: transferSizeChartInstance,
    title: "Transfer Size Over Time",
    yAxisLabel: "Size (KB)",
    datasets: [
      {
        label: "Transfer Size",
        valueFormatter: (item) => (item.transferSize ? Math.round(item.transferSize / 1024) : 0),
        tooltipFormatter: (value) => `${value} KB`,
      },
    ],
  },
  {
    canvasRef: clsChartCanvas,
    chartInstance: clsChartInstance,
    title: "Cumulative Layout Shift Over Time",
    yAxisLabel: "CLS Score",
    datasets: [
      {
        label: "Cumulative Layout Shift",
        valueFormatter: (item) => item.cumulativeLayoutShift ?? 0,
        tooltipFormatter: (value) => `${value}`,
      },
    ],
  },
];

const createChart = (config: ChartConfig) => {
  if (!config.canvasRef.value || !shop.value?.sitespeed) return;

  if (config.chartInstance.value) {
    config.chartInstance.value.destroy();
  }

  const ctx = config.canvasRef.value.getContext("2d");
  if (!ctx) return;

  const sortedData = [...shop.value.sitespeed].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  );

  const datasets = config.datasets.map((dataset) => ({
    label: dataset.label,
    data: sortedData.map((item) => ({
      x: new Date(item.createdAt).getTime(),
      y: dataset.valueFormatter?.(item) ?? 0,
    })),
  }));

  const annotations: Record<
    string,
    {
      type: string;
      xMin: number;
      xMax: number;
      borderColor: string;
      borderWidth: number;
      borderDash: number[];
      label: {
        display: boolean;
        content: string;
        position: string;
        backgroundColor: string;
        color: string;
        font: { size: number };
        rotation: number;
        yAdjust: number;
      };
    }
  > = {};
  deploymentMarkers.value.forEach((deployment) => {
    annotations[`deployment-${deployment.id}`] = {
      type: "line",
      xMin: deployment.timestamp,
      xMax: deployment.timestamp,
      borderColor: "rgba(99, 102, 241, 0.8)",
      borderWidth: 2,
      borderDash: [5, 5],
      label: {
        display: true,
        content: deployment.name,
        position: "start",
        backgroundColor: "rgba(99, 102, 241, 0.8)",
        color: "white",
        font: {
          size: 10,
        },
        rotation: 0,
        yAdjust: -10,
      },
    };
  });

  config.chartInstance.value = new Chart(ctx, {
    type: "line",
    data: {
      datasets,
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: "index",
        intersect: false,
      },
      plugins: {
        annotation: {
          annotations,
        },
        title: {
          display: true,
          text: config.title,
        },
        legend: {
          display: true,
          position: "top",
        },
        tooltip: {
          callbacks: {
            label: function (context) {
              const dataset = config.datasets[context.datasetIndex];
              const value = context.parsed.y;
              return `${context.dataset.label}: ${dataset.tooltipFormatter?.(value) ?? value}`;
            },
            title: function (tooltipItems) {
              if (tooltipItems.length > 0) {
                const timestamp = tooltipItems[0].parsed.x;
                return new Date(timestamp).toLocaleString();
              }
              return "";
            },
          },
        },
      },
      scales: {
        x: {
          type: "time",
          time: {
            unit: "day",
            displayFormats: {
              day: "MMM dd",
            },
          },
          title: {
            display: true,
            text: "Date",
          },
        },
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: config.yAxisLabel,
          },
        },
      },
    },
  });
};

const createAllCharts = () => {
  chartConfigs.forEach((config) => createChart(config));
};

onMounted(async () => {
  await nextTick();
  if (shop.value?.sitespeed) {
    createAllCharts();
  }
});

onUnmounted(() => {
  chartConfigs.forEach((config) => {
    if (config.chartInstance.value) {
      config.chartInstance.value.destroy();
    }
  });
});

watch(
  () => shop.value?.sitespeed,
  async () => {
    await nextTick();
    createAllCharts();
  },
  { deep: true },
);
</script>
