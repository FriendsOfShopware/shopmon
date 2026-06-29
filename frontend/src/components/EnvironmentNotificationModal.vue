<template>
  <Dialog :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <DialogContent class="max-w-2xl">
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

        <!-- Per-environment overrides (only meaningful while watching) -->
        <div v-if="isSubscribed && environment" class="space-y-2">
          <p class="text-xs text-muted-foreground">{{ $t("environment.perEventHint") }}</p>
          <NotificationMatrix :environment-id="environment.id" />
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

import NotificationMatrix from "@/components/NotificationMatrix.vue";
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
const { loadAll } = useNotificationPreferences();

// Load when the modal opens (and whenever it reopens) so it reflects current state.
watch(
  () => props.open,
  (open) => {
    if (open) loadAll();
  },
);
</script>
