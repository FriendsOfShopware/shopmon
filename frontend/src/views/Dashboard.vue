<template>
  <div v-if="environments">
    <header-container :title="$t('dashboard.title')" />
    <Panel>
      <template #title>
        <icon-fa6-solid:shop />
        {{ $t("dashboard.myEnvironments") }}
      </template>

      <element-empty
        v-if="environments.length === 0"
        :route="{ name: 'account.environments.new' }"
        title="No environments yet"
        button="Add Environment"
      >
        Add your first Shopware environment to start monitoring.
      </element-empty>

      <div v-else class="item-grid">
        <div v-for="env in environments" :key="env.id" class="item">
          <router-link
            :to="{
              name: 'account.environments.detail',
              params: {
                organizationId: env.organizationId,
                environmentId: env.id,
              },
            }"
            class="item-link item-wrapper"
          >
            <div class="item-logo">
              <img
                v-if="env.favicon"
                :src="env.favicon"
                :alt="$t('dashboard.environmentLogo')"
                class="item-logo-img"
              />
            </div>

            <div class="item-info">
              <div class="item-name">
                {{ env.name }}
              </div>

              <div class="item-content">
                {{ env.shopName }}<br />
                {{ env.shopwareVersion }}
              </div>

              <div class="item-state">
                <status-icon :status="env.status" />
              </div>
            </div>
          </router-link>
        </div>
      </div>
    </Panel>

    <Panel>
      <template #title>
        <icon-fa6-solid:building />
        {{ $t("dashboard.myOrganizations") }}
      </template>

      <div class="item-grid">
        <div v-for="organization in organizations" :key="organization.id" class="item">
          <router-link
            :to="{
              name: 'account.organizations.detail',
              params: { organizationId: organization.id },
            }"
            class="item-link item-wrapper"
          >
            <div class="item-logo">
              <icon-fa6-solid:building class="item-logo-icon" />
            </div>

            <div class="item-info">
              <div class="item-name">
                {{ organization.name }}
              </div>

              <div class="item-content">
                {{ $t("dashboard.nMembers", { count: organization.memberCount }) }},
                {{ $t("dashboard.nEnvironments", { count: organization.environmentCount }) }}
              </div>
            </div>
          </router-link>
        </div>
      </div>
    </Panel>

    <Panel v-if="changelogs.length > 0">
      <template #title>
        <icon-fa6-solid:file-waveform />
        {{ $t("dashboard.lastChanges") }}
      </template>

      <data-table
        :columns="[
          { key: 'environmentName', name: $t('dashboard.environment'), sortable: true },
          { key: 'extensions', name: $t('dashboard.changes'), sortable: true },
          { key: 'date', name: $t('common.date'), sortable: true, sortPath: 'date' },
        ]"
        :data="changelogs"
      >
        <template #cell-environmentName="{ row }">
          <router-link
            :to="{
              name: 'account.environments.detail',
              params: {
                organizationId: row.environmentOrganizationId,
                environmentId: row.environmentId,
              },
            }"
          >
            {{ row.environmentName }}
          </router-link>
        </template>

        <template #cell-extensions="{ row }">
          {{ sumChanges(row) }}
        </template>

        <template #cell-date="{ row }">
          {{ formatDateTime(row.date) }}
        </template>
      </data-table>
    </Panel>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { sumChanges } from "@/helpers/changelog";
import { formatDateTime } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { ref } from "vue";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

const { t } = useI18n();

const organizations = ref<components["schemas"]["AccountOrganization"][]>();
api.GET("/account/organizations").then(({ data }) => {
  organizations.value = data;
});

const { environments } = useAccountEnvironments();

const changelogs = ref<components["schemas"]["AccountChangelog"][]>([]);

api.GET("/account/changelogs").then(({ data }) => {
  if (data) changelogs.value = data;
});
</script>

<style scoped></style>
