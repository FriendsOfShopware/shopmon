<template>
    <header-container title="My Shops">
        <router-link
            to="/account/shops/new"
            type="button"
            class="group btn btn-primary flex items-center align-middle"
        >
            <icon-fa6-solid:plus
                class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50"
                aria-hidden="true"
            />
            Add Shop
        </router-link>
    </header-container>
    <main-container v-if="!shopStore.isLoading">
        <div
            v-if="shopStore.shops.length === 0"
            class="text-center"
        >
            <svg
                class="mx-auto h-12 w-12 text-gray-400 dark:text-neutral-500"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
            >
                <path
                    vector-effect="non-scaling-stroke"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                />
            </svg>
            <h3 class="mt-2 font-medium">
                No shops
            </h3>
            <p class="mt-1 text-gray-500 dark:text-neutral-500">
                Get started by adding your first Shop.
            </p>
            <div class="mt-6">
                <router-link
                    to="/account/shops/new"
                    class="btn btn-primary group flex items-center"
                >
                    <icon-fa6-solid:plus
                        class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50"
                        aria-hidden="true"
                    />
                    Add Shop
                </router-link>
            </div>
        </div>

        <div
            v-else
            class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden bg-white dark:bg-neutral-800 dark:shadow-none"
        >
            <data-table
                :columns="[
                    { key: 'favicon', name: 'Logo' },
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
                    <img
                        v-if="row.favicon"
                        :src="row.favicon"
                        alt="Shop Logo"
                        class="inline-block w-5 h-5 align-middle"
                    >
                </template>

                <template #cell-name="{ row }">
                    <icon-fa6-solid:circle-xmark
                        v-if="row.status == 'red'"
                        class="text-red-600 mr-2 text-base dark:text-red-400 "
                    />
                    <icon-fa6-solid:circle-info
                        v-else-if="row.status === 'yellow'"
                        class="text-yellow-400 mr-2 text-base dark:text-yellow-200"
                    />
                    <icon-fa6-solid:circle-check
                        v-else
                        class="text-green-400 mr-2 text-base dark:text-green-300"
                    />
                    <router-link
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                organizationId: row.organizationId,
                                shopId: row.id
                            }
                        }"
                    >
                        {{ row.name }}
                    </router-link>
                </template>

                <template #cell-url="{ row }">
                    <a
                        :href="row.url"
                        data-tooltip="Go to shopware storefront"
                        target="_blank"
                    >
                        <icon-fa6-solid:store />
                    </a>
          &nbsp;
                    <a
                        :href="row.url + '/admin'"
                        data-tooltip="Go to shopware admin"
                        target="_blank"
                    >
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
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

import { formatDate, formatDateTime } from '@/helpers/formatter';

import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();

shopStore.loadShops();
</script>
