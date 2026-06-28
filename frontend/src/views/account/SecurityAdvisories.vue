<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("securityAdvisories.title") }}</h1>
      <p class="text-sm text-muted-foreground">{{ $t("securityAdvisories.description") }}</p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center gap-2 py-16 text-muted-foreground">
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("securityAdvisories.loading") }}
    </div>

    <template v-if="!loading && !error">
      <div v-if="packages.length" class="flex items-center gap-2">
        <Select v-model="packageFilter">
          <SelectTrigger class="w-72">
            <SelectValue :placeholder="$t('securityAdvisories.allPackages')" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{{ $t("securityAdvisories.allPackages") }}</SelectItem>
            <SelectItem v-for="pkg in packages" :key="pkg" :value="pkg">{{ pkg }}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Card>
        <CardContent class="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ $t("securityAdvisories.colSeverity") }}</TableHead>
                <TableHead>{{ $t("securityAdvisories.colTitle") }}</TableHead>
                <TableHead>{{ $t("securityAdvisories.colPackage") }}</TableHead>
                <TableHead>{{ $t("securityAdvisories.colAffectedVersions") }}</TableHead>
                <TableHead>{{ $t("securityAdvisories.colCve") }}</TableHead>
                <TableHead>{{ $t("securityAdvisories.colReportedAt") }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="!filteredAdvisories.length" :colspan="6">
                {{ $t("securityAdvisories.empty") }}
              </TableEmpty>
              <TableRow v-for="advisory in filteredAdvisories" :key="advisory.advisoryId">
                <TableCell>
                  <Badge :variant="severityVariant(advisory.severity)" class="capitalize">
                    {{ advisory.severity || $t("securityAdvisories.unknownSeverity") }}
                  </Badge>
                </TableCell>
                <TableCell class="max-w-md">
                  <a
                    v-if="advisory.link"
                    :href="advisory.link"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="font-medium text-primary hover:underline"
                  >
                    {{ advisory.title }}
                  </a>
                  <span v-else class="font-medium">{{ advisory.title }}</span>
                </TableCell>
                <TableCell>
                  <span class="font-mono text-xs">{{ advisory.packageName }}</span>
                </TableCell>
                <TableCell>
                  <span class="font-mono text-xs">{{ advisory.affectedVersions }}</span>
                </TableCell>
                <TableCell>
                  <a
                    v-if="advisory.cve"
                    :href="`https://nvd.nist.gov/vuln/detail/${advisory.cve}`"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="font-mono text-xs text-primary hover:underline"
                  >
                    {{ advisory.cve }}
                  </a>
                  <span v-else class="text-xs text-muted-foreground">—</span>
                </TableCell>
                <TableCell class="whitespace-nowrap text-xs text-muted-foreground">
                  {{ formatDate(advisory.reportedAt) }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, onMounted, ref } from "vue";

type SecurityAdvisory = components["schemas"]["SecurityAdvisory"];

const advisories = ref<SecurityAdvisory[]>([]);
const loading = ref(true);
const error = ref("");
const packageFilter = ref("all");

const packages = computed(() => [...new Set(advisories.value.map((a) => a.packageName))].sort());

const filteredAdvisories = computed(() => {
  if (packageFilter.value === "all") return advisories.value;
  return advisories.value.filter((a) => a.packageName === packageFilter.value);
});

function severityVariant(severity: string | null | undefined) {
  switch (severity?.toLowerCase()) {
    case "critical":
    case "high":
      return "destructive" as const;
    case "medium":
      return "default" as const;
    default:
      return "secondary" as const;
  }
}

function formatDate(value: string | null | undefined) {
  if (!value) return "—";
  return new Date(value).toLocaleDateString();
}

async function loadAdvisories() {
  loading.value = true;
  error.value = "";

  try {
    const { data, error: resError } = await api.GET("/info/security-advisories");
    if (resError || !data) {
      throw new Error("request failed");
    }
    advisories.value = data;
  } catch (err) {
    error.value = `Failed to load security advisories: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading.value = false;
  }
}

onMounted(loadAdvisories);
</script>
