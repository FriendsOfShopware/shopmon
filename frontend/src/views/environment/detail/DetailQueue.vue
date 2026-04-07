<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <StatCard :icon="IconLayerGroup" :value="queues.length" :label="$t('common.queues')" />
      <StatCard
        :icon="IconEnvelope"
        :value="totalSize.toLocaleString()"
        :label="$t('shopDetail.pendingMessages')"
        :color="totalSize > 0 ? 'warning' : 'success'"
      />
      <StatCard
        :icon="IconGaugeHigh"
        :value="busyQueues"
        :label="$t('shopDetail.queuesWithBacklog')"
        :color="busyQueues > 0 ? 'warning' : 'success'"
      />
    </div>

    <!-- Queue list -->
    <div v-if="sortedQueues.length > 0" class="space-y-2">
      <div
        v-for="queue in sortedQueues"
        :key="queue.name"
        :class="[
          'flex items-center gap-4 rounded-xl border px-4 py-3',
          queue.size > 0 ? 'border-warning/20 bg-warning/5' : '',
        ]"
      >
        <!-- Indicator -->
        <div
          class="flex size-8 shrink-0 items-center justify-center rounded-lg"
          :class="queue.size > 0 ? 'bg-warning/10' : 'bg-success/10'"
        >
          <icon-fa6-solid:circle
            class="size-2"
            :class="queue.size > 0 ? 'text-warning' : 'text-success'"
          />
        </div>

        <!-- Queue name -->
        <div class="min-w-0 flex-1">
          <div class="truncate text-sm font-medium font-mono">{{ queue.name }}</div>
        </div>

        <!-- Size -->
        <div class="shrink-0 text-right">
          <span
            class="font-mono text-sm font-semibold tabular-nums"
            :class="queue.size > 0 ? 'text-warning' : 'text-muted-foreground'"
          >
            {{ queue.size.toLocaleString() }}
          </span>
          <span class="ml-1 text-xs text-muted-foreground">{{ $t("common.messages") }}</span>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState
      :icon="IconLayerGroup"
      :title="$t('shopDetail.noQueues')"
      :description="$t('shopDetail.noQueueDescription')"
      size="sm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import StatCard from "@/components/StatCard.vue";
import EmptyState from "@/components/EmptyState.vue";

import IconLayerGroup from "~icons/fa6-solid/layer-group";
import IconEnvelope from "~icons/fa6-solid/envelope";
import IconGaugeHigh from "~icons/fa6-solid/gauge-high";

const { environment } = useEnvironmentDetail();

const queues = computed(() => environment.value?.queues ?? []);

const totalSize = computed(() => queues.value.reduce((sum, q) => sum + q.size, 0));
const busyQueues = computed(() => queues.value.filter((q) => q.size > 0).length);

// Sort: queues with messages first (desc by size), then empty alphabetical
const sortedQueues = computed(() =>
  [...queues.value].sort((a, b) => {
    if (a.size > 0 && b.size === 0) return -1;
    if (a.size === 0 && b.size > 0) return 1;
    if (a.size !== b.size) return b.size - a.size;
    return a.name.localeCompare(b.name);
  }),
);
</script>
