<template>
  <Card class="p-0 overflow-hidden">
    <data-table
      v-if="environment"
      :columns="[
        { key: 'name', name: $t('common.name'), sortable: true },
        { key: 'size', name: $t('shopDetail.size'), sortable: true },
      ]"
      :data="environment.queues || []"
    />
  </Card>
</template>

<script setup lang="ts">
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import { Card } from "@/components/ui/card";

const { environment } = useEnvironmentDetail();
</script>
