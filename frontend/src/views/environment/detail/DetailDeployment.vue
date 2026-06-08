<template>
  <div v-if="deployment" class="max-w-[1200px] space-y-6">
    <h2 class="text-2xl font-semibold">{{ deployment.name }}</h2>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t("nav.deploymentDetails") }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="flex flex-col">
          <div class="flex justify-between items-center py-2 border-b">
            <span class="text-muted-foreground">{{ $t("deployments.duration") }}:</span>
            <span class="font-medium text-right">{{
              formatDuration(deployment.executionTime)
            }}</span>
          </div>

          <div class="flex justify-between items-center py-2 border-b">
            <span class="text-muted-foreground">{{ $t("deployments.command") }}:</span>
            <code
              class="font-medium text-right font-mono bg-muted px-2 py-1 rounded text-sm max-w-[60%] overflow-hidden text-ellipsis whitespace-nowrap"
              >{{ deployment.command }}</code
            >
          </div>

          <div class="flex justify-between items-center py-2 border-b">
            <span class="text-muted-foreground">{{ $t("deployments.exitCode") }}:</span>
            <span
              class="inline-flex items-center gap-1.5 font-medium"
              :class="deployment.returnCode === 0 ? 'text-green-500' : 'text-destructive'"
            >
              <icon-fa6-solid:check v-if="deployment.returnCode === 0" class="size-4" />
              <icon-fa6-solid:xmark v-else class="size-4" />
              {{ deployment.returnCode }}
            </span>
          </div>

          <div class="flex justify-between items-center py-2 border-b">
            <span class="text-muted-foreground">{{ $t("deployments.started") }}:</span>
            <span class="font-medium text-right">{{ formatDateTime(deployment.startDate) }}</span>
          </div>

          <div class="flex justify-between items-center py-2 border-b last:border-b-0">
            <span class="text-muted-foreground">{{ $t("deployments.completed") }}:</span>
            <span class="font-medium text-right">{{ formatDateTime(deployment.endDate) }}</span>
          </div>

          <div
            v-if="deployment.reference && deployment.gitUrl"
            class="flex justify-between items-center py-2"
          >
            <span class="text-muted-foreground">{{ $t("deployments.reference") }}:</span>
            <a
              :href="`${deployment.gitUrl}/commit/${deployment.reference}`"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 font-medium text-primary hover:underline"
            >
              <icon-fa6-solid:arrow-up-right-from-square class="size-4" />
              {{ deployment.reference.substring(0, 8) }}
            </a>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card v-if="deployment.composer && Object.keys(deployment.composer).length > 0">
      <CardHeader>
        <CardTitle class="flex items-center gap-1.5">
          <icon-fa6-solid:box class="size-4" />
          {{ $t("deployments.composerPackages") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div class="overflow-x-auto">
          <table class="w-full border-collapse">
            <thead>
              <tr class="bg-muted border-b-2">
                <th
                  class="px-4 py-2 text-left font-semibold text-sm text-muted-foreground uppercase tracking-wide"
                >
                  {{ $t("deployments.package") }}
                </th>
                <th
                  class="px-4 py-2 text-left font-semibold text-sm text-muted-foreground uppercase tracking-wide"
                >
                  {{ $t("common.version") }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="[packageName, version] in sortedComposer"
                :key="packageName"
                class="hover:bg-muted/50 border-b last:border-b-0"
              >
                <td class="px-4 py-2 font-mono text-sm">{{ packageName }}</td>
                <td class="px-4 py-2 font-mono text-sm text-muted-foreground font-semibold">
                  {{ version }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-1.5">
          <icon-fa6-solid:terminal class="size-4" />
          {{ $t("deployments.output") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div class="bg-[#0d1117] overflow-hidden p-4 rounded-md">
          <pre
            class="m-0 p-0 overflow-x-auto font-mono text-[0.8125rem] leading-relaxed text-[#c9d1d9] whitespace-pre break-normal"
            style="tab-size: 4"
            v-html="formattedOutput"
          ></pre>
        </div>
      </CardContent>
    </Card>
  </div>
  <div v-else class="p-12 text-center text-muted-foreground">{{ $t("common.loading") }}</div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { useRoute } from "vue-router";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const route = useRoute();
const { environment } = useEnvironmentDetail();

const deployment = ref<any>(null);

const loadDeployment = async () => {
  if (!environment.value) return;

  try {
    const deploymentId = parseInt(route.params.deploymentId as string, 10);
    const { data } = await api.GET("/environments/{environmentId}/deployments/{deploymentId}", {
      params: { path: { environmentId: environment.value.id, deploymentId } },
    });
    deployment.value = data ?? null;
  } catch (error) {
    console.error("Failed to load deployment:", error);
  }
};

const formatDuration = (seconds: string) => {
  const num = parseFloat(seconds);
  if (num < 1) {
    return `${(num * 1000).toFixed(0)}ms`;
  }
  if (num < 60) {
    return `${num.toFixed(2)}s`;
  }
  const minutes = Math.floor(num / 60);
  const secs = (num % 60).toFixed(0);
  return `${minutes}m ${secs}s`;
};

const ansiToHtml = (text: string) => {
  if (!text) return "";

  const ansiColors: Record<number, string> = {
    0: "</span>",
    30: "#24292e",
    31: "#f85149",
    32: "#3fb950",
    33: "#d29922",
    34: "#58a6ff",
    35: "#bc8cff",
    36: "#39c5cf",
    37: "#c9d1d9",
    39: "#c9d1d9",
    90: "#6e7681",
    91: "#ff7b72",
    92: "#56d364",
    93: "#e3b341",
    94: "#79c0ff",
    95: "#d2a8ff",
    96: "#56d4dd",
    97: "#f0f6fc",
  };

  let html = "";
  let currentColor = "";

  const parts = text.split(/\x1b\[/);

  for (let i = 0; i < parts.length; i++) {
    if (i === 0) {
      html += parts[i];
      continue;
    }

    const match = parts[i].match(/^(\d+(?:;\d+)*)m(.*)$/s);
    if (match) {
      const codes = match[1].split(";").map(Number);
      const content = match[2];

      for (const code of codes) {
        if (code === 0) {
          if (currentColor) {
            html += "</span>";
            currentColor = "";
          }
        } else if (ansiColors[code]) {
          if (currentColor) {
            html += "</span>";
          }
          currentColor = ansiColors[code];
          html += `<span style="color: ${currentColor}">`;
        }
      }

      html += content;
    } else {
      html += parts[i];
    }
  }

  if (currentColor) {
    html += "</span>";
  }

  return html;
};

const sortedComposer = computed(() => {
  if (!deployment.value?.composer) return [];
  return Object.entries(deployment.value.composer as Record<string, string>).sort((a, b) =>
    a[0].localeCompare(b[0]),
  );
});

const formattedOutput = computed(() => {
  if (!deployment.value?.output) return "";
  return ansiToHtml(deployment.value.output);
});

watch(
  environment,
  (newEnvironment) => {
    if (newEnvironment) {
      loadDeployment();
    }
  },
  { immediate: true },
);
</script>
