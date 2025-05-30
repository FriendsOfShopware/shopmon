<template>
    <data-table
        :columns="[
            { key: 'name', name: 'Name', sortable: true },
            { key: 'size', name: 'Size', sortable: true },
        ]"
        :data="shop.queueInfo || []"
    />
</template>

<script setup lang="ts">
import type { RouterOutput } from '@/helpers/trpc';

const { shop } = defineProps<{
    shop: RouterOutput['organization']['shop']['get'];
}>();
</script>
