<template>
  <Card v-if="loadFailed">
    <CardContent class="flex flex-col items-center gap-3 py-10 text-center">
      <p class="text-sm text-muted-foreground">{{ $t("uptime.loadFailed") }}</p>
      <Button variant="outline" @click="load">{{ $t("uptime.retry") }}</Button>
    </CardContent>
  </Card>

  <div v-else-if="!view" class="space-y-6">
    <Skeleton class="h-32 w-full" />
    <Skeleton class="h-64 w-full" />
  </div>

  <div v-else class="space-y-6">
    <!-- Status overview -->
    <Card>
      <CardHeader class="pb-2">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <CardTitle class="text-base">{{ $t("uptime.title") }}</CardTitle>
          <div class="flex items-center gap-2">
            <Badge :variant="statusBadgeVariant" class="gap-1.5">
              <span class="size-1.5 rounded-full" :class="statusDotClass" />
              {{ statusLabel }}
            </Badge>
            <Select v-model="range" @update:model-value="load">
              <SelectTrigger class="h-8 w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="24h">{{ $t("uptime.range24h") }}</SelectItem>
                <SelectItem value="7d">{{ $t("uptime.range7d") }}</SelectItem>
                <SelectItem value="30d">{{ $t("uptime.range30d") }}</SelectItem>
                <SelectItem value="90d">{{ $t("uptime.range90d") }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <template v-if="!view.settings.enabled">
          <Alert variant="default">
            <AlertDescription>{{ $t("uptime.enableDesc") }}</AlertDescription>
          </Alert>
        </template>

        <template v-else>
          <!-- Open incident banner -->
          <Alert
            v-if="view.openIncident"
            variant="destructive"
            class="mb-4 border-destructive/30 bg-destructive/10"
          >
            <CircleX class="size-4" />
            <AlertDescription>
              {{ $t("uptime.statusDown") }} —
              {{ formatDuration(view.openIncident.durationSeconds) }}
              <template v-if="view.openIncident.error"> ({{ view.openIncident.error }})</template>
            </AlertDescription>
          </Alert>

          <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div>
              <div class="text-xs font-medium text-muted-foreground">
                {{ $t("uptime.availability") }}
              </div>
              <div class="mt-1 text-2xl font-bold tabular-nums">
                {{ formatAvailability(view.availability) }}
              </div>
            </div>
            <div>
              <div class="text-xs font-medium text-muted-foreground">
                {{ $t("uptime.lastChecked") }}
              </div>
              <div class="mt-1 text-sm font-semibold">
                {{
                  view.settings.lastCheckedAt ? formatDateTime(view.settings.lastCheckedAt) : "—"
                }}
              </div>
            </div>
            <div>
              <div class="text-xs font-medium text-muted-foreground">
                {{ $t("uptime.lastStatusCode") }}
              </div>
              <div class="mt-1 text-sm font-semibold tabular-nums">
                {{ view.settings.lastStatusCode ?? "—" }}
              </div>
            </div>
            <div>
              <div class="text-xs font-medium text-muted-foreground">
                {{ $t("uptime.latency") }}
              </div>
              <div class="mt-1 text-sm font-semibold tabular-nums">
                {{ view.settings.lastLatencyMs != null ? `${view.settings.lastLatencyMs}ms` : "—" }}
              </div>
            </div>
          </div>

          <p v-if="view.settings.lastError" class="mt-3 text-sm text-destructive">
            {{ view.settings.lastError }}
          </p>
        </template>
      </CardContent>
    </Card>

    <template v-if="view.settings.enabled">
      <!-- Day strip (multi-day ranges) -->
      <Card v-if="view.days.length">
        <CardHeader class="pb-2">
          <CardTitle class="text-base">{{ $t("uptime.dayAvailability") }} (UTC)</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex gap-0.5">
            <div
              v-for="day in view.days"
              :key="day.day"
              class="h-8 flex-1 rounded-sm"
              :class="dayColor(day)"
              :title="dayTooltip(day)"
            />
          </div>
          <div class="mt-2 flex justify-between text-xs text-muted-foreground">
            <span>{{ view.days[0]?.day }}</span>
            <span>{{ view.days[view.days.length - 1]?.day }}</span>
          </div>
        </CardContent>
      </Card>

      <!-- Latency chart (24h range) -->
      <Card v-if="range === '24h' && view.latency.length">
        <CardHeader class="pb-2">
          <CardTitle class="text-base">{{ $t("uptime.latencyOverTime") }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="h-56">
            <canvas ref="latencyCanvas" />
          </div>
        </CardContent>
      </Card>

      <!-- Incidents -->
      <Card class="overflow-hidden p-0">
        <CardHeader class="p-4 pb-0">
          <CardTitle class="text-base">{{ $t("uptime.incidents") }}</CardTitle>
        </CardHeader>
        <CardContent class="p-0 pt-3">
          <p v-if="!view.incidents.length" class="px-4 pb-4 text-sm text-muted-foreground">
            {{ $t("uptime.noIncidents") }}
          </p>
          <div v-else class="divide-y">
            <div
              v-for="incident in view.incidents"
              :key="incident.id"
              class="flex flex-wrap items-center justify-between gap-2 px-4 py-3"
            >
              <div class="flex items-center gap-3">
                <CircleX v-if="!incident.resolvedAt" class="size-4 shrink-0 text-destructive" />
                <CircleCheck v-else class="size-4 shrink-0 text-success" />
                <div>
                  <div class="text-sm font-medium">
                    {{ formatDateTime(incident.startedAt) }}
                    <Badge v-if="!incident.resolvedAt" variant="destructive" class="ml-2">
                      {{ $t("uptime.ongoing") }}
                    </Badge>
                  </div>
                  <div class="text-xs text-muted-foreground">
                    {{ formatDuration(incident.durationSeconds) }}
                    <template v-if="incident.statusCode">
                      · HTTP {{ incident.statusCode }}</template
                    >
                    <template v-if="incident.error"> · {{ incident.error }}</template>
                  </div>
                </div>
              </div>
              <div v-if="incident.resolvedAt" class="text-xs text-muted-foreground">
                {{ $t("uptime.resolved") }}: {{ formatDateTime(incident.resolvedAt) }}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>

    <!-- Settings -->
    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-base">{{ $t("uptime.settingsTitle") }}</CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="flex items-center justify-between">
          <Label for="uptime-enabled">{{ $t("uptime.enabled") }}</Label>
          <Switch id="uptime-enabled" v-model="settings.enabled" />
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-1.5">
            <Label for="uptime-url">{{ $t("uptime.urlOverride") }}</Label>
            <Input
              id="uptime-url"
              v-model="settings.url"
              type="url"
              :placeholder="environment?.url ?? ''"
            />
            <p class="text-xs text-muted-foreground">{{ $t("uptime.urlOverrideDesc") }}</p>
          </div>
          <div class="space-y-1.5">
            <Label for="uptime-interval">{{ $t("uptime.interval") }}</Label>
            <Input
              id="uptime-interval"
              v-model.number="settings.intervalSeconds"
              type="number"
              min="30"
              max="3600"
            />
          </div>
          <div class="space-y-1.5">
            <Label for="uptime-expected-status">{{ $t("uptime.expectedStatus") }}</Label>
            <Input
              id="uptime-expected-status"
              v-model.number="settings.expectedStatus"
              type="number"
              min="0"
              max="599"
            />
            <p class="text-xs text-muted-foreground">{{ $t("uptime.expectedStatusDesc") }}</p>
          </div>
          <div class="space-y-1.5">
            <Label for="uptime-content-match">{{ $t("uptime.contentMatch") }}</Label>
            <Input id="uptime-content-match" v-model="settings.contentMatch" type="text" />
            <p class="text-xs text-muted-foreground">{{ $t("uptime.contentMatchDesc") }}</p>
          </div>
          <div class="space-y-1.5">
            <Label for="uptime-failure-threshold">{{ $t("uptime.failureThreshold") }}</Label>
            <Input
              id="uptime-failure-threshold"
              v-model.number="settings.failureThreshold"
              type="number"
              min="1"
              max="10"
            />
            <p class="text-xs text-muted-foreground">{{ $t("uptime.failureThresholdDesc") }}</p>
          </div>
          <div class="space-y-1.5">
            <Label for="uptime-recovery-threshold">{{ $t("uptime.recoveryThreshold") }}</Label>
            <Input
              id="uptime-recovery-threshold"
              v-model.number="settings.recoveryThreshold"
              type="number"
              min="1"
              max="10"
            />
            <p class="text-xs text-muted-foreground">{{ $t("uptime.recoveryThresholdDesc") }}</p>
          </div>
        </div>

        <Button :disabled="isSaving" @click="save">
          <icon-fa6-solid:rotate v-if="isSaving" class="mr-1.5 size-3 animate-spin" />
          {{ $t("uptime.save") }}
        </Button>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { Chart, registerables } from "chart.js";
import "chartjs-adapter-date-fns";
import { CircleCheck, CircleX } from "lucide-vue-next";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useAlert } from "@/composables/useAlert";
import { formatDateTime } from "@/helpers/formatter";
import {
  getEnvironmentUptime,
  updateUptimeSettings,
  type UptimeDay,
  type UptimeResponse,
} from "@/api/generated";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

Chart.register(...registerables);

type Range = "24h" | "7d" | "30d" | "90d";

const { t } = useI18n();
const { success, error } = useAlert();
const { environment } = useEnvironmentDetail();

const environmentId = computed(() => environment.value?.id ?? 0);

const view = ref<UptimeResponse | null>(null);
const range = ref<Range>("24h");
const isSaving = ref(false);
const loadFailed = ref(false);
const latencyCanvas = ref<HTMLCanvasElement | null>(null);
let chart: Chart | null = null;
// Guards against stale responses: when the range or environment changes while
// a request is in flight, only the newest request may update the view.
let loadToken = 0;

const settings = reactive({
  enabled: false,
  url: "",
  intervalSeconds: 60,
  expectedStatus: 0,
  contentMatch: "",
  failureThreshold: 3,
  recoveryThreshold: 2,
});

function applySettingsToForm() {
  if (!view.value) return;
  const s = view.value.settings;
  settings.enabled = s.enabled;
  settings.url = s.url ?? "";
  settings.intervalSeconds = s.intervalSeconds;
  settings.expectedStatus = s.expectedStatus;
  settings.contentMatch = s.contentMatch ?? "";
  settings.failureThreshold = s.failureThreshold;
  settings.recoveryThreshold = s.recoveryThreshold;
}

const numericFieldBounds = [
  { key: "intervalSeconds", min: 30, max: 3600, labelKey: "uptime.interval" },
  { key: "expectedStatus", min: 0, max: 599, labelKey: "uptime.expectedStatus" },
  { key: "failureThreshold", min: 1, max: 10, labelKey: "uptime.failureThreshold" },
  { key: "recoveryThreshold", min: 1, max: 10, labelKey: "uptime.recoveryThreshold" },
] as const;

// Number inputs yield "" when cleared and v-model.number cannot revive them;
// validate before sending so the API never receives junk.
function validateSettingsForm(): string | null {
  for (const field of numericFieldBounds) {
    const value = settings[field.key];
    if (
      typeof value !== "number" ||
      Number.isNaN(value) ||
      value < field.min ||
      value > field.max
    ) {
      return t("uptime.invalidNumber", {
        field: t(field.labelKey),
        min: field.min,
        max: field.max,
      });
    }
  }
  return null;
}

async function load() {
  if (!environmentId.value) return;
  const token = ++loadToken;
  loadFailed.value = false;
  try {
    const { data, error: fetchError } = await getEnvironmentUptime({
      path: { environmentId: environmentId.value },
      query: { range: range.value },
    });
    if (token !== loadToken) return;
    if (fetchError) {
      loadFailed.value = !view.value;
      error(fetchError instanceof Error ? fetchError.message : String(fetchError));
      return;
    }
    if (data) {
      view.value = data;
      applySettingsToForm();
      await nextTick();
      renderLatencyChart();
    }
  } catch (e) {
    if (token !== loadToken) return;
    loadFailed.value = !view.value;
    error(e instanceof Error ? e.message : String(e));
  }
}

async function save() {
  if (!environmentId.value) return;
  const invalid = validateSettingsForm();
  if (invalid) {
    error(invalid);
    return;
  }
  try {
    isSaving.value = true;
    await updateUptimeSettings({
      path: { environmentId: environmentId.value },
      body: {
        enabled: settings.enabled,
        url: settings.url || null,
        intervalSeconds: settings.intervalSeconds,
        expectedStatus: settings.expectedStatus,
        contentMatch: settings.contentMatch || null,
        failureThreshold: settings.failureThreshold,
        recoveryThreshold: settings.recoveryThreshold,
      },
    });
    success(t("uptime.saved"));
    await load();
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  } finally {
    isSaving.value = false;
  }
}

const statusLabel = computed(() => {
  if (!view.value) return "";
  if (!view.value.settings.enabled) return t("uptime.statusDisabled");
  switch (view.value.settings.status) {
    case "up":
      return t("uptime.statusUp");
    case "down":
      return t("uptime.statusDown");
    case "paused":
      return t("uptime.statusPaused");
    default:
      return t("uptime.statusUnknown");
  }
});

const statusBadgeVariant = computed(() => {
  const status = view.value?.settings.status;
  if (!view.value?.settings.enabled) return "secondary";
  if (status === "down") return "destructive";
  if (status === "up") return "default";
  return "secondary";
});

const statusDotClass = computed(() => {
  const status = view.value?.settings.status;
  if (!view.value?.settings.enabled) return "bg-muted-foreground";
  if (status === "down") return "bg-destructive";
  if (status === "up") return "bg-success";
  return "bg-muted-foreground";
});

function formatAvailability(value: number | null | undefined): string {
  if (value == null) return "—";
  return `${(value * 100).toFixed(value >= 0.999 ? 3 : 2)}%`;
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}

function dayColor(day: UptimeDay): string {
  if (day.availability == null) return "bg-muted";
  if (day.availability >= 0.999) return "bg-success";
  if (day.availability >= 0.99) return "bg-warning";
  return "bg-destructive";
}

function dayTooltip(day: UptimeDay): string {
  const availability =
    day.availability == null ? t("uptime.dayNoData") : `${(day.availability * 100).toFixed(2)}%`;
  return `${day.day}: ${availability}`;
}

function renderLatencyChart() {
  if (chart) {
    chart.destroy();
    chart = null;
  }
  if (range.value !== "24h" || !latencyCanvas.value || !view.value?.latency.length) return;

  const points = view.value.latency;
  chart = new Chart(latencyCanvas.value, {
    type: "line",
    data: {
      datasets: [
        {
          label: t("uptime.avg"),
          data: points.map((p) => ({ x: new Date(p.timestamp).getTime(), y: p.avgMs ?? null })),
          borderColor: "rgb(59, 130, 246)",
          backgroundColor: "rgba(59, 130, 246, 0.1)",
          tension: 0.3,
        },
        {
          label: t("uptime.p95"),
          data: points.map((p) => ({ x: new Date(p.timestamp).getTime(), y: p.p95Ms ?? null })),
          borderColor: "rgb(168, 85, 247)",
          backgroundColor: "rgba(168, 85, 247, 0.1)",
          tension: 0.3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      interaction: { mode: "index", intersect: false },
      plugins: {
        legend: { display: true, position: "top" },
        tooltip: {
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${ctx.parsed.y}ms`,
            title: (items) =>
              items.length && items[0].parsed.x != null
                ? (formatDateTime(new Date(items[0].parsed.x)) ?? "")
                : "",
          },
        },
      },
      scales: {
        x: {
          type: "time",
          time: { unit: "hour", displayFormats: { hour: "HH:mm" } },
          title: { display: false },
        },
        y: {
          beginAtZero: true,
          title: { display: true, text: "ms" },
        },
      },
    },
  });
}

watch(environmentId, () => load());
onMounted(() => load());
onUnmounted(() => {
  if (chart) {
    chart.destroy();
    chart = null;
  }
});
</script>
