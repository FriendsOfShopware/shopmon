<template>
  <Panel variant="table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'date', name: $t('common.date'), class: 'changelog-date', sortable: true },
        { key: 'extensions', name: $t('shopDetail.log'), sortable: false },
      ]"
      :data="shop.changelogs"
    >
      <template #cell-date="{ row }">
        {{ formatDateTime(row.date) }}
      </template>

      <template #cell-extensions="{ row }">
        <span class="modal-open-changelog" @click="openShopChangelog(row as AccountChangelog)">
          {{ sumChanges(row as AccountChangelog) }}
        </span>
      </template>
    </data-table>
  </Panel>

  <!-- Changelog Modal -->
  <shop-changelog
    :show="viewShopChangelogDialog"
    :changelog="dialogShopChangelog"
    @close="closeShopChangelog"
  />
</template>

<script setup lang="ts">
import { sumChanges } from "@/helpers/changelog";
import { formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useShopDetail } from "@/composables/useShopDetail";
import { useShopChangelogModal } from "@/composables/useShopChangelogModal";
import ShopChangelog from "@/components/modal/ShopChangelog.vue";

type AccountChangelog = components["schemas"]["AccountChangelog"];

const { shop } = useShopDetail();
const { viewShopChangelogDialog, dialogShopChangelog, openShopChangelog, closeShopChangelog } =
  useShopChangelogModal();
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
