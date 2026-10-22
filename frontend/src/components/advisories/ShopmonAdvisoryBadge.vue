<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { isLinkedShopmonAdvisory, isShopmonAdvisory } from "@/helpers/advisory";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  advisory: { origin?: string | null; cve?: string | null; ghsaId?: string | null };
}>();

const { t } = useI18n();

// The badge only exists to distinguish Shopmon reports from official
// advisories, so synced advisories render nothing here.
const visible = computed(() => isShopmonAdvisory(props.advisory));
const label = computed(() =>
  isLinkedShopmonAdvisory(props.advisory)
    ? t("advisories.shopmonFirstReported")
    : t("advisories.shopmonBadge"),
);
</script>

<template>
  <Badge
    v-if="visible"
    variant="outline"
    class="gap-1 border-primary/30 bg-primary/10 text-primary"
  >
    <icon-fa6-solid:shield-halved class="size-2.5" />
    {{ label }}
  </Badge>
</template>
