<template>
  <Panel variant="table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'name', name: $t('common.name'), sortable: true },
        { key: 'runInterval', name: $t('shopDetail.interval') },
        { key: 'lastExecutionTime', name: $t('shopDetail.lastExecuted'), sortable: true },
        { key: 'nextExecutionTime', name: $t('shopDetail.nextExecution'), sortable: true },
        { key: 'status', name: $t('common.status') },
      ]"
      :data="shop.scheduledTasks || []"
      :default-sorting="{ by: 'nextExecutionTime' }"
    >
      <template #cell-lastExecutionTime="{ row }">
        {{ row.lastExecutionTime ? formatDateTime(row.lastExecutionTime) : "-" }}
      </template>

      <template #cell-nextExecutionTime="{ row }">
        {{ row.nextExecutionTime ? formatDateTime(row.nextExecutionTime) : "-" }}
      </template>

      <template #cell-status="{ row }">
        <span v-if="row.status === 'scheduled' && !row.overdue" class="pill pill-success">
          <icon-fa6-solid:check />
          {{ row.status }}
        </span>
        <span
          v-else-if="row.status === 'queued' || (row.status === 'running' && !row.overdue)"
          class="pill pill-info"
        >
          <icon-fa6-solid:rotate />
          {{ row.status }}
        </span>
        <span
          v-else-if="
            row.status === 'scheduled' ||
            row.status === 'queued' ||
            (row.status === 'running' && row.overdue)
          "
          class="pill pill-warning"
        >
          <icon-fa6-solid:info class="align-[-0.1em]" />
          {{ row.status }}
        </span>
        <span v-else-if="row.status === 'inactive'" class="pill">
          <icon-fa6-solid:pause />
          {{ row.status }}
        </span>
        <span v-else class="pill pill-error">
          <icon-fa6-solid:xmark />
          {{ row.status }}
        </span>
      </template>

      <template #cell-actions="{ row }">
        <button
          class="tooltip-top-left"
          :data-tooltip="$t('shopDetail.rescheduleTask')"
          @click="onReScheduleTask(row.id)"
        >
          <icon-fa6-solid:arrow-rotate-right class="icon" />
        </button>
      </template>
    </data-table>
  </Panel>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { formatDateTime } from "@/helpers/formatter";
import { useShopDetail } from "@/composables/useShopDetail";
import { api } from "@/helpers/api";

const { t } = useI18n();

const { shop } = useShopDetail();

const { success, error } = useAlert();

async function onReScheduleTask(taskId: string) {
  try {
    await api.POST("/shops/{shopId}/tasks/{taskId}/reschedule", {
      params: { path: { shopId: shop.value?.id ?? 0, taskId } },
    });
    success(t("shopDetail.taskRescheduled"));
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
}
</script>
