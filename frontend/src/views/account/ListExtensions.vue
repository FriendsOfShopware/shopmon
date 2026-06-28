<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">{{ $t("extensions.title") }}</h1>
        <p class="mt-1 text-sm text-muted-foreground">{{ $t("extensions.subtitle") }}</p>
      </div>
      <div
        v-if="extensions.length > 0"
        class="flex items-center gap-4 text-sm text-muted-foreground"
      >
        <span>{{ $t("extensions.summary", { count: extensions.length }, extensions.length) }}</span>
        <span v-if="outdatedCount > 0" class="text-warning">
          {{ $t("extensions.updatesSummary", { count: outdatedCount }, outdatedCount) }}
        </span>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Skeleton v-for="i in 8" :key="i" class="h-40 rounded-xl" />
    </div>

    <!-- Empty: no extensions at all -->
    <EmptyState
      v-else-if="extensions.length === 0"
      :icon="IconPuzzlePiece"
      :title="$t('extensions.noExtensions')"
      :description="$t('extensions.noExtensionsHint')"
    >
      <Button as-child>
        <RouterLink :to="{ name: 'account.environments.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("environment.addEnvironment") }}
        </RouterLink>
      </Button>
    </EmptyState>

    <template v-else>
      <!-- Toolbar: search + filters -->
      <FilterSearchBar
        v-model:filter="activeFilter"
        v-model:search="term"
        :filters="filters"
        :search-placeholder="$t('shopDetail.searchExtensions')"
      />

      <!-- Card grid -->
      <div
        v-if="filteredExtensions.length > 0"
        class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
      >
        <RouterLink
          v-for="ext in filteredExtensions"
          :key="ext.name"
          :to="{ name: 'account.extension.detail', params: { name: ext.name } }"
          class="group flex h-full flex-col rounded-xl border bg-card p-4 text-left shadow-sm transition-all hover:-translate-y-0.5 hover:border-primary/30 hover:shadow-md"
          :class="{ 'opacity-60': isInactive(ext) }"
        >
          <div class="flex items-start gap-3">
            <ExtensionIcon :src="ext.iconUrl" :alt="ext.label" :size="44" />
            <div class="min-w-0 flex-1">
              <h3 class="truncate font-semibold transition-colors group-hover:text-primary">
                {{ ext.label }}
              </h3>
              <p class="truncate font-mono text-xs text-muted-foreground">{{ ext.name }}</p>
            </div>
            <Badge
              v-if="isInactive(ext)"
              variant="secondary"
              class="shrink-0 text-[10px] uppercase"
            >
              {{ $t("extensions.inactive") }}
            </Badge>
          </div>

          <div class="mt-3 min-h-5">
            <RatingStars v-if="ext.ratingAverage !== null" :rating="ext.ratingAverage" />
            <span v-else class="text-xs text-muted-foreground/70">{{
              $t("extensions.noRating")
            }}</span>
          </div>

          <div class="mt-auto flex items-center justify-between gap-2 pt-4">
            <span class="text-xs text-muted-foreground">
              {{
                $t("extensions.usedIn", { count: ext.environments.length }, ext.environments.length)
              }}
            </span>
            <Badge
              v-if="updateCount(ext) > 0"
              variant="outline"
              class="border-warning/30 bg-warning/10 text-warning"
            >
              <icon-fa6-solid:arrow-up class="mr-1 size-2.5" />
              {{ $t("extensions.updatesAvailable", { count: updateCount(ext) }, updateCount(ext)) }}
            </Badge>
            <Badge v-else variant="outline" class="border-success/30 bg-success/10 text-success">
              <icon-fa6-solid:check class="mr-1 size-2.5" />
              {{ $t("extensions.upToDate") }}
            </Badge>
          </div>
        </RouterLink>
      </div>

      <!-- Empty: no filter/search match -->
      <EmptyState
        v-else
        :icon="IconMagnifyingGlass"
        :title="$t('extensions.noResults')"
        :description="$t('extensions.noResultsHint')"
        size="sm"
      >
        <Button variant="outline" size="sm" @click="clearFilters">
          {{ $t("extensions.clearFilters") }}
        </Button>
      </EmptyState>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import RatingStars from "@/components/RatingStars.vue";
import EmptyState from "@/components/EmptyState.vue";
import FilterSearchBar from "@/components/FilterSearchBar.vue";
import ExtensionIcon from "@/components/ExtensionIcon.vue";
import IconPuzzlePiece from "~icons/fa6-solid/puzzle-piece";
import IconMagnifyingGlass from "~icons/fa6-solid/magnifying-glass";
import { api, apiLanguage } from "@/helpers/api";
import {
  type AccountExtension,
  hasUpdate,
  isInactive,
  updateCount,
} from "@/composables/useAccountExtensions";

const { t } = useI18n();

const term = ref("");
const activeFilter = ref<"all" | "updates" | "inactive">("all");
const extensions = ref<AccountExtension[]>([]);
const loading = ref(true);

const filters = [
  { label: t("extensions.filterAll"), value: "all" },
  { label: t("extensions.filterUpdates"), value: "updates" },
  { label: t("extensions.filterInactive"), value: "inactive" },
];

api
  .GET("/account/extensions", { params: { query: { language: apiLanguage() } } })
  .then(({ data }) => {
    if (data) extensions.value = data;
    loading.value = false;
  });

const outdatedCount = computed(() => extensions.value.filter((e) => hasUpdate(e)).length);

function clearFilters() {
  term.value = "";
  activeFilter.value = "all";
}

const filteredExtensions = computed(() => {
  let list = [...extensions.value];

  list.sort((a, b) => {
    if (a.active !== b.active) return a.active ? -1 : 1;
    if (a.installed !== b.installed) return a.installed ? -1 : 1;
    return a.label.localeCompare(b.label);
  });

  if (activeFilter.value === "updates") list = list.filter((e) => updateCount(e) > 0);
  if (activeFilter.value === "inactive") list = list.filter((e) => isInactive(e));

  if (term.value.length >= 2) {
    const q = term.value.toLowerCase();
    list = list.filter(
      (e) => e.label.toLowerCase().includes(q) || e.name.toLowerCase().includes(q),
    );
  }

  return list;
});
</script>
