<template>
  <Panel variant="table">
    <data-table
      v-if="environment"
      :columns="[
        { key: 'name', name: $t('common.name'), sortable: true },
        { key: 'size', name: $t('shopDetail.size'), sortable: true },
      ]"
      :data="environment.queues || []"
    />
  </Panel>
</template>

<script setup lang="ts">
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

const { environment } = useEnvironmentDetail();
</script>
