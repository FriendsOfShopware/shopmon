<template>
  <Panel variant="table">
    <data-table
      v-if="environment"
      :columns="[
        { key: 'date', name: $t('common.date'), class: 'changelog-date', sortable: true },
        { key: 'extensions', name: $t('shopDetail.log'), sortable: false },
      ]"
      :data="environment.changelogs"
    >
      <template #cell-date="{ row }">
        {{ formatDateTime(row.date) }}
      </template>

      <template #cell-extensions="{ row }">
        <span
          class="modal-open-changelog"
          @click="openEnvironmentChangelog(row as AccountChangelog)"
        >
          {{ sumChanges(row as AccountChangelog) }}
        </span>
      </template>
    </data-table>
  </Panel>

  <!-- Changelog Modal -->
  <environment-changelog
    :show="viewEnvironmentChangelogDialog"
    :changelog="dialogEnvironmentChangelog"
    @close="closeEnvironmentChangelog"
  />
</template>

<script setup lang="ts">
import { sumChanges } from "@/helpers/changelog";
import { formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useEnvironmentChangelogModal } from "@/composables/useEnvironmentChangelogModal";
import EnvironmentChangelog from "@/components/modal/ShopChangelog.vue";

type AccountChangelog = components["schemas"]["AccountChangelog"];

const { environment } = useEnvironmentDetail();
const {
  viewEnvironmentChangelogDialog,
  dialogEnvironmentChangelog,
  openEnvironmentChangelog,
  closeEnvironmentChangelog,
} = useEnvironmentChangelogModal();
</script>

<style>
.changelog-date {
  width: 200px;
}
</style>

<style scoped>
.modal-open-changelog {
  cursor: pointer;
}
</style>
