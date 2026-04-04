<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary cards -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg bg-primary/10">
            <icon-fa6-solid:puzzle-piece class="size-4 text-primary" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ extensions.length }}</div>
            <div class="text-xs text-muted-foreground">Total</div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg bg-success/10">
            <icon-fa6-solid:circle-check class="size-4 text-success" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ counts.active }}</div>
            <div class="text-xs text-muted-foreground">Active</div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg bg-warning/10">
            <icon-fa6-solid:arrow-up class="size-4 text-warning" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ counts.outdated }}</div>
            <div class="text-xs text-muted-foreground">Updates</div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-4">
          <div class="flex size-9 items-center justify-center rounded-lg bg-muted">
            <icon-fa6-solid:power-off class="size-4 text-muted-foreground" />
          </div>
          <div>
            <div class="text-2xl font-bold tabular-nums">{{ counts.inactive }}</div>
            <div class="text-xs text-muted-foreground">Inactive</div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Filter + search bar -->
    <div class="flex flex-wrap items-center justify-between gap-3 max-sm:flex-col max-sm:w-full">
      <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
        <button
          v-for="f in filters"
          :key="f.value"
          :class="[
            'rounded-md px-3 py-1 text-sm font-medium transition-colors',
            activeFilter === f.value ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="activeFilter = f.value"
        >
          {{ f.label }}
        </button>
      </div>

      <div class="relative">
        <icon-fa6-solid:magnifying-glass class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="searchQuery"
          type="search"
          placeholder="Search extensions..."
          class="h-8 w-full sm:w-56 pl-8 text-sm"
        />
      </div>
    </div>

    <!-- Extension list -->
    <div v-if="filteredExtensions.length > 0" class="space-y-2">
      <div
        v-for="ext in filteredExtensions"
        :key="ext.name"
        :class="[
          'flex items-center gap-4 rounded-xl border px-4 py-3 transition-colors',
          !ext.active && ext.installed ? 'opacity-60' : '',
          !ext.installed ? 'opacity-40' : '',
          hasUpdate(ext) ? 'border-warning/20 bg-warning/5' : '',
        ]"
      >
        <!-- Status indicator -->
        <StatusIcon :status="getExtensionState(ext)" class="shrink-0" />

        <!-- Name + technical name -->
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <component
              :is="ext.storeLink ? 'a' : 'span'"
              v-bind="ext.storeLink ? { href: ext.storeLink, target: '_blank', class: 'hover:text-primary transition-colors' } : {}"
              class="font-medium truncate"
            >
              {{ ext.label }}
              <icon-fa6-solid:arrow-up-right-from-square v-if="ext.storeLink" class="ml-1 inline size-2.5 text-muted-foreground" />
            </component>
          </div>
          <div class="text-xs text-muted-foreground">{{ ext.name }}</div>
        </div>

        <!-- Version info -->
        <div class="flex shrink-0 items-center gap-2">
          <Badge variant="secondary" class="font-mono text-xs">{{ ext.version }}</Badge>
          <template v-if="hasUpdate(ext)">
            <icon-fa6-solid:arrow-right class="size-2.5 text-muted-foreground" />
            <Badge
              class="cursor-pointer font-mono text-xs bg-warning/10 text-warning border-warning/30 hover:bg-warning/20"
              @click="openExtensionChangelog(ext)"
            >
              {{ ext.latestVersion }}
            </Badge>
          </template>
        </div>

        <!-- Rating -->
        <div class="hidden shrink-0 lg:block">
          <RatingStars :rating="ext.ratingAverage ?? null" />
        </div>

        <!-- Installed date -->
        <div v-if="ext.installedAt" class="hidden shrink-0 text-xs tabular-nums text-muted-foreground xl:block">
          {{ formatDate(ext.installedAt) }}
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center">
      <icon-fa6-solid:puzzle-piece class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">No extensions found</h3>
      <p class="text-sm text-muted-foreground">
        <template v-if="searchQuery">No extensions match "{{ searchQuery }}".</template>
        <template v-else>No extensions match the current filter.</template>
      </p>
    </div>
  </div>

  <!-- Extension Changelog Modal -->
  <ExtensionChangelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { formatDate } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import StatusIcon from "@/components/StatusIcon.vue";
import RatingStars from "@/components/RatingStars.vue";

type Extension = components["schemas"]["EnvironmentExtension"];

const { environment } = useEnvironmentDetail();

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

const activeFilter = ref<"all" | "active" | "inactive" | "updates">("all");
const searchQuery = ref("");

const filters = [
  { label: "All", value: "all" as const },
  { label: "Active", value: "active" as const },
  { label: "Inactive", value: "inactive" as const },
  { label: "Updates", value: "updates" as const },
];

const extensions = computed(() => environment.value?.extensions ?? []);

const counts = computed(() => ({
  active: extensions.value.filter((e) => e.active && e.installed).length,
  inactive: extensions.value.filter((e) => !e.active || !e.installed).length,
  outdated: extensions.value.filter((e) => hasUpdate(e)).length,
}));

function hasUpdate(ext: Extension): boolean {
  return !!(ext.installed && ext.latestVersion && ext.version !== ext.latestVersion);
}

function getExtensionState(extension: Extension) {
  if (!extension.installed) return "not installed";
  if (extension.active) return "active";
  return "inactive";
}

const filteredExtensions = computed(() => {
  let list = [...extensions.value];

  // Sort: active first, then by label
  list.sort((a, b) => {
    if (a.active !== b.active) return a.active ? -1 : 1;
    if (a.installed !== b.installed) return a.installed ? -1 : 1;
    return a.label.localeCompare(b.label);
  });

  switch (activeFilter.value) {
    case "active":
      list = list.filter((e) => e.active && e.installed);
      break;
    case "inactive":
      list = list.filter((e) => !e.active || !e.installed);
      break;
    case "updates":
      list = list.filter((e) => hasUpdate(e));
      break;
  }

  if (searchQuery.value.length >= 2) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter((e) => e.label.toLowerCase().includes(q) || e.name.toLowerCase().includes(q));
  }

  return list;
});
</script>
