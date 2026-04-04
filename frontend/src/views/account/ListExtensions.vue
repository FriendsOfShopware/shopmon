<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t('nav.myExtensions') }}</h1>
  </div>

  <div v-if="extensions" class="space-y-6">
    <Card v-if="extensions.length === 0">
      <CardContent>
        <element-empty
          :title="$t('shopDetail.extensions')"
          :button="$t('environment.addEnvironment')"
          :route="{ name: 'account.environments.new' }"
        >
          {{ $t("common.getStartedElement") }}
        </element-empty>
      </CardContent>
    </Card>

    <template v-else>
      <Card>
        <CardContent>
          <Input v-model="term" :placeholder="$t('common.search')" />
        </CardContent>
      </Card>

      <Card class="p-0 overflow-hidden">
        <data-table
          :columns="[
            {
              key: 'label',
              name: $t('common.name'),
              sortable: true,
              searchable: true,
            },
            { key: 'environments', name: $t('dashboard.environment') },
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
              <div class="font-bold whitespace-normal">{{ row.label }}</div>
              <span class="font-normal opacity-60">{{ row.name }}</span>
            </component>
          </template>

          <template #cell-environments="{ row }">
            <div
              v-for="(env, rowIndex) in row.environments"
              :key="rowIndex"
              class="whitespace-nowrap leading-tight [&:not(:last-child)]:mb-1"
            >
              <router-link
                :to="{
                  name: 'account.environments.detail',
                  params: {
                    organizationId: env.environmentOrganizationId,
                    environmentId: env.environmentId,
                  },
                }"
              >
                <status-icon :status="getExtensionState(row)" :tooltip="true" />
                {{ env.environmentName }}
              </router-link>
            </div>
          </template>

          <template #cell-version="{ row }">
            <div
              v-for="(env, rowIndex) in row.environments"
              :key="rowIndex"
              class="whitespace-nowrap leading-tight [&:not(:last-child)]:mb-1"
            >
              <span :title="env.version.length > 6 ? env.version : ''">{{
                env.version.replace(/(.{6})..+/, "$1&hellip;")
              }}</span>
              <span
                v-if="row.latestVersion && env.version < row.latestVersion"
                :title="$t('shopDetail.updateAvailableClick')"
                class="cursor-pointer"
                @click="openExtensionChangelog(row)"
              >
                <icon-fa6-solid:rotate class="inline-block size-3 text-amber-500 align-[-0.1em] ml-1" />
              </span>
            </div>
          </template>

          <template #cell-latestVersion="{ row }">
            <span
              v-if="row.latestVersion"
              :title="row.latestVersion.length > 6 ? row.latestVersion : ''"
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
      </Card>
    </template>
  </div>

  <!-- Extension Changelog Modal -->
  <extension-changelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import ElementEmpty from "@/components/layout/ElementEmpty.vue";
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import ExtensionChangelog from "@/components/modal/ExtensionChangelog.vue";
import { ref } from "vue";

const { t } = useI18n();
const term = ref("");
const extensions = ref<components["schemas"]["AccountExtension"][]>([]);

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

api.GET("/account/extensions").then(({ data }) => {
  if (data) extensions.value = data;
});

function getExtensionState(extension: components["schemas"]["AccountExtension"]) {
  if (!extension.installed) {
    return "not installed";
  }
  if (extension.active) {
    return "active";
  }

  return "inactive";
}
</script>
