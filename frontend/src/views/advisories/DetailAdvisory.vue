<template>
  <div class="space-y-6">
    <Button as-child variant="ghost" size="sm" class="-ml-2">
      <RouterLink :to="{ name: 'account.advisories.list' }">
        <icon-fa6-solid:arrow-left class="mr-1.5 size-3.5" />
        {{ $t("advisories.backToList") }}
      </RouterLink>
    </Button>

    <div v-if="loading" class="space-y-6" aria-hidden="true">
      <div class="space-y-3">
        <Skeleton class="h-8 w-2/3" />
        <Skeleton class="h-4 w-1/2" />
        <Skeleton class="h-4 w-1/3" />
      </div>
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Skeleton v-for="i in 4" :key="i" class="h-[68px] rounded-xl" />
      </div>
      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
        <Skeleton class="h-72 rounded-xl" />
        <Skeleton class="h-72 rounded-xl" />
      </div>
    </div>

    <Alert v-else-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <template v-else-if="advisory">
      <div>
        <div class="flex flex-wrap items-center gap-3">
          <h1 class="text-2xl font-bold tracking-tight">{{ advisory.title }}</h1>
          <SeverityBadge :severity="advisory.effectiveSeverity" />
        </div>
        <p v-if="advisory.summary" class="mt-2 max-w-3xl text-sm text-muted-foreground">
          {{ advisory.summary }}
        </p>
        <div
          class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1.5 text-sm text-muted-foreground"
        >
          <span class="flex flex-wrap gap-1.5">
            <Badge
              v-for="pkg in advisory.packages"
              :key="pkg.packageName"
              variant="secondary"
              class="font-mono text-xs"
              >{{ pkg.packageName }}</Badge
            >
          </span>
          <span class="font-mono text-xs">{{ advisory.advisoryId }}</span>
          <span v-if="advisory.reportedAt" class="inline-flex items-center gap-1.5 text-xs">
            <icon-fa6-regular:calendar class="size-3" />
            {{ formatDate(advisory.reportedAt) }}
          </span>
        </div>
      </div>

      <!-- Key facts at a glance -->
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          :icon="IconShield"
          :value="severityLabel"
          :label="$t('advisories.kpi.severity')"
          :color="severityStatColor(advisory.effectiveSeverity)"
        />
        <StatCard
          :icon="IconGauge"
          :value="advisory.cvssScore != null ? advisory.cvssScore.toFixed(1) : '—'"
          :label="$t('advisories.kpi.cvss')"
          :color="cvssStatColor(advisory.cvssScore)"
        />
        <!-- "—" rather than 0 when the check failed: a confident zero here
             would be a false all-clear. Counts distinct environments, which is
             what the API returns, so the label says environments. -->
        <StatCard
          :icon="IconStore"
          :value="affectedLoading ? '…' : affectedError ? '—' : affectedTotal"
          :label="$t('advisories.kpi.affectedEnvironments')"
          :color="affectedError ? 'muted' : affectedTotal > 0 ? 'destructive' : 'success'"
        />
        <StatCard
          :icon="IconCalendar"
          :value="advisory.reportedAt ? formatDate(advisory.reportedAt) : '—'"
          :label="$t('advisories.kpi.reported')"
          color="muted"
        />
      </div>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
        <!-- Main column -->
        <div class="min-w-0 space-y-6">
          <Card v-if="descriptionHtml">
            <CardHeader>
              <CardTitle class="text-base">{{ $t("advisories.fullDescription") }}</CardTitle>
            </CardHeader>
            <CardContent>
              <!-- eslint-disable vue/no-v-html -->
              <div class="richtext text-sm" v-html="descriptionHtml" />
              <!-- eslint-enable vue/no-v-html -->
            </CardContent>
          </Card>
          <Alert v-else-if="!advisory.description">
            <AlertDescription>{{ $t("advisories.descriptionPending") }}</AlertDescription>
          </Alert>

          <Card>
            <CardHeader>
              <CardTitle class="text-base">{{ $t("advisories.affected.title") }}</CardTitle>
              <CardDescription>
                {{ $t("advisories.affected.description") }}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div v-if="affectedLoading" class="space-y-2" aria-hidden="true">
                <Skeleton v-for="i in 2" :key="i" class="h-10 w-full" />
              </div>

              <!-- A failed check must never render as "nothing affected". -->
              <Alert v-else-if="affectedError" variant="destructive">
                <AlertDescription class="flex flex-wrap items-center gap-3">
                  {{ affectedError }}
                  <Button variant="outline" size="sm" @click="reloadAffected">
                    {{ $t("common.retry") }}
                  </Button>
                </AlertDescription>
              </Alert>

              <p v-else-if="!affected.length" class="py-2 text-sm text-muted-foreground">
                {{ $t("advisories.affected.none") }}
              </p>

              <div v-else class="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{{ $t("advisories.affected.shop") }}</TableHead>
                      <TableHead>{{ $t("advisories.affected.package") }}</TableHead>
                      <TableHead>{{ $t("advisories.affected.installed") }}</TableHead>
                      <TableHead>{{ $t("advisories.affected.range") }}</TableHead>
                      <TableHead class="text-right" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow
                      v-for="row in affected"
                      :key="`${row.environmentId}-${row.packageName}`"
                      :class="row.suppressed ? 'opacity-60' : ''"
                    >
                      <TableCell>
                        <RouterLink
                          :to="{
                            name: 'account.environments.detail',
                            params: { environmentId: row.environmentId },
                          }"
                          class="font-medium text-primary hover:underline"
                        >
                          {{ row.shopName }}
                        </RouterLink>
                        <div class="text-xs text-muted-foreground">
                          {{ row.environmentName }} · {{ row.organizationName }}
                        </div>
                      </TableCell>
                      <TableCell class="font-mono text-xs">{{ row.packageName }}</TableCell>
                      <TableCell>
                        <Badge
                          :variant="row.suppressed ? 'secondary' : 'destructive'"
                          class="font-mono text-[10px]"
                        >
                          {{ row.installedVersion }}
                        </Badge>
                      </TableCell>
                      <TableCell class="font-mono text-xs text-muted-foreground">
                        {{ row.affectedVersions }}
                      </TableCell>
                      <TableCell class="text-right">
                        <Badge v-if="row.suppressed" variant="secondary" class="gap-1 text-[10px]">
                          <icon-fa6-solid:bell-slash class="size-2.5" />
                          {{ $t("advisories.suppressedBadge") }}
                        </Badge>
                        <Button
                          v-else
                          variant="ghost"
                          size="sm"
                          class="text-xs"
                          @click="openSuppress(row)"
                        >
                          {{ $t("advisories.suppress") }}
                        </Button>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </div>

              <p v-if="globalTotal !== null" class="mt-3 text-xs text-muted-foreground">
                {{ $t("advisories.affected.globalTotal", { count: globalTotal }) }}
              </p>
            </CardContent>
          </Card>

          <Dialog v-model:open="suppressOpen">
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{{ $t("advisories.suppressTitle") }}</DialogTitle>
                <DialogDescription>
                  {{ $t("advisories.suppressDescription") }}
                </DialogDescription>
              </DialogHeader>

              <div class="space-y-4">
                <div>
                  <Label for="suppress-scope" class="text-xs text-muted-foreground">
                    {{ $t("advisories.suppressScope") }}
                  </Label>
                  <!-- The action starts from one environment row, so that is
                       the default. Shop-wide is offered explicitly because it
                       is the wider decision and requires an elevated role. -->
                  <Select id="suppress-scope" v-model="suppressScope">
                    <SelectTrigger class="mt-1">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="environment">
                        {{
                          $t("advisories.suppressScopeEnvironment", {
                            name: suppressTarget?.environmentName ?? "",
                          })
                        }}
                      </SelectItem>
                      <SelectItem value="shop">
                        {{
                          $t("advisories.suppressScopeShopOption", {
                            name: suppressTarget?.shopName ?? "",
                          })
                        }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <p class="mt-1 text-xs text-muted-foreground">
                    {{
                      suppressScope === "shop"
                        ? $t("advisories.suppressScopeShop")
                        : $t("advisories.suppressScopeEnvironmentHint")
                    }}
                  </p>
                </div>

                <div>
                  <Label for="suppress-reason">{{ $t("advisories.suppressReason") }}</Label>
                  <Textarea
                    id="suppress-reason"
                    v-model="suppressReason"
                    :placeholder="$t('advisories.suppressReasonPlaceholder')"
                    rows="3"
                    class="mt-1"
                  />
                </div>

                <div>
                  <Label for="suppress-expiry">{{ $t("advisories.suppressExpiry") }}</Label>
                  <Select id="suppress-expiry" v-model="suppressExpiry">
                    <SelectTrigger class="mt-1">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="30">{{ $t("advisories.suppressExpiry30") }}</SelectItem>
                      <SelectItem value="60">{{ $t("advisories.suppressExpiry60") }}</SelectItem>
                      <SelectItem value="90">{{ $t("advisories.suppressExpiry90") }}</SelectItem>
                      <SelectItem value="never">
                        {{ $t("advisories.suppressExpiryNever") }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <Alert v-if="suppressError" variant="destructive">
                  <AlertDescription>{{ suppressError }}</AlertDescription>
                </Alert>
              </div>

              <DialogFooter>
                <Button variant="outline" @click="suppressOpen = false">
                  {{ $t("common.cancel") }}
                </Button>
                <Button
                  :disabled="!suppressReason.trim() || suppressSaving"
                  @click="submitSuppress"
                >
                  {{ $t("advisories.suppressSubmit") }}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>

        <!-- Sidebar -->
        <div class="space-y-6">
          <Card v-if="hasEnrichment">
            <CardHeader>
              <CardTitle class="text-base">{{ $t("advisories.shopmonContext") }}</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4 text-sm">
              <div v-if="advisory.shopwareImpactSummary">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.impact") }}
                </div>
                <p class="mt-0.5 whitespace-pre-wrap">{{ advisory.shopwareImpactSummary }}</p>
              </div>
              <div v-if="advisory.remediationSummary">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.remediation") }}
                </div>
                <p class="mt-0.5 whitespace-pre-wrap">{{ advisory.remediationSummary }}</p>
              </div>
              <div v-if="advisory.recommendedUpgrade">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.recommendedUpgrade") }}
                </div>
                <code class="mt-0.5 block rounded bg-muted px-2 py-1 font-mono text-xs">{{
                  advisory.recommendedUpgrade
                }}</code>
              </div>
              <div v-if="advisory.remediationUrl">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.remediationUrl") }}
                </div>
                <a
                  v-if="isSafeHttpUrl(advisory.remediationUrl)"
                  :href="advisory.remediationUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="mt-0.5 inline-flex items-center gap-1 break-all text-primary hover:underline"
                >
                  {{ advisory.remediationUrl }}
                  <icon-fa6-solid:arrow-up-right-from-square class="size-3 shrink-0" />
                </a>
                <!-- Non-http(s) values are shown but never made clickable. -->
                <span v-else class="mt-0.5 block break-all text-xs text-muted-foreground">
                  {{ advisory.remediationUrl }}
                </span>
              </div>
              <div v-if="advisory.affectedComponents?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.components") }}
                </div>
                <div class="mt-1 flex flex-wrap gap-1">
                  <Badge
                    v-for="c in advisory.affectedComponents"
                    :key="c"
                    variant="secondary"
                    class="text-[10px]"
                    >{{ c }}</Badge
                  >
                </div>
              </div>
              <div v-if="advisory.notesPublic">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.notes") }}
                </div>
                <p class="mt-0.5 whitespace-pre-wrap">{{ advisory.notesPublic }}</p>
              </div>
              <div v-if="advisory.tags?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.tags") }}
                </div>
                <div class="mt-1 flex flex-wrap gap-1">
                  <Badge
                    v-for="tag in advisory.tags"
                    :key="tag"
                    variant="outline"
                    class="text-[10px]"
                    >{{ tag }}</Badge
                  >
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="text-base">{{ $t("advisories.packagistDetails") }}</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4 text-sm">
              <div v-if="advisory.packages?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.affectedPackages") }}
                </div>
                <ul class="mt-1.5 space-y-1.5">
                  <li
                    v-for="pkg in advisory.packages"
                    :key="pkg.packageName"
                    class="rounded-md bg-muted px-2.5 py-2"
                  >
                    <div class="font-mono text-xs font-medium">{{ pkg.packageName }}</div>
                    <code class="mt-0.5 block font-mono text-[11px] text-muted-foreground">{{
                      pkg.affectedVersions
                    }}</code>
                  </li>
                </ul>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div v-if="advisory.cve">
                  <div class="text-xs font-medium text-muted-foreground">CVE</div>
                  <div class="mt-0.5 font-mono text-xs">{{ advisory.cve }}</div>
                </div>
                <div v-if="advisory.ghsaId">
                  <div class="text-xs font-medium text-muted-foreground">GHSA</div>
                  <div class="mt-0.5 font-mono text-xs">{{ advisory.ghsaId }}</div>
                </div>
                <div v-if="advisory.cvssVector" class="col-span-2">
                  <div class="text-xs font-medium text-muted-foreground">
                    {{ $t("advisories.cvssVector") }}
                  </div>
                  <div class="mt-0.5 font-mono text-xs break-all text-muted-foreground">
                    {{ advisory.cvssVector }}
                  </div>
                </div>
              </div>

              <div v-if="advisory.cwes?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.cwes") }}
                </div>
                <ul class="mt-1 space-y-1">
                  <li v-for="cwe in advisory.cwes" :key="cwe.id" class="text-xs">
                    <span class="font-mono">{{ cwe.id }}</span>
                    <span v-if="cwe.name && cwe.name !== cwe.id" class="text-muted-foreground">
                      — {{ cwe.name }}</span
                    >
                  </li>
                </ul>
              </div>

              <div v-if="advisory.externalReferences?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.references") }}
                </div>
                <ul class="mt-1 space-y-1.5">
                  <li v-for="(ref, i) in advisory.externalReferences" :key="i">
                    <a
                      v-if="isSafeHttpUrl(ref)"
                      :href="ref"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-flex items-start gap-1.5 break-all text-xs text-primary hover:underline"
                    >
                      <icon-fa6-solid:arrow-up-right-from-square class="mt-0.5 size-3 shrink-0" />
                      {{ ref }}
                    </a>
                    <span v-else class="block break-all text-xs text-muted-foreground">{{
                      ref
                    }}</span>
                  </li>
                </ul>
              </div>
              <div v-else-if="advisory.link">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.disclosure") }}
                </div>
                <a
                  v-if="isSafeHttpUrl(advisory.link)"
                  :href="advisory.link"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="mt-0.5 inline-flex items-start gap-1.5 break-all text-xs text-primary hover:underline"
                >
                  <icon-fa6-solid:arrow-up-right-from-square class="mt-0.5 size-3 shrink-0" />
                  {{ advisory.link }}
                </a>
                <span v-else class="mt-0.5 block break-all text-xs text-muted-foreground">{{
                  advisory.link
                }}</span>
              </div>

              <div v-if="advisory.sources?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.sources") }}
                </div>
                <div class="mt-1 flex flex-wrap gap-1">
                  <Badge
                    v-for="(src, i) in advisory.sources"
                    :key="i"
                    variant="outline"
                    class="font-mono text-[10px]"
                    >{{ src.name }}: {{ src.remoteId }}</Badge
                  >
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import StatCard from "@/components/StatCard.vue";
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  type Advisory,
  type AdvisoryAffectedEnvironment,
  createAdvisorySuppression,
  getAdvisory,
  listAdvisoryAffectedEnvironments,
} from "@/api/generated";
import { cvssStatColor, isSafeHttpUrl, severityStatColor } from "@/helpers/advisory";
import { formatDate } from "@/helpers/formatter";
import { sanitizeHtml } from "@/helpers/sanitize";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import IconCalendar from "~icons/fa6-solid/calendar-days";
import IconGauge from "~icons/fa6-solid/gauge-high";
import IconShield from "~icons/fa6-solid/shield-halved";
import IconStore from "~icons/fa6-solid/store";

type AffectedEnvironment = AdvisoryAffectedEnvironment;

const route = useRoute();
const { t } = useI18n();
const advisory = ref<Advisory | null>(null);
const loading = ref(true);
const error = ref("");
const affected = ref<AffectedEnvironment[]>([]);
const affectedLoading = ref(true);
const affectedError = ref("");
const affectedTotal = ref(0);
const globalTotal = ref<number | null>(null);

const descriptionHtml = computed(() => sanitizeHtml(advisory.value?.descriptionHtml));

const severityLabel = computed(() => {
  const level = advisory.value?.effectiveSeverity;
  return level ? t(`advisories.severity.${level}`) : t("advisories.severity.unknown");
});

const hasEnrichment = computed(() => {
  const a = advisory.value;
  if (!a) return false;
  return !!(
    a.shopwareImpactSummary ||
    a.remediationSummary ||
    a.recommendedUpgrade ||
    a.remediationUrl ||
    a.notesPublic ||
    (a.affectedComponents && a.affectedComponents.length) ||
    (a.tags && a.tags.length)
  );
});

const suppressOpen = ref(false);
const suppressTarget = ref<AffectedEnvironment | null>(null);
const suppressReason = ref("");
// Default to a bounded snooze: an unbounded default is how "temporary"
// mitigations quietly become permanent blind spots.
const suppressExpiry = ref("30");
// Default to the environment the action was started from: it is the narrower
// decision, and shop-wide acceptance requires an owner/admin role that a
// member does not have.
const suppressScope = ref<"environment" | "shop">("environment");
const suppressSaving = ref(false);
const suppressError = ref("");

function openSuppress(row: AffectedEnvironment) {
  suppressTarget.value = row;
  suppressReason.value = "";
  suppressExpiry.value = "30";
  suppressScope.value = "environment";
  suppressError.value = "";
  suppressOpen.value = true;
}

function expiryIso(): string | undefined {
  if (suppressExpiry.value === "never") return undefined;
  const days = Number.parseInt(suppressExpiry.value, 10);
  const at = new Date();
  at.setDate(at.getDate() + days);
  return at.toISOString();
}

async function submitSuppress() {
  const target = suppressTarget.value;
  const id = advisory.value?.advisoryId;
  if (!target || !id) return;

  suppressSaving.value = true;
  suppressError.value = "";
  try {
    const { error: apiError, response } = await createAdvisorySuppression({
      path: { advisoryId: id },
      body: {
        shopId: target.shopId,
        // Omitted for shop-wide scope; set narrows it to this environment.
        environmentId: suppressScope.value === "shop" ? undefined : target.environmentId,
        reason: suppressReason.value.trim(),
        expiresAt: expiryIso(),
      },
    });

    if (apiError) {
      // Distinguish the cases a user can act on: a member choosing shop-wide
      // scope, or a scope already suppressed by someone else.
      if (response?.status === 403) {
        suppressError.value = t("advisories.suppressForbidden");
      } else if (response?.status === 409) {
        suppressError.value = t("advisories.suppressConflict");
      } else {
        suppressError.value = t("advisories.suppressFailed");
      }
      return;
    }

    suppressOpen.value = false;
    await loadAffected(id);
  } finally {
    // Cleared in finally so a network throw cannot leave the dialog stuck.
    suppressSaving.value = false;
  }
}

async function loadAffected(id: string) {
  affectedLoading.value = true;
  affectedError.value = "";
  const { data, error: apiError } = await listAdvisoryAffectedEnvironments({
    path: { advisoryId: id },
  });
  affectedLoading.value = false;
  if (apiError || !data) {
    // Never fall through to the empty state: on a security page, "we could not
    // check" rendered as "none of your shops are affected" is the one wrong
    // answer that matters.
    affected.value = [];
    affectedTotal.value = 0;
    globalTotal.value = null;
    affectedError.value = t("advisories.affectedLoadFailed");
    return;
  }

  affected.value = data.environments;
  affectedTotal.value = data.total;
  globalTotal.value = data.globalTotal ?? null;
}

function reloadAffected() {
  const id = advisory.value?.advisoryId;
  if (id) void loadAffected(id);
}

onMounted(async () => {
  const id = String(route.params.id);
  const { data, error: apiError } = await getAdvisory({
    path: { advisoryId: id },
  });
  loading.value = false;
  if (apiError || !data) {
    error.value = t("advisories.notFound");
    affectedLoading.value = false;
    return;
  }
  advisory.value = data;

  await loadAffected(id);
});
</script>
