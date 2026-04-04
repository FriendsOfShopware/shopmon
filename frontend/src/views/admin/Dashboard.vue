<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.dashboard") }}</h1>
  </div>

  <Card>
    <CardContent>
      <Alert v-if="error" variant="destructive">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <div
        v-if="loading"
        class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
      >
        <icon-line-md:loading-twotone-loop class="size-6 animate-spin" />
        {{ $t("admin.loadingStats") }}
      </div>

      <div
        v-if="!loading && stats"
        class="grid grid-cols-[repeat(auto-fit,minmax(250px,1fr))] gap-8 mb-8"
      >
        <div
          class="flex flex-col rounded-lg border bg-card p-6 transition-all hover:border-primary hover:shadow-md"
        >
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {{ $t("admin.totalUsers") }}
            </h3>
            <icon-fa6-solid:users class="size-6 text-primary opacity-70" />
          </div>
          <div class="text-2xl sm:text-4xl font-bold my-2">{{ stats.totalUsers }}</div>
          <p class="text-sm text-muted-foreground">{{ $t("admin.totalUsersDesc") }}</p>
        </div>

        <div
          class="flex flex-col rounded-lg border bg-card p-6 transition-all hover:border-primary hover:shadow-md"
        >
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {{ $t("admin.totalOrgs") }}
            </h3>
            <icon-fa6-solid:building class="size-6 text-primary opacity-70" />
          </div>
          <div class="text-2xl sm:text-4xl font-bold my-2">{{ stats.totalOrganizations }}</div>
          <p class="text-sm text-muted-foreground">{{ $t("admin.totalOrgsDesc") }}</p>
        </div>

        <div
          class="flex flex-col rounded-lg border bg-card p-6 transition-all hover:border-primary hover:shadow-md"
        >
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {{ $t("admin.totalEnvironments") }}
            </h3>
            <icon-fa6-solid:store class="size-6 text-primary opacity-70" />
          </div>
          <div class="text-2xl sm:text-4xl font-bold my-2">{{ stats.totalEnvironments }}</div>
          <p class="text-sm text-muted-foreground">{{ $t("admin.totalEnvironmentsDesc") }}</p>
        </div>

        <div
          class="flex flex-col rounded-lg border bg-card p-6 transition-all hover:border-primary hover:shadow-md"
        >
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {{ $t("admin.environmentStatus") }}
            </h3>
            <icon-fa6-solid:chart-bar class="size-6 text-primary opacity-70" />
          </div>
          <div class="flex flex-col gap-3">
            <div
              v-for="(count, status) in stats.environmentsByStatus"
              :key="status"
              class="flex items-center justify-between rounded-md bg-background p-3"
            >
              <div class="flex-1">
                <Badge
                  :class="{
                    'bg-emerald-500 text-white border-transparent': status === 'green',
                    'bg-amber-500 text-white border-transparent': status === 'yellow',
                    'bg-red-500 text-white border-transparent': status === 'red',
                  }"
                >
                  {{ status }}
                </Badge>
              </div>
              <div class="text-xl font-bold">{{ count }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Links -->
      <div v-if="!loading && stats" class="flex flex-wrap gap-4">
        <Button as-child>
          <RouterLink to="/admin/organizations">
            {{ $t("admin.manageOrgs") }}
          </RouterLink>
        </Button>
        <Button as-child>
          <RouterLink to="/admin/environments">
            {{ $t("admin.manageEnvironments") }}
          </RouterLink>
        </Button>
      </div>
    </CardContent>
  </Card>

  <!-- Growth Charts -->
  <div v-if="growthData" class="grid grid-cols-[repeat(auto-fit,minmax(300px,1fr))] gap-8">
    <Card>
      <CardHeader>
        <CardTitle>User Growth</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="relative h-[300px]">
          <canvas ref="userChartCanvas" />
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Shop Growth</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="relative h-[300px]">
          <canvas ref="shopChartCanvas" />
        </div>
      </CardContent>
    </Card>
  </div>

  <!-- Shopware Version Distribution -->
  <Card v-if="versionData && versionData.length > 0">
    <CardHeader>
      <CardTitle>Shopware Version Distribution</CardTitle>
    </CardHeader>
    <CardContent>
      <div class="relative h-[300px]">
        <canvas ref="versionChartCanvas" />
      </div>
    </CardContent>
  </Card>

  <!-- Recent Activity -->
  <div v-if="activity" class="grid grid-cols-[repeat(auto-fit,minmax(300px,1fr))] gap-8">
    <Card>
      <CardHeader>
        <CardTitle>Recent Signups</CardTitle>
      </CardHeader>
      <CardContent>
        <div
          v-if="activity.recentUsers.length === 0"
          class="text-center py-8 text-muted-foreground"
        >
          No recent signups
        </div>
        <div v-else class="flex flex-col">
          <div
            v-for="user in activity.recentUsers"
            :key="user.id"
            class="flex items-center justify-between py-3 border-b last:border-b-0"
          >
            <div class="flex flex-col gap-0.5 min-w-0">
              <span class="font-semibold truncate">{{ user.displayName }}</span>
              <span class="text-xs text-muted-foreground truncate">{{ user.email }}</span>
            </div>
            <span class="text-xs text-muted-foreground whitespace-nowrap ml-4">{{
              formatDateTime(user.createdAt)
            }}</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Recent Environments</CardTitle>
      </CardHeader>
      <CardContent>
        <div
          v-if="activity.recentEnvironments.length === 0"
          class="text-center py-8 text-muted-foreground"
        >
          No recent environments
        </div>
        <div v-else class="flex flex-col">
          <div
            v-for="env in activity.recentEnvironments"
            :key="env.id"
            class="flex items-center justify-between py-3 border-b last:border-b-0"
          >
            <div class="flex flex-col gap-0.5 min-w-0">
              <span class="font-semibold truncate">{{ env.name }}</span>
              <span class="text-xs text-muted-foreground truncate">{{ env.organizationName }}</span>
            </div>
            <span class="text-xs text-muted-foreground whitespace-nowrap ml-4">{{
              env.lastScrapedAt ? formatDateTime(env.lastScrapedAt) : ""
            }}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
