<template>
  <div v-if="shops">
    <header-container :title="$t('dashboard.title')" />
    <Panel>
      <template #title>
        <icon-fa6-solid:shop />
        {{ $t("dashboard.myShops") }}
      </template>

      <element-empty
        v-if="shops.length === 0"
        :route="{ name: 'account.shops.new' }"
        title="No shops yet"
        button="Add Shop"
      >
        Add your first Shopware shop to start monitoring.
      </element-empty>

      <div v-else class="item-grid">
        <div v-for="shop in shops" :key="shop.id" class="item">
          <router-link
            :to="{
              name: 'account.shops.detail',
              params: {
                organizationId: shop.organizationId,
                shopId: shop.id,
              },
            }"
            class="item-link item-wrapper"
          >
            <div class="item-logo">
              <img
                v-if="shop.favicon"
                :src="shop.favicon"
                :alt="$t('dashboard.shopLogo')"
                class="item-logo-img"
              />
            </div>

            <div class="item-info">
              <div class="item-name">
                {{ shop.name }}
              </div>

              <div class="item-content">
                {{ shop.projectName }}<br />
                {{ shop.shopwareVersion }}
              </div>

              <div class="item-state">
                <status-icon :status="shop.status" />
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
                {{ $t("dashboard.nShops", { count: organization.shopCount }) }}
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
          { key: 'shopName', name: $t('dashboard.shop'), sortable: true },
          { key: 'extensions', name: $t('dashboard.changes'), sortable: true },
          { key: 'date', name: $t('common.date'), sortable: true, sortPath: 'date' },
        ]"
        :data="changelogs"
      >
        <template #cell-shopName="{ row }">
          <router-link
            :to="{
              name: 'account.shops.detail',
              params: {
                organizationId: row.shopOrganizationId,
                shopId: row.shopId,
              },
            }"
          >
            {{ row.shopName }}
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
import { useAccountShops } from "@/composables/useAccountShops";

const { t } = useI18n();

const organizations = ref<components["schemas"]["AccountOrganization"][]>();
api.GET("/account/organizations").then(({ data }) => {
  organizations.value = data;
});

const { shops } = useAccountShops();

const changelogs = ref<components["schemas"]["AccountChangelog"][]>([]);

api.GET("/account/changelogs").then(({ data }) => {
  if (data) changelogs.value = data;
});
</script>

<style scoped></style>
