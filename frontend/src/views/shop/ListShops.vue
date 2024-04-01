<template>
    <header-container title="My Shops">
        <router-link to="/account/shops/new" class="btn btn-primary">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Add Shop
        </router-link>
    </header-container>
    <main-container v-if="!shopStore.isLoading">
        <div v-if="shopStore.shops.length === 0" class="shops-empty">
            <svg class="shops-empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
            </svg>

            <h3 class="shops-empty-title">No shops</h3>

            <p class="shops-empty-description">Get started by adding your first Shop.</p>
            
            <div class="shops-empty-cta">
                <router-link to="/account/shops/new" class="btn btn-primary">
                    <icon-fa6-solid:plus class="icon" aria-hidden="true" />
                    Add Shop
                </router-link>
            </div>
        </div>

        <div v-else class="shops-table-container">
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
                    <icon-fa6-solid:circle-xmark v-if="row.status == 'red'" class="icon-error" />
                    <icon-fa6-solid:circle-info v-else-if="row.status === 'yellow'" class="icon-warning" />
                    <icon-fa6-solid:circle-check v-else class="icon-success" />
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

import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();

shopStore.loadShops();
</script>

<style>

.shops-empty {
    text-align: center;
    
    &-icon {
        width: 3rem;
        height: 3rem;
        margin: 0 auto 0.5rem;
    }

    &-title {
        font-size: 1.25rem;
        line-height: 1.75rem;
        font-weight: 500;
    }

    &-description {
        margin-top: 0.25rem;
    }

    &-cta {
        margin-top: 1.5rem;
    }
}

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

    [class^="icon"] {
        font-size: 1rem;
        margin-right: .5rem;
    }
}
</style>
