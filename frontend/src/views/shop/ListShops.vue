<template>
    <header-container title="My Shops">
        <router-link :to="{ name: 'account.shops.new' }" class="btn btn-primary">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Add Shop
        </router-link>
    </header-container>

    <main-container v-if="!shopStore.isLoading">
        <template v-if="shopStore.shops.length === 0">
            <element-empty title="No Shops" button="Add Shop" :route="{ name: 'account.shops.new' }">
                Get started by adding your first Shop.
            </element-empty>
        </template>

        <div v-else class="panel panel-table">
            <data-table
                :columns="[
                    { key: 'favicon', name: '', class: 'favicon-col' },
                    { key: 'name', name: 'Name', sortable: true },
                    { key: 'url', name: 'URL' },
                    { key: 'shopwareVersion', name: 'version', sortable: true },
                    { key: 'lastChangelog', name: 'Last Update', sortable: true, sortPath: 'lastUpdated.date' },
                    { key: 'organizationName', name: 'Organization', sortable: true },
                    { key: 'lastScrapedAt', name: 'Last checked at', sortable: true }
                ]"
                :data="shopStore.shops"
                :default-sorting="{by: 'name'}"
            >
                <template #cell-favicon="{ row }">
                    <img v-if="row.favicon" :src="row.favicon" alt="Shop Logo" class="favicon" />
                </template>

                <template #cell-name="{ row }">
                    <status-icon :status="row.status" />
                    <router-link
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                organizationId: row.organizationId,
                                shopId: row.id
                            }
                        }"
                        class="shop-name-link"
                    >
                        {{ row.name }}
                    </router-link>
                </template>

                <template #cell-url="{ row }">
                    <a :href="row.url" data-tooltip="Go to shopware storefront" target="_blank">
                        <icon-fa6-solid:store />
                    </a>
                    &nbsp;
                    <a :href="row.url + '/admin'" data-tooltip="Go to shopware admin" target="_blank">
                        <icon-fa6-solid:shield-halved />
                    </a>
                </template>

                <template #cell-lastChangelog="{ row }">
                    <template v-if="row.lastChangelog?.date">
                        <span :data-tooltip="'from ' + row.lastChangelog.from + ' to ' + row.lastChangelog.to">
                            {{ formatDate(row.lastChangelog.date) }}
                        </span>
                    </template>
                    <template v-else>
                        {{ null }}
                    </template>
                </template>

                <template #cell-lastScrapedAt="{ row }">
                    {{ formatDateTime(row.lastScrapedAt) }}
                </template>
            </data-table>
        </div>
    </main-container>
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from '@/helpers/formatter';

import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();

shopStore.loadShops();
</script>

<style>
.data-table td {
    &.favicon-col {
        width: 2.75rem;
        min-width: 2.75rem;
        padding: 0.875rem 0.75rem;
        
        .favicon {
            width: 1.25rem;
            height: 1.25rem;
            display: inline-block;
            vertical-align: middle;
        }
    }
}
</style>
