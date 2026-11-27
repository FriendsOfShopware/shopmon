<template>
  <div class="min-w-0 space-y-2">
    <p
      v-if="destination && destinationPrefs.loading.value"
      role="status"
      class="text-sm text-muted-foreground"
    >
      {{ $t("destinations.loadingPreferences") }}
    </p>
    <div
      v-else-if="destination && destinationPrefs.failed.value"
      role="alert"
      class="text-sm text-destructive"
    >
      {{ $t("destinations.failedLoadPreferences") }}
      <button type="button" class="ml-2 underline" @click="destinationPrefs.load">
        {{ $t("destinations.retry") }}
      </button>
    </div>
    <div v-else class="min-w-0 overflow-x-auto rounded-xl border">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b bg-muted/40">
            <th class="px-4 py-2 text-left font-medium">{{ $t("settings.eventTypes") }}</th>
            <th
              v-for="ch in matrixChannels"
              :key="ch.channel"
              class="px-4 py-2 text-center font-medium whitespace-nowrap"
            >
              {{ $t(ch.shortLabelKey) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="et in eventTypes" :key="et.type" class="border-b last:border-0">
            <td class="px-4 py-2.5 font-medium">{{ eventTypeLabel(et.type) }}</td>
            <td v-for="ch in matrixChannels" :key="ch.channel" class="px-4 py-2.5 text-center">
              <template v-if="et.defaultChannels.includes(ch.channel)">
                <!-- Global scope: a simple on/off switch per cell. -->
                <Switch
                  v-if="environmentId == null"
                  class="align-middle"
                  :disabled="controlsDisabled"
                  :aria-label="`${eventTypeLabel(et.type)} — ${$t(ch.shortLabelKey)}`"
                  :model-value="enabled(et.type, ch.channel)"
                  @update:model-value="(v: boolean) => setEnabled(et.type, ch.channel, v)"
                />
                <!-- Environment scope: inherit / on / off, overriding the global cell. -->
                <div v-else class="inline-flex rounded-lg border bg-muted/50 p-0.5">
                  <button
                    v-for="opt in triStateOptions"
                    :key="opt.value"
                    type="button"
                    :disabled="controlsDisabled"
                    :aria-pressed="state(et.type, ch.channel) === opt.value"
                    :class="[
                      'rounded-md px-2 py-0.5 text-xs font-medium transition-colors',
                      state(et.type, ch.channel) === opt.value
                        ? 'bg-background text-foreground shadow-sm'
                        : 'text-muted-foreground hover:text-foreground',
                    ]"
                    @click="setState(et.type, ch.channel, opt.value)"
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
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from "vue";

import { Switch } from "@/components/ui/switch";
import {
  useNotificationPreferences,
  type ChannelName,
  type TriState,
} from "@/composables/useNotificationPreferences";

import {
  useDestinationNotificationPreferences,
  type PreferenceDestination,
} from "@/composables/useDestinationNotificationPreferences";

// When environmentId is set the matrix renders per-environment tri-state
// overrides; otherwise it renders the global on/off cells.
const props = defineProps<{
  environmentId?: number;
  destination?: PreferenceDestination;
  disabled?: boolean;
}>();

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
  if (props.destination)
    return [{ channel: "destination", shortLabelKey: "settings.channelDestination" }];
  const used = new Set<string>();
  for (const et of eventTypes.value) {
    for (const ch of et.defaultChannels) used.add(ch);
  }
  return channelList.filter((c) => used.has(c.channel));
});
const destinationPrefs = useDestinationNotificationPreferences(
  () => props.destination,
  () => props.environmentId,
);
const controlsDisabled = computed(
  () =>
    props.disabled ||
    (props.destination &&
      (destinationPrefs.busy.value ||
        destinationPrefs.loading.value ||
        destinationPrefs.failed.value)),
);
watch(
  () => [props.destination?.id, props.destination?.organizationId, props.environmentId],
  () => {
    if (props.destination) void destinationPrefs.load();
  },
  { immediate: true },
);
function enabled(event: string, channel: string) {
  return props.destination
    ? destinationPrefs.enabled(event)
    : eventChannelEnabled(event, channel as ChannelName);
}
function state(event: string, channel: string) {
  return props.destination
    ? destinationPrefs.state(event)
    : envEventChannelState(props.environmentId!, event, channel as ChannelName);
}
function setEnabled(event: string, channel: string, value: boolean) {
  if (controlsDisabled.value) return;
  return props.destination
    ? destinationPrefs.set(event, value)
    : setEventChannel(event, channel as ChannelName, value);
}
function setState(event: string, channel: string, value: TriState) {
  if (controlsDisabled.value) return;
  return props.destination
    ? destinationPrefs.set(event, value)
    : setEnvEventChannel(props.environmentId!, event, channel as ChannelName, value);
}
</script>
