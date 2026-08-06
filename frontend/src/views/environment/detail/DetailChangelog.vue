<template>
  <div v-if="environment" class="space-y-6">
    <!-- Empty state — only once a page has actually been loaded, so the initial
         fetch does not flash "no changes recorded". -->
    <div
      v-if="hasLoaded && !entries.length"
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:clock-rotate-left class="size-10 text-muted-foreground" />
      <h3 class="text-lg font-semibold">{{ $t("shopDetail.noChangesRecorded") }}</h3>
      <p class="text-sm text-muted-foreground">{{ $t("shopDetail.changesWillAppear") }}</p>
    </div>

    <!-- Timeline -->
    <div v-if="entries.length" class="space-y-3" :class="{ 'opacity-60': isLoading }">
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="group rounded-xl border bg-card transition-colors hover:border-primary/20"
      >
        <!-- Header row — always visible -->
        <button
          class="flex w-full cursor-pointer items-start gap-4 px-4 py-3 text-left sm:items-center"
          @click="toggle(entry.id)"
        >
          <!-- Date -->
          <div class="shrink-0 pt-0.5 text-sm tabular-nums text-muted-foreground sm:pt-0">
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
              class="rounded-lg bg-muted/50 px-3 py-2 text-sm"
            >
              <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
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
                <div class="min-w-0 flex-1 basis-full sm:basis-0">
                  <span class="font-medium wrap-break-word">{{ ext.label }}</span>
                  <span class="ml-1 text-muted-foreground break-all">({{ ext.name }})</span>
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

              <!-- Extension changelog text (only for updates) -->
              <div
                v-if="ext.state === 'updated' && ext.changelog && ext.changelog.length > 0"
                class="mt-3 space-y-3 border-t border-border/60 pt-3"
              >
                <div v-for="entryLog in ext.changelog" :key="entryLog.version">
                  <div class="mb-1 flex items-center gap-2 text-xs font-semibold">
                    <Badge variant="secondary" class="font-mono text-[10px]">{{
                      entryLog.version
                    }}</Badge>
                    <span class="font-normal text-muted-foreground tabular-nums">{{
                      formatDate(entryLog.creationDate)
                    }}</span>
                  </div>
                  <!-- eslint-disable vue/no-v-html -->
                  <div class="richtext text-xs" v-html="changelogText(entryLog)" />
                  <!-- eslint-enable vue/no-v-html -->
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-center gap-4">
      <Button
        size="sm"
        variant="outline"
        :disabled="currentPage === 1 || isLoading"
        @click="loadPage(currentPage - 1)"
      >
        {{ $t("common.previous") }}
      </Button>
      <span class="text-sm text-muted-foreground tabular-nums">{{
        $t("common.pageOf", { current: currentPage, total: totalPages })
      }}</span>
      <Button
        size="sm"
        variant="outline"
        :disabled="currentPage === totalPages || isLoading"
        @click="loadPage(currentPage + 1)"
      >
        {{ $t("common.next") }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { formatDate } from "@/helpers/formatter";
import { useChangelogText } from "@/helpers/changelog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getEnvironmentChangelogs, type AccountChangelog } from "@/api/generated";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useAlert } from "@/composables/useAlert";

const { environment } = useEnvironmentDetail();
const changelogText = useChangelogText();
const { error } = useAlert();

const PAGE_SIZE = 10;

const expanded = reactive(new Set<number>());
const entries = ref<AccountChangelog[]>([]);
const total = ref(0);
const currentPage = ref(1);
const isLoading = ref(false);
const hasLoaded = ref(false);

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)));

async function loadPage(page: number) {
  const id = environment.value?.id;
  if (!id) return;

  isLoading.value = true;
  try {
    const { data } = await getEnvironmentChangelogs({
      path: { environmentId: id },
      query: { limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE },
    });
    entries.value = data?.entries ?? [];
    total.value = data?.total ?? 0;
    currentPage.value = page;
    // Expansion state refers to entries of the previous page.
    expanded.clear();
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  } finally {
    isLoading.value = false;
    hasLoaded.value = true;
  }
}

watch(
  () => environment.value?.id,
  (id) => {
    if (id) void loadPage(1);
  },
  { immediate: true },
);

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
