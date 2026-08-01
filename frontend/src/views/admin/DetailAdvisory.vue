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
                  {{ $t("advisories.packagistSeverity") }}
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
                  :href="advisory.link"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="mt-0.5 inline-flex items-start gap-1.5 break-all text-xs text-primary hover:underline"
                >
                  <icon-fa6-solid:arrow-up-right-from-square class="mt-0.5 size-3 shrink-0" />
                  {{ advisory.link }}
                </a>
              </div>
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
              <Label>{{ $t("admin.severityOverride") }}</Label>
              <Select v-model="form.severityOverride">
                <SelectTrigger>
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
              <Switch v-model="form.isVisible" />
              <Label>{{ $t("admin.visibleToUsers") }}</Label>
            </div>

            <div class="space-y-2">
              <Label>{{ $t("advisories.impact") }}</Label>
              <Textarea v-model="form.shopwareImpactSummary" rows="3" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.remediation") }}</Label>
              <Textarea v-model="form.remediationSummary" rows="3" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.recommendedUpgrade") }}</Label>
              <Input v-model="form.recommendedUpgrade" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.remediationUrl") }}</Label>
              <Input v-model="form.remediationUrl" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.components") }} ({{ $t("admin.commaSeparated") }})</Label>
              <Input v-model="form.affectedComponents" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.tags") }} ({{ $t("admin.commaSeparated") }})</Label>
              <Input v-model="form.tags" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("advisories.notes") }}</Label>
              <Textarea v-model="form.notesPublic" rows="3" />
            </div>
            <div class="space-y-2">
              <Label>{{ $t("admin.notesInternal") }}</Label>
              <Textarea v-model="form.notesInternal" rows="3" />
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
import SeverityBadge from "@/components/advisories/SeverityBadge.vue";
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
  type SeverityLevel,
  adminGetAdvisory,
  adminUpdateAdvisory,
} from "@/api/generated";
import { formatDate } from "@/helpers/formatter";
import { sanitizeHtml } from "@/helpers/sanitize";
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const route = useRoute();
const { t } = useI18n();
const advisory = ref<AdminAdvisory | null>(null);
const loading = ref(true);
const error = ref("");
const saving = ref(false);
const saveError = ref("");
const saveSuccess = ref(false);
const descriptionHtml = computed(() => sanitizeHtml(advisory.value?.descriptionHtml));

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
      form.severityOverride === "none-override" ? null : (form.severityOverride as SeverityLevel);

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

onMounted(async () => {
  const id = String(route.params.id);
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
});
</script>
