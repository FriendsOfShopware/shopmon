<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("ecosystem.title") }}</h1>
      <p class="text-sm text-muted-foreground">{{ $t("ecosystem.description") }}</p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-16 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("ecosystem.loading") }}
    </div>

    <template v-if="!loading && !error">
      <!-- Growth charts -->
      <div v-if="growthData" class="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">{{ $t("admin.userGrowth") }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="userChartCanvas" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">{{ $t("admin.environmentGrowth") }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="shopChartCanvas" />
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Version distribution -->
      <Card v-if="versionData?.length">
        <CardHeader class="flex-row items-center justify-between gap-4 space-y-0 pb-3">
          <CardTitle class="text-base">{{ $t("admin.shopwareVersions") }}</CardTitle>
          <label class="flex items-center gap-2 text-sm font-normal text-muted-foreground">
            <Switch v-model="groupByMinor" />
            {{ $t("ecosystem.groupByMinor") }}
          </label>
        </CardHeader>
        <CardContent>
          <div class="space-y-2">
            <div v-for="v in aggregatedVersions" :key="v.version" class="flex items-center gap-3">
              <Badge variant="secondary" class="min-w-[5rem] justify-center font-mono text-xs">{{
                v.version
              }}</Badge>
              <div class="flex-1">
                <div class="h-2 overflow-hidden rounded-full bg-muted">
                  <div
                    class="h-full rounded-full bg-primary transition-all"
                    :style="{ width: `${(v.count / totalVersionCount) * 100}%` }"
                  />
                </div>
              </div>
              <span class="w-10 text-right text-xs tabular-nums text-muted-foreground">{{
                v.count
              }}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { onMounted, onUnmounted, ref, nextTick, watch, computed } from "vue";
import { Chart, registerables } from "chart.js";

Chart.register(...registerables);

type GrowthData = components["schemas"]["AdminGrowth"];
type VersionData = components["schemas"]["ShopwareVersionCount"][];

const growthData = ref<GrowthData | null>(null);
const versionData = ref<VersionData | null>(null);
const loading = ref(true);
const error = ref("");

// Default: aggregate by major.minor (e.g. "6.7"). When enabled: by
// major.minor.patch (e.g. "6.7.8").
const groupByMinor = ref(false);

// Truncate a full version string (e.g. "6.7.8.2") to the desired precision.
function versionKey(version: string, segments: number): string {
  return version.split(".").slice(0, segments).join(".");
}

const aggregatedVersions = computed(() => {
  const segments = groupByMinor.value ? 3 : 2;
  const counts = new Map<string, number>();
  for (const v of versionData.value ?? []) {
    const key = versionKey(v.version, segments);
    counts.set(key, (counts.get(key) ?? 0) + v.count);
  }
  return [...counts.entries()]
    .map(([version, count]) => ({ version, count }))
    .sort((a, b) => b.count - a.count);
});

const totalVersionCount = computed(() =>
  Math.max(versionData.value?.reduce((sum, v) => sum + v.count, 0) ?? 0, 1),
);

const userChartCanvas = ref<HTMLCanvasElement | null>(null);
const shopChartCanvas = ref<HTMLCanvasElement | null>(null);
let userChartInstance: Chart | null = null;
let shopChartInstance: Chart | null = null;

function createGrowthChart(
  canvas: HTMLCanvasElement,
  data: { month: string; count: number }[],
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
      plugins: { legend: { display: false }, title: { display: false } },
      scales: {
        x: { title: { display: false } },
        y: { beginAtZero: true, ticks: { precision: 0 } },
      },
    },
  });
}

function renderCharts() {
  userChartInstance?.destroy();
  shopChartInstance?.destroy();

  if (growthData.value) {
    if (userChartCanvas.value) {
      userChartInstance = createGrowthChart(
        userChartCanvas.value,
        growthData.value.users,
        "#6366f1",
      );
    }
    if (shopChartCanvas.value) {
      shopChartInstance = createGrowthChart(
        shopChartCanvas.value,
        growthData.value.environments,
        "#10b981",
      );
    }
  }
}

async function loadStats() {
  loading.value = true;
  error.value = "";

  try {
    const { data, error: resError } = await api.GET("/info/ecosystem");
    if (resError || !data) {
      throw new Error("request failed");
    }
    growthData.value = data.growth;
    versionData.value = data.shopwareVersions;
  } catch (err) {
    error.value = `Failed to load ecosystem stats: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await loadStats();
  await nextTick();
  renderCharts();
});

watch(growthData, async () => {
  await nextTick();
  renderCharts();
});

onUnmounted(() => {
  userChartInstance?.destroy();
  shopChartInstance?.destroy();
});
</script>
