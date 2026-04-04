<template>
  <header-container :title="$t('shop.listTitle')">
    <div class="header-actions">
      <router-link :to="{ name: 'account.shops.new' }" class="btn btn-primary">
        <icon-fa6-solid:folder-plus class="icon" aria-hidden="true" />
        {{ $t("shop.addShop") }}
      </router-link>
    </div>
  </header-container>

  <main-container v-if="!loading">
    <Panel v-if="shops.length === 0">
      <element-empty
        :title="$t('shop.noShops')"
        :button="$t('shop.addShop')"
        :route="{ name: 'account.shops.new' }"
      >
        {{ $t("shop.getStarted") }}
      </element-empty>
    </Panel>

    <div v-else>
      <!-- Shops -->
      <Panel
        v-for="shop in shops"
        :key="shop.id"
        class="shop-panel"
        :description="shop.description || undefined"
      >
        <template #title>{{ shop.name }}</template>
        <template #action>
          <router-link
            :to="{ name: 'account.shops.edit', params: { shopId: shop.id } }"
            class="btn btn-secondary"
          >
            <icon-fa6-solid:pen-to-square class="icon" aria-hidden="true" />
            {{ $t("nav.editShop") }}
          </router-link>
        </template>

        <p class="shop-meta">
          <span class="environment-count">{{
            $t("shop.nEnvironments", { count: shopEnvironments[shop.id]?.length || 0 })
          }}</span>
        </p>

        <div v-if="shopEnvironments[shop.id]?.length > 0" class="item-grid">
          <div v-for="env in shopEnvironments[shop.id]" :key="env.id" class="item">
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
                  alt="Environment favicon"
                  class="item-logo-img"
                />
                <icon-fa6-solid:store v-else />
              </div>

              <div class="item-info">
                <div class="item-name">
                  {{ env.name }}
                </div>

                <div class="item-content">{{ env.shopwareVersion }}</div>
                <status-icon :status="env.status" class="item-state" />
              </div>
            </router-link>

            <div class="item-actions">
              <a
                :href="env.url"
                target="_blank"
                class=""
                :data-tooltip="$t('shop.visitEnvironment')"
              >
                <icon-fa6-solid:arrow-up-right-from-square />
              </a>
            </div>
          </div>

          <router-link
            :to="{ name: 'account.environments.new', query: { shopId: shop.id } }"
            class="item item-link item-wrapper add-environment-card"
          >
            <div class="item-logo">
              <icon-fa6-solid:plus />
            </div>
            <div class="item-info">
              <div class="item-name">{{ $t("environment.addEnvironment") }}</div>
              <div class="item-content">{{ $t("shop.addToShop") }}</div>
            </div>
          </router-link>
        </div>

        <element-empty
          v-else
          :route="{ name: 'account.environments.new', query: { shopId: shop.id } }"
          title=""
          :button="$t('environment.addEnvironment')"
        >
          {{ $t("shop.noEnvironmentsYet") }}
        </element-empty>
      </Panel>
    </div>
  </main-container>
</template>

<script setup lang="ts">
import ElementEmpty from "@/components/layout/ElementEmpty.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { computed, ref } from "vue";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

const loading = ref(true);
const { environments } = useAccountEnvironments();
const shops = ref<components["schemas"]["AccountShop"][]>([]);

// Group environments by shop
const shopEnvironments = computed(() => {
  const grouped: Record<number, typeof environments.value> = {};

  for (const env of environments.value) {
    const shopId = env.shopId;
    if (shopId) {
      grouped[shopId] ??= [];
      grouped[shopId].push(env);
    }
  }

  return grouped;
});

// Load shops
api
  .GET("/account/shops")
  .then((shopsRes) => {
    if (shopsRes.data) shops.value = shopsRes.data;
    loading.value = false;
  })
  .catch(() => {
    loading.value = false;
  });
</script>

<style scoped>
.shop-info {
  flex: 1;
}

.shop-description {
  margin: 0 0 0.5rem 0;
}

.shop-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.separator {
  opacity: 0.5;
}

.shop-panel {
  .item-info {
    padding-right: 1.75rem;
  }
}

.add-environment-card {
  border: 1px dashed var(--panel-border-color);
  min-height: 70px;

  .item-logo {
    color: var(--text-color-muted);
  }
}
</style>
