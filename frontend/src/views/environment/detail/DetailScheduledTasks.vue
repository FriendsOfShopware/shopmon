<template>
  <Card class="p-0 overflow-hidden">
    <data-table
      v-if="environment"
      :columns="[
        { key: 'name', name: $t('common.name'), sortable: true },
        { key: 'runInterval', name: $t('shopDetail.interval') },
        { key: 'lastExecutionTime', name: $t('shopDetail.lastExecuted'), sortable: true },
        { key: 'nextExecutionTime', name: $t('shopDetail.nextExecution'), sortable: true },
        { key: 'status', name: $t('common.status') },
      ]"
      :data="environment.scheduledTasks || []"
      :default-sorting="{ by: 'nextExecutionTime' }"
    >
      <template #cell-lastExecutionTime="{ row }">
        {{ row.lastExecutionTime ? formatDateTime(row.lastExecutionTime) : "-" }}
      </template>

      <template #cell-nextExecutionTime="{ row }">
        {{ row.nextExecutionTime ? formatDateTime(row.nextExecutionTime) : "-" }}
      </template>

      <template #cell-status="{ row }">
        <Badge
          v-if="row.status === 'scheduled' && !row.overdue"
          variant="default"
          class="bg-green-600 text-white gap-1"
        >
          <icon-fa6-solid:check class="size-2.5" />
          {{ row.status }}
        </Badge>
        <Badge
          v-else-if="row.status === 'queued' || (row.status === 'running' && !row.overdue)"
          variant="default"
          class="bg-blue-600 text-white gap-1"
        >
          <icon-fa6-solid:rotate class="size-2.5" />
          {{ row.status }}
        </Badge>
        <Badge
          v-else-if="
            row.status === 'scheduled' ||
            row.status === 'queued' ||
            (row.status === 'running' && row.overdue)
          "
          variant="default"
          class="bg-yellow-600 text-white gap-1"
        >
          <icon-fa6-solid:info class="size-2.5 align-[-0.1em]" />
          {{ row.status }}
        </Badge>
        <Badge
          v-else-if="row.status === 'inactive'"
          variant="secondary"
          class="gap-1"
        >
          <icon-fa6-solid:pause class="size-2.5" />
          {{ row.status }}
        </Badge>
        <Badge
          v-else
          variant="destructive"
          class="gap-1"
        >
          <icon-fa6-solid:xmark class="size-2.5" />
          {{ row.status }}
        </Badge>
      </template>

      <template #cell-actions="{ row }">
        <button
          :title="$t('shopDetail.rescheduleTask')"
          @click="onReScheduleTask(row.id)"
        >
          <icon-fa6-solid:arrow-rotate-right class="size-4" />
        </button>
      </template>
    </data-table>
  </Card>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { formatDateTime } from "@/helpers/formatter";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { api } from "@/helpers/api";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const { t } = useI18n();

const { environment } = useEnvironmentDetail();

const { success, error } = useAlert();

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
