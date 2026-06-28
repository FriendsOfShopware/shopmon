<template>
  <div class="space-y-6">
    <!-- Back link -->
    <RouterLink
      :to="{ name: 'account.extension.list' }"
      class="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
    >
      <icon-fa6-solid:chevron-left class="size-3" />
      {{ $t("extensions.back") }}
    </RouterLink>

    <!-- Loading -->
    <div v-if="loading" class="space-y-6">
      <Skeleton class="h-32 w-full rounded-xl" />
      <Skeleton class="h-64 w-full rounded-xl" />
    </div>

    <!-- Not found -->
    <EmptyState
      v-else-if="!extension"
      :icon="IconPuzzlePiece"
      :title="$t('extensions.notFound')"
      :description="$t('extensions.notFoundHint')"
    >
      <Button as-child variant="outline">
        <RouterLink :to="{ name: 'account.extension.list' }">
          {{ $t("extensions.back") }}
        </RouterLink>
      </Button>
    </EmptyState>

    <template v-else>
      <!-- Header card -->
      <Card>
        <CardContent class="flex flex-col gap-5 p-5 sm:flex-row sm:items-start sm:p-6">
          <ExtensionIcon :src="extension.iconUrl" :alt="extension.label" :size="64" />

          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
              <h1 class="text-2xl font-bold tracking-tight">{{ extension.label }}</h1>
              <Badge
                v-if="outdated > 0"
                class="border-warning/30 bg-warning/10 text-warning"
                variant="outline"
              >
                <icon-fa6-solid:arrow-up class="mr-1 size-2.5" />
                {{ $t("extensions.updatesAvailable", { count: outdated }, outdated) }}
              </Badge>
              <Badge v-else class="border-success/30 bg-success/10 text-success" variant="outline">
                <icon-fa6-solid:check class="mr-1 size-2.5" />
                {{ $t("extensions.upToDate") }}
              </Badge>
            </div>

            <div
              class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground"
            >
              <a
                v-if="producerWebsite"
                :href="producerWebsite"
                target="_blank"
                rel="noopener noreferrer"
                class="font-medium text-primary hover:underline"
              >
                {{ extension.producerName }}
              </a>
              <span v-else-if="extension.producerName" class="font-medium">{{
                extension.producerName
              }}</span>
              <span v-if="extension.producerName" class="text-muted-foreground/40">·</span>
              <span class="font-mono text-xs">{{ extension.name }}</span>
            </div>

            <p v-if="shortDescription" class="mt-2 text-sm text-muted-foreground">
              {{ shortDescription }}
            </p>

            <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2">
              <RatingStars
                v-if="extension.ratingAverage !== null"
                :rating="extension.ratingAverage"
              />
              <span v-else class="text-sm text-muted-foreground">{{
                $t("extensions.noRating")
              }}</span>

              <span class="hidden h-4 w-px bg-border sm:block" />

              <div class="flex items-center gap-1.5 text-sm">
                <span class="text-muted-foreground">{{ $t("extensions.latestVersion") }}</span>
                <Badge variant="secondary" class="font-mono text-xs">{{
                  extension.latestVersion
                }}</Badge>
              </div>

              <template v-if="extension.installedAt">
                <span class="hidden h-4 w-px bg-border sm:block" />
                <span class="text-sm text-muted-foreground">
                  {{ $t("shopDetail.installedAt") }} {{ formatDate(extension.installedAt) }}
                </span>
              </template>
            </div>
          </div>

          <div class="flex shrink-0 flex-col items-stretch gap-2 sm:w-44">
            <Button v-if="extension.storeLink" as-child>
              <a :href="extension.storeLink" target="_blank" rel="noopener noreferrer">
                {{ $t("extensions.openInStore") }}
                <icon-fa6-solid:arrow-up-right-from-square class="ml-1.5 size-3" />
              </a>
            </Button>
            <div
              class="flex items-center justify-center gap-1.5 rounded-lg border bg-muted/40 px-3 py-1.5 text-xs text-muted-foreground"
            >
              <icon-fa6-solid:server class="size-3" />
              {{ $t("extensions.usedIn", { count: environments.length }, environments.length) }}
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Body: content + screenshots -->
      <div class="grid gap-6" :class="screenshots.length > 0 ? 'lg:grid-cols-[1fr_340px]' : ''">
        <div class="min-w-0">
          <Tabs default-value="environments">
            <TabsList>
              <TabsTrigger value="environments">
                {{ $t("extensions.tabEnvironments") }}
                <Badge
                  v-if="outdated > 0"
                  variant="outline"
                  class="ml-1.5 border-warning/30 bg-warning/10 text-warning"
                >
                  {{ outdated }}
                </Badge>
              </TabsTrigger>
              <TabsTrigger v-if="description" value="description">
                {{ $t("extensions.tabDescription") }}
              </TabsTrigger>
              <TabsTrigger v-if="installationManual" value="installation">
                {{ $t("extensions.tabInstallation") }}
              </TabsTrigger>
              <TabsTrigger value="changelog">{{ $t("extensions.tabChangelog") }}</TabsTrigger>
            </TabsList>

            <!-- Environments tab -->
            <TabsContent value="environments" class="mt-4">
              <Card>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{{ $t("extensions.colEnvironment") }}</TableHead>
                      <TableHead>{{ $t("extensions.colVersion") }}</TableHead>
                      <TableHead>{{ $t("extensions.colState") }}</TableHead>
                      <TableHead>{{ $t("extensions.colUpdate") }}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <template v-for="group in environmentsByShop" :key="group.shopId">
                      <!-- Shop heading -->
                      <TableRow class="bg-muted/30 hover:bg-muted/30">
                        <TableCell colspan="4" class="py-2">
                          <span class="flex items-center gap-1.5 text-xs font-semibold">
                            <icon-fa6-solid:folder class="size-3 text-muted-foreground" />
                            {{ group.shopName }}
                            <span class="font-normal text-muted-foreground">
                              ({{
                                $t(
                                  "extensions.usedIn",
                                  { count: group.environments.length },
                                  group.environments.length,
                                )
                              }})
                            </span>
                          </span>
                        </TableCell>
                      </TableRow>
                      <TableRow v-for="env in group.environments" :key="env.environmentId">
                        <TableCell>
                          <RouterLink
                            :to="{
                              name: 'account.environments.detail.extensions',
                              params: { environmentId: env.environmentId },
                            }"
                            class="flex items-center gap-2.5 pl-4 font-medium hover:text-primary"
                          >
                            <StatusIcon :status="envState(env)" class="shrink-0" />
                            {{ env.environmentName }}
                          </RouterLink>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" class="font-mono text-xs">{{
                            env.version
                          }}</Badge>
                        </TableCell>
                        <TableCell>
                          <span
                            v-if="env.active"
                            class="inline-flex items-center gap-1.5 text-sm text-success"
                          >
                            <span class="size-1.5 rounded-full bg-success" />
                            {{ $t("extensions.stateActive") }}
                          </span>
                          <span
                            v-else
                            class="inline-flex items-center gap-1.5 text-sm text-muted-foreground"
                          >
                            <span class="size-1.5 rounded-full bg-muted-foreground/50" />
                            {{ $t("extensions.stateInactive") }}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge
                            v-if="envHasUpdate(env)"
                            variant="outline"
                            class="border-warning/30 bg-warning/10 font-mono text-warning"
                          >
                            <icon-fa6-solid:arrow-up class="mr-1 size-2.5" />
                            {{ env.latestVersion }}
                          </Badge>
                          <span v-else class="inline-flex items-center gap-1 text-sm text-success">
                            <icon-fa6-solid:check class="size-3" />
                            {{ $t("extensions.current") }}
                          </span>
                        </TableCell>
                      </TableRow>
                    </template>
                  </TableBody>
                </Table>
              </Card>
            </TabsContent>

            <!-- Description tab -->
            <TabsContent v-if="description" value="description" class="mt-4">
              <Card>
                <CardContent class="p-6">
                  <!-- eslint-disable vue/no-v-html -->
                  <div class="richtext" v-html="description" />
                  <!-- eslint-enable vue/no-v-html -->
                </CardContent>
              </Card>
            </TabsContent>

            <!-- Installation tab -->
            <TabsContent v-if="installationManual" value="installation" class="mt-4">
              <Card>
                <CardContent class="p-6">
                  <!-- eslint-disable vue/no-v-html -->
                  <div class="richtext" v-html="installationManual" />
                  <!-- eslint-enable vue/no-v-html -->
                </CardContent>
              </Card>
            </TabsContent>

            <!-- Changelog tab -->
            <TabsContent value="changelog" class="mt-4">
              <Card v-if="changelog.length > 0">
                <CardContent class="space-y-6 p-6">
                  <div
                    v-for="(entry, i) in changelog"
                    :key="entry.version"
                    class="relative border-l-2 pl-5"
                    :class="i === 0 ? 'border-primary/40' : 'border-border'"
                  >
                    <span
                      class="absolute -left-[7px] top-1 size-3 rounded-full border-2 border-card"
                      :class="i === 0 ? 'bg-primary' : 'bg-border'"
                    />
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge variant="secondary" class="font-mono text-xs font-semibold">{{
                        entry.version
                      }}</Badge>
                      <Badge v-if="i === 0" class="bg-primary/10 text-primary" variant="outline">
                        {{ $t("extensions.latestVersion") }}
                      </Badge>
                      <span class="text-xs text-muted-foreground">{{
                        formatDate(entry.creationDate)
                      }}</span>
                    </div>
                    <!-- eslint-disable vue/no-v-html -->
                    <div class="richtext mt-2" v-html="changelogText(entry)" />
                    <!-- eslint-enable vue/no-v-html -->
                  </div>
                </CardContent>
              </Card>
              <EmptyState
                v-else
                :icon="IconClockRotateLeft"
                :title="$t('extensions.noChangelog')"
                size="sm"
              />
            </TabsContent>
          </Tabs>
        </div>

        <!-- Screenshots sidebar -->
        <aside v-if="screenshots.length > 0" class="lg:sticky lg:top-6 lg:self-start">
          <h2 class="mb-3 text-sm font-semibold">{{ $t("extensions.screenshots") }}</h2>
          <ScreenshotGallery
            :images="screenshots.map((s) => s.url)"
            :alt="$t('extensions.screenshots')"
          />
        </aside>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { api, apiLanguage } from "@/helpers/api";
import { formatDate } from "@/helpers/formatter";
import { useChangelogText } from "@/helpers/changelog";
import { sanitizeHtml } from "@/helpers/sanitize";
import {
  type AccountExtension,
  envHasUpdate,
  envState,
  normalizeWebsite,
  updateCount,
} from "@/composables/useAccountExtensions";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import StatusIcon from "@/components/StatusIcon.vue";
import RatingStars from "@/components/RatingStars.vue";
import EmptyState from "@/components/EmptyState.vue";
import ExtensionIcon from "@/components/ExtensionIcon.vue";
import ScreenshotGallery from "@/components/ScreenshotGallery.vue";
import IconPuzzlePiece from "~icons/fa6-solid/puzzle-piece";
import IconClockRotateLeft from "~icons/fa6-solid/clock-rotate-left";

const route = useRoute();
const changelogText = useChangelogText();

const extension = ref<AccountExtension | null>(null);
const loading = ref(true);

const name = computed(() => String(route.params.name));

api
  .GET("/account/extensions/{name}", {
    params: { path: { name: name.value }, query: { language: apiLanguage() } },
  })
  .then(({ data }) => {
    extension.value = data ?? null;
    loading.value = false;
  });

const environments = computed(() => {
  const list = [...(extension.value?.environments ?? [])];
  list.sort((a, b) => {
    if (a.active !== b.active) return a.active ? -1 : 1;
    return a.environmentName.localeCompare(b.environmentName);
  });
  return list;
});

// Group the environments by shop so an extension that spans multiple shops shows
// a heading per shop (environments can share a name across shops). Groups are
// ordered by shop name; environments within a group keep the active-first order.
const environmentsByShop = computed(() => {
  const groups = new Map<
    number,
    { shopId: number; shopName: string; environments: typeof environments.value }
  >();
  for (const env of environments.value) {
    let group = groups.get(env.shopId);
    if (!group) {
      group = { shopId: env.shopId, shopName: env.shopName, environments: [] };
      groups.set(env.shopId, group);
    }
    group.environments.push(env);
  }
  return [...groups.values()].sort((a, b) => a.shopName.localeCompare(b.shopName));
});

const changelog = computed(() => extension.value?.changelog ?? []);
const outdated = computed(() => (extension.value ? updateCount(extension.value) : 0));

// Localized text fields are already resolved to the requested language by the
// API (with English fallback), so the frontend just reads them directly.
const producerWebsite = computed(() => normalizeWebsite(extension.value?.producerWebsite));
const shortDescription = computed(() => extension.value?.shortDescription ?? null);
// description + installationManual are third-party store HTML rendered via
// v-html, so sanitize them. shortDescription is rendered as plain text ({{ }})
// and needs no sanitization.
const description = computed(() => sanitizeHtml(extension.value?.description));
const installationManual = computed(() => sanitizeHtml(extension.value?.installationManual));

// Sort screenshots so the store's preview image comes first.
const screenshots = computed(() => {
  const list = [...(extension.value?.screenshots ?? [])];
  list.sort((a, b) => (b.preview ? 1 : 0) - (a.preview ? 1 : 0));
  return list;
});
</script>
