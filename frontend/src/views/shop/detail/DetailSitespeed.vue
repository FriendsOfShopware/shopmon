<template>
    <Alert v-if="shop && !shop.sitespeedEnabled" type="info">
        Sitespeed monitoring is not enabled for this shop. Please enable it in the shop settings to view performance metrics.
    </Alert>
    <div class="mb-1">
        <a v-if="shop && shop.sitespeed.length > 0" class="btn btn-primary-outline" :href="'/sitespeed/' + shop.id + '/index.html'" target="_blank">
            <i class="fa fa-chart-line"></i> View Sitespeed Report
        </a>
    </div>
    <div class="panel panel-table">
        <data-table
            v-if="shop"
            :columns="[
                { key: 'createdAt', name: 'Checked At' },
                { key: 'ttfb', name: 'TTFB' },
                { key: 'fullyLoaded', name: 'Fully Loaded' },
                { key: 'largestContentfulPaint', name: 'Largest Contentful Paint' },
                { key: 'firstContentfulPaint', name: 'First Contentful Paint' },
                { key: 'cumulativeLayoutShift', name: 'Cumulative Layout Shift' },
                { key: 'transferSize', name: 'Transfer Size' },
            ]"
            :data="shop.sitespeed || []"
            >
            <template #cell-ttfb="{ row }">
                {{ row.ttfb ? `${row.ttfb} ms` : '-' }}
            </template>
            <template #cell-fullyLoaded="{ row }">
                {{ row.fullyLoaded ? `${row.fullyLoaded} ms` : '-' }}
            </template>
            <template #cell-largestContentfulPaint="{ row }">
                {{ row.largestContentfulPaint ? `${row.largestContentfulPaint} ms` : '-' }}
            </template>
            <template #cell-firstContentfulPaint="{ row }">
                {{ row.firstContentfulPaint ? `${row.firstContentfulPaint} ms` : '-' }}
            </template>
            <template #cell-cumulativeLayoutShift="{ row }">
                {{ row.cumulativeLayoutShift ? row.cumulativeLayoutShift : '-' }}
            </template>
            <template #cell-transferSize="{ row }">
                {{ row.transferSize ? formatBytes(row.transferSize) : '-' }}
            </template>
            <template #cell-createdAt="{ row }">
                {{ formatDateTime(row.createdAt) }}
            </template>
        </data-table>
    </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/helpers/formatter';
import {useShopDetail} from "@/composables/useShopDetail";

const {
    shop,
} = useShopDetail();

function formatBytes(bytes: number, decimals = 2): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}
</script>
