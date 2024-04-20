<template>
    <header-container title="My Apps" />

    <main-container v-if="!extensionStore.isLoading">
        <div
            v-if="extensionStore.extensions.length === 0"
            class="text-center"
        >
            <shops-empty />
        </div>

        <template v-else>
            <input
                v-model="term"
                class="field field-search mb-3"
                placeholder="Search ..."
            >

            <div
                class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden bg-white
            dark:bg-neutral-800 dark:shadow-none"
            >
                <data-table
                    :columns="[
                        { key: 'label', name: 'Name', sortable: true, searchable: true},
                        { key: 'shops', name: 'Shop' },
                        { key: 'version', name: 'Version' },
                        { key: 'latestVersion', name: 'Latest' },
                        { key: 'ratingAverage', name: 'Rating', sortable: true },
                        { key: 'installedAt', name: 'Installed At', sortable: true },
                    ]"
                    :data="extensionStore.extensions"
                    :default-sort="{ key: 'label', desc: false }"
                    :search-term="term"
                >
                    <template #cell-actions-header>
                        Known Issues
                    </template>
                    <template #cell-label="{ row }">
                        <div v-if="row.storeLink">
                            <a
                                :href="row.storeLink"
                                target="_blank"
                            >
                                <div
                                    v-if="row.label"
                                    class="font-bold whitespace-normal"
                                >{{ row.label }}</div>
                                <div class="dark:opacity-50">{{ row.name }}</div>
                            </a>
                        </div>
                        <div v-else>
                            <div
                                v-if="row.label"
                                class="font-bold whitespace-normal"
                            >
                                {{ row.label }}
                            </div>
                            <div class="dark:opacity-50">
                                {{ row.name }}
                            </div>
                        </div>
                    </template>

                    <template #cell-shops="{ row }">
                        <div
                            v-for="(shop, rowIndex) in row.shops"
                            :key="rowIndex"
                            class="mb-1"
                        >
                            <router-link
                                :to="{
                                    name: 'account.shops.detail',
                                    params: {
                                        organizationId: shop.organizationId,
                                        shopId: shop.id
                                    }
                                }"
                            >
                                <span
                                    v-if="!shop.installed"
                                    class="leading-5 text-gray-400 mr-1 text-base dark:text-neutral-500"
                                    data-tooltip="Not installed"
                                >
                                    <icon-fa6-regular:circle />
                                </span>
                                <span
                                    v-else-if="shop.active"
                                    class="leading-5 text-green-400 mr-1 text-base dark:text-green-300"
                                    data-tooltip="Active"
                                >
                                    <icon-fa6-solid:circle-check />
                                </span>
                                <span
                                    v-else
                                    class="leading-5 text-gray-300 mr-1 text-base dark:text-neutral-500"
                                    data-tooltip="Inactive"
                                >
                                    <icon-fa6-solid:circle-xmark />
                                </span>
                                {{ shop.name }}
                            </router-link>
                        </div>
                    </template>

                    <template #cell-version="{ row }">
                        <div
                            v-for="(shop, rowIndex) in row.shops"
                            :key="rowIndex"
                            class="mb-1"
                        >
                            {{ shop.version }}
                            <span
                                v-if="shop.latestVersion && shop.version < shop.latestVersion"
                                data-tooltip="Update available"
                            >
                                <icon-fa6-solid:rotate class="ml-1 text-base text-amber-600 dark:text-amber-400" />
                            </span>
                        </div>
                    </template>

                    <template #cell-ratingAverage="{ row }">
                        <RatingStars :rating="row.ratingAverage" />
                    </template>

                    <template #cell-installedAt="{ row }">
                        <template v-if="row.installedAt">
                            {{ formatDateTime(row.installedAt) }}
                        </template>
                    </template>

                    <template #cell-actions>
                        No known issues. <a href="#">Report issue</a>
                    </template>
                </data-table>
            </div>
        </template>
    </main-container>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useExtensionStore } from '@/stores/extension.store';

const extensionStore = useExtensionStore();

const term = ref('');

extensionStore.loadExtensions();

import { formatDateTime } from '@/helpers/formatter';
</script>
