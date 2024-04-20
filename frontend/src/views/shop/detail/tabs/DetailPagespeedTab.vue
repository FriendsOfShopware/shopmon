<template>
    <data-table
        :columns="[
            { key: 'createdAt', name: 'Checked At' },
            { key: 'performance', name: 'Performance' },
            { key: 'accessibility', name: 'Accessibility' },
            { key: 'bestPractices', name: 'Best Practices' },
            { key: 'seo', name: 'SEO' },
        ]"
        :data="shopStore.shop.pageSpeed"
    >
        <template #cell-createdAt="{ row }">
            <a target="_blank" :href="'https://pagespeed.web.dev/analysis?url=' + shopStore.shop.url">
                {{ formatDateTime(row.createdAt) }}
            </a>
        </template>

        <template
            v-for="(cell, cellKey) in {
                'performance': 'cell-performance',
                'accessibility': 'cell-accessibility',
                'bestPractices': 'cell-bestPractices',
                'seo': 'cell-seo'
            } as const"
            #[cell]="{ row, rowIndex }"
            :key="cellKey"
        >
            <template v-if="
                shopStore.shop.pageSpeed[(rowIndex + 1)] &&
                shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] !== row[cellKey]"
            >
                <icon-fa6-solid:arrow-right
                    :class="[{
                        'text-green-400 -rotate-45 dark:text-green-300':
                            shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] < row[cellKey],
                        'text-red-600 rotate-45 dark:text-red-400':
                            shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] > row[cellKey],
                    }]"
                />
            </template>
            <icon-fa6-solid:minus v-else />

            <span class="ml-2">{{ row[cellKey] }}</span>
        </template>
    </data-table>
</template>

<script setup lang="ts">
import { formatDateTime } from "@/helpers/formatter";
import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();
</script>
