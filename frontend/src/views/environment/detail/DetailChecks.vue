<template>
  <Card class="p-0 overflow-hidden">
    <data-table
      v-if="environment"
      :columns="[{ key: 'message', name: $t('shopDetail.message') }]"
      :data="sortedChecks"
    >
      <template #cell-message="{ row }">
        <status-icon :status="row.level" />

        <component
          :is="row.link ? 'a' : 'span'"
          v-bind="row.link ? { href: row.link, target: '_blank' } : {}"
        >
          {{ row.message }}
          <icon-fa6-solid:up-right-from-square v-if="row.link" class="inline size-3 ml-1" />
        </component>
      </template>

      <template #cell-actions="{ row }">
        <button
          v-if="environment.ignores?.includes(row.id)"
          :title="$t('shopDetail.checkIgnored')"
          type="button"
          @click="removeIgnore(row.id)"
        >
          <icon-fa6-solid:eye-slash class="size-4 text-destructive" />
        </button>

        <button
          v-else
          :title="$t('shopDetail.checkUsed')"
          type="button"
          @click="ignoreCheck(row.id)"
        >
          <icon-fa6-solid:eye class="size-4" />
        </button>
      </template>
    </data-table>
  </Card>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { computed } from "vue";

import { Card } from "@/components/ui/card";

const { t } = useI18n();

const { environment } = useEnvironmentDetail();

// Sort checks by status: red first, then yellow, then green
const sortedChecks = computed(() => {
  if (!environment.value?.checks) return [];
  return [...environment.value.checks].sort((a, b) => {
    if (a.level === "red" && b.level !== "red") return -1;
    if (a.level !== "red" && b.level === "red") return 1;
    if (a.level === "yellow" && b.level === "green") return -1;
    if (a.level === "green" && b.level === "yellow") return 1;
    return 0;
  });
});

const { info } = useAlert();

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

  // Update environment data
  environment.value = { ...environment.value, ignores: updatedIgnores };
  notificateIgnoreUpdate();
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

  // Update environment data
  environment.value = { ...environment.value, ignores: updatedIgnores };
  notificateIgnoreUpdate();
}

async function notificateIgnoreUpdate() {
  info(t("shopDetail.ignoreStateUpdated"));
}
</script>
