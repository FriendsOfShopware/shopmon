<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-3">
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg bg-primary/10">
            <icon-fa6-solid:layer-group class="size-4 text-primary" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ queues.length }}</div>
            <div class="text-xs text-muted-foreground">Queues</div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg" :class="totalSize > 0 ? 'bg-warning/10' : 'bg-success/10'">
            <icon-fa6-solid:envelope class="size-4" :class="totalSize > 0 ? 'text-warning' : 'text-success'" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ totalSize.toLocaleString() }}</div>
            <div class="text-xs text-muted-foreground">Pending messages</div>
          </div>
        </CardContent>
      </Card>
      <Card class="col-span-2 lg:col-span-1">
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg" :class="busyQueues > 0 ? 'bg-warning/10' : 'bg-success/10'">
            <icon-fa6-solid:gauge-high class="size-4" :class="busyQueues > 0 ? 'text-warning' : 'text-success'" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ busyQueues }}</div>
            <div class="text-xs text-muted-foreground">Queues with backlog</div>
          </div>
        </CardContent>
      </Card>
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
        <div class="flex size-8 shrink-0 items-center justify-center rounded-lg" :class="queue.size > 0 ? 'bg-warning/10' : 'bg-success/10'">
          <icon-fa6-solid:circle class="size-2" :class="queue.size > 0 ? 'text-warning' : 'text-success'" />
        </div>

        <!-- Queue name -->
        <div class="min-w-0 flex-1">
          <div class="truncate text-sm font-medium font-mono">{{ queue.name }}</div>
        </div>

        <!-- Size -->
        <div class="shrink-0 text-right">
          <span class="font-mono text-sm font-semibold tabular-nums" :class="queue.size > 0 ? 'text-warning' : 'text-muted-foreground'">
            {{ queue.size.toLocaleString() }}
          </span>
          <span class="ml-1 text-xs text-muted-foreground">messages</span>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center">
      <icon-fa6-solid:layer-group class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">No queues</h3>
      <p class="text-sm text-muted-foreground">This environment has no message queues.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import { Card, CardContent } from "@/components/ui/card";

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
