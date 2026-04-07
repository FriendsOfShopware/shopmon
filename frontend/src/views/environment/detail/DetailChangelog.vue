<template>
  <div v-if="environment" class="space-y-6">
    <!-- Empty state -->
    <div
      v-if="!environment.changelogs?.length"
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:clock-rotate-left class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">{{ $t("shopDetail.noChangesRecorded") }}</h3>
      <p class="text-sm text-muted-foreground">{{ $t("shopDetail.changesWillAppear") }}</p>
    </div>

    <!-- Timeline -->
    <div v-else class="space-y-3">
      <div
        v-for="entry in environment.changelogs"
        :key="entry.id"
        class="group rounded-xl border bg-card transition-colors hover:border-primary/20"
      >
        <!-- Header row — always visible -->
        <button
          class="flex w-full items-center gap-4 px-4 py-3 text-left"
          @click="toggle(entry.id)"
        >
          <!-- Date -->
          <div class="shrink-0 text-sm tabular-nums text-muted-foreground">
            {{ formatDate(entry.date) }}
          </div>

          <!-- Summary -->
          <div class="min-w-0 flex-1">
            <!-- Shopware version change -->
            <div
              v-if="entry.oldShopwareVersion && entry.newShopwareVersion"
              class="flex items-center gap-2"
            >
              <Badge class="bg-primary/10 text-primary border-primary/20 font-mono text-xs">
                {{ entry.oldShopwareVersion }} → {{ entry.newShopwareVersion }}
              </Badge>
            </div>

            <!-- Extension change summary -->
            <div
              class="flex flex-wrap items-center gap-2"
              :class="{ 'mt-1': entry.oldShopwareVersion }"
            >
              <Badge
                v-for="(count, state) in groupedStates(entry)"
                :key="state"
                variant="secondary"
                class="gap-1 text-xs capitalize"
              >
                <icon-fa6-solid:arrow-up v-if="state === 'updated'" class="size-2.5 text-info" />
                <icon-fa6-solid:plus
                  v-else-if="state === 'installed'"
                  class="size-2.5 text-success"
                />
                <icon-fa6-solid:trash
                  v-else-if="state === 'removed'"
                  class="size-2.5 text-destructive"
                />
                <icon-fa6-solid:toggle-on
                  v-else-if="state === 'activated'"
                  class="size-2.5 text-success"
                />
                <icon-fa6-solid:toggle-off
                  v-else-if="state === 'deactivated'"
                  class="size-2.5 text-warning"
                />
                <icon-fa6-solid:circle v-else class="size-2 text-muted-foreground" />
                {{ count }} {{ state }}
              </Badge>
            </div>
          </div>

          <!-- Expand chevron -->
          <icon-fa6-solid:chevron-down
            :class="[
              'size-3 shrink-0 text-muted-foreground transition-transform',
              expanded.has(entry.id) ? 'rotate-180' : '',
            ]"
          />
        </button>

        <!-- Expanded detail -->
        <div v-if="expanded.has(entry.id)" class="border-t px-4 py-3">
          <div class="space-y-2">
            <div
              v-for="ext in entry.extensions"
              :key="ext.name"
              class="flex items-center gap-3 rounded-lg bg-muted/50 px-3 py-2 text-sm"
            >
              <!-- State icon -->
              <div class="shrink-0">
                <icon-fa6-solid:arrow-up
                  v-if="ext.state === 'updated'"
                  class="size-3.5 text-info"
                />
                <icon-fa6-solid:plus
                  v-else-if="ext.state === 'installed'"
                  class="size-3.5 text-success"
                />
                <icon-fa6-solid:trash
                  v-else-if="ext.state === 'removed'"
                  class="size-3.5 text-destructive"
                />
                <icon-fa6-solid:toggle-on
                  v-else-if="ext.state === 'activated'"
                  class="size-3.5 text-success"
                />
                <icon-fa6-solid:toggle-off
                  v-else-if="ext.state === 'deactivated'"
                  class="size-3.5 text-warning"
                />
                <icon-fa6-solid:circle v-else class="size-2.5 text-muted-foreground" />
              </div>

              <!-- Extension info -->
              <div class="min-w-0 flex-1">
                <span class="font-medium">{{ ext.label }}</span>
                <span class="ml-1 text-muted-foreground">({{ ext.name }})</span>
              </div>

              <!-- Version change -->
              <div
                v-if="ext.state === 'updated' && ext.oldVersion && ext.newVersion"
                class="shrink-0"
              >
                <Badge variant="secondary" class="font-mono text-xs">
                  {{ ext.oldVersion }} → {{ ext.newVersion }}
                </Badge>
              </div>
              <div v-else-if="ext.newVersion || ext.oldVersion" class="shrink-0">
                <Badge variant="secondary" class="font-mono text-xs">
                  {{ ext.newVersion || ext.oldVersion }}
                </Badge>
              </div>

              <!-- State label -->
              <Badge variant="outline" class="shrink-0 capitalize text-xs">{{ ext.state }}</Badge>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import { formatDate } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import { Badge } from "@/components/ui/badge";

type AccountChangelog = components["schemas"]["AccountChangelog"];

const { environment } = useEnvironmentDetail();

const expanded = reactive(new Set<number>());

function toggle(id: number) {
  if (expanded.has(id)) {
    expanded.delete(id);
  } else {
    expanded.add(id);
  }
}

function groupedStates(entry: AccountChangelog): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const ext of entry.extensions) {
    counts[ext.state] = (counts[ext.state] ?? 0) + 1;
  }
  return counts;
}
</script>
