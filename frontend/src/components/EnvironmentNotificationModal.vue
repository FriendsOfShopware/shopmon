<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="max-w-md">
      <DialogHeader>
        <DialogTitle>{{ $t("environment.notificationSettings") }}</DialogTitle>
        <DialogDescription>{{ environment?.name }}</DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <!-- Watch toggle -->
        <div class="flex items-center justify-between gap-4 rounded-xl border px-4 py-3">
          <span class="min-w-0">
            <span class="block text-sm font-medium">{{ $t("environment.watchEnvironment") }}</span>
            <span class="block text-xs text-muted-foreground">{{
              $t("environment.watchHint")
            }}</span>
          </span>
          <Switch
            :model-value="isSubscribed"
            :disabled="isSubscribing"
            @update:model-value="toggleNotificationSubscription"
          />
        </div>

        <!-- Per-event-type overrides (only meaningful while watching) -->
        <div v-if="isSubscribed && environment" class="space-y-2">
          <h4 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            {{ $t("settings.eventTypes") }}
          </h4>
          <p class="text-xs text-muted-foreground">{{ $t("environment.perEventHint") }}</p>
          <div
            v-for="et in eventTypes"
            :key="et.type"
            class="flex items-center justify-between gap-3 rounded-xl border px-4 py-3"
          >
            <span class="text-sm font-medium">{{ eventTypeLabel(et.type) }}</span>
            <div class="flex rounded-lg border bg-muted/50 p-0.5">
              <button
                v-for="opt in triStateOptions"
                :key="opt.value"
                type="button"
                :class="[
                  'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                  envEventState(environment.id, et.type) === opt.value
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground',
                ]"
                @click="setEnvEvent(environment.id, et.type, opt.value)"
              >
                {{ $t(opt.labelKey) }}
              </button>
            </div>
          </div>
        </div>
        <p v-else class="text-sm text-muted-foreground">
          {{ $t("environment.watchToConfigure") }}
        </p>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="emit('update:open', false)">{{
          $t("common.close")
        }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { watch } from "vue";

import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useNotificationPreferences } from "@/composables/useNotificationPreferences";

import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ "update:open": [boolean] }>();

const { environment, isSubscribed, isSubscribing, toggleNotificationSubscription } =
  useEnvironmentDetail();
const { eventTypes, triStateOptions, loadAll, eventTypeLabel, envEventState, setEnvEvent } =
  useNotificationPreferences();

// Load when the modal opens (and whenever it reopens) so it reflects current state.
watch(
  () => props.open,
  (open) => {
    if (open) loadAll();
  },
);
</script>
