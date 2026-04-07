<template>
  <div v-if="environment" class="space-y-6">
    <!-- Summary bar -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <StatCard
        :icon="IconCircleXmark"
        :value="counts.red"
        :label="$t('common.errors')"
        color="destructive"
      />
      <StatCard
        :icon="IconCircleInfo"
        :value="counts.yellow"
        :label="$t('common.warnings')"
        color="warning"
      />
      <StatCard
        :icon="IconCircleCheck"
        :value="counts.green"
        :label="$t('common.passed')"
        color="success"
      />
    </div>

    <!-- Filter tabs -->
    <div class="flex items-center justify-between gap-4">
      <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
        <button
          v-for="f in filters"
          :key="f.value"
          :class="[
            'rounded-md px-3 py-1 text-sm font-medium transition-colors',
            activeFilter === f.value
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground',
          ]"
          @click="activeFilter = f.value"
        >
          {{ f.label }}
        </button>
      </div>

      <div class="flex items-center gap-2">
        <label class="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground">
          <Switch :checked="showIgnored" @update:checked="showIgnored = $event" />
          {{ $t("common.showIgnored") }}
        </label>
      </div>
    </div>

    <!-- Check list -->
    <div v-if="filteredChecks.length > 0" class="space-y-2">
      <div
        v-for="check in filteredChecks"
        :key="check.id"
        :class="[
          'flex items-start gap-3 rounded-xl border px-4 py-3 transition-colors',
          isIgnored(check.id) ? 'opacity-50' : '',
          check.level === 'red' ? 'border-destructive/20 bg-destructive/5' : '',
          check.level === 'yellow' ? 'border-warning/20 bg-warning/5' : '',
          check.level === 'green' ? 'border-success/20 bg-success/5' : '',
        ]"
      >
        <StatusIcon :status="check.level" class="mt-0.5 shrink-0" />

        <div class="min-w-0 flex-1">
          <div class="text-sm">
            {{ check.message }}
          </div>
          <a
            v-if="check.link"
            :href="check.link"
            target="_blank"
            class="mt-1 inline-flex items-center gap-1 text-xs text-primary hover:underline"
          >
            {{ $t("shopDetail.learnMore") }}
            <icon-fa6-solid:arrow-up-right-from-square class="size-2.5" />
          </a>
        </div>

        <Button
          variant="ghost"
          size="icon"
          class="size-7 shrink-0"
          :title="isIgnored(check.id) ? $t('shopDetail.checkIgnored') : $t('shopDetail.checkUsed')"
          @click="isIgnored(check.id) ? removeIgnore(check.id) : ignoreCheck(check.id)"
        >
          <icon-fa6-solid:eye-slash v-if="isIgnored(check.id)" class="size-3.5 text-destructive" />
          <icon-fa6-solid:eye v-else class="size-3.5 text-muted-foreground" />
        </Button>
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-else
      class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-16 text-center"
    >
      <icon-fa6-solid:circle-check class="size-10 text-success" />
      <h3 class="text-lg font-semibold">{{ $t("shopDetail.allClear") }}</h3>
      <p class="text-sm text-muted-foreground">{{ $t("shopDetail.noChecksMatch") }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";

import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import StatusIcon from "@/components/StatusIcon.vue";
import StatCard from "@/components/StatCard.vue";

import IconCircleXmark from "~icons/fa6-solid/circle-xmark";
import IconCircleInfo from "~icons/fa6-solid/circle-info";
import IconCircleCheck from "~icons/fa6-solid/circle-check";

const { t } = useI18n();
const { info } = useAlert();
const { environment } = useEnvironmentDetail();

const activeFilter = ref<"all" | "red" | "yellow" | "green">("all");
const showIgnored = ref(false);

const filters = [
  { label: t("common.all"), value: "all" as const },
  { label: t("common.errors"), value: "red" as const },
  { label: t("common.warnings"), value: "yellow" as const },
  { label: t("common.passed"), value: "green" as const },
];

const counts = computed(() => {
  const checks = environment.value?.checks ?? [];
  return {
    red: checks.filter((c) => c.level === "red").length,
    yellow: checks.filter((c) => c.level === "yellow").length,
    green: checks.filter((c) => c.level === "green").length,
  };
});

function isIgnored(id: string): boolean {
  return environment.value?.ignores?.includes(id) ?? false;
}

const sortedChecks = computed(() => {
  if (!environment.value?.checks) return [];
  return [...environment.value.checks].sort((a, b) => {
    const order: Record<string, number> = { red: 0, yellow: 1, green: 2 };
    return (order[a.level] ?? 3) - (order[b.level] ?? 3);
  });
});

const filteredChecks = computed(() => {
  let checks = sortedChecks.value;

  if (activeFilter.value !== "all") {
    checks = checks.filter((c) => c.level === activeFilter.value);
  }

  if (!showIgnored.value) {
    checks = checks.filter((c) => !isIgnored(c.id));
  }

  return checks;
});

async function ignoreCheck(id: string) {
  if (!environment.value) return;

  const updatedIgnores = [...(environment.value.ignores ?? []), id];

  await api.PATCH("/environments/{environmentId}", {
    params: { path: { environmentId: environment.value.id } },
    body: {
      shopId: environment.value.shopId ?? 0,
      ignores: updatedIgnores,
    },
  });

  environment.value = { ...environment.value, ignores: updatedIgnores };
  info(t("shopDetail.ignoreStateUpdated"));
}

async function removeIgnore(id: string) {
  if (!environment.value) return;

  const updatedIgnores = (environment.value.ignores ?? []).filter((aid: string) => aid !== id);

  await api.PATCH("/environments/{environmentId}", {
    params: { path: { environmentId: environment.value.id } },
    body: {
      shopId: environment.value.shopId ?? 0,
      ignores: updatedIgnores,
    },
  });

  environment.value = { ...environment.value, ignores: updatedIgnores };
  info(t("shopDetail.ignoreStateUpdated"));
}
</script>
