<script setup lang="ts">
import { trpcClient } from '@/helpers/trpc';
import HeaderContainer from '@/components/layout/HeaderContainer.vue';

import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

import { formatDate, formatDateTime } from '@/helpers/formatter';

const shops = await trpcClient.account.currentUserShops.query();
</script>

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
    <main-container>
        <div
            v-if="shops.length === 0"
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
                    { key: 'favicon', name: '', classOverride: true, class: 'w-11 min-w-[44px] py-3.5 px-3'},
                    { key: 'name', name: 'Name', sortable: true},
                    { key: 'url', name: 'URL'},
                    { key: 'shopwareVersion', name: 'Version', sortable: true},
                    { key: 'lastUpdated', name: 'Last update', sortable: true, sortPath: 'lastUpdated.date'},
                    { key: 'organizationName', name: 'Organization', sortable: true},
                    { key: 'lastScrapedAt', name: 'last checked at', sortable: true}
                ]"
                :data="shops"
                :sorting="{sortBy: 'name', sortDesc: false}"
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

                <template #cell-lastUpdated="{ row }">
                    <p
                        v-if="row.lastUpdated?.date"
                        :data-tooltip="'from ' + row.lastUpdated.from + ' to ' + row.lastUpdated.to"
                    >
                        {{ formatDate(row.lastUpdated.date) }}
                    </p>
                    <p v-else>
                        {{ null }}
                    </p>
                </template>

                <template #cell-lastScrapedAt="{ row }">
                    {{ formatDateTime(row.lastScrapedAt) }}
                </template>
            </data-table>
        </div>
    </main-container>
</template>
