<template>
  <Panel variant="table">
    <data-table
      v-if="shop"
      :columns="[
        { key: 'label', name: $t('common.name'), sortable: true },
        { key: 'version', name: $t('common.version'), class: 'extension-version-column' },
        { key: 'latestVersion', name: $t('shopDetail.latest'), class: 'extension-version-column' },
        { key: 'ratingAverage', name: $t('shopDetail.rating'), sortable: true },
        { key: 'installedAt', name: $t('shopDetail.installedAt'), sortable: true },
      ]"
      :data="shop?.extensions || []"
      :default-sort="{ key: 'label', desc: false }"
    >
      <template #cell-label="{ row }">
        <div class="extension-label">
          <status-icon :status="getExtensionState(row)" :tooltip="true" />

          <component
            :is="row.storeLink ? 'a' : 'span'"
            v-bind="row.storeLink ? { href: row.storeLink, target: '_blank' } : {}"
          >
            <div class="extension-name">{{ row.label }}</div>
            <span class="extension-technical-name">{{ row.name }}</span>
          </component>
        </div>
      </template>

      <template #cell-version="{ row }">
        <span :data-tooltip="row.version.length > 6 ? row.version : ''">{{
          row.version.replace(/(.{6})..+/, "$1&hellip;")
        }}</span>
        <span
          v-if="row.latestVersion && row.version < row.latestVersion"
          :data-tooltip="$t('shopDetail.updateAvailable')"
          class="extension-update-available"
          @click="openExtensionChangelog(row)"
        >
          <icon-fa6-solid:rotate class="icon icon-warning" />
        </span>
      </template>

      <template #cell-latestVersion="{ row }">
        <span
          v-if="row.latestVersion"
          :data-tooltip="row.latestVersion.length > 6 ? row.latestVersion : ''"
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
  </Panel>

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
import { useShopDetail } from "@/composables/useShopDetail";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";

type Extension = components["schemas"]["ShopExtension"];

const { shop } = useShopDetail();

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

<style>
.extension-version-column {
  white-space: nowrap;
}
</style>

<style scoped>
.extension-label {
  display: flex;
  justify-content: flex-start;
}

.extension-name {
  font-weight: bold;
  white-space: normal;
}

.extension-technical-name,
.extension-changelog-name {
  font-weight: normal;
  color: var(--text-color-muted);
}

.extension-update-available {
  margin-left: 0.4rem;
}

.extension-changelog-item {
  margin-bottom: 0.75rem;
}

.extension-changelog-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.extension-changelog-date {
  color: var(--text-color-muted);
  font-weight: normal;
}

.extension-changelog-content {
  padding-left: 1.5rem;
}

.extension-update-available {
  cursor: pointer;

  &:hover {
    opacity: 0.7;
  }
}
</style>
