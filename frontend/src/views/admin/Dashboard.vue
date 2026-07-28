<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.dashboard") }}</h1>
        <p class="text-xs text-muted-foreground mt-0.5">
          System overview, metrics, and administration insights
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          class="h-8 gap-1.5 text-xs"
          :disabled="loading"
          @click="loadStats"
        >
          <icon-fa6-solid:rotate class="size-3" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
      </div>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div
      v-if="loading && !stats"
      class="flex items-center justify-center gap-2 py-16 text-muted-foreground"
    >
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("admin.loadingStats") }}
    </div>

    <template v-if="stats">
      <!-- Stat cards -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card
          v-for="stat in statCards"
          :key="stat.label"
          class="transition-all duration-200 hover:border-primary/30 hover:shadow-md"
        >
          <CardContent class="flex items-center gap-3.5 p-4">
            <div
              :class="[
                'flex size-11 shrink-0 items-center justify-center rounded-xl transition-transform duration-200 group-hover:scale-105',
                stat.bgClass,
              ]"
            >
              <component :is="stat.icon" :class="['size-5', stat.textClass]" />
            </div>
            <div>
              <div class="text-2xl font-bold tabular-nums tracking-tight">{{ stat.value }}</div>
              <div class="text-xs font-medium text-muted-foreground">{{ stat.label }}</div>
            </div>
          </CardContent>
        </Card>

        <Card class="transition-all duration-200 hover:border-primary/30 hover:shadow-md">
          <CardContent class="flex items-center gap-3.5 p-4">
            <div
              class="flex size-11 shrink-0 items-center justify-center rounded-xl bg-rose-500/10 text-rose-500"
            >
              <FaHeartPulse class="size-5" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="text-2xl font-bold tabular-nums tracking-tight">
                {{ stats.totalEnvironments }}
              </div>
              <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                <span class="text-xs font-medium text-muted-foreground">{{
                  $t("admin.environmentHealth")
                }}</span>
                <div v-if="stats.environmentsByStatus" class="flex items-center gap-1.5">
                  <div
                    v-for="(count, status) in stats.environmentsByStatus"
                    :key="status"
                    class="flex items-center gap-1 rounded-md px-1.5 py-0.5 bg-muted/60"
                  >
                    <StatusIcon :status="String(status)" />
                    <span class="text-[11px] font-semibold tabular-nums">{{ count }}</span>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Charts row -->
      <div v-if="growthData" class="grid gap-4 lg:grid-cols-2">
        <Card class="transition-all duration-200 hover:border-border/80">
          <CardHeader class="pb-2">
            <div class="flex items-center justify-between">
              <CardTitle class="text-base font-semibold">{{ $t("admin.userGrowth") }}</CardTitle>
              <Badge variant="outline" class="text-[10px] text-muted-foreground">Monthly</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div class="h-56">
              <canvas ref="userChartCanvas" />
            </div>
          </CardContent>
        </Card>

        <Card class="transition-all duration-200 hover:border-border/80">
          <CardHeader class="pb-2">
            <div class="flex items-center justify-between">
              <CardTitle class="text-base font-semibold">{{
                $t("admin.environmentGrowth")
              }}</CardTitle>
              <Badge variant="outline" class="text-[10px] text-muted-foreground">Monthly</Badge>
            </div>
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
            <CardTitle class="text-base font-semibold">{{
              $t("admin.shopwareVersions")
            }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="space-y-3">
              <div v-for="v in versionData" :key="v.version" class="flex items-center gap-3">
                <Badge
                  variant="secondary"
                  class="min-w-20 justify-center font-mono text-xs font-semibold"
                  >{{ v.version }}</Badge
                >
                <div class="flex-1">
                  <div class="h-2.5 overflow-hidden rounded-full bg-muted">
                    <div
                      class="h-full rounded-full bg-primary transition-all duration-500"
                      :style="{ width: `${(v.count / totalEnvCount) * 100}%` }"
                    />
                  </div>
                </div>
                <span class="w-8 text-right text-xs font-semibold tabular-nums text-foreground">{{
                  v.count
                }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Recent signups -->
        <Card>
          <CardHeader class="pb-3">
            <div class="flex items-center justify-between">
              <CardTitle class="text-base font-semibold">{{ $t("admin.recentSignups") }}</CardTitle>
              <RouterLink
                to="/admin/users"
                class="text-xs font-medium text-primary hover:underline"
              >
                View all
              </RouterLink>
            </div>
          </CardHeader>
          <CardContent>
            <div
              v-if="!activity?.recentUsers?.length"
              class="flex flex-col items-center gap-2 py-6 text-center text-muted-foreground"
            >
              <icon-fa6-solid:users class="size-6 opacity-30" />
              <p class="text-sm">{{ $t("admin.noRecentSignups") }}</p>
            </div>
            <div v-else class="space-y-1.5">
              <RouterLink
                v-for="user in activity.recentUsers"
                :key="user.id"
                :to="{ name: 'admin.users.detail', params: { id: user.id } }"
                class="group flex items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-accent/70"
              >
                <div
                  class="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 transition-transform group-hover:scale-105"
                >
                  <icon-fa6-solid:user class="size-3 text-primary" />
                </div>
                <div class="min-w-0 flex-1">
                  <div
                    class="truncate text-sm font-medium transition-colors group-hover:text-primary"
                  >
                    {{ user.displayName }}
                  </div>
                  <div class="truncate text-xs text-muted-foreground">{{ user.email }}</div>
                </div>
                <span class="shrink-0 text-[11px] tabular-nums text-muted-foreground">{{
                  formatDate(user.createdAt)
                }}</span>
              </RouterLink>
            </div>
          </CardContent>
        </Card>

        <!-- Recent environments -->
        <Card>
          <CardHeader class="pb-3">
            <div class="flex items-center justify-between">
              <CardTitle class="text-base font-semibold">{{
                $t("admin.recentEnvironments")
              }}</CardTitle>
              <RouterLink
                to="/admin/environments"
                class="text-xs font-medium text-primary hover:underline"
              >
                View all
              </RouterLink>
            </div>
          </CardHeader>
          <CardContent>
            <div
              v-if="!activity?.recentEnvironments?.length"
              class="flex flex-col items-center gap-2 py-6 text-center text-muted-foreground"
            >
              <icon-fa6-solid:earth-americas class="size-6 opacity-30" />
              <p class="text-sm">{{ $t("admin.noRecentEnvironments") }}</p>
            </div>
            <div v-else class="space-y-1.5">
              <RouterLink
                v-for="env in activity.recentEnvironments"
                :key="env.id"
                :to="{ name: 'admin.environments.detail', params: { id: env.id } }"
                class="group flex items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-accent/70"
              >
                <div
                  class="flex size-8 shrink-0 items-center justify-center rounded-lg border bg-muted transition-transform group-hover:scale-105"
                >
                  <icon-fa6-solid:earth-americas
                    class="size-3 text-muted-foreground group-hover:text-primary"
                  />
                </div>
                <div class="min-w-0 flex-1">
                  <div
                    class="truncate text-sm font-medium transition-colors group-hover:text-primary"
                  >
                    {{ env.name }}
                  </div>
                  <div class="truncate text-xs text-muted-foreground">
                    {{ env.organizationName }}
                  </div>
                </div>
                <StatusIcon :status="env.status" />
              </RouterLink>
            </div>
          </CardContent>
        </Card>
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
import {
  adminGetGrowth,
  adminGetRecentActivity,
  adminGetShopwareVersions,
  adminGetStats,
  type AdminGrowth,
  type AdminRecentActivity,
  type AdminStats,
  type ShopwareVersionCount,
} from "@/api/generated";
import { formatDate } from "@/helpers/formatter";
import { onMounted, onUnmounted, ref, nextTick, watch, computed } from "vue";
import { useI18n } from "vue-i18n";
import { Chart, registerables } from "chart.js";

import FaUsers from "~icons/fa6-solid/users";
import FaBuilding from "~icons/fa6-solid/building";
import FaEarth from "~icons/fa6-solid/earth-americas";
import FaHeartPulse from "~icons/fa6-solid/heart-pulse";
import FaRotate from "~icons/fa6-solid/rotate";
import FaUser from "~icons/fa6-solid/user";
import FaClockRotateLeft from "~icons/fa6-solid/clock-rotate-left";

Chart.register(...registerables);

const { t } = useI18n();

type Stats = AdminStats;
type GrowthData = AdminGrowth;
type Activity = AdminRecentActivity;
type VersionData = ShopwareVersionCount[];

const stats = ref<Stats | null>(null);
const growthData = ref<GrowthData | null>(null);
const activity = ref<Activity | null>(null);
const versionData = ref<VersionData | null>(null);
const loading = ref(true);
const error = ref("");

const statCards = computed(() => {
  if (!stats.value) return [];
  return [
    {
      icon: FaUsers,
      value: stats.value.totalUsers,
      label: t("common.users"),
      bgClass: "bg-indigo-500/10",
      textClass: "text-indigo-500",
    },
    {
      icon: FaBuilding,
      value: stats.value.totalOrganizations,
      label: t("common.organizations"),
      bgClass: "bg-sky-500/10",
      textClass: "text-sky-500",
    },
    {
      icon: FaEarth,
      value: stats.value.totalEnvironments,
      label: t("common.environments"),
      bgClass: "bg-emerald-500/10",
      textClass: "text-emerald-500",
    },
  ];
});

const totalEnvCount = computed(() => stats.value?.totalEnvironments ?? 1);

const userChartCanvas = ref<HTMLCanvasElement | null>(null);
const shopChartCanvas = ref<HTMLCanvasElement | null>(null);
let userChartInstance: Chart | null = null;
let shopChartInstance: Chart | null = null;

function createGrowthChart(
  canvas: HTMLCanvasElement,
  data: { month: string; count: number }[],
  primaryColor: string,
  gradientColor: string,
) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;

  const gradient = ctx.createLinearGradient(0, 0, 0, 220);
  gradient.addColorStop(0, gradientColor);
  gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

  return new Chart(ctx, {
    type: "line",
    data: {
      labels: data.map((d) => d.month),
      datasets: [
        {
          data: data.map((d) => d.count),
          borderColor: primaryColor,
          backgroundColor: gradient,
          fill: true,
          tension: 0.4,
          borderWidth: 2,
          pointRadius: 4,
          pointHoverRadius: 6,
          pointBackgroundColor: primaryColor,
          pointBorderColor: "#fff",
          pointBorderWidth: 1.5,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: "index",
        intersect: false,
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          padding: 10,
          cornerRadius: 8,
          titleFont: { size: 12, weight: "bold" },
          bodyFont: { size: 12 },
        },
      },
      scales: {
        x: {
          grid: { display: false },
          ticks: { font: { size: 11 } },
        },
        y: {
          beginAtZero: true,
          ticks: { precision: 0, font: { size: 11 } },
          grid: { color: "rgba(150, 150, 150, 0.1)" },
        },
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
        "rgba(99, 102, 241, 0.25)",
      );
    }
    if (shopChartCanvas.value) {
      shopChartInstance = createGrowthChart(
        shopChartCanvas.value,
        growthData.value.environments,
        "#10b981",
        "rgba(16, 185, 129, 0.25)",
      );
    }
  }
}

async function loadStats() {
  loading.value = true;
  error.value = "";

  try {
    const [statsRes, growthRes, activityRes, versionRes] = await Promise.all([
      adminGetStats(),
      adminGetGrowth(),
      adminGetRecentActivity(),
      adminGetShopwareVersions(),
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
