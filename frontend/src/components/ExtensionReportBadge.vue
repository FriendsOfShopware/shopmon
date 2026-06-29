<template>
  <Tooltip v-if="total > 0">
    <TooltipTrigger as-child>
      <Badge
        variant="outline"
        class="cursor-default gap-1 border-warning/40 bg-warning/10 text-warning"
      >
        <icon-fa6-solid:triangle-exclamation class="size-2.5" />
        {{ total }}
      </Badge>
    </TooltipTrigger>
    <TooltipContent class="max-w-xs">
      <p class="mb-1 font-medium">{{ $t("reportExtension.communityReports") }}</p>
      <ul class="space-y-0.5">
        <li v-for="r in reports" :key="r.category">
          {{ $t(`reportExtension.categories.${r.category}`) }}: {{ r.count }}
        </li>
      </ul>
    </TooltipContent>
  </Tooltip>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { components } from "@/types/api";

type ExtensionReportSummary = components["schemas"]["ExtensionReportSummary"];

const props = defineProps<{
  reports?: ExtensionReportSummary[] | null;
}>();

const reports = computed(() => props.reports ?? []);
const total = computed(() => reports.value.reduce((sum, r) => sum + r.count, 0));
</script>
