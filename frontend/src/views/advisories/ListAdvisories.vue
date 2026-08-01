<template>
  <div class="space-y-6">
    <PageHeader :title="$t('advisories.title')" :description="$t('advisories.subtitle')" />

    <!-- Scope is the primary axis: the catalog is large and mostly irrelevant
         to any one tenant, so "affects my shops" leads rather than sitting in
         the filter row. Counts are always shown so the other tab is honest. -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div
        role="tablist"
        :aria-label="$t('advisories.scope.label')"
        class="inline-flex items-center gap-1 rounded-lg bg-muted p-1"
      >
        <button
          v-for="option in scopeOptions"
          :key="option.value"
          role="tab"
          type="button"
          :aria-selected="scope === option.value"
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
            {{ option.count }}
          </Badge>
        </button>
      </div>
      <p v-if="!loading" class="text-xs text-muted-foreground">
        {{ $t("advisories.totalCount", total, { count: total }) }}
      </p>
    </div>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <div class="relative sm:max-w-xs sm:flex-1">
        <icon-fa6-solid:magnifying-glass
          class="pointer-events-none absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground"
        />
        <Input
          v-model="search"
          :placeholder="$t('advisories.searchPlaceholder')"
          class="pl-9"
          @keyup.enter="applyFilters"
        />
      </div>
      <Select v-model="packageFilter" @update:model-value="applyFilters">
        <SelectTrigger class="sm:w-56">
          <SelectValue :placeholder="$t('advisories.allPackages')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">{{ $t("advisories.allPackages") }}</SelectItem>
          <SelectItem v-for="pkg in packages" :key="pkg" :value="pkg">{{ pkg }}</SelectItem>
        </SelectContent>
      </Select>
      <Select v-model="severityFilter" @update:model-value="applyFilters">
        <SelectTrigger class="sm:w-40">
          <SelectValue :placeholder="$t('advisories.allSeverities')" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">{{ $t("advisories.allSeverities") }}</SelectItem>
          <SelectItem v-for="s in severities" :key="s" :value="s">
            {{ $t(`advisories.severity.${s}`) }}
          </SelectItem>
        </SelectContent>
      </Select>
      <Select v-model="sort" @update:model-value="applyFilters">
        <SelectTrigger class="sm:w-48" :aria-label="$t('advisories.sort.label')">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="s in sortOptions" :key="s" :value="s">
            {{ $t(`advisories.sort.${s}`) }}
          </SelectItem>
        </SelectContent>
      </Select>
      <Button variant="outline" size="sm" @click="applyFilters">
        {{ $t("advisories.search") }}
      </Button>
      <Button v-if="hasActiveFilters" variant="ghost" size="sm" @click="clearFilters">
        {{ $t("advisories.clearFilters") }}
      </Button>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

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
        class="group flex items-start gap-3 rounded-xl border bg-card p-4 transition-all hover:border-primary/30 hover:shadow-sm sm:items-center sm:gap-4"
      >
        <div
          class="flex size-9 shrink-0 items-center justify-center rounded-lg"
          :class="severityTileClass(adv.effectiveSeverity)"
        >
          <icon-fa6-solid:shield-halved class="size-4" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
            <span class="font-medium transition-colors group-hover:text-primary">{{
              adv.title
            }}</span>
            <!-- Blast radius. Icon + text, never colour alone, so the signal
                 survives greyscale and colour-vision differences. -->
            <Badge
              v-if="adv.affectedEnvironmentCount"
              variant="destructive"
              class="gap-1 text-[10px]"
            >
              <icon-fa6-solid:triangle-exclamation class="size-2.5" />
              {{
                $t("advisories.affectedCount", adv.affectedEnvironmentCount, {
                  count: adv.affectedEnvironmentCount,
                })
              }}
            </Badge>
            <!-- An acknowledged advisory is labelled rather than hidden, so it
                 never becomes an invisible blind spot. -->
            <Badge
              v-if="adv.suppressedEnvironmentCount"
              variant="secondary"
              class="gap-1 text-[10px]"
            >
              <icon-fa6-solid:bell-slash class="size-2.5" />
              {{
                $t("advisories.suppressedCount", adv.suppressedEnvironmentCount, {
                  count: adv.suppressedEnvironmentCount,
                })
              }}
            </Badge>
          </div>
          <p v-if="adv.summary" class="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
            {{ adv.summary }}
          </p>
          <div
            class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground"
          >
            <span class="flex flex-wrap gap-1">
              <Badge
                v-for="pkg in adv.packages"
                :key="pkg.packageName"
                variant="secondary"
                class="font-mono text-[10px]"
                >{{ pkg.packageName }}</Badge
              >
            </span>
            <span v-if="adv.cve" class="font-mono">{{ adv.cve }}</span>
            <span v-if="adv.ghsaId" class="font-mono">{{ adv.ghsaId }}</span>
            <span v-if="adv.reportedAt" class="inline-flex items-center gap-1">
              <icon-fa6-regular:calendar class="size-3" />
              {{ formatDate(adv.reportedAt) }}
            </span>
          </div>
          <div v-if="adv.tags?.length" class="mt-1.5 flex flex-wrap gap-1">
            <Badge v-for="tag in adv.tags" :key="tag" variant="outline" class="text-[10px]">{{
              tag
            }}</Badge>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-3">
          <SeverityBadge :severity="adv.effectiveSeverity" class="text-[10px]" />
          <icon-fa6-solid:chevron-right
            class="hidden size-3 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 sm:block"
          />
        </div>
      </RouterLink>
    </div>

    <div v-if="totalPages > 1" class="flex items-center justify-center gap-4">
      <Button size="sm" variant="outline" :disabled="page <= 1" @click="goToPage(page - 1)">
        {{ $t("common.previous") }}
      </Button>
      <span class="text-sm tabular-nums text-muted-foreground">{{ page }} / {{ totalPages }}</span>
      <Button
        size="sm"
        variant="outline"
        :disabled="page >= totalPages"
        @click="goToPage(page + 1)"
      >
        {{ $t("common.next") }}
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import EmptyState from "@/components/EmptyState.vue";
import PageHeader from "@/components/PageHeader.vue";
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  type Advisory,
  type ListAdvisoriesData,
  listAdvisories,
  listAdvisoryPackages,
} from "@/api/generated";
import { severityTileClass } from "@/helpers/advisory";
import { formatDate } from "@/helpers/formatter";
import { computed, onMounted, ref } from "vue";
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

const advisories = ref<Advisory[]>([]);
const packages = ref<string[]>([]);
const total = ref(0);
const scopeCounts = ref({ all: 0, affected: 0, suppressed: 0 });
const loading = ref(true);
const error = ref("");
const search = ref("");
const packageFilter = ref("all");
const severityFilter = ref<Severity | "all">("all");
const scope = ref<Scope>("affected");
const sort = ref<SortOption>("reported");
const page = ref(1);
const perPage = 25;

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
    search.value.trim() !== "" || packageFilter.value !== "all" || severityFilter.value !== "all",
);

function setScope(next: Scope) {
  if (scope.value === next) return;
  scope.value = next;
  page.value = 1;
  reload();
}

function applyFilters() {
  page.value = 1;
  reload();
}

function clearFilters() {
  search.value = "";
  packageFilter.value = "all";
  severityFilter.value = "all";
  applyFilters();
}

function goToPage(next: number) {
  page.value = next;
  reload();
}

// Mirror state into the URL so a filtered view can be bookmarked, shared, and
// survives a reload or back-navigation.
function syncUrl() {
  // Scope is always written, even at its default: a shared link should keep
  // showing what the sender saw if the default ever changes.
  const query: Record<string, string> = { scope: scope.value };
  if (sort.value !== "reported") query.sort = sort.value;
  if (search.value.trim()) query.q = search.value.trim();
  if (packageFilter.value !== "all") query.package = packageFilter.value;
  if (severityFilter.value !== "all") query.severity = severityFilter.value;
  if (page.value > 1) query.page = String(page.value);

  router.replace({ query });
}

function readUrl() {
  const q = route.query;
  scope.value = q.scope === "all" || q.scope === "suppressed" ? (q.scope as Scope) : "affected";
  sort.value =
    typeof q.sort === "string" && (sortOptions as readonly string[]).includes(q.sort)
      ? (q.sort as SortOption)
      : "reported";
  search.value = typeof q.q === "string" ? q.q : "";
  packageFilter.value = typeof q.package === "string" ? q.package : "all";
  severityFilter.value =
    typeof q.severity === "string" && (severities as readonly string[]).includes(q.severity)
      ? (q.severity as Severity)
      : "all";
  const parsedPage = Number.parseInt(String(q.page ?? "1"), 10);
  page.value = Number.isFinite(parsedPage) && parsedPage > 0 ? parsedPage : 1;
}

async function reload() {
  loading.value = true;
  error.value = "";
  syncUrl();
  try {
    const query: ListAdvisoriesData["query"] = {
      limit: perPage,
      offset: (page.value - 1) * perPage,
      scope: scope.value,
      sort: sort.value,
    };
    if (search.value.trim()) query.q = search.value.trim();
    if (packageFilter.value !== "all") query.package = packageFilter.value;
    if (severityFilter.value !== "all") query.severity = severityFilter.value;

    const { data, error: apiError } = await listAdvisories({ query });
    if (apiError || !data) {
      error.value = t("advisories.loadFailed");
      return;
    }
    advisories.value = data.advisories;
    total.value = data.total;
    if (data.scopeCounts) {
      scopeCounts.value = data.scopeCounts;
    }
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  readUrl();
  const { data } = await listAdvisoryPackages();
  packages.value = data ?? [];
  await reload();
});
</script>
