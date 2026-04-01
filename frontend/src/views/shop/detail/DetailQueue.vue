<template>
  <Panel variant="table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'name', name: $t('common.name'), sortable: true },
        { key: 'size', name: $t('shopDetail.size'), sortable: true },
      ]"
      :data="shop.queueInfo || []"
    />
  </Panel>
</template>

<script setup lang="ts">
import { useShopDetail } from "@/composables/useShopDetail";

const { shop } = useShopDetail();
</script>
