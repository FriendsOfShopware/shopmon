<template>
  <Panel variant="table">
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
          <icon-fa6-solid:up-right-from-square v-if="row.link" class="icon icon-xs" />
        </component>
      </template>

      <template #cell-actions="{ row }">
        <button
          v-if="environment.ignores?.includes(row.id)"
          :data-tooltip="$t('shopDetail.checkIgnored')"
          class="tooltip-top-left"
          type="button"
          @click="removeIgnore(row.id)"
        >
          <icon-fa6-solid:eye-slash class="icon icon-error" />
        </button>

        <button
          v-else
          :data-tooltip="$t('shopDetail.checkUsed')"
          class="tooltip-top-left"
          type="button"
          @click="ignoreCheck(row.id)"
        >
          <icon-fa6-solid:eye class="icon" />
        </button>
      </template>
    </data-table>
  </Panel>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { computed } from "vue";

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
