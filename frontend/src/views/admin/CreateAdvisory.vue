<template>
  <div class="space-y-6">
    <Button as-child variant="ghost" size="sm" class="-ml-2">
      <RouterLink :to="{ name: 'admin.advisories' }">
        <icon-fa6-solid:arrow-left class="mr-1.5 size-3.5" />
        {{ $t("admin.backToAdvisories") }}
      </RouterLink>
    </Button>

    <div>
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("admin.createAdvisory") }}</h1>
      <p class="mt-1 max-w-3xl text-sm text-muted-foreground">
        {{ $t("admin.createAdvisoryHint") }}
      </p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">{{ $t("admin.importGhsaTitle") }}</CardTitle>
      </CardHeader>
      <CardContent class="space-y-3">
        <p class="text-sm text-muted-foreground">{{ $t("admin.importGhsaHint") }}</p>
        <div class="flex flex-wrap items-center gap-2">
          <Input
            v-model="importInput"
            :aria-label="$t('admin.importGhsaTitle')"
            class="min-w-64 flex-1 font-mono text-xs"
            :placeholder="$t('admin.importGhsaPlaceholder')"
            @keyup.enter.prevent="fetchPreview"
          />
          <Button
            type="button"
            variant="outline"
            :disabled="importing || !importInput.trim()"
            @click="fetchPreview"
          >
            <icon-line-md:loading-twotone-loop v-if="importing" class="mr-1.5 size-3.5" />
            <icon-fa6-brands:github v-else class="mr-1.5 size-3.5" />
            {{ $t("admin.importGhsaButton") }}
          </Button>
        </div>
        <Alert v-if="importError" variant="destructive">
          <AlertDescription>{{ importError }}</AlertDescription>
        </Alert>
        <Alert v-if="existingAdvisoryId">
          <AlertDescription class="flex flex-wrap items-center gap-2">
            {{ $t("admin.importExisting", { id: existingAdvisoryId }) }}
            <RouterLink
              :to="{ name: 'admin.advisories.detail', params: { id: existingAdvisoryId } }"
              class="font-medium text-primary hover:underline"
            >
              {{ $t("admin.importExistingView") }}
            </RouterLink>
          </AlertDescription>
        </Alert>
        <p v-if="importMode" class="text-xs text-muted-foreground">
          {{ $t("admin.importModeNote", { ghsa: importedGhsaId }) }}
          <span v-if="patchedNote">{{ patchedNote }}</span>
        </p>
      </CardContent>
    </Card>

    <form class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,420px)]" @submit.prevent="submit">
      <div class="min-w-0 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle class="text-base">{{ $t("admin.coreContent") }}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="create-advisory-form-title">{{ $t("admin.advisoryTitle") }}</Label>
              <Input id="create-advisory-form-title" v-model="form.title" required />
            </div>

            <fieldset class="space-y-2">
              <legend class="text-sm font-medium">{{ $t("advisories.affectedPackages") }}</legend>
              <AdvisoryPackageEditor v-model="form.packages" />
            </fieldset>

            <div class="space-y-2">
              <Label for="create-advisory-form-summary">{{ $t("advisories.summary") }}</Label>
              <Textarea id="create-advisory-form-summary" v-model="form.summary" rows="2" />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-description"
                >{{ $t("advisories.fullDescription") }} (Markdown)</Label
              >
              <Textarea
                id="create-advisory-form-description"
                v-model="form.description"
                rows="8"
                class="font-mono text-xs"
              />
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <div class="space-y-2">
                <Label for="create-advisory-form-cve">CVE</Label>
                <Input
                  id="create-advisory-form-cve"
                  v-model="form.cve"
                  class="font-mono text-xs"
                  placeholder="CVE-2026-1234"
                />
              </div>
              <div class="space-y-2">
                <Label for="create-advisory-form-ghsaId">GHSA</Label>
                <Input
                  id="create-advisory-form-ghsaId"
                  v-model="form.ghsaId"
                  class="font-mono text-xs"
                  placeholder="GHSA-xxxx-xxxx-xxxx"
                />
              </div>
              <div class="space-y-2">
                <Label for="create-advisory-form-link">{{ $t("advisories.disclosure") }}</Label>
                <Input id="create-advisory-form-link" v-model="form.link" placeholder="https://…" />
              </div>
              <div class="space-y-2">
                <Label for="create-advisory-form-reportedAt">{{ $t("admin.reportedAt") }}</Label>
                <Input id="create-advisory-form-reportedAt" v-model="form.reportedAt" type="date" />
              </div>
            </div>

            <div class="space-y-2">
              <Label for="create-advisory-form-externalReferences"
                >{{ $t("advisories.references") }} ({{ $t("admin.referencesHint") }})</Label
              >
              <Textarea
                id="create-advisory-form-externalReferences"
                v-model="form.externalReferences"
                rows="3"
                class="font-mono text-xs"
              />
            </div>
          </CardContent>
        </Card>
      </div>

      <div class="space-y-6 lg:sticky lg:top-24 lg:self-start">
        <Card>
          <CardHeader>
            <CardTitle class="text-base">{{ $t("advisories.kpi.severity") }}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="create-advisory-form-severity">{{ $t("advisories.kpi.severity") }}</Label>
              <Select v-model="form.severity">
                <SelectTrigger id="create-advisory-form-severity">
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
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="create-advisory-form-cvssScore">{{ $t("advisories.cvss") }}</Label>
                <Input
                  id="create-advisory-form-cvssScore"
                  v-model="form.cvssScore"
                  type="number"
                  step="0.1"
                  min="0"
                  max="10"
                  placeholder="0.0–10.0"
                />
              </div>
              <div class="space-y-2">
                <Label for="create-advisory-form-cvssVector">{{
                  $t("advisories.cvssVector")
                }}</Label>
                <Input
                  id="create-advisory-form-cvssVector"
                  v-model="form.cvssVector"
                  class="font-mono text-xs"
                />
              </div>
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-cwes"
                >{{ $t("advisories.cwes") }} ({{ $t("admin.commaSeparated") }})</Label
              >
              <Input
                id="create-advisory-form-cwes"
                v-model="form.cwes"
                class="font-mono text-xs"
                placeholder="CWE-79"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="text-base">{{ $t("advisories.shopmonContext") }}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="create-advisory-form-shopwareImpactSummary">{{
                $t("advisories.impact")
              }}</Label>
              <Textarea
                id="create-advisory-form-shopwareImpactSummary"
                v-model="form.shopwareImpactSummary"
                rows="3"
              />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-remediationSummary">{{
                $t("advisories.remediation")
              }}</Label>
              <Textarea
                id="create-advisory-form-remediationSummary"
                v-model="form.remediationSummary"
                rows="3"
              />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-recommendedUpgrade">{{
                $t("advisories.recommendedUpgrade")
              }}</Label>
              <Input
                id="create-advisory-form-recommendedUpgrade"
                v-model="form.recommendedUpgrade"
              />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-remediationUrl">{{
                $t("advisories.remediationUrl")
              }}</Label>
              <Input id="create-advisory-form-remediationUrl" v-model="form.remediationUrl" />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-affectedComponents"
                >{{ $t("advisories.components") }} ({{ $t("admin.commaSeparated") }})</Label
              >
              <Input
                id="create-advisory-form-affectedComponents"
                v-model="form.affectedComponents"
              />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-tags"
                >{{ $t("advisories.tags") }} ({{ $t("admin.commaSeparated") }})</Label
              >
              <Input id="create-advisory-form-tags" v-model="form.tags" />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-notesPublic">{{ $t("advisories.notes") }}</Label>
              <Textarea id="create-advisory-form-notesPublic" v-model="form.notesPublic" rows="3" />
            </div>
            <div class="space-y-2">
              <Label for="create-advisory-form-notesInternal">{{
                $t("admin.notesInternal")
              }}</Label>
              <Textarea
                id="create-advisory-form-notesInternal"
                v-model="form.notesInternal"
                rows="3"
              />
            </div>

            <div class="space-y-2 border-t pt-4">
              <div class="flex items-center gap-3">
                <Switch id="create-advisory-form-isVisible" v-model="form.isVisible" />
                <Label for="create-advisory-form-isVisible">{{
                  $t("admin.publishImmediately")
                }}</Label>
              </div>
              <p class="text-xs text-muted-foreground">{{ $t("admin.publishHint") }}</p>
            </div>

            <Button type="submit" :disabled="saving" class="w-full">
              <icon-line-md:loading-twotone-loop v-if="saving" class="mr-1.5 size-3.5" />
              {{ $t("admin.createAdvisory") }}
            </Button>
          </CardContent>
        </Card>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import AdvisoryPackageEditor from "@/components/advisories/AdvisoryPackageEditor.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  type AdvisoryCwe,
  type AdvisoryPackageInput,
  type CreateAdvisoryRequest,
  type SeverityLevel,
  adminCreateAdvisory,
  adminImportGhsaAdvisory,
  adminPreviewGhsaImport,
} from "@/api/generated";
import { advisoryCwesFromInput } from "@/helpers/advisory";
import { computed, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

const { t } = useI18n();
const router = useRouter();
const saving = ref(false);
const error = ref("");

const severities = ["critical", "high", "medium", "low", "none"] as const;

const form = reactive({
  title: "",
  packages: [{ packageName: "", affectedVersions: "" }] as AdvisoryPackageInput[],
  summary: "",
  description: "",
  cve: "",
  ghsaId: "",
  link: "",
  reportedAt: "",
  externalReferences: "",
  severity: "unknown-severity",
  cvssScore: "",
  cvssVector: "",
  cwes: "",
  isVisible: false,
  notesPublic: "",
  notesInternal: "",
  remediationSummary: "",
  remediationUrl: "",
  recommendedUpgrade: "",
  shopwareImpactSummary: "",
  affectedComponents: "",
  tags: "",
});

const importInput = ref("");
const importing = ref(false);
const importError = ref("");
const existingAdvisoryId = ref("");
const importedGhsaId = ref("");
const patchedNote = ref("");
const importedCwes = ref<AdvisoryCwe[]>([]);
const importedUpgrade = ref("");

// Submitting creates a GitHub import exactly while the form still carries
// the fetched GHSA id; editing it away falls back to a Shopmon advisory.
const importMode = computed(
  () => importedGhsaId.value !== "" && form.ghsaId.trim() === importedGhsaId.value,
);

async function fetchPreview() {
  importing.value = true;
  importError.value = "";
  existingAdvisoryId.value = "";
  try {
    const { data, error: apiError } = await adminPreviewGhsaImport({
      body: { input: importInput.value.trim() },
    });
    if (apiError || !data) {
      importError.value =
        (apiError as { message?: string } | undefined)?.message ?? t("admin.importGhsaFailed");
      return;
    }
    if (data.existingAdvisoryId) {
      existingAdvisoryId.value = data.existingAdvisoryId;
      return;
    }

    form.title = data.title;
    form.packages = data.packages.map((p) => ({
      packageName: p.packageName,
      affectedVersions: p.affectedVersions,
    }));
    form.description = data.description ?? "";
    form.severity = data.severity ?? "unknown-severity";
    form.cvssScore = data.cvssScore != null ? String(data.cvssScore) : "";
    form.cvssVector = data.cvssVector ?? "";
    form.cwes = data.cwes.map((c) => c.id).join(", ");
    importedCwes.value = data.cwes;
    form.link = data.link ?? "";
    form.externalReferences = data.externalReferences.join("\n");
    form.reportedAt = data.reportedAt ? data.reportedAt.slice(0, 10) : "";
    form.ghsaId = data.ghsaId;
    form.cve = data.cve ?? "";

    const patched = [...new Set(data.packages.map((p) => p.patchedVersion ?? ""))];
    const nextUpgrade = patched.length === 1 ? patched[0]! : "";
    patchedNote.value = nextUpgrade ? t("admin.importPatched", { version: nextUpgrade }) : "";
    if (form.recommendedUpgrade === "" || form.recommendedUpgrade === importedUpgrade.value) {
      form.recommendedUpgrade = nextUpgrade;
    }
    importedUpgrade.value = nextUpgrade;
    importedGhsaId.value = data.ghsaId;
  } finally {
    importing.value = false;
  }
}

function splitList(value: string): string[] {
  return value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

function optional(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

async function submit() {
  saving.value = true;
  error.value = "";
  try {
    const score = Number.parseFloat(form.cvssScore);
    const body: CreateAdvisoryRequest = {
      title: form.title.trim(),
      packages: form.packages
        .map((p) => ({
          packageName: p.packageName.trim(),
          affectedVersions: p.affectedVersions.trim(),
        }))
        .filter((p) => p.packageName !== "" || p.affectedVersions !== ""),
      severity: form.severity === "unknown-severity" ? undefined : (form.severity as SeverityLevel),
      summary: optional(form.summary),
      description: optional(form.description),
      cvssScore: Number.isFinite(score) ? score : undefined,
      cvssVector: optional(form.cvssVector),
      cwes: advisoryCwesFromInput(form.cwes, importedCwes.value),
      link: optional(form.link),
      externalReferences: form.externalReferences
        .split("\n")
        .map((s) => s.trim())
        .filter(Boolean),
      cve: optional(form.cve),
      ghsaId: optional(form.ghsaId),
      reportedAt: form.reportedAt
        ? new Date(`${form.reportedAt}T00:00:00Z`).toISOString()
        : undefined,
      isVisible: form.isVisible,
      notesPublic: optional(form.notesPublic),
      notesInternal: optional(form.notesInternal),
      remediationSummary: optional(form.remediationSummary),
      remediationUrl: optional(form.remediationUrl),
      recommendedUpgrade: optional(form.recommendedUpgrade),
      shopwareImpactSummary: optional(form.shopwareImpactSummary),
      affectedComponents: splitList(form.affectedComponents),
      tags: splitList(form.tags),
    };

    const { data, error: apiError } = importMode.value
      ? await adminImportGhsaAdvisory({ body })
      : await adminCreateAdvisory({ body });
    if (apiError || !data) {
      // The API explains rejections (bad constraint, duplicate CVE, …) in its
      // message; surfacing it beats a generic failure the admin cannot act on.
      error.value =
        (apiError as { message?: string } | undefined)?.message ?? t("admin.createFailed");
      return;
    }
    await router.push({
      name: "admin.advisories.detail",
      params: { id: data.advisoryId },
    });
  } finally {
    saving.value = false;
  }
}
</script>
