<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';
import { useDashboardStore } from '@/stores/dashboard.store';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';

import MainContainer from '@/components/layout/MainContainer.vue';

import { sumChanges } from '@/helpers/changelog';
import { formatDateTime } from '@/helpers/formatter';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const organizations = user.value?.organizations.map(organization => ({
    ...organization,
    initials: organization.name.substring(0, 2),
}));

const shopStore = useShopStore();
shopStore.loadShops();

const dashboardStore = useDashboardStore();
dashboardStore.loadChangelogs();
</script>

<template>
    <div v-if="user && !shopStore.isLoading">
        <header-container title="Dashboard" />
        <main-container>
            <h2 class="text-gray-500 text-lg font-medium pb-1 mb-4 border-b dark:border-neutral-800">
                <icon-fa6-solid:shop />
                My Shops's
            </h2>
            <ul
                role="list"
                class="grid grid-cols-1 gap-5 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-10"
            >
                <li
                    v-for="shop in shopStore.shops"
                    :key="shop.id"
                    class="col-span-1 shadow-sm rounded-md bg-white dark:shadow-none
                     dark:bg-neutral-800 hover:bg-sky-50 dark:hover:bg-[#2a2b2f]"
                >
                    <router-link
                        :to="{ 
                            name: 'account.shops.detail',
                            params: { 
                                organizationId: shop.organizationId, 
                                shopId: shop.id 
                            }
                        }"
                        class="flex"
                    >
                        <div
                            class="flex-shrink-0 flex items-center justify-center w-16 rounded-l-md bg-gray-100
                         bg-opacity-50 dark:bg-neutral-700 dark:bg-opacity-25"
                        >
                            <img
                                v-if="shop.favicon"
                                :src="shop.favicon"
                                alt="Shop Logo"
                                class="inline-block w-5 h-5 align-middle"
                            >
                        </div>
                        <div class="flex-1 items-center justify-between px-4 py-2 truncate">
                            <div class="text-gray-900 font-medium truncate dark:text-neutral-400">
                                {{ shop.name }}
                            </div>
                            <div class="text-gray-500">
                                {{ shop.organizationName }}<br>
                                {{ shop.shopwareVersion }}
                            </div>
                        </div>
                    </router-link>
                </li>
            </ul>

            <h2 class="text-gray-500 text-lg font-medium pb-1 mb-4 border-b dark:border-neutral-800">
                <icon-fa6-solid:people-group />
                My Organizations
            </h2>
            <ul
                role="list"
                class="grid grid-cols-1 gap-5 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-10"
            >
                <li
                    v-for="organization in organizations"
                    :key="organization.id"
                    class="col-span-1 shadow-sm rounded-md bg-white dark:shadow-none dark:bg-neutral-800
                    hover:bg-sky-50 dark:hover:bg-[#2a2b2f]"
                >
                    <router-link
                        :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
                        class="flex"
                    >
                        <div
                            class="flex-shrink-0 flex items-center justify-center w-16 rounded-l-md bg-gray-100
                         bg-opacity-50 dark:bg-neutral-700 dark:bg-opacity-25"
                        >
                            <icon-fa6-solid:people-group class="text-lg text-gray-900 dark:text-neutral-400" />
                        </div>
                        <div class="flex-1 items-center justify-between px-4 py-2 truncate">
                            <div class="text-gray-900 font-medium truncate dark:text-neutral-400">
                                {{ organization.name }}
                            </div>
                            <div class="text-gray-500">
                                {{ organization.memberCount }} Members, {{ organization.shopCount }} Shops
                            </div>
                        </div>
                    </router-link>
                </li>
            </ul>

            <template v-if="dashboardStore.changelogs.length > 0">
                <h2 class="text-gray-500 text-lg font-medium pb-1 mb-4 border-b dark:border-neutral-800">
                    <icon-fa6-solid:file-waveform />
                    Last Changes
                </h2>
                <div
                    class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden bg-white dark:bg-neutral-800
                 dark:shadow-none mb-10"
                >
                    <DataTable
                        :labels="{
                            name: {name: 'Shop'},
                            changes: {name: 'Changes'},
                            date: {name: 'Date'}
                        }"
                        :data="dashboardStore.changelogs"
                    >
                        <template #cell(name)="{ item }">
                            <router-link
                                :to="{ 
                                    name: 'account.shops.detail', 
                                    params: { 
                                        organizationId: item.organizationId,
                                        shopId: item.shopId
                                    }
                                }"
                            >
                                {{ item.shopName }}
                            </router-link>
                        </template>

                        <template #cell(changes)="{ item }">
                            {{ sumChanges(item) }}
                        </template>          

                        <template #cell(date)="{ item }">
                            {{ formatDateTime(item.date) }}
                        </template>
                    </DataTable>
                </div>
            </template>
        </main-container>
    </div>
</template>
