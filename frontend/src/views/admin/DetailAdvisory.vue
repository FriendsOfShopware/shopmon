<template>
  <div class="space-y-6">
    <Button as-child variant="ghost" size="sm" class="-ml-2">
      <RouterLink :to="{ name: 'admin.advisories' }">
        <icon-fa6-solid:arrow-left class="mr-1.5 size-3.5" />
        {{ $t("admin.backToAdvisories") }}
      </RouterLink>
    </Button>

    <div v-if="loading" class="space-y-6" aria-hidden="true">
      <div class="space-y-3">
        <Skeleton class="h-8 w-2/3" />
        <Skeleton class="h-4 w-1/3" />
      </div>
      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,420px)]">
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
          <ShopmonAdvisoryBadge :advisory="advisory" />
          <Badge v-if="advisory.origin === 'github'" variant="outline" class="gap-1">
            <icon-fa6-brands:github class="size-2.5" />
            {{ $t("admin.originGithub") }}
          </Badge>
          <Badge v-if="!advisory.isVisible" variant="outline">{{ $t("admin.hidden") }}</Badge>
        </div>
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

      <Card>
        <CardHeader>
          <CardTitle class="text-base">{{ $t("admin.advisoryImpact.title") }}</CardTitle>
          <p class="text-sm text-muted-foreground">{{ $t("admin.advisoryImpact.hint") }}</p>
        </CardHeader>
        <CardContent>
          <dl class="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div v-for="metric in impactMetrics" :key="metric.key">
              <dt class="text-xs font-medium text-muted-foreground">{{ metric.label }}</dt>
              <dd class="mt-1 text-2xl font-semibold tabular-nums">{{ metric.value }}</dd>
            </div>
          </dl>
          <p class="mt-4 text-xs text-muted-foreground">
            {{ $t("admin.advisoryImpact.suppressionHint") }}
          </p>
        </CardContent>
      </Card>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,420px)]">
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

          <Card>
            <CardHeader>
              <CardTitle class="text-base">{{ $t("advisories.packagistDetails") }}</CardTitle>
            </CardHeader>
            <CardContent class="grid gap-4 text-sm sm:grid-cols-2">
              <div class="sm:col-span-2" v-if="advisory.packages?.length">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.affectedPackages") }}
                </div>
                <ul class="mt-1.5 space-y-1.5">
                  <li
                    v-for="pkg in advisory.packages"
                    :key="pkg.packageName"
                    class="rounded-md bg-muted px-2.5 py-2"
                  >
                    <span class="font-mono text-xs font-medium">{{ pkg.packageName }}</span>
                    <code class="ml-2 font-mono text-[11px] text-muted-foreground">{{
                      pkg.affectedVersions
                    }}</code>
                  </li>
                </ul>
              </div>
              <div>
                <div class="text-xs font-medium text-muted-foreground">
                  {{
                    isManual ? $t("advisories.kpi.severity") : $t("advisories.packagistSeverity")
                  }}
                </div>
                <SeverityBadge
                  v-if="advisory.severity"
                  :severity="advisory.severity"
                  class="mt-1 text-[10px]"
                />
                <span v-else>—</span>
              </div>
              <div v-if="advisory.cve">
                <div class="text-xs font-medium text-muted-foreground">CVE</div>
                <div class="mt-0.5 font-mono text-xs">{{ advisory.cve }}</div>
              </div>
              <div v-if="advisory.ghsaId">
                <div class="text-xs font-medium text-muted-foreground">GHSA</div>
                <div class="mt-0.5 font-mono text-xs">{{ advisory.ghsaId }}</div>
              </div>
              <div v-if="advisory.cvssScore != null">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.cvss") }}
                </div>
                <div class="mt-0.5 font-mono text-xs">{{ advisory.cvssScore.toFixed(1) }}</div>
              </div>
              <div v-if="advisory.summary" class="sm:col-span-2">
                <div class="text-xs font-medium text-muted-foreground">
                  {{ $t("advisories.summary") }}
                </div>
                <p class="mt-0.5">{{ advisory.summary }}</p>
              </div>
              <div v-if="advisory.link" class="sm:col-span-2">
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
            </CardContent>
          </Card>
          <!-- Shopmon-authored content stays editable until the official
               CVE/GHSA is linked; from then on the sync owns these fields. -->
          <Card v-if="isManual">
            <CardHeader>
              <CardTitle class="text-base">{{ $t("admin.coreContent") }}</CardTitle>
            </CardHeader>
            <CardContent v-if="isLocked" class="text-sm text-muted-foreground">
              {{ $t("admin.coreLocked") }}
            </CardContent>
            <CardContent v-else class="space-y-4">
              <Alert v-if="contentError" variant="destructive">
                <AlertDescription>{{ contentError }}</AlertDescription>
              </Alert>
              <Alert v-if="contentSuccess">
                <AlertDescription>{{ $t("admin.coreSaved") }}</AlertDescription>
              </Alert>

              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-title">{{ $t("admin.advisoryTitle") }}</Label>
                <Input id="edit-advisory-contentForm-title" v-model="contentForm.title" />
              </div>
              <fieldset class="space-y-2">
                <legend class="text-sm font-medium">{{ $t("advisories.affectedPackages") }}</legend>
                <AdvisoryPackageEditor v-model="contentForm.packages" />
              </fieldset>
              <div class="grid gap-4 sm:grid-cols-2">
                <div class="space-y-2">
                  <Label for="edit-advisory-contentForm-severity">{{
                    $t("advisories.kpi.severity")
                  }}</Label>
                  <Select v-model="contentForm.severity">
                    <SelectTrigger id="edit-advisory-contentForm-severity">
                      <SelectValue :placeholder="$t('advisories.severity.unknown')" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="unknown-severity">
                        {{ $t("advisories.severity.unknown") }}
                      </SelectItem>
                      <SelectItem v-for="s in severities" :key="s" :value="s">
                        {{ $t(`advisories.severity.${s}`) }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-2">
                  <Label for="edit-advisory-contentForm-reportedAt">{{
                    $t("admin.reportedAt")
                  }}</Label>
                  <Input
                    id="edit-advisory-contentForm-reportedAt"
                    v-model="contentForm.reportedAt"
                    type="date"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="edit-advisory-contentForm-cvssScore">{{
                    $t("advisories.cvss")
                  }}</Label>
                  <Input
                    id="edit-advisory-contentForm-cvssScore"
                    v-model="contentForm.cvssScore"
                    type="number"
                    step="0.1"
                    min="0"
                    max="10"
                    placeholder="0.0–10.0"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="edit-advisory-contentForm-cvssVector">{{
                    $t("advisories.cvssVector")
                  }}</Label>
                  <Input
                    id="edit-advisory-contentForm-cvssVector"
                    v-model="contentForm.cvssVector"
                    class="font-mono text-xs"
                  />
                </div>
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-summary">{{
                  $t("advisories.summary")
                }}</Label>
                <Textarea
                  id="edit-advisory-contentForm-summary"
                  v-model="contentForm.summary"
                  rows="2"
                />
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-description"
                  >{{ $t("advisories.fullDescription") }} (Markdown)</Label
                >
                <Textarea
                  id="edit-advisory-contentForm-description"
                  v-model="contentForm.description"
                  rows="8"
                  class="font-mono text-xs"
                />
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-cwes"
                  >{{ $t("advisories.cwes") }} ({{ $t("admin.commaSeparated") }})</Label
                >
                <Input
                  id="edit-advisory-contentForm-cwes"
                  v-model="contentForm.cwes"
                  class="font-mono text-xs"
                  placeholder="CWE-79"
                />
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-link">{{
                  $t("advisories.disclosure")
                }}</Label>
                <Input
                  id="edit-advisory-contentForm-link"
                  v-model="contentForm.link"
                  placeholder="https://…"
                />
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-contentForm-externalReferences">
                  {{ $t("advisories.references") }} ({{ $t("admin.referencesHint") }})
                </Label>
                <Textarea
                  id="edit-advisory-contentForm-externalReferences"
                  v-model="contentForm.externalReferences"
                  rows="3"
                  class="font-mono text-xs"
                />
              </div>

              <Button :disabled="savingContent" class="w-full" @click="saveContent">
                <icon-line-md:loading-twotone-loop v-if="savingContent" class="mr-1.5 size-3.5" />
                {{ $t("common.save") }}
              </Button>
            </CardContent>
          </Card>

          <Card v-if="isManual">
            <CardHeader>
              <CardTitle class="text-base">{{ $t("admin.linkOfficial") }}</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <p class="text-sm text-muted-foreground">{{ $t("admin.linkOfficialHint") }}</p>
              <Alert v-if="linkError" variant="destructive">
                <AlertDescription>{{ linkError }}</AlertDescription>
              </Alert>
              <Alert v-if="linkSuccess">
                <AlertDescription>{{ $t("admin.linkSaved") }}</AlertDescription>
              </Alert>
              <div class="grid gap-4 sm:grid-cols-2">
                <div class="space-y-2">
                  <Label for="edit-advisory-linkForm-cve">CVE</Label>
                  <Input
                    id="edit-advisory-linkForm-cve"
                    v-model="linkForm.cve"
                    class="font-mono text-xs"
                    placeholder="CVE-2026-1234"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="edit-advisory-linkForm-ghsaId">GHSA</Label>
                  <Input
                    id="edit-advisory-linkForm-ghsaId"
                    v-model="linkForm.ghsaId"
                    class="font-mono text-xs"
                    placeholder="GHSA-xxxx-xxxx-xxxx"
                  />
                </div>
              </div>
              <div class="space-y-2">
                <Label for="edit-advisory-linkForm-link">{{ $t("advisories.disclosure") }}</Label>
                <Input
                  id="edit-advisory-linkForm-link"
                  v-model="linkForm.link"
                  placeholder="https://…"
                />
              </div>
              <Button
                :disabled="linking || (!linkForm.cve.trim() && !linkForm.ghsaId.trim())"
                class="w-full"
                @click="linkOfficial"
              >
                <icon-line-md:loading-twotone-loop v-if="linking" class="mr-1.5 size-3.5" />
                {{ $t("admin.linkOfficial") }}
              </Button>
            </CardContent>
          </Card>
        </div>

        <!-- Enrichment form -->
        <Card class="self-start lg:sticky lg:top-24">
          <CardHeader>
            <CardTitle class="text-base">{{ $t("admin.enrichment") }}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <Alert v-if="saveError" variant="destructive">
              <AlertDescription>{{ saveError }}</AlertDescription>
            </Alert>
            <Alert v-if="saveSuccess">
              <AlertDescription>{{ $t("admin.enrichmentSaved") }}</AlertDescription>
            </Alert>

            <div class="space-y-2">
              <Label for="edit-advisory-form-severityOverride">{{
                $t("admin.severityOverride")
              }}</Label>
              <Select v-model="form.severityOverride">
                <SelectTrigger id="edit-advisory-form-severityOverride">
                  <SelectValue :placeholder="$t('admin.noOverride')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none-override">{{ $t("admin.noOverride") }}</SelectItem>
                  <SelectItem v-for="s in severities" :key="s" :value="s">
                    {{ $t(`advisories.severity.${s}`) }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="flex items-center gap-3">
              <Switch id="edit-advisory-form-isVisible" v-model="form.isVisible" />
              <Label for="edit-advisory-form-isVisible">{{ $t("admin.visibleToUsers") }}</Label>
            </div>

            <div class="space-y-2">
              <Label for="edit-advisory-form-shopwareImpactSummary">{{
                $t("advisories.impact")
              }}</Label>
              <Textarea
                id="edit-advisory-form-shopwareImpactSummary"
                v-model="form.shopwareImpactSummary"
                rows="3"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-remediationSummary">{{
                $t("advisories.remediation")
              }}</Label>
              <Textarea
                id="edit-advisory-form-remediationSummary"
                v-model="form.remediationSummary"
                rows="3"
              />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-recommendedUpgrade">{{
                $t("advisories.recommendedUpgrade")
              }}</Label>
              <Input id="edit-advisory-form-recommendedUpgrade" v-model="form.recommendedUpgrade" />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-remediationUrl">{{
                $t("advisories.remediationUrl")
              }}</Label>
              <Input id="edit-advisory-form-remediationUrl" v-model="form.remediationUrl" />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-affectedComponents"
                >{{ $t("advisories.components") }} ({{ $t("admin.commaSeparated") }})</Label
              >
              <Input id="edit-advisory-form-affectedComponents" v-model="form.affectedComponents" />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-tags"
                >{{ $t("advisories.tags") }} ({{ $t("admin.commaSeparated") }})</Label
              >
              <Input id="edit-advisory-form-tags" v-model="form.tags" />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-notesPublic">{{ $t("advisories.notes") }}</Label>
              <Textarea id="edit-advisory-form-notesPublic" v-model="form.notesPublic" rows="3" />
            </div>
            <div class="space-y-2">
              <Label for="edit-advisory-form-notesInternal">{{ $t("admin.notesInternal") }}</Label>
              <Textarea
                id="edit-advisory-form-notesInternal"
                v-model="form.notesInternal"
                rows="3"
              />
            </div>

            <Button :disabled="saving" class="w-full" @click="save">
              <icon-line-md:loading-twotone-loop v-if="saving" class="mr-1.5 size-3.5" />
              {{ $t("common.save") }}
            </Button>
          </CardContent>
        </Card>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import AdvisoryPackageEditor from "@/components/advisories/AdvisoryPackageEditor.vue";
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
import ShopmonAdvisoryBadge from "@/components/advisories/ShopmonAdvisoryBadge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  type AdminAdvisory,
  type AdvisoryPackageInput,
  type SeverityLevel,
  adminGetAdvisory,
  adminLinkAdvisoryOfficial,
  adminUpdateAdvisory,
} from "@/api/generated";
import {
  advisoryCwesFromInput,
  isAdvisoryContentLocked,
  isManualAdvisory,
  isSafeHttpUrl,
} from "@/helpers/advisory";
import { formatDate } from "@/helpers/formatter";
import { sanitizeHtml } from "@/helpers/sanitize";
import { computed, reactive, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const route = useRoute();
const { t } = useI18n();
const advisory = ref<AdminAdvisory | null>(null);
const impactMetrics = computed(() => {
  const impact = advisory.value?.impact;
  return (
    ["shopCount", "environmentCount", "activeShopCount", "suppressedEnvironmentCount"] as const
  ).map((key) => ({
    key,
    label: t(`admin.advisoryImpact.${key}`),
    value: impact?.[key] ?? "—",
  }));
});
const loading = ref(true);
const error = ref("");
const saving = ref(false);
const saveError = ref("");
const saveSuccess = ref(false);
const descriptionHtml = computed(() => sanitizeHtml(advisory.value?.descriptionHtml));
const isManual = computed(() => (advisory.value ? isManualAdvisory(advisory.value) : false));
const isLocked = computed(() => (advisory.value ? isAdvisoryContentLocked(advisory.value) : false));

const severities = ["critical", "high", "medium", "low", "none"] as const;

const form = reactive({
  severityOverride: "none-override",
  isVisible: true,
  notesPublic: "",
  notesInternal: "",
  remediationSummary: "",
  remediationUrl: "",
  recommendedUpgrade: "",
  shopwareImpactSummary: "",
  affectedComponents: "",
  tags: "",
});

function splitList(value: string): string[] {
  return value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

const contentForm = reactive({
  title: "",
  packages: [] as AdvisoryPackageInput[],
  severity: "unknown-severity",
  reportedAt: "",
  cvssScore: "",
  cvssVector: "",
  summary: "",
  description: "",
  cwes: "",
  link: "",
  externalReferences: "",
});
const savingContent = ref(false);
const contentError = ref("");
const contentSuccess = ref(false);

const linkForm = reactive({ cve: "", ghsaId: "", link: "" });
const linking = ref(false);
const linkError = ref("");
const linkSuccess = ref(false);

function loadContentForm(a: AdminAdvisory) {
  contentForm.title = a.title;
  contentForm.packages = (a.packages ?? []).map((p) => ({
    packageName: p.packageName,
    affectedVersions: p.affectedVersions,
  }));
  if (contentForm.packages.length === 0) {
    contentForm.packages = [{ packageName: "", affectedVersions: "" }];
  }
  contentForm.severity = a.severity ?? "unknown-severity";
  contentForm.reportedAt = a.reportedAt ? a.reportedAt.slice(0, 10) : "";
  contentForm.cvssScore = a.cvssScore != null ? String(a.cvssScore) : "";
  contentForm.cvssVector = a.cvssVector ?? "";
  contentForm.summary = a.summary ?? "";
  contentForm.description = a.description ?? "";
  contentForm.cwes = (a.cwes ?? []).map((c) => c.id).join(", ");
  contentForm.link = a.link ?? "";
  contentForm.externalReferences = (a.externalReferences ?? []).join("\n");
  contentError.value = "";
  contentSuccess.value = false;
}

function loadLinkForm(a: AdminAdvisory) {
  linkForm.cve = a.cve ?? "";
  linkForm.ghsaId = a.ghsaId ?? "";
  linkForm.link = a.link ?? "";
  linkError.value = "";
  linkSuccess.value = false;
}

async function saveContent() {
  if (!advisory.value) return;
  savingContent.value = true;
  contentError.value = "";
  contentSuccess.value = false;
  try {
    const score = Number.parseFloat(contentForm.cvssScore);
    const { data, error: apiError } = await adminUpdateAdvisory({
      path: { advisoryId: advisory.value.advisoryId },
      body: {
        title: contentForm.title,
        packages: contentForm.packages
          .map((p) => ({
            packageName: p.packageName.trim(),
            affectedVersions: p.affectedVersions.trim(),
          }))
          .filter((p) => p.packageName !== "" || p.affectedVersions !== ""),
        // An explicit empty severity clears it server-side.
        severity: contentForm.severity === "unknown-severity" ? "" : contentForm.severity,
        summary: contentForm.summary,
        description: contentForm.description,
        cvssScore: Number.isFinite(score) ? score : null,
        cvssVector: contentForm.cvssVector,
        cwes: advisoryCwesFromInput(contentForm.cwes, advisory.value.cwes ?? []),
        link: contentForm.link,
        externalReferences: contentForm.externalReferences
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
        reportedAt: contentForm.reportedAt
          ? new Date(`${contentForm.reportedAt}T00:00:00Z`).toISOString()
          : undefined,
      },
    });
    if (apiError || !data) {
      contentError.value =
        (apiError as { message?: string } | undefined)?.message ?? t("admin.coreSaveFailed");
      return;
    }
    // Both cards edit the disclosure URL. Keep an untouched linking field
    // current without replacing an independent, unsaved edit in that card.
    if (linkForm.link === (advisory.value.link ?? "")) {
      linkForm.link = data.link ?? "";
    }
    advisory.value = data;
    loadContentForm(data);
    contentSuccess.value = true;
  } finally {
    savingContent.value = false;
  }
}

async function linkOfficial() {
  if (!advisory.value) return;
  linking.value = true;
  linkError.value = "";
  linkSuccess.value = false;
  try {
    const { data, error: apiError } = await adminLinkAdvisoryOfficial({
      path: { advisoryId: advisory.value.advisoryId },
      body: {
        cve: linkForm.cve.trim() || undefined,
        ghsaId: linkForm.ghsaId.trim() || undefined,
        link: linkForm.link.trim() || undefined,
      },
    });
    if (apiError || !data) {
      linkError.value =
        (apiError as { message?: string } | undefined)?.message ?? t("admin.linkFailed");
      return;
    }
    advisory.value = data;
    loadContentForm(data);
    loadLinkForm(data);
    linkSuccess.value = true;
  } finally {
    linking.value = false;
  }
}

function loadForm(a: AdminAdvisory) {
  form.severityOverride = a.severityOverride ?? "none-override";
  form.isVisible = a.isVisible;
  form.notesPublic = a.notesPublic ?? "";
  form.notesInternal = a.notesInternal ?? "";
  form.remediationSummary = a.remediationSummary ?? "";
  form.remediationUrl = a.remediationUrl ?? "";
  form.recommendedUpgrade = a.recommendedUpgrade ?? "";
  form.shopwareImpactSummary = a.shopwareImpactSummary ?? "";
  form.affectedComponents = (a.affectedComponents ?? []).join(", ");
  form.tags = (a.tags ?? []).join(", ");
}

async function save() {
  if (!advisory.value) return;
  saving.value = true;
  saveError.value = "";
  saveSuccess.value = false;
  try {
    const severityOverride =
      form.severityOverride === "none-override"
        ? undefined
        : (form.severityOverride as SeverityLevel);

    const { data, error: apiError } = await adminUpdateAdvisory({
      path: { advisoryId: advisory.value.advisoryId },
      body: {
        severityOverride,
        isVisible: form.isVisible,
        notesPublic: form.notesPublic,
        notesInternal: form.notesInternal,
        remediationSummary: form.remediationSummary,
        remediationUrl: form.remediationUrl,
        recommendedUpgrade: form.recommendedUpgrade,
        shopwareImpactSummary: form.shopwareImpactSummary,
        affectedComponents: splitList(form.affectedComponents),
        tags: splitList(form.tags),
      },
    });
    if (apiError || !data) {
      saveError.value = t("admin.enrichmentSaveFailed");
      return;
    }
    advisory.value = data;
    loadForm(data);
    saveSuccess.value = true;
  } finally {
    saving.value = false;
  }
}

// Watched rather than loaded once on mount: the router reuses this component
// between /admin/advisories/:id routes, so onMounted alone would leave the
// previous advisory on screen — and its form bound to the wrong row.
watch(
  () => route.params.id,
  async (rawID) => {
    const id = String(rawID ?? "");
    if (!id) return;

    loading.value = true;
    error.value = "";
    advisory.value = null;
    saveError.value = "";
    saveSuccess.value = false;

    const { data, error: apiError } = await adminGetAdvisory({
      path: { advisoryId: id },
    });
    loading.value = false;
    if (apiError || !data) {
      error.value = t("advisories.notFound");
      return;
    }
    advisory.value = data;
    loadForm(data);
    loadContentForm(data);
    loadLinkForm(data);
  },
  { immediate: true },
);
</script>
