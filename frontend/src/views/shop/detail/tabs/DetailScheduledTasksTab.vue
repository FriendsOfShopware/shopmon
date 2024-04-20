<template>
    <data-table
        :columns="[
            { key: 'name', name: 'Name', sortable: true },
            { key: 'interval', name: 'Interval' },
            { key: 'lastExecutionTime', name: 'Last Executed', sortable: true },
            { key: 'nextExecutionTime', name: 'Next Execution', sortable: true },
            { key: 'status', name: 'Status' },
        ]"
        :data="shopStore.shop.scheduledTask || []"
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
            <span
                class="cursor-pointer opacity-25 hover:opacity-100 tooltip-position-left"
                data-tooltip="Re-schedule task"
                @click="onReScheduleTask(row.id)"
            >
                <icon-fa6-solid:arrow-rotate-right class="text-base leading-3" />
            </span>
        </template>
    </data-table>
</template>

<script setup lang="ts">
import { formatDateTime } from "@/helpers/formatter";
import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';

const shopStore = useShopStore();
const alertStore = useAlertStore();

async function onReScheduleTask(taskId: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.reScheduleTask(shopStore.shop.organizationId, shopStore.shop.id, taskId);
            alertStore.success('Task is re-scheduled');
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}
</script>
