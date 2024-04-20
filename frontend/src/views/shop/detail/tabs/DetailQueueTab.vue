<template>
    <data-table
        :columns="[
            { key: 'name', name: 'Name', sortable: true },
            { key: 'size', name: 'Size', sortable: true },
        ]"
        :data="shopStore.shop.queueInfo || []"
    />
</template>

<script setup lang="ts">
import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();
</script>
