<template>
  <header-container :title="$t('nav.myExtensions')" />

  <main-container v-if="extensions">
    <Panel v-if="extensions.length === 0">
      <element-empty
        :title="$t('shopDetail.extensions')"
        :button="$t('shop.addShop')"
        :route="{ name: 'account.shops.new' }"
      >
        {{ $t("common.getStartedElement") }}
      </element-empty>
    </Panel>

    <template v-else>
      <Panel>
        <input v-model="term" class="field field-search" :placeholder="$t('common.search')" />
      </Panel>

      <Panel variant="table">
        <data-table
          :columns="[
            {
              key: 'label',
              name: $t('common.name'),
              class: 'extension-label',
              sortable: true,
              searchable: true,
            },
            { key: 'shops', name: $t('dashboard.shop') },
            { key: 'version', name: $t('common.version') },
            { key: 'latestVersion', name: $t('shopDetail.latest') },
            { key: 'ratingAverage', name: $t('shopDetail.rating'), sortable: true },
            { key: 'installedAt', name: $t('shopDetail.installedAt'), sortable: true },
          ]"
          :data="extensions"
          :default-sort="{ key: 'label', desc: false }"
          :search-term="term"
        >
          <template #cell-label="{ row }">
            <component
              :is="row.storeLink ? 'a' : 'span'"
              v-bind="row.storeLink ? { href: row.storeLink, target: '_blank' } : {}"
            >
              <div class="extension-name">{{ row.label }}</div>
              <span class="extension-technical-name">{{ row.name }}</span>
            </component>
          </template>

          <template #cell-shops="{ row }">
            <div v-for="(shop, rowIndex) in row.shops" :key="rowIndex" class="shops-row">
              <router-link
                :to="{
                  name: 'account.shops.detail',
                  params: {
                    slug: shop.organizationSlug,
                    shopId: shop.id,
                  },
                }"
              >
                <status-icon :status="getExtensionState(row)" :tooltip="true" />
                {{ shop.name }}
              </router-link>
            </div>
          </template>

          <template #cell-version="{ row }">
            <div v-for="(shop, rowIndex) in row.shops" :key="rowIndex" class="shops-row">
              <span :data-tooltip="shop.version.length > 6 ? shop.version : ''">{{
                shop.version.replace(/(.{6})..+/, "$1&hellip;")
              }}</span>
              <span
                v-if="row.latestVersion && shop.version < row.latestVersion"
                :data-tooltip="$t('shopDetail.updateAvailableClick')"
                class="extension-update-available"
                @click="openExtensionChangelog(row)"
              >
                <icon-fa6-solid:rotate class="icon icon-warning icon-update" />
              </span>
            </div>
          </template>

          <template #cell-latestVersion="{ row }">
            <span
              v-if="row.latestVersion"
              :data-tooltip="row.latestVersion.length > 6 ? row.latestVersion : ''"
              >{{ row.latestVersion.replace(/(.{6})..+/, "$1&hellip;") }}</span
            >
          </template>

          <template #cell-ratingAverage="{ row }">
            <RatingStars :rating="row.ratingAverage" />
          </template>

          <template #cell-installedAt="{ row }">
            <template v-if="row.installedAt">
              {{ formatDateTime(row.installedAt) }}
            </template>
          </template>
        </data-table>
      </Panel>
    </template>
  </main-container>

  <!-- Extension Changelog Modal -->
  <extension-changelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import ElementEmpty from "@/components/layout/ElementEmpty.vue";
import { formatDateTime } from "@/helpers/formatter";
import { type RouterOutput, trpcClient } from "@/helpers/trpc";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";
import { ref } from "vue";

const { t } = useI18n();
const term = ref("");
const extensions = ref<RouterOutput["account"]["currentUserExtensions"]>([]);

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

trpcClient.account.currentUserExtensions.query().then((data) => {
  extensions.value = data;
});

function getExtensionState(extension: RouterOutput["account"]["currentUserExtensions"][number]) {
  if (!extension.installed) {
    return "not installed";
  }
  if (extension.active) {
    return "active";
  }

  return "inactive";
}
</script>

<style scoped>
.extension-label {
  .extension-name {
    font-weight: bold;
    white-space: normal;
  }

  .extension-technical-name {
    font-weight: normal;
    opacity: 0.6;
  }
}

.extension-update-available {
  cursor: pointer;
}

.shops-row {
  white-space: nowrap;
  line-height: 1.2rem;

  &:not(:last-child) {
    margin-bottom: 0.35rem;
  }

  .icon-update {
    vertical-align: -0.1em;
    margin-left: 0.3rem;
  }
}
</style>

<style>
.shops-row {
  .icon-status {
    vertical-align: -0.2em;
  }
}
</style>
