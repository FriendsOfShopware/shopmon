<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("ecosystem.title") }}</h1>

    <div v-if="loading" class="grid gap-6 md:grid-cols-2">
      <Card v-for="i in 3" :key="i" class="animate-pulse">
        <CardHeader class="h-16" />
        <CardContent class="h-48" />
      </Card>
    </div>

    <div
      v-else-if="error"
      class="rounded-md border border-destructive/50 bg-destructive/10 p-4 text-destructive"
    >
      {{ error }}
    </div>

    <template v-else>
      <div class="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center justify-between text-lg">
              <span>{{ $t("ecosystem.environmentsGrowth") }}</span>
              <span
                v-if="growthData?.environments.length"
                class="text-2xl font-bold text-emerald-500"
              >
                {{ growthData.environments[growthData.environments.length - 1]?.count ?? 0 }}
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-64">
              <canvas ref="shopChartCanvas" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="flex items-center justify-between text-lg">
              <span>{{ $t("ecosystem.userGrowth") }}</span>
              <span v-if="growthData?.users.length" class="text-2xl font-bold text-indigo-500">
                {{ growthData.users[growthData.users.length - 1]?.count ?? 0 }}
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-64">
              <canvas ref="userChartCanvas" />
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader class="pb-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <CardTitle class="text-lg">{{ $t("ecosystem.shopwareDistribution") }}</CardTitle>
            <div class="flex items-center gap-2">
              <span class="text-xs text-muted-foreground">{{ $t("ecosystem.groupByMinor") }}</span>
              <Switch v-model:checked="groupByMinor" />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div v-if="!versionData?.length" class="py-6 text-center text-muted-foreground">
            {{ $t("ecosystem.noVersionData") }}
          </div>

          <div v-else class="space-y-3">
            <div v-for="v in aggregatedVersions" :key="v.version" class="space-y-1">
              <div class="flex items-center justify-between text-sm">
                <span class="font-mono font-medium">Shopware {{ v.version }}</span>
                <span class="text-xs text-muted-foreground">
                  {{ v.count }} ({{ Math.round((v.count / totalVersionCount) * 100) }}%)
                </span>
              </div>
              <div class="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  class="h-full bg-primary transition-all duration-300"
                  :style="{ width: `${(v.count / totalVersionCount) * 100}%` }"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { getEcosystemStats, type AdminGrowth, type ShopwareVersionCount } from "@/api/generated";
import { onMounted, onUnmounted, ref, nextTick, computed } from "vue";
import { Chart, registerables } from "chart.js";

Chart.register(...registerables);

type GrowthData = AdminGrowth;
type VersionData = ShopwareVersionCount[];

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
    const { data, error: resError } = await getEcosystemStats();
    if (resError || !data) {
      throw new Error("request failed");
    }
    growthData.value = data.growth;
    versionData.value = data.shopwareVersions;

    await nextTick();
    renderCharts();
  } catch {
    error.value = "Failed to load ecosystem statistics";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadStats();
});

onUnmounted(() => {
  userChartInstance?.destroy();
  shopChartInstance?.destroy();
});
</script>
