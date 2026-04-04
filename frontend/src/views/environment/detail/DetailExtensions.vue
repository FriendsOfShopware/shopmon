<template>
  <Card class="p-0 overflow-hidden">
    <data-table
      v-if="environment"
      :columns="[
        { key: 'label', name: $t('common.name'), sortable: true },
        { key: 'version', name: $t('common.version'), class: 'whitespace-nowrap' },
        { key: 'latestVersion', name: $t('shopDetail.latest'), class: 'whitespace-nowrap' },
        { key: 'ratingAverage', name: $t('shopDetail.rating'), sortable: true },
        { key: 'installedAt', name: $t('shopDetail.installedAt'), sortable: true },
      ]"
      :data="environment?.extensions || []"
      :default-sort="{ key: 'label', desc: false }"
    >
      <template #cell-label="{ row }">
        <div class="flex justify-start">
          <status-icon :status="getExtensionState(row)" :tooltip="true" />

          <component
            :is="row.storeLink ? 'a' : 'span'"
            v-bind="row.storeLink ? { href: row.storeLink, target: '_blank' } : {}"
          >
            <div class="font-bold whitespace-normal">{{ row.label }}</div>
            <span class="font-normal text-muted-foreground">{{ row.name }}</span>
          </component>
        </div>
      </template>

      <template #cell-version="{ row }">
        <span :title="row.version.length > 6 ? row.version : ''">{{
          row.version.replace(/(.{6})..+/, "$1&hellip;")
        }}</span>
        <span
          v-if="row.latestVersion && row.version < row.latestVersion"
          :title="$t('shopDetail.updateAvailable')"
          class="ml-1.5 cursor-pointer hover:opacity-70"
          @click="openExtensionChangelog(row)"
        >
          <icon-fa6-solid:rotate class="inline size-3.5 text-yellow-500" />
        </span>
      </template>

      <template #cell-latestVersion="{ row }">
        <span
          v-if="row.latestVersion"
          :title="row.latestVersion.length > 6 ? row.latestVersion : ''"
          >{{ row.latestVersion.replace(/(.{6})..+/, "$1&hellip;") }}</span
        >
      </template>

      <template #cell-ratingAverage="{ row }">
        <rating-stars :rating="row.ratingAverage ?? null" />
      </template>

      <template #cell-installedAt="{ row }">
        <template v-if="row.installedAt">
          {{ formatDateTime(row.installedAt) }}
        </template>
      </template>
    </data-table>
  </Card>

  <!-- Extension Changelog Modal -->
  <extension-changelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />
</template>

<script setup lang="ts">
import { formatDateTime } from "@/helpers/formatter";
import type { components } from "@/types/api";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";

import { Card } from "@/components/ui/card";

type Extension = components["schemas"]["EnvironmentExtension"];

const { environment } = useEnvironmentDetail();

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

function getExtensionState(extension: Extension) {
  if (!extension.installed) {
    return "not installed";
  }
  if (extension.active) {
    return "active";
  }

  return "inactive";
}
</script>
