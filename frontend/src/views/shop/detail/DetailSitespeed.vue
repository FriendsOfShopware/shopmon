<template>
  <Alert v-if="shop && !shop.sitespeedEnabled" type="info">
    {{ $t('shop.sitespeedActivateDesc') }}
  </Alert>
  <div class="mb-1">
    <a
      v-if="showSitespeedData"
      class="btn btn-primary-outline"
      :href="shop.sitespeedReportUrl"
      target="_blank"
    >
      <i class="fa fa-chart-line" /> {{ $t('nav.sitespeed') }}
    </a>
  </div>

  <Panel v-if="showSitespeedData" class="chart-panel">
    <canvas ref="timeChartCanvas" width="800" height="400" />
  </Panel>

  <Panel v-if="showSitespeedData" class="chart-panel">
    <canvas ref="transferSizeChartCanvas" width="800" height="400" />
  </Panel>

  <Panel v-if="showSitespeedData" class="chart-panel">
    <canvas ref="clsChartCanvas" width="800" height="400" />
  </Panel>

  <Panel v-if="shop && shop.sitespeedEnabled" variant="table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'createdAt', name: $t('sitespeed.checkedAt') },
        { key: 'deployment', name: $t('sitespeed.deployment') },
        { key: 'ttfb', name: $t('sitespeed.ttfb') },
        { key: 'fullyLoaded', name: $t('sitespeed.fullyLoaded') },
        { key: 'largestContentfulPaint', name: $t('sitespeed.lcpFull') },
        { key: 'firstContentfulPaint', name: $t('sitespeed.fcpFull') },
        { key: 'cumulativeLayoutShift', name: $t('sitespeed.cls') },
        { key: 'transferSize', name: $t('sitespeed.transferSize') },
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
  </Panel>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, Ref, computed } from "vue";
import { Chart, registerables } from "chart.js";
import annotationPlugin from "chartjs-plugin-annotation";
import "chartjs-adapter-date-fns";
import { formatDateTime } from "@/helpers/formatter";
import { useShopDetail } from "@/composables/useShopDetail";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
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
  { key: "ttfb", label: t("sitespeed.ttfb") },
  { key: "fullyLoaded", label: t("sitespeed.fullyLoaded") },
  { key: "largestContentfulPaint", label: t("sitespeed.lcp") },
  { key: "firstContentfulPaint", label: t("sitespeed.fcp") },
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
    title: t("sitespeed.performanceOverTime"),
    yAxisLabel: t("sitespeed.timeMs"),
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
    title: t("sitespeed.transferSizeOverTime"),
    yAxisLabel: t("sitespeed.sizeKb"),
    datasets: [
      {
        label: t("sitespeed.transferSize"),
        valueFormatter: (item) => (item.transferSize ? Math.round(item.transferSize / 1024) : 0),
        tooltipFormatter: (value) => `${value} KB`,
      },
    ],
  },
  {
    canvasRef: clsChartCanvas,
    chartInstance: clsChartInstance,
    title: t("sitespeed.clsOverTime"),
    yAxisLabel: t("sitespeed.clsScore"),
    datasets: [
      {
        label: t("sitespeed.cls"),
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
            text: t("sitespeed.date"),
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
