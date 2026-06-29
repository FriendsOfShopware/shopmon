<template>
  <div class="overflow-x-auto rounded-xl border">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b bg-muted/40">
          <th class="px-4 py-2 text-left font-medium">{{ $t("settings.eventTypes") }}</th>
          <th
            v-for="ch in matrixChannels"
            :key="ch.channel"
            class="px-4 py-2 text-center font-medium whitespace-nowrap"
          >
            {{ $t(ch.labelKey) }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="et in eventTypes" :key="et.type" class="border-b last:border-0">
          <td class="px-4 py-2.5 font-medium whitespace-nowrap">{{ eventTypeLabel(et.type) }}</td>
          <td v-for="ch in matrixChannels" :key="ch.channel" class="px-4 py-2.5 text-center">
            <template v-if="et.defaultChannels.includes(ch.channel)">
              <!-- Global scope: a simple on/off switch per cell. -->
              <Switch
                v-if="environmentId == null"
                class="align-middle"
                :model-value="eventChannelEnabled(et.type, ch.channel)"
                @update:model-value="(v: boolean) => setEventChannel(et.type, ch.channel, v)"
              />
              <!-- Environment scope: inherit / on / off, overriding the global cell. -->
              <div v-else class="inline-flex rounded-lg border bg-muted/50 p-0.5">
                <button
                  v-for="opt in triStateOptions"
                  :key="opt.value"
                  type="button"
                  :class="[
                    'rounded-md px-2 py-0.5 text-xs font-medium transition-colors',
                    envEventChannelState(environmentId, et.type, ch.channel) === opt.value
                      ? 'bg-background text-foreground shadow-sm'
                      : 'text-muted-foreground hover:text-foreground',
                  ]"
                  @click="setEnvEventChannel(environmentId, et.type, ch.channel, opt.value)"
                >
                  {{ $t(opt.labelKey) }}
                </button>
              </div>
            </template>
            <span v-else class="text-muted-foreground">—</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

import { Switch } from "@/components/ui/switch";
import { useNotificationPreferences } from "@/composables/useNotificationPreferences";

// When environmentId is set the matrix renders per-environment tri-state
// overrides; otherwise it renders the global on/off cells.
const props = defineProps<{ environmentId?: number }>();

const {
  eventTypes,
  channelList,
  triStateOptions,
  eventTypeLabel,
  eventChannelEnabled,
  setEventChannel,
  envEventChannelState,
  setEnvEventChannel,
} = useNotificationPreferences();

// Columns are the union of every event's default channels, ordered to match
// channelList. Channels that no event uses are omitted entirely.
const matrixChannels = computed(() => {
  const used = new Set<string>();
  for (const et of eventTypes.value) {
    for (const ch of et.defaultChannels) used.add(ch);
  }
  return channelList.filter((c) => used.has(c.channel));
});
</script>
