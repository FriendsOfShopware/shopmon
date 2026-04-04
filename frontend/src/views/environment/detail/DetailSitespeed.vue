<template>
  <!-- Not enabled -->
  <Alert v-if="environment && !environment.sitespeedEnabled" variant="default">
    <AlertDescription>{{ $t("environment.sitespeedActivateDesc") }}</AlertDescription>
  </Alert>

  <div v-else-if="environment && environment.sitespeedEnabled" class="space-y-6">
    <!-- Latest metrics summary -->
    <div v-if="latest" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
      <Card v-for="metric in metricCards" :key="metric.label">
        <CardContent class="p-4">
          <div class="text-xs font-medium text-muted-foreground">{{ metric.label }}</div>
          <div class="mt-1 text-xl font-bold tabular-nums" :class="metric.color">{{ metric.value }}</div>
          <div v-if="metric.delta" class="mt-0.5 text-xs tabular-nums" :class="metric.deltaColor">
            {{ metric.delta }}
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- External link -->
    <div v-if="environment.sitespeedDetailUrl" class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">Performance over time</h2>
      <Button variant="outline" size="sm" as-child>
        <a :href="environment.sitespeedDetailUrl" target="_blank">
          <icon-fa6-solid:chart-line class="mr-1.5 size-3" />
          {{ $t("nav.sitespeed") }} details
          <icon-fa6-solid:arrow-up-right-from-square class="ml-1.5 size-2.5" />
        </a>
      </Button>
    </div>

    <!-- Charts -->
    <template v-if="showSitespeedData">
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-base">{{ $t("sitespeed.performanceOverTime") }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="h-72">
            <canvas ref="timeChartCanvas" />
          </div>
        </CardContent>
      </Card>

      <div class="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">{{ $t("sitespeed.transferSizeOverTime") }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="transferSizeChartCanvas" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">{{ $t("sitespeed.clsOverTime") }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="clsChartCanvas" />
            </div>
          </CardContent>
        </Card>
      </div>
    </template>

    <!-- History table -->
    <Card v-if="environment.sitespeeds?.length" class="overflow-hidden p-0">
      <CardHeader class="p-4 pb-0">
        <CardTitle class="text-base">History</CardTitle>
      </CardHeader>
      <CardContent class="p-0 pt-3">
        <DataTable
          :columns="[
            { key: 'createdAt', name: $t('sitespeed.checkedAt'), sortable: true },
            { key: 'deployment', name: $t('sitespeed.deployment') },
            { key: 'ttfb', name: 'TTFB', sortable: true },
            { key: 'fullyLoaded', name: 'Loaded', sortable: true },
            { key: 'largestContentfulPaint', name: 'LCP', sortable: true },
            { key: 'firstContentfulPaint', name: 'FCP', sortable: true },
            { key: 'cumulativeLayoutShift', name: 'CLS', sortable: true },
            { key: 'transferSize', name: 'Size', sortable: true },
          ]"
          :data="environment.sitespeeds"
        >
          <template #cell-createdAt="{ row }">
            <span class="tabular-nums text-muted-foreground">{{ formatDateTime(row.createdAt) }}</span>
          </template>
          <template #cell-deployment="{ row }">
            <RouterLink
              v-if="row.deployment"
              :to="{
                name: 'account.environments.detail.deployment',
                params: {
                  organizationId: $route.params.organizationId,
                  shopId: $route.params.environmentId,
                  deploymentId: row.deployment.id,
                },
              }"
              class="text-primary hover:underline"
            >
              {{ row.deployment.name || `#${row.deployment.id}` }}
            </RouterLink>
            <span v-else class="text-muted-foreground">—</span>
          </template>
          <template #cell-ttfb="{ row }">
            <span :class="msColor(row.ttfb, 800, 1800)" class="tabular-nums">{{ row.ttfb ? `${row.ttfb}ms` : "—" }}</span>
          </template>
          <template #cell-fullyLoaded="{ row }">
            <span :class="msColor(row.fullyLoaded, 3000, 6000)" class="tabular-nums">{{ row.fullyLoaded ? `${row.fullyLoaded}ms` : "—" }}</span>
          </template>
          <template #cell-largestContentfulPaint="{ row }">
            <span :class="msColor(row.largestContentfulPaint, 2500, 4000)" class="tabular-nums">{{ row.largestContentfulPaint ? `${row.largestContentfulPaint}ms` : "—" }}</span>
          </template>
          <template #cell-firstContentfulPaint="{ row }">
            <span :class="msColor(row.firstContentfulPaint, 1800, 3000)" class="tabular-nums">{{ row.firstContentfulPaint ? `${row.firstContentfulPaint}ms` : "—" }}</span>
          </template>
          <template #cell-cumulativeLayoutShift="{ row }">
            <span :class="clsColor(row.cumulativeLayoutShift)" class="tabular-nums">{{ row.cumulativeLayoutShift ?? "—" }}</span>
          </template>
          <template #cell-transferSize="{ row }">
            <span class="tabular-nums">{{ row.transferSize ? formatBytes(row.transferSize) : "—" }}</span>
          </template>
        </DataTable>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from "vue";
import type { Ref } from "vue";
import { Chart, registerables } from "chart.js";
import annotationPlugin from "chartjs-plugin-annotation";
import "chartjs-adapter-date-fns";
import { formatDateTime } from "@/helpers/formatter";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useI18n } from "vue-i18n";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import DataTable from "@/components/layout/DataTable.vue";

const { t } = useI18n();
const { environment } = useEnvironmentDetail();

Chart.register(...registerables, annotationPlugin);

const showSitespeedData = computed(() =>
  environment.value?.sitespeedEnabled && (environment.value?.sitespeeds?.length ?? 0) > 0,
);

// Latest entry
const latest = computed(() => {
  const entries = environment.value?.sitespeeds;
  if (!entries?.length) return null;
  return entries[0]; // already sorted newest first
});

const previous = computed(() => {
  const entries = environment.value?.sitespeeds;
  if (!entries || entries.length < 2) return null;
  return entries[1];
});

function formatMs(ms: number | null | undefined): string {
  if (ms == null) return "—";
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
  return `${ms}ms`;
}

function formatBytes(bytes: number, decimals = 1): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(decimals))} ${sizes[i]}`;
}

function delta(current: number | null | undefined, prev: number | null | undefined): { text: string; good: boolean } | null {
  if (current == null || prev == null || prev === 0) return null;
  const diff = current - prev;
  const pct = Math.round((diff / prev) * 100);
  if (pct === 0) return null;
  return { text: `${pct > 0 ? "+" : ""}${pct}%`, good: diff < 0 };
}

// Color helpers for Core Web Vitals thresholds
function msColor(val: number | null | undefined, good: number, poor: number): string {
  if (val == null) return "text-muted-foreground";
  if (val <= good) return "text-success";
  if (val <= poor) return "text-warning";
  return "text-destructive";
}

function clsColor(val: number | null | undefined): string {
  if (val == null) return "text-muted-foreground";
  if (val <= 0.1) return "text-success";
  if (val <= 0.25) return "text-warning";
  return "text-destructive";
}

const metricCards = computed(() => {
  if (!latest.value) return [];
  const prev = previous.value;

  const cards = [
    {
      label: "TTFB",
      value: formatMs(latest.value.ttfb),
      color: msColor(latest.value.ttfb, 800, 1800),
      ...(() => { const d = delta(latest.value!.ttfb, prev?.ttfb); return d ? { delta: d.text, deltaColor: d.good ? "text-success" : "text-destructive" } : {}; })(),
    },
    {
      label: "Fully Loaded",
      value: formatMs(latest.value.fullyLoaded),
      color: msColor(latest.value.fullyLoaded, 3000, 6000),
      ...(() => { const d = delta(latest.value!.fullyLoaded, prev?.fullyLoaded); return d ? { delta: d.text, deltaColor: d.good ? "text-success" : "text-destructive" } : {}; })(),
    },
    {
      label: "LCP",
      value: formatMs(latest.value.largestContentfulPaint),
      color: msColor(latest.value.largestContentfulPaint, 2500, 4000),
      ...(() => { const d = delta(latest.value!.largestContentfulPaint, prev?.largestContentfulPaint); return d ? { delta: d.text, deltaColor: d.good ? "text-success" : "text-destructive" } : {}; })(),
    },
    {
      label: "FCP",
      value: formatMs(latest.value.firstContentfulPaint),
      color: msColor(latest.value.firstContentfulPaint, 1800, 3000),
      ...(() => { const d = delta(latest.value!.firstContentfulPaint, prev?.firstContentfulPaint); return d ? { delta: d.text, deltaColor: d.good ? "text-success" : "text-destructive" } : {}; })(),
    },
    {
      label: "CLS",
      value: latest.value.cumulativeLayoutShift != null ? String(latest.value.cumulativeLayoutShift) : "—",
      color: clsColor(latest.value.cumulativeLayoutShift),
    },
    {
      label: "Transfer Size",
      value: latest.value.transferSize ? formatBytes(latest.value.transferSize) : "—",
      color: "",
      ...(() => { const d = delta(latest.value!.transferSize, prev?.transferSize); return d ? { delta: d.text, deltaColor: d.good ? "text-success" : "text-destructive" } : {}; })(),
    },
  ];
  return cards;
});

// ── Charts ──

const timeChartCanvas = ref<HTMLCanvasElement | null>(null);
const timeChartInstance = ref<Chart | null>(null);
const transferSizeChartCanvas = ref<HTMLCanvasElement | null>(null);
const transferSizeChartInstance = ref<Chart | null>(null);
const clsChartCanvas = ref<HTMLCanvasElement | null>(null);
const clsChartInstance = ref<Chart | null>(null);

const deploymentMarkers = computed(() => {
  if (!environment.value?.sitespeeds) return [];
  const map = new Map<number, { id: number; name: string; timestamp: number }>();
  environment.value.sitespeeds.forEach((item) => {
    if (item.deployment?.id && !map.has(item.deployment.id)) {
      map.set(item.deployment.id, {
        id: item.deployment.id,
        name: item.deployment.name || `#${item.deployment.id}`,
        timestamp: new Date(item.createdAt).getTime(),
      });
    }
  });
  return Array.from(map.values()).sort((a, b) => a.timestamp - b.timestamp);
});

interface SitespeedDataItem {
  createdAt: string;
  deployment?: { id: number; name: string } | null;
  ttfb?: number | null;
  fullyLoaded?: number | null;
  largestContentfulPaint?: number | null;
  firstContentfulPaint?: number | null;
  cumulativeLayoutShift?: number | null;
  transferSize?: number | null;
}

interface ChartConfig {
  canvasRef: Ref<HTMLCanvasElement | null>;
  chartInstance: Ref<Chart | null>;
  title: string;
  yAxisLabel: string;
  datasets: Array<{
    label: string;
    valueFormatter: (item: SitespeedDataItem) => number;
    tooltipFormatter: (value: number) => string;
  }>;
}

const chartConfigs: ChartConfig[] = [
  {
    canvasRef: timeChartCanvas,
    chartInstance: timeChartInstance,
    title: t("sitespeed.performanceOverTime"),
    yAxisLabel: t("sitespeed.timeMs"),
    datasets: [
      { label: t("sitespeed.ttfb"), valueFormatter: (i) => i.ttfb ?? 0, tooltipFormatter: (v) => `${v}ms` },
      { label: t("sitespeed.fullyLoaded"), valueFormatter: (i) => i.fullyLoaded ?? 0, tooltipFormatter: (v) => `${v}ms` },
      { label: t("sitespeed.lcp"), valueFormatter: (i) => i.largestContentfulPaint ?? 0, tooltipFormatter: (v) => `${v}ms` },
      { label: t("sitespeed.fcp"), valueFormatter: (i) => i.firstContentfulPaint ?? 0, tooltipFormatter: (v) => `${v}ms` },
    ],
  },
  {
    canvasRef: transferSizeChartCanvas,
    chartInstance: transferSizeChartInstance,
    title: t("sitespeed.transferSizeOverTime"),
    yAxisLabel: t("sitespeed.sizeKb"),
    datasets: [
      { label: t("sitespeed.transferSize"), valueFormatter: (i) => i.transferSize ? Math.round(i.transferSize / 1024) : 0, tooltipFormatter: (v) => `${v} KB` },
    ],
  },
  {
    canvasRef: clsChartCanvas,
    chartInstance: clsChartInstance,
    title: t("sitespeed.clsOverTime"),
    yAxisLabel: t("sitespeed.clsScore"),
    datasets: [
      { label: t("sitespeed.cls"), valueFormatter: (i) => i.cumulativeLayoutShift ?? 0, tooltipFormatter: (v) => `${v}` },
    ],
  },
];

function createChart(config: ChartConfig) {
  if (!config.canvasRef.value || !environment.value?.sitespeeds) return;
  config.chartInstance.value?.destroy();

  const ctx = config.canvasRef.value.getContext("2d");
  if (!ctx) return;

  const sorted = [...environment.value.sitespeeds].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  );

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const annotations: Record<string, any> = {};
  deploymentMarkers.value.forEach((d) => {
    annotations[`deployment-${d.id}`] = {
      type: "line",
      xMin: d.timestamp,
      xMax: d.timestamp,
      borderColor: "rgba(99, 102, 241, 0.8)",
      borderWidth: 2,
      borderDash: [5, 5],
      label: { display: true, content: d.name, position: "start", backgroundColor: "rgba(99, 102, 241, 0.8)", color: "white", font: { size: 10 }, rotation: 0, yAdjust: -10 },
    };
  });

  config.chartInstance.value = new Chart(ctx, {
    type: "line",
    data: {
      datasets: config.datasets.map((ds) => ({
        label: ds.label,
        data: sorted.map((item) => ({ x: new Date(item.createdAt).getTime(), y: ds.valueFormatter(item) })),
      })),
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        annotation: { annotations },
        title: { display: false },
        legend: { display: true, position: "top" },
        tooltip: {
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${config.datasets[ctx.datasetIndex].tooltipFormatter(ctx.parsed.y ?? 0)}`,
            title: (items) => items.length ? new Date(items[0].parsed.x).toLocaleString() : "",
          },
        },
      },
      scales: {
        x: { type: "time", time: { unit: "day", displayFormats: { day: "MMM dd" } }, title: { display: false } },
        y: { beginAtZero: true, title: { display: true, text: config.yAxisLabel } },
      },
    },
  });
}

function createAllCharts() {
  chartConfigs.forEach((c) => createChart(c));
}

onMounted(async () => {
  await nextTick();
  if (environment.value?.sitespeeds) createAllCharts();
});

onUnmounted(() => {
  chartConfigs.forEach((c) => c.chartInstance.value?.destroy());
});

watch(() => environment.value?.sitespeeds, async () => {
  await nextTick();
  createAllCharts();
}, { deep: true });
</script>
