<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary cards -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard :icon="IconPuzzlePiece" :value="extensions.length" label="Total" />
      <StatCard :icon="IconCircleCheck" :value="counts.active" label="Active" color="success" />
      <StatCard :icon="IconArrowUp" :value="counts.outdated" label="Updates" color="warning" />
      <StatCard :icon="IconPowerOff" :value="counts.inactive" label="Inactive" color="muted" />
    </div>

    <!-- Filter + search bar -->
    <FilterSearchBar
      v-model:filter="activeFilter"
      v-model:search="searchQuery"
      :filters="filters"
      search-placeholder="Search extensions..."
    />

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
              v-bind="
                ext.storeLink
                  ? {
                      href: ext.storeLink,
                      target: '_blank',
                      class: 'hover:text-primary transition-colors',
                    }
                  : {}
              "
              class="font-medium truncate"
            >
              {{ ext.label }}
              <icon-fa6-solid:arrow-up-right-from-square
                v-if="ext.storeLink"
                class="ml-1 inline size-2.5 text-muted-foreground"
              />
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
        <div
          v-if="ext.installedAt"
          class="hidden shrink-0 text-xs tabular-nums text-muted-foreground xl:block"
        >
          {{ formatDate(ext.installedAt) }}
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <EmptyState v-else :icon="IconPuzzlePiece" title="No extensions found" size="sm">
      <p class="text-sm text-muted-foreground">
        <template v-if="searchQuery">No extensions match "{{ searchQuery }}".</template>
        <template v-else>No extensions match the current filter.</template>
      </p>
    </EmptyState>
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

import { Badge } from "@/components/ui/badge";
import StatusIcon from "@/components/StatusIcon.vue";
import RatingStars from "@/components/RatingStars.vue";
import StatCard from "@/components/StatCard.vue";
import EmptyState from "@/components/EmptyState.vue";
import FilterSearchBar from "@/components/FilterSearchBar.vue";

import IconPuzzlePiece from "~icons/fa6-solid/puzzle-piece";
import IconCircleCheck from "~icons/fa6-solid/circle-check";
import IconArrowUp from "~icons/fa6-solid/arrow-up";
import IconPowerOff from "~icons/fa6-solid/power-off";

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
    list = list.filter(
      (e) => e.label.toLowerCase().includes(q) || e.name.toLowerCase().includes(q),
    );
  }

  return list;
});
</script>
