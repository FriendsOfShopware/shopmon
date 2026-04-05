<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary cards -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard :icon="IconListCheck" :value="tasks.length" label="Total" />
      <StatCard :icon="IconCircleCheck" :value="counts.healthy" label="Healthy" color="success" />
      <StatCard :icon="IconClock" :value="counts.overdue" label="Overdue" color="warning" />
      <StatCard :icon="IconPause" :value="counts.inactive" label="Inactive" color="muted" />
    </div>

    <!-- Filter + search -->
    <FilterSearchBar
      v-model:filter="activeFilter"
      v-model:search="searchQuery"
      :filters="filters"
      search-placeholder="Search tasks..."
    />

    <!-- Task list -->
    <div v-if="filteredTasks.length > 0" class="space-y-2">
      <div
        v-for="task in filteredTasks"
        :key="task.id"
        :class="[
          'flex items-center gap-4 rounded-xl border px-4 py-3 transition-colors',
          task.overdue && task.status !== 'inactive' ? 'border-warning/20 bg-warning/5' : '',
          task.status === 'inactive' ? 'opacity-50' : '',
        ]"
      >
        <!-- Status badge -->
        <Badge :class="statusClass(task)" class="shrink-0 gap-1 capitalize">
          <component :is="statusIcon(task)" class="size-2.5" />
          {{ task.overdue && task.status !== "inactive" ? "overdue" : task.status }}
        </Badge>

        <!-- Task name -->
        <div class="min-w-0 flex-1">
          <div class="truncate text-sm font-medium font-mono">{{ task.name }}</div>
        </div>

        <!-- Interval -->
        <div
          class="hidden shrink-0 text-xs text-muted-foreground sm:block"
          :title="`Runs every ${task.runInterval} seconds`"
        >
          every {{ formatInterval(task.runInterval) }}
        </div>

        <!-- Timing info -->
        <div class="hidden shrink-0 text-right text-xs tabular-nums lg:block">
          <div class="text-muted-foreground">
            last: {{ task.lastExecutionTime ? formatDateTime(task.lastExecutionTime) : "—" }}
          </div>
          <div
            :class="
              task.overdue && task.status !== 'inactive'
                ? 'text-warning font-medium'
                : 'text-muted-foreground'
            "
          >
            next: {{ task.nextExecutionTime ? formatDateTime(task.nextExecutionTime) : "—" }}
          </div>
        </div>

        <!-- Reschedule action -->
        <Button
          variant="ghost"
          size="icon"
          class="size-7 shrink-0"
          :title="$t('shopDetail.rescheduleTask')"
          @click="onReScheduleTask(task.id)"
        >
          <icon-fa6-solid:arrow-rotate-right class="size-3.5" />
        </Button>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState v-else :icon="IconListCheck" title="No tasks found" size="sm">
      <p class="text-sm text-muted-foreground">
        <template v-if="searchQuery">No tasks match "{{ searchQuery }}".</template>
        <template v-else>No scheduled tasks match the current filter.</template>
      </p>
    </EmptyState>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { formatDateTime } from "@/helpers/formatter";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import StatCard from "@/components/StatCard.vue";
import EmptyState from "@/components/EmptyState.vue";
import FilterSearchBar from "@/components/FilterSearchBar.vue";

import IconListCheck from "~icons/fa6-solid/list-check";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconClock from "~icons/fa6-solid/clock";
import IconPause from "~icons/fa6-solid/pause";

import FaCheck from "~icons/fa6-solid/check";
import FaRotate from "~icons/fa6-solid/rotate";
import FaClock from "~icons/fa6-solid/clock";
import FaPause from "~icons/fa6-solid/pause";
import FaXmark from "~icons/fa6-solid/xmark";

type ScheduledTask = components["schemas"]["ScheduledTask"];

const { t } = useI18n();
const { success, error } = useAlert();
const { environment } = useEnvironmentDetail();

const activeFilter = ref<"all" | "healthy" | "overdue" | "inactive">("all");
const searchQuery = ref("");

const filters = [
  { label: "All", value: "all" as const },
  { label: "Healthy", value: "healthy" as const },
  { label: "Overdue", value: "overdue" as const },
  { label: "Inactive", value: "inactive" as const },
];

const tasks = computed(() => environment.value?.scheduledTasks ?? []);

const counts = computed(() => ({
  healthy: tasks.value.filter((t) => !t.overdue && t.status !== "inactive").length,
  overdue: tasks.value.filter((t) => t.overdue && t.status !== "inactive").length,
  inactive: tasks.value.filter((t) => t.status === "inactive").length,
}));

function statusClass(task: ScheduledTask): string {
  if (task.status === "inactive") return "bg-muted text-muted-foreground";
  if (task.overdue) return "bg-warning/10 text-warning border-warning/30";
  if (task.status === "queued" || task.status === "running")
    return "bg-info/10 text-info border-info/30";
  return "bg-success/10 text-success border-success/30";
}

function statusIcon(task: ScheduledTask) {
  if (task.status === "inactive") return FaPause;
  if (task.overdue) return FaClock;
  if (task.status === "queued" || task.status === "running") return FaRotate;
  if (task.status === "failed") return FaXmark;
  return FaCheck;
}

function formatInterval(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
  if (seconds < 86400) return `${Math.round(seconds / 3600)}h`;
  return `${Math.round(seconds / 86400)}d`;
}

const filteredTasks = computed(() => {
  let list = [...tasks.value];

  // Sort: overdue first, then by next execution
  list.sort((a, b) => {
    // Overdue active tasks first
    const aOverdue = a.overdue && a.status !== "inactive" ? 0 : 1;
    const bOverdue = b.overdue && b.status !== "inactive" ? 0 : 1;
    if (aOverdue !== bOverdue) return aOverdue - bOverdue;
    // Inactive last
    if (a.status === "inactive" && b.status !== "inactive") return 1;
    if (a.status !== "inactive" && b.status === "inactive") return -1;
    // Then by name
    return a.name.localeCompare(b.name);
  });

  switch (activeFilter.value) {
    case "healthy":
      list = list.filter((t) => !t.overdue && t.status !== "inactive");
      break;
    case "overdue":
      list = list.filter((t) => t.overdue && t.status !== "inactive");
      break;
    case "inactive":
      list = list.filter((t) => t.status === "inactive");
      break;
  }

  if (searchQuery.value.length >= 2) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter((t) => t.name.toLowerCase().includes(q));
  }

  return list;
});

async function onReScheduleTask(taskId: string) {
  try {
    await api.POST("/environments/{environmentId}/tasks/{taskId}/reschedule", {
      params: { path: { environmentId: environment.value?.id ?? 0, taskId } },
    });
    success(t("shopDetail.taskRescheduled"));
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
}
</script>
