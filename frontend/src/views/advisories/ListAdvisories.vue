<template>
  <div class="space-y-6">
    <PageHeader :title="$t('advisories.title')" :description="$t('advisories.subtitle')" />

    <!-- Scope is the primary axis: the catalog is large and mostly irrelevant
         to any one tenant, so "affects my shops" leads rather than sitting in
         the filter row. Counts are always shown so the other tab is honest. -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div
        v-if="session"
        role="group"
        :aria-label="$t('advisories.scope.label')"
        class="inline-flex flex-wrap items-center gap-1 rounded-lg bg-muted p-1"
      >
        <button
          v-for="option in scopeOptions"
          :key="option.value"
          type="button"
          :aria-pressed="scope === option.value"
          class="inline-flex cursor-pointer items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :class="
            scope === option.value
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          "
          @click="setScope(option.value)"
        >
          {{ option.label }}
          <Badge
            variant="secondary"
            class="min-w-5 justify-center px-1.5 text-[10px] tabular-nums"
            :class="
              option.value === 'affected' && option.count > 0
                ? 'bg-destructive/15 text-destructive'
                : ''
            "
          >
            {{ countsLoaded ? option.count : "—" }}
          </Badge>
        </button>
      </div>
    </div>

    <form class="space-y-4 rounded-xl border bg-card p-4" @submit.prevent="applyFilters">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-12">
        <div class="space-y-1.5 sm:col-span-2 xl:col-span-4">
          <Label for="advisory-search">{{ $t("advisories.search") }}</Label>
          <div class="relative">
            <icon-fa6-solid:magnifying-glass
              class="pointer-events-none absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="advisory-search"
              v-model="search"
              type="search"
              :placeholder="$t('advisories.searchPlaceholder')"
              class="pl-9"
            />
          </div>
        </div>
        <div class="min-w-0 space-y-1.5 xl:col-span-3">
          <Label for="advisory-package">{{ $t("advisories.affected.package") }}</Label>
          <FilterSelect
            id="advisory-package"
            v-model="packageFilter"
            searchable
            :label="$t('advisories.affected.package')"
            :options="packageOptions"
            :search-placeholder="$t('advisories.searchPackages')"
            @update:model-value="applyFilters"
          />
        </div>
        <div class="min-w-0 space-y-1.5 xl:col-span-2">
          <Label for="advisory-severity">{{ $t("advisories.kpi.severity") }}</Label>
          <FilterSelect
            id="advisory-severity"
            v-model="severityFilter"
            :label="$t('advisories.kpi.severity')"
            :options="severityOptions"
            @update:model-value="applyFilters"
          />
        </div>
        <div class="min-w-0 space-y-1.5 sm:col-span-2 xl:col-span-3">
          <Label for="advisory-sort">{{ $t("advisories.sort.label") }}</Label>
          <FilterSelect
            id="advisory-sort"
            v-model="sort"
            :label="$t('advisories.sort.label')"
            :options="sortingOptions"
            @update:model-value="applyFilters"
          />
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <Button type="submit" size="sm">
          {{ $t("advisories.search") }}
        </Button>
        <Button
          v-if="hasActiveFilters || search.trim()"
          type="button"
          variant="ghost"
          size="sm"
          @click="clearFilters"
        >
          {{ $t("advisories.clearFilters") }}
        </Button>
      </div>
    </form>

    <p role="status" class="text-sm text-muted-foreground">
      <template v-if="loading">{{ $t("advisories.loading") }}</template>
      <template v-else-if="!error">{{
        $t("advisories.totalCount", total, { count: total })
      }}</template>
    </p>

    <div v-if="loading" class="space-y-2" aria-hidden="true">
      <div v-for="i in 6" :key="i" class="flex items-center gap-4 rounded-xl border bg-card p-4">
        <Skeleton class="size-9 shrink-0 rounded-lg" />
        <div class="flex-1 space-y-2">
          <Skeleton class="h-4 w-2/3" />
          <Skeleton class="h-3 w-1/3" />
        </div>
        <Skeleton class="hidden h-5 w-16 rounded-full sm:block" />
      </div>
    </div>

    <Alert v-else-if="error" variant="destructive">
      <AlertDescription class="flex flex-wrap items-center justify-between gap-3">
        <span>{{ error }}</span>
        <Button variant="outline" size="sm" @click="reload">{{ $t("common.retry") }}</Button>
      </AlertDescription>
    </Alert>

    <!-- Distinct zero-results cases: nothing suppressed, nothing affects you
         (good news), filters excluded everything, or the catalog is empty. -->
    <EmptyState
      v-else-if="advisories.length === 0 && scope === 'suppressed' && !hasActiveFilters"
      :icon="IconShield"
      :title="$t('advisories.emptySuppressed')"
      :description="$t('advisories.emptySuppressedHint')"
    >
      <Button variant="outline" size="sm" @click="setScope('affected')">
        {{ $t("advisories.scope.affected") }}
      </Button>
    </EmptyState>

    <EmptyState
      v-else-if="advisories.length === 0 && scope === 'affected' && !hasActiveFilters"
      :icon="IconShieldCheck"
      :title="$t('advisories.emptyAffected')"
      :description="
        scopeCounts.all > 0
          ? $t('advisories.emptyAffectedHint', { count: scopeCounts.all })
          : $t('advisories.emptyHint')
      "
    >
      <Button v-if="scopeCounts.all > 0" variant="outline" size="sm" @click="setScope('all')">
        {{ $t("advisories.viewAll") }}
      </Button>
    </EmptyState>

    <EmptyState
      v-else-if="advisories.length === 0 && hasActiveFilters"
      :icon="IconShield"
      :title="$t('advisories.emptyFiltered')"
      :description="$t('advisories.emptyFilteredHint')"
    >
      <Button variant="outline" size="sm" @click="clearFilters">
        {{ $t("advisories.clearFilters") }}
      </Button>
    </EmptyState>

    <EmptyState
      v-else-if="advisories.length === 0"
      :icon="IconShield"
      :title="$t('advisories.empty')"
      :description="$t('advisories.emptyHint')"
    />

    <div v-else class="space-y-2">
      <RouterLink
        v-for="adv in advisories"
        :key="adv.advisoryId"
        :to="{ name: 'account.advisories.detail', params: { id: adv.advisoryId } }"
        class="group flex items-start gap-3 rounded-xl border bg-card p-4 transition-all hover:border-primary/30 hover:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:gap-4 sm:p-5"
      >
        <div
          class="flex size-9 shrink-0 items-center justify-center rounded-lg"
          :class="severityTileClass(adv.effectiveSeverity)"
        >
          <icon-fa6-solid:shield-halved class="size-4" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="mb-2 flex flex-wrap items-center gap-2">
            <SeverityBadge :severity="adv.effectiveSeverity" class="text-xs" />
            <ShopmonAdvisoryBadge :advisory="adv" class="text-xs" />
            <span v-if="adv.cvssScore != null" class="text-xs tabular-nums text-muted-foreground">
              {{ $t("advisories.cvss") }} {{ adv.cvssScore.toFixed(1) }}
            </span>
            <span
              v-if="adv.reportedAt"
              class="inline-flex items-center gap-1 text-xs text-muted-foreground sm:ml-auto"
            >
              <icon-fa6-regular:calendar class="size-3" />
              <time :datetime="adv.reportedAt">{{ formatDate(adv.reportedAt) }}</time>
            </span>
          </div>
          <h2 class="font-medium break-words transition-colors group-hover:text-primary">
            {{ adv.title }}
          </h2>
          <p v-if="adv.summary" class="mt-1 line-clamp-2 text-sm break-words text-muted-foreground">
            {{ adv.summary }}
          </p>
          <div
            v-if="adv.affectedEnvironmentCount || adv.suppressedEnvironmentCount"
            class="mt-2 flex flex-wrap items-center gap-2"
          >
            <!-- Blast radius. Icon + text, never colour alone, so the signal
                 survives greyscale and colour-vision differences. -->
            <Badge v-if="adv.affectedEnvironmentCount" variant="destructive" class="gap-1 text-xs">
              <icon-fa6-solid:triangle-exclamation class="size-2.5" />
              {{
                $t("advisories.affectedCount", adv.affectedEnvironmentCount, {
                  count: adv.affectedEnvironmentCount,
                })
              }}
            </Badge>
            <!-- An acknowledged advisory is labelled rather than hidden, so it
                 never becomes an invisible blind spot. -->
            <Badge v-if="adv.suppressedEnvironmentCount" variant="secondary" class="gap-1 text-xs">
              <icon-fa6-solid:bell-slash class="size-2.5" />
              {{
                $t("advisories.suppressedCount", adv.suppressedEnvironmentCount, {
                  count: adv.suppressedEnvironmentCount,
                })
              }}
            </Badge>
          </div>
          <div
            class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-2 text-xs text-muted-foreground"
          >
            <span class="flex flex-wrap gap-1">
              <Badge
                v-for="pkg in adv.packages"
                :key="pkg.packageName"
                variant="secondary"
                class="max-w-full font-mono text-xs whitespace-normal break-all"
                >{{ pkg.packageName }}</Badge
              >
            </span>
            <span v-if="adv.cve" class="font-mono">{{ adv.cve }}</span>
            <span v-if="adv.ghsaId" class="font-mono">{{ adv.ghsaId }}</span>
          </div>
          <div v-if="adv.tags?.length" class="mt-1.5 flex flex-wrap gap-1">
            <Badge v-for="tag in adv.tags" :key="tag" variant="outline" class="text-[10px]">{{
              tag
            }}</Badge>
          </div>
        </div>
        <div class="hidden shrink-0 self-center sm:block">
          <icon-fa6-solid:chevron-right
            class="hidden size-3 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 sm:block"
          />
        </div>
      </RouterLink>
    </div>

    <nav
      v-if="!loading && !error && total > 0"
      :aria-label="$t('advisories.pagination')"
      class="flex flex-wrap items-center justify-between gap-3 border-t pt-4"
    >
      <p class="text-sm tabular-nums text-muted-foreground">
        {{
          $t("advisories.resultRange", {
            start: (page - 1) * perPage + 1,
            end: Math.min(page * perPage, total),
            total,
          })
        }}
      </p>
      <div v-if="totalPages > 1" class="flex items-center gap-3">
        <Button size="sm" variant="outline" :disabled="page <= 1" @click="goToPage(page - 1)">
          {{ $t("common.previous") }}
        </Button>
        <span class="text-sm tabular-nums text-muted-foreground">{{
          $t("common.pageOf", { current: page, total: totalPages })
        }}</span>
        <Button
          size="sm"
          variant="outline"
          :disabled="page >= totalPages"
          @click="goToPage(page + 1)"
        >
          {{ $t("common.next") }}
        </Button>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import EmptyState from "@/components/EmptyState.vue";
import PageHeader from "@/components/PageHeader.vue";
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
import ShopmonAdvisoryBadge from "@/components/advisories/ShopmonAdvisoryBadge.vue";
import { useSession } from "@/composables/useSession";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import FilterSelect from "@/components/FilterSelect.vue";
import { Skeleton } from "@/components/ui/skeleton";
import {
  type Advisory,
  type ListAdvisoriesData,
  listAdvisories,
  listAdvisoryPackages,
} from "@/api/generated";
import { severityTileClass } from "@/helpers/advisory";
import { formatDate } from "@/helpers/formatter";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import IconShield from "~icons/fa6-solid/shield-halved";
import IconShieldCheck from "~icons/fa6-solid/shield-heart";

type Scope = "affected" | "suppressed" | "all";

const severities = ["critical", "high", "medium", "low", "none"] as const;
const sortOptions = ["reported", "severity", "affected", "cvss"] as const;
type Severity = (typeof severities)[number];
type SortOption = (typeof sortOptions)[number];

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { session } = useSession();

const advisories = ref<Advisory[]>([]);
const packages = ref<string[]>([]);
const total = ref(0);
const scopeCounts = ref({ all: 0, affected: 0, suppressed: 0 });
const countsLoaded = ref(false);
const loading = ref(true);
const error = ref("");
const search = ref("");
const appliedSearch = ref("");
const packageFilter = ref("all");
const severityFilter = ref<Severity | "all">("all");
const scope = ref<Scope>(session.value ? "affected" : "all");
const sort = ref<SortOption>("reported");
const page = ref(1);
const perPage = 25;

const packageOptions = computed(() => [
  { value: "all", label: t("advisories.allPackages") },
  ...packages.value.map((pkg) => ({ value: pkg, label: pkg })),
]);
const severityOptions = computed(() => [
  { value: "all" as const, label: t("advisories.allSeverities") },
  ...severities.map((value) => ({ value, label: t(`advisories.severity.${value}`) })),
]);
const availableSortOptions = computed(() =>
  session.value ? sortOptions : sortOptions.filter((s) => s !== "affected"),
);
const sortingOptions = computed(() =>
  availableSortOptions.value.map((value) => ({ value, label: t(`advisories.sort.${value}`) })),
);

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / perPage)));

// The suppressed tab only appears once something is in it (or while it is the
// active scope), so it stays out of the way for the common case of a user who
// has never suppressed anything.
const scopeOptions = computed(() => {
  const options = [
    {
      value: "affected" as Scope,
      label: t("advisories.scope.affected"),
      count: scopeCounts.value.affected,
    },
  ];
  if (scopeCounts.value.suppressed > 0 || scope.value === "suppressed") {
    options.push({
      value: "suppressed" as Scope,
      label: t("advisories.scope.suppressed"),
      count: scopeCounts.value.suppressed,
    });
  }
  options.push({
    value: "all" as Scope,
    label: t("advisories.scope.all"),
    count: scopeCounts.value.all,
  });
  return options;
});

// "Active filters" excludes scope and sort: those always have a value, so
// counting them would keep the clear-filters button permanently visible.
const hasActiveFilters = computed(
  () =>
    appliedSearch.value !== "" || packageFilter.value !== "all" || severityFilter.value !== "all",
);

function setScope(next: Scope) {
  if (scope.value === next) return;
  scope.value = next;
  page.value = 1;
  syncUrl();
}

function applyFilters() {
  page.value = 1;
  appliedSearch.value = search.value.trim();
  syncUrl();
}

function clearFilters() {
  search.value = "";
  packageFilter.value = "all";
  severityFilter.value = "all";
  applyFilters();
}

function goToPage(next: number) {
  page.value = next;
  syncUrl();
}

// Mirror state into the URL so a filtered view can be bookmarked, shared, and
// survives a reload or back-navigation.
function syncUrl() {
  // Preserve scope for signed-in users. Anonymous visitors always see the catalog.
  const query = { ...route.query };
  for (const key of ["scope", "sort", "q", "package", "severity", "page"]) delete query[key];
  if (session.value) query.scope = scope.value;
  if (sort.value !== "reported") query.sort = sort.value;
  if (appliedSearch.value) query.q = appliedSearch.value;
  if (packageFilter.value !== "all") query.package = packageFilter.value;
  if (severityFilter.value !== "all") query.severity = severityFilter.value;
  if (page.value > 1) query.page = String(page.value);

  const target = { query, hash: route.hash };
  if (router.resolve(target).fullPath === route.fullPath) {
    void reload();
  } else {
    void router.push(target);
  }
}

function readUrl() {
  const q = route.query;
  scope.value = !session.value
    ? "all"
    : q.scope === "all" || q.scope === "suppressed"
      ? (q.scope as Scope)
      : "affected";
  sort.value =
    typeof q.sort === "string" && (availableSortOptions.value as readonly string[]).includes(q.sort)
      ? (q.sort as SortOption)
      : "reported";
  search.value = typeof q.q === "string" ? q.q : "";
  appliedSearch.value = search.value.trim();
  packageFilter.value = typeof q.package === "string" ? q.package : "all";
  severityFilter.value =
    typeof q.severity === "string" && (severities as readonly string[]).includes(q.severity)
      ? (q.severity as Severity)
      : "all";
  const parsedPage = typeof q.page === "string" ? Number(q.page) : 1;
  page.value =
    Number.isSafeInteger(parsedPage) && parsedPage > 0 && (parsedPage - 1) * perPage <= 2147483647
      ? parsedPage
      : 1;
}

// Only the latest request may update the results or finish the loading state.
let requestId = 0;
async function reload() {
  const currentRequest = ++requestId;
  loading.value = true;
  error.value = "";
  try {
    const query: ListAdvisoriesData["query"] = {
      limit: perPage,
      offset: (page.value - 1) * perPage,
      scope: scope.value,
      sort: sort.value,
    };
    if (appliedSearch.value) query.q = appliedSearch.value;
    if (packageFilter.value !== "all") query.package = packageFilter.value;
    if (severityFilter.value !== "all") query.severity = severityFilter.value;

    const { data, error: apiError } = await listAdvisories({ query });
    if (currentRequest !== requestId) return;
    if (apiError || !data) {
      error.value = t("advisories.loadFailed");
      return;
    }
    advisories.value = data.advisories;
    total.value = data.total;
    if (data.scopeCounts) {
      scopeCounts.value = data.scopeCounts;
      countsLoaded.value = true;
    }
    if (page.value > totalPages.value) {
      page.value = totalPages.value;
      await router.replace({
        query: { ...route.query, page: page.value > 1 ? String(page.value) : undefined },
        hash: route.hash,
      });
    }
  } catch {
    if (currentRequest === requestId) error.value = t("advisories.loadFailed");
  } finally {
    if (currentRequest === requestId) loading.value = false;
  }
}

const listRouteName = route.name;
watch(
  () => route.query,
  () => {
    if (route.name !== listRouteName) return;
    readUrl();
    void reload();
  },
  { immediate: true },
);

onMounted(async () => {
  try {
    const { data } = await listAdvisoryPackages();
    packages.value = data ?? [];
  } catch {
    // Package suggestions are optional; keep the advisory list usable.
    packages.value = [];
  }
});

onBeforeUnmount(() => {
  requestId++;
});
</script>
