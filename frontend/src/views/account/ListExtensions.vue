<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t('nav.myExtensions') }}</h1>

    <!-- Empty state -->
    <div v-if="extensions && extensions.length === 0" class="flex flex-col items-center gap-4 rounded-xl border border-dashed py-20 text-center">
      <div class="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
        <icon-fa6-solid:puzzle-piece class="size-6 text-primary" />
      </div>
      <h2 class="text-xl font-semibold">{{ $t('shopDetail.extensions') }}</h2>
      <p class="max-w-md text-muted-foreground">{{ $t("common.getStartedElement") }}</p>
      <Button as-child>
        <RouterLink :to="{ name: 'account.environments.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("environment.addEnvironment") }}
        </RouterLink>
      </Button>
    </div>

    <template v-else-if="extensions && extensions.length > 0">
      <!-- Summary + search -->
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-4 text-sm text-muted-foreground">
          <span><strong class="text-foreground tabular-nums">{{ extensions.length }}</strong> extensions</span>
          <span><strong class="tabular-nums" :class="outdatedCount > 0 ? 'text-warning' : 'text-foreground'">{{ outdatedCount }}</strong> updates</span>
        </div>
        <div class="relative">
          <icon-fa6-solid:magnifying-glass class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground" />
          <Input v-model="term" :placeholder="$t('common.search')" class="h-8 w-56 pl-8 text-sm" />
        </div>
      </div>

      <!-- Extension list -->
      <div class="space-y-2">
        <div
          v-for="ext in filteredExtensions"
          :key="ext.name"
          :class="[
            'rounded-xl border transition-colors',
            hasUpdate(ext) ? 'border-warning/20 bg-warning/5' : '',
            !ext.active && ext.installed ? 'opacity-60' : '',
            !ext.installed ? 'opacity-40' : '',
          ]"
        >
          <!-- Main row -->
          <div class="flex items-center gap-4 px-4 py-3">
            <StatusIcon :status="getExtensionState(ext)" class="shrink-0" />

            <!-- Name -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <component
                  :is="ext.storeLink ? 'a' : 'span'"
                  v-bind="ext.storeLink ? { href: ext.storeLink, target: '_blank', class: 'hover:text-primary transition-colors' } : {}"
                  class="truncate font-medium"
                >
                  {{ ext.label }}
                  <icon-fa6-solid:arrow-up-right-from-square v-if="ext.storeLink" class="ml-1 inline size-2.5 text-muted-foreground" />
                </component>
              </div>
              <div class="text-xs text-muted-foreground">{{ ext.name }}</div>
            </div>

            <!-- Version -->
            <div class="hidden shrink-0 items-center gap-2 sm:flex">
              <Badge variant="secondary" class="font-mono text-xs">{{ ext.latestVersion || ext.version }}</Badge>
              <template v-if="hasUpdate(ext)">
                <button class="text-xs text-warning hover:underline" @click="openExtensionChangelog(ext)">
                  update available
                </button>
              </template>
            </div>

            <!-- Rating -->
            <div class="hidden shrink-0 lg:block">
              <RatingStars :rating="ext.ratingAverage" />
            </div>

            <!-- Expand -->
            <button class="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-accent" @click="toggle(ext.name)">
              <icon-fa6-solid:chevron-down :class="['size-2.5 transition-transform', expanded.has(ext.name) ? 'rotate-180' : '']" />
            </button>
          </div>

          <!-- Expanded: environments using this extension -->
          <div v-if="expanded.has(ext.name)" class="border-t px-4 py-3">
            <div class="mb-2 text-xs font-medium text-muted-foreground">Installed in {{ ext.environments.length }} environment{{ ext.environments.length !== 1 ? 's' : '' }}</div>
            <div class="space-y-1.5">
              <RouterLink
                v-for="env in ext.environments"
                :key="env.environmentId"
                :to="{
                  name: 'account.environments.detail',
                  params: { organizationId: env.environmentOrganizationId, environmentId: env.environmentId },
                }"
                class="flex items-center gap-3 rounded-lg bg-muted/50 px-3 py-2 text-sm transition-colors hover:bg-accent"
              >
                <StatusIcon :status="getEnvExtensionState(env)" class="shrink-0" />
                <span class="min-w-0 flex-1 truncate">{{ env.environmentName }}</span>
                <Badge variant="secondary" class="font-mono text-[10px]">{{ env.version }}</Badge>
                <icon-fa6-solid:arrow-up v-if="env.version !== env.latestVersion && env.latestVersion" class="size-3 text-warning" :title="'Update to ' + env.latestVersion" />
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>

  <ExtensionChangelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />
</template>

<script setup lang="ts">
import { ref, computed, reactive } from "vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import StatusIcon from "@/components/StatusIcon.vue";
import RatingStars from "@/components/RatingStars.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";

type AccountExtension = components["schemas"]["AccountExtension"];
type AccountExtensionEnvironment = components["schemas"]["AccountExtensionEnvironment"];

const term = ref("");
const extensions = ref<AccountExtension[]>([]);
const expanded = reactive(new Set<string>());

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

api.GET("/account/extensions").then(({ data }) => {
  if (data) extensions.value = data;
});

function hasUpdate(ext: AccountExtension): boolean {
  return !!(ext.installed && ext.latestVersion && ext.version !== ext.latestVersion);
}

function getExtensionState(ext: AccountExtension) {
  if (!ext.installed) return "not installed";
  if (ext.active) return "active";
  return "inactive";
}

function getEnvExtensionState(env: AccountExtensionEnvironment) {
  if (!env.installed) return "not installed";
  if (env.active) return "active";
  return "inactive";
}

function toggle(name: string) {
  if (expanded.has(name)) {
    expanded.delete(name);
  } else {
    expanded.add(name);
  }
}

const outdatedCount = computed(() => extensions.value.filter((e) => hasUpdate(e)).length);

const filteredExtensions = computed(() => {
  let list = [...extensions.value];

  list.sort((a, b) => {
    if (a.active !== b.active) return a.active ? -1 : 1;
    if (a.installed !== b.installed) return a.installed ? -1 : 1;
    return a.label.localeCompare(b.label);
  });

  if (term.value.length >= 2) {
    const q = term.value.toLowerCase();
    list = list.filter((e) => e.label.toLowerCase().includes(q) || e.name.toLowerCase().includes(q));
  }

  return list;
});
</script>
