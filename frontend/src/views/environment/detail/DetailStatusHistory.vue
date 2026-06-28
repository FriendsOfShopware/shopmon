<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-16 text-center text-sm text-muted-foreground">
      {{ $t("common.loading") }}
    </div>

    <EmptyState
      v-else-if="!events.length"
      :icon="IconClock"
      :title="$t('statusHistory.empty')"
      :description="$t('statusHistory.emptyHint')"
      size="sm"
    />

    <div v-else class="space-y-3">
      <div v-for="event in events" :key="event.id" class="rounded-xl border px-4 py-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2 text-sm font-medium">
            <StatusIcon :status="event.oldStatus" />
            <icon-fa6-solid:arrow-right class="size-3 text-muted-foreground" />
            <StatusIcon :status="event.newStatus" />
            <span class="capitalize">{{ event.oldStatus }} → {{ event.newStatus }}</span>
          </div>
          <div class="text-xs text-muted-foreground">
            {{ formatDateTime(event.createdAt) }}
          </div>
        </div>

        <ul v-if="event.reasons.length" class="mt-3 space-y-1.5 border-t pt-3">
          <li
            v-for="(reason, index) in event.reasons"
            :key="index"
            class="flex items-start gap-2 text-sm"
          >
            <StatusIcon :status="reason.level" class="mt-0.5 shrink-0" />
            <span>{{ translateReason(reason) }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";

import { api } from "@/helpers/api";
import { formatDateTime } from "@/helpers/formatter";
import { translateReason } from "@/helpers/i18n";
import { useAlert } from "@/composables/useAlert";
import type { components } from "@/types/api";

import StatusIcon from "@/components/StatusIcon.vue";
import EmptyState from "@/components/EmptyState.vue";
import IconClock from "~icons/fa6-solid/clock-rotate-left";

const route = useRoute();
const { error } = useAlert();

const events = ref<components["schemas"]["StatusEvent"][]>([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const { data } = await api.GET("/environments/{environmentId}/status-events", {
      params: { path: { environmentId: Number(route.params.environmentId) } },
    });
    events.value = data ?? [];
  } catch (err) {
    error(err instanceof Error ? err.message : String(err));
  } finally {
    loading.value = false;
  }
});
</script>
