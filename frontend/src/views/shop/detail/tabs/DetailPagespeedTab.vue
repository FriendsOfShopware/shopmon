<template>
    <data-table
        :columns="[
            { key: 'createdAt', name: 'Checked At' },
            { key: 'performance', name: 'Performance' },
            { key: 'accessibility', name: 'Accessibility' },
            { key: 'bestPractices', name: 'Best Practices' },
            { key: 'seo', name: 'SEO' },
        ]"
        :data="shop.pageSpeed || []"
    >
        <template #cell-createdAt="{ row }">
            <a target="_blank" :href="'https://pagespeed.web.dev/analysis?url=' + shop.url">
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
                shop.pageSpeed?.[(rowIndex + 1)] &&
                shop.pageSpeed[(rowIndex + 1)][cellKey] !== row[cellKey]"
            >
                <icon-fa6-solid:arrow-right
                    :class="[{
                        'icon' : true,
                        'icon-success':
                            shop.pageSpeed?.[(rowIndex + 1)]?.[cellKey] != null && 
                            row[cellKey] != null &&
                            shop.pageSpeed[(rowIndex + 1)][cellKey]! < row[cellKey],
                        'icon-error':
                            shop.pageSpeed?.[(rowIndex + 1)]?.[cellKey] != null && 
                            row[cellKey] != null &&
                            shop.pageSpeed[(rowIndex + 1)][cellKey]! > row[cellKey],
                    }]"
                />
            </template>
            <icon-fa6-solid:minus v-else class="icon" />

            <span class="ml-2">{{ row[cellKey] ?? '-' }}</span>
        </template>
    </data-table>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/helpers/formatter';
import type { RouterOutput } from '@/helpers/trpc';

const { shop } = defineProps<{
    shop: RouterOutput['organization']['shop']['get'];
}>();
</script>

<style scoped>
.icon {
    margin-right: .25rem;
}

.icon-success {
    transform: rotate(-45deg);
}

.icon-error {
    transform: rotate(45deg);
}
</style>
