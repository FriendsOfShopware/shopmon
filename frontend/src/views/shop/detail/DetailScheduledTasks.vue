<template>
    <div class="panel panel-table">
        <data-table
        v-if="shop"
        :columns="[
            { key: 'name', name: 'Name', sortable: true },
            { key: 'interval', name: 'Interval' },
            { key: 'lastExecutionTime', name: 'Last Executed', sortable: true },
            { key: 'nextExecutionTime', name: 'Next Execution', sortable: true },
            { key: 'status', name: 'Status' },
        ]"
        :data="shop.scheduledTask || []"
        :default-sorting="{ by: 'nextExecutionTime' }"
    >
        <template #cell-lastExecutionTime="{ row }">
            {{ formatDateTime(row.lastExecutionTime) }}
        </template>

        <template #cell-nextExecutionTime="{ row }">
            {{ formatDateTime(row.nextExecutionTime) }}
        </template>

        <template #cell-status="{ row }">
            <span
                v-if="row.status === 'scheduled' && !row.overdue"
                class="pill pill-success"
            >
                <icon-fa6-solid:check />
                {{ row.status }}
            </span>
            <span
                v-else-if="row.status === 'queued' || row.status === 'running' && !row.overdue"
                class="pill pill-info"
            >
                <icon-fa6-solid:rotate />
                {{ row.status }}
            </span>
            <span
                v-else-if="
                    row.status === 'scheduled' ||
                    row.status === 'queued' ||
                    row.status === 'running' &&
                    row.overdue"
                class="pill pill-warning"
            >
                <icon-fa6-solid:info class="align-[-0.1em]" />
                {{ row.status }}
            </span>
            <span
                v-else-if="row.status === 'inactive'"
                class="pill"
            >
                <icon-fa6-solid:pause />
                {{ row.status }}
            </span>
            <span
                v-else
                class="pill pill-error"
            >
                <icon-fa6-solid:xmark />
                {{ row.status }}
            </span>
        </template>

        <template #cell-actions="{ row }">
            <button
                class="tooltip-position-left"
                data-tooltip="Re-schedule task"
                @click="onReScheduleTask(row.id)"
            >
                <icon-fa6-solid:arrow-rotate-right class="icon" />
            </button>
        </template>
    </data-table>
    </div>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { formatDateTime } from '@/helpers/formatter';
import {useShopDetail} from '@/composables/useShopDetail';
import { trpcClient } from '@/helpers/trpc';

const {
    shop,
} = useShopDetail();

const { success, error } = useAlert();

async function onReScheduleTask(taskId: string) {
    try {
        await trpcClient.organization.shop.rescheduleTask.mutate({
            shopId: shop.value?.id ?? 0,
            taskId,
        });
        success('Task is re-scheduled');
    } catch (e) {
        error(e instanceof Error ? e.message : String(e));
    }
}
</script>
