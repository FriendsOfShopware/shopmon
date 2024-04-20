<template>
    <data-table
        :columns="[
            { key: 'label', name: 'Name', sortable: true },
            { key: 'version', name: 'Version' },
            { key: 'latestVersion', name: 'Latest' },
            { key: 'ratingAverage', name: 'Rating', sortable: true },
            { key: 'installedAt', name: 'Installed at', sortable: true },
        ]"
        :data="shopStore.shop.extensions || []"
        :default-sort="{ key: 'label', desc: false }"
    >
        <template #cell-actions-header>
            Known Issues
        </template>

        <template #cell-label="{ row }">
            <div class="flex items-start">
                <span
                    v-if="!row.installed"
                    class="leading-5 text-gray-400 mr-2 text-base dark:text-neutral-500"
                    data-tooltip="Not installed"
                >
                    <icon-fa6-regular:circle />
                </span>
                <span
                    v-else-if="row.active"
                    class="leading-5 text-green-400 mr-2 text-base dark:text-green-300"
                    data-tooltip="Active"
                >
                    <icon-fa6-solid:circle-check />
                </span>
                <span
                    v-else
                    class="leading-5 text-gray-300 mr-2 text-base dark:text-neutral-500"
                    data-tooltip="Inactive"
                >
                    <icon-fa6-solid:circle-xmark />
                </span>

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
            </div>
        </template>

        <template #cell-version="{ row }">
            {{ row.version }}
            <span
                v-if="row.latestVersion && row.version < row.latestVersion"
                data-tooltip="Update available"
                @click="openExtensionChangelog(row as Extension)"
            >
                <icon-fa6-solid:rotate
                    class="ml-1 text-base text-amber-600 dark:text-amber-400 cursor-pointer"
                />
            </span>
        </template>

        <template #cell-ratingAverage="{ row }">
            <rating-stars :rating="row.ratingAverage" />
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

    <modal
        :show="viewExtensionChangelogDialog"
        close-x-mark
        @close="viewExtensionChangelogDialog = false"
    >
        <template #title>
            Changelog - <span class="font-normal">{{ dialogExtension?.name }}</span>
        </template>
        <template #content>
            <ul v-if="dialogExtension?.changelog?.length ?? 0 > 0">
                <li
                    v-for="changeLog in dialogExtension.changelog"
                    :key="changeLog.version"
                    class="mb-2"
                >
                    <div class="font-medium mb-1">
                            <span
                                v-if="!changeLog.isCompatible"
                                data-tooltip="not compatible with your version"
                            >
                                <icon-fa6-solid:circle-info class="text-yellow-400 text-base dark:text-yellow-200" />
                            </span>
                        {{ changeLog.version }} -
                        <span class="text-xs font-normal text-gray-500">
                                {{ formatDate(changeLog.creationDate) }}
                            </span>
                    </div>
                    <!-- eslint-disable vue/no-v-html -->
                    <div
                        class="pl-3"
                        v-html="changeLog.text"
                    />
                    <!-- eslint-enable vue/no-v-html -->
                </li>
            </ul>
            <div
                v-else
                class="rounded-md bg-red-50 p-4 border border-red-200"
            >
                <div class="flex">
                    <div class="flex-shrink-0">
                        <icon-fa6-solid:circle-xmark
                            class="h-5 w-5 text-red-600 dark:text-red-400 "
                            aria-hidden="true"
                        />
                    </div>
                    <div class="ml-3 text-red-900">
                        No Changelog data provided
                    </div>
                </div>
            </div>
        </template>
    </modal>
</template>

<script setup lang="ts">
import type { Extension } from "@/types/shop";
import { formatDate, formatDateTime } from "@/helpers/formatter";
import { useShopStore } from '@/stores/shop.store';
import { ref, Ref } from "vue";

const shopStore = useShopStore();

const viewExtensionChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);

function openExtensionChangelog(extension: Extension | null) {
    dialogExtension.value = extension;
    viewExtensionChangelogDialog.value = true;
}
</script>
