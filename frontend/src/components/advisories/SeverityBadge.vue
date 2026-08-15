<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { severityBadgeClass } from "@/helpers/advisory";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  severity: string | null | undefined;
}>();

const { t } = useI18n();

const label = computed(() =>
  props.severity ? t(`advisories.severity.${props.severity}`) : t("advisories.severity.unknown"),
);
const badgeClass = computed(() => severityBadgeClass(props.severity));
</script>

<template>
  <Badge :class="badgeClass" class="uppercase">{{ label }}</Badge>
</template>
