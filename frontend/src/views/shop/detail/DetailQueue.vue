<template>
    <div class="panel panel-table">
        <data-table
            v-if="shop"
            :columns="[
                { key: 'name', name: 'Name', sortable: true },
                { key: 'size', name: 'Size', sortable: true },
            ]"
            :data="shop.queueInfo || []"
        />
    </div>
</template>

<script setup lang="ts">
import { useShopDetail } from '@/composables/useShopDetail';

const {
    shop,
} = useShopDetail();
</script>
