<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.dashboard") }}</h1>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-16 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingStats") }}
    </div>

    <template v-if="!loading && stats">
      <!-- Stat cards -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card v-for="stat in statCards" :key="stat.label">
          <CardContent class="flex items-center gap-3 p-4">
            <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
              <component :is="stat.icon" class="size-4 text-primary" />
            </div>
            <div>
              <div class="text-2xl font-bold tabular-nums">{{ stat.value }}</div>
              <div class="text-xs text-muted-foreground">{{ stat.label }}</div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Status breakdown -->
      <div class="flex flex-wrap items-center gap-3">
        <span class="text-sm text-muted-foreground">Environment health:</span>
        <div v-for="(count, status) in stats.environmentsByStatus" :key="status" class="flex items-center gap-1.5">
          <StatusIcon :status="String(status)" />
          <span class="text-sm font-medium tabular-nums">{{ count }}</span>
        </div>
      </div>

      <!-- Charts row -->
      <div v-if="growthData" class="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">User Growth</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="userChartCanvas" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-2">
            <CardTitle class="text-base">Environment Growth</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="shopChartCanvas" />
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Version distribution + recent activity -->
      <div class="grid gap-4 lg:grid-cols-3">
        <!-- Version distribution -->
        <Card v-if="versionData?.length">
          <CardHeader class="pb-3">
            <CardTitle class="text-base">Shopware Versions</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="space-y-2">
              <div v-for="v in versionData" :key="v.version" class="flex items-center gap-3">
                <Badge variant="secondary" class="min-w-[5rem] justify-center font-mono text-xs">{{ v.version }}</Badge>
                <div class="flex-1">
                  <div class="h-2 overflow-hidden rounded-full bg-muted">
                    <div class="h-full rounded-full bg-primary transition-all" :style="{ width: `${(v.count / totalEnvCount) * 100}%` }" />
                  </div>
                </div>
                <span class="w-6 text-right text-xs tabular-nums text-muted-foreground">{{ v.count }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Recent signups -->
        <Card>
          <CardHeader class="pb-3">
            <CardTitle class="text-base">Recent Signups</CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="!activity?.recentUsers?.length" class="flex flex-col items-center gap-2 py-6 text-center text-muted-foreground">
              <icon-fa6-solid:users class="size-6 opacity-30" />
              <p class="text-sm">No recent signups</p>
            </div>
            <div v-else class="space-y-1.5">
              <div v-for="user in activity.recentUsers" :key="user.id" class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-accent">
                <div class="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
                  <icon-fa6-solid:user class="size-3 text-primary" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium">{{ user.displayName }}</div>
                  <div class="truncate text-xs text-muted-foreground">{{ user.email }}</div>
                </div>
                <span class="shrink-0 text-xs tabular-nums text-muted-foreground">{{ formatDate(user.createdAt) }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Recent environments -->
        <Card>
          <CardHeader class="pb-3">
            <CardTitle class="text-base">Recent Environments</CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="!activity?.recentEnvironments?.length" class="flex flex-col items-center gap-2 py-6 text-center text-muted-foreground">
              <icon-fa6-solid:earth-americas class="size-6 opacity-30" />
              <p class="text-sm">No recent environments</p>
            </div>
            <div v-else class="space-y-1.5">
              <div v-for="env in activity.recentEnvironments" :key="env.id" class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-accent">
                <div class="flex size-8 shrink-0 items-center justify-center rounded-lg border bg-muted">
                  <icon-fa6-solid:earth-americas class="size-3 text-muted-foreground" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium">{{ env.name }}</div>
                  <div class="truncate text-xs text-muted-foreground">{{ env.organizationName }}</div>
                </div>
                <StatusIcon :status="env.status" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Quick actions -->
      <div class="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" as-child>
          <RouterLink to="/admin/users">
            <icon-fa6-solid:users class="mr-1.5 size-3" />
            {{ $t("admin.manageOrgs") }}
          </RouterLink>
        </Button>
        <Button variant="outline" size="sm" as-child>
          <RouterLink to="/admin/environments">
            <icon-fa6-solid:earth-americas class="mr-1.5 size-3" />
            {{ $t("admin.manageEnvironments") }}
          </RouterLink>
        </Button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import StatusIcon from "@/components/StatusIcon.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { formatDate } from "@/helpers/formatter";
import { onMounted, onUnmounted, ref, nextTick, watch, computed } from "vue";
import { Chart, registerables } from "chart.js";

import FaUsers from "~icons/fa6-solid/users";
import FaBuilding from "~icons/fa6-solid/building";
import FaEarth from "~icons/fa6-solid/earth-americas";

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

const statCards = computed(() => {
  if (!stats.value) return [];
  return [
    { icon: FaUsers, value: stats.value.totalUsers, label: "Users" },
    { icon: FaBuilding, value: stats.value.totalOrganizations, label: "Organizations" },
    { icon: FaEarth, value: stats.value.totalEnvironments, label: "Environments" },
  ];
});

const totalEnvCount = computed(() => stats.value?.totalEnvironments ?? 1);

const userChartCanvas = ref<HTMLCanvasElement | null>(null);
const shopChartCanvas = ref<HTMLCanvasElement | null>(null);
let userChartInstance: Chart | null = null;
let shopChartInstance: Chart | null = null;

function createGrowthChart(canvas: HTMLCanvasElement, data: { month: string; count: number }[], color: string) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;

  return new Chart(ctx, {
    type: "line",
    data: {
      labels: data.map((d) => d.month),
      datasets: [{
        data: data.map((d) => d.count),
        borderColor: color,
        backgroundColor: color + "20",
        fill: true,
        tension: 0.3,
        pointRadius: 3,
        pointHoverRadius: 5,
      }],
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
      userChartInstance = createGrowthChart(userChartCanvas.value, growthData.value.users, "#6366f1");
    }
    if (shopChartCanvas.value) {
      shopChartInstance = createGrowthChart(shopChartCanvas.value, growthData.value.environments, "#10b981");
    }
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

watch(growthData, async () => {
  await nextTick();
  renderCharts();
});

onUnmounted(() => {
  userChartInstance?.destroy();
  shopChartInstance?.destroy();
});
</script>
