<template>
    <header-container
        v-if="shopStore.shop"
        :title="shopStore.shop.name"
    >
        <div class="flex gap-2">
            <button
                class="btn"
                data-tooltip="Clear shop cache"
                :disabled="shopStore.isCacheClearing"
                type="button"
                @click="onCacheClear"
            >
                <icon-ic:baseline-cleaning-services
                    :class="{ 'animate-pulse': shopStore.isCacheClearing }"
                    class="w-4 h-4"
                />
            </button>
            <button
                class="btn"
                data-tooltip="Refresh shop data"
                :disabled="shopStore.isRefreshing"
                type="button"
                @click="showShopRefreshModal = true"
            >
                <icon-fa6-solid:rotate :class="{ 'animate-spin': shopStore.isRefreshing }" />
            </button>

            <router-link
                :to="{
                    name: 'account.shops.edit',
                    params: {
                        organizationId: shopStore.shop.organizationId,
                        shopId: shopStore.shop.id
                    }
                }"
                type="button"
                class="group btn btn-primary flex items-center"
            >
                <icon-fa6-solid:pencil
                    class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
                    aria-hidden="true"
                />
                Edit Shop
            </router-link>
        </div>
    </header-container>

    <main-container v-if="shopStore.shop && shopStore.shop.lastScrapedAt">
        <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg dark:bg-neutral-800 dark:shadow-none">
            <div class="py-5 px-4 sm:px-6 lg:px-8">
                <h3 class="text-lg leading-6 font-medium">
                    <icon-fa6-solid:circle-xmark
                        v-if="shopStore.shop.status == 'red'"
                        class="text-red-600 mr-1 dark:text-red-400 "
                    />
                    <icon-fa6-solid:circle-info
                        v-else-if="shopStore.shop.status === 'yellow'"
                        class="text-yellow-400 mr-1 dark:text-yellow-200"
                    />
                    <icon-fa6-solid:circle-check
                        v-else
                        class="text-green-400 mr-1 dark:text-green-300"
                    />
                    Shop Information
                </h3>
            </div>
            <div
                class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8 dark:border-neutral-700
                 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3"
            >
                <dl class="grid grid-cols-1 auto-rows-min gap-x-6 gap-y-2 md:col-span-2 md:grid-cols-2">
                    <div class="md:col-span-1">
                        <dt class="text-sm font-medium">
                            Shopware Version
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shopStore.shop.shopwareVersion }}
                            <template
                                v-if="latestShopwareVersion && latestShopwareVersion != shopStore.shop.shopwareVersion"
                            >
                                <a
                                    class="badge badge-warning"
                                    :href="'https://github.com/shopware/platform/releases/tag/v' + latestShopwareVersion"
                                    target="_blank"
                                >
                                    {{ latestShopwareVersion }}
                                </a>
                                <button
                                    class="ml-2 badge badge-info"
                                    type="button"
                                    @click="openUpdateWizard"
                                >
                                    <icon-fa6-solid:rotate />
                                    Compatibility Check
                                </button>
                            </template>
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Last Shop Update
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            <template v-if="shopStore.shop.lastChangelog && shopStore.shop.lastChangelog.date">
                                {{ formatDate(shopStore.shop.lastChangelog.date) }}
                            </template>
                            <template v-else>
                                never
                            </template>
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Organization
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shopStore.shop.organizationName }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Last Checked At
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ formatDateTime(shopStore.shop.lastScrapedAt) }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Environment
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shopStore.shop.cacheInfo?.environment }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            HTTP Cache
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shopStore.shop.cacheInfo?.httpCache ? 'Enabled' : 'Disabled' }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            URL
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            <a
                                :href="shopStore.shop.url"
                                data-tooltip="Go to storefront"
                                target="_blank"
                            >
                                <icon-fa6-solid:store /> Storefront
                            </a>
                            &nbsp;/&nbsp;
                            <a
                                :href="shopStore.shop.url + '/admin'"
                                data-tooltip="Go to shopware admin"
                                target="_blank"
                            >
                                <icon-fa6-solid:shield-halved /> Admin
                            </a>
                        </dd>
                    </div>
                </dl>
                <div
                    class="mt-6 sm:mt-0 sm:col-span-1 sm:row-span-full sm:col-start-2 sm:row-start-1
                     md:col-start-3 md:row-span-full md:row-start-1"
                >
                    <img
                        v-if="shopStore.shop.shopImage"
                        :src="`/${shopStore.shop.shopImage}`"
                        class="h-[200px] sm:h-[400px] md:h-[200px]"
                    >
                    <icon-fa6-solid:image
                        v-else
                        class="text-gray-100 text-9xl"
                    />
                </div>
            </div>
        </div>

        <tabs
            :labels="[
                { key: 'checks', title: 'Checks', icon: FaCircleCheck },
                { key: 'extensions', title: 'Extensions', icon: FaPlug },
                { key: 'tasks', title: 'Scheduled Tasks', icon: FaListCheck },
                { key: 'queue', title: 'Queue', icon: FaCircleCheck },
                { key: 'pagespeed', title: 'Pagespeed', icon: FaRocket },
                { key: 'changelog', title: 'Changelog', icon: FaFileWaverform }
            ]"
        >
            <template #panel-checks>
                <data-table
                    :columns="[
                        { key: 'message', name: 'Message' },
                    ]"
                    :data="shopStore.shop.checks || []"
                >
                    <template #cell-message="{ row }">
                        <icon-fa6-solid:circle-xmark
                            v-if="row.level == 'red'"
                            class="text-red-600 mr-2 text-base dark:text-red-400 "
                        />
                        <icon-fa6-solid:circle-info
                            v-else-if="row.level === 'yellow'"
                            class="text-yellow-400 mr-2 text-base dark:text-yellow-200 "
                        />
                        <icon-fa6-solid:circle-check
                            v-else
                            class="text-green-400 mr-2 text-base dark:text-green-300"
                        />
                        <a
                            v-if="row.link"
                            :href="row.link"
                            target="_blank"
                        >
                            {{ row.message }}
                            <icon-fa6-solid:up-right-from-square class="text-xs" />
                        </a>
                        <template v-else>
                            {{ row.message }}
                        </template>
                    </template>
                    <template #cell-actions="{ row }">
                        <button
                            v-if="shopStore.shop.ignores.includes(row.id)"
                            data-tooltip="check is ignored"
                            class="text-red-600 opacity-25 tooltip-position-left dark:text-red-400 group-hover:opacity-100"
                            type="button"
                            @click="removeIgnore(row.id)"
                        >
                            <icon-fa6-solid:eye-slash />
                        </button>
                        <button
                            v-else
                            data-tooltip="check used"
                            class="opacity-25 tooltip-position-left group-hover:opacity-100"
                            type="button"
                            @click="ignoreCheck(row.id)"
                        >
                            <icon-fa6-solid:eye />
                        </button>
                    </template>
                </data-table>
            </template>

            <template #panel-extensions>
                <data-table
                    :columns="[
                        { key: 'label', name: 'Name', sortable: true },
                        { key: 'version', name: 'Version' },
                        { key: 'latestVersion', name: 'Latest' },
                        { key: 'ratingAverage', name: 'Rating', sortable: true },
                        { key: 'installedAt', name: 'Installed at', sortable: true },
                    ]"
                    :data="shopStore.shop.extensions || []"
                    :default-sorting="{ by: 'label' }"
                >
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
            </template>

            <template #panel-tasks>
                <data-table
                    :columns="[
                        { key: 'name', name: 'Name', sortable: true },
                        { key: 'interval', name: 'Interval' },
                        { key: 'lastExecutionTime', name: 'Last Executed', sortable: true },
                        { key: 'nextExecutionTime', name: 'Next Execution', sortable: true },
                        { key: 'status', name: 'Status' },
                    ]"
                    :data="shopStore.shop.scheduledTask || []"
                    :default-sorting="{ by: 'nextExecutionTime' }"
                >
                    <template #cell-lastExecutionTime="{ row }">
                        {{ formatDateTime(row.lastExecutionTime) }}
                    </template>

                    <template #cell-nextExecutionTime="{ row }">
                        {{ formatDateTime(row.nextExecutionTime) }}
                    </template>

                    <template #cell-status="{ row }">
                        <span
                            v-if="row.status === 'scheduled' && !row.overdue"
                            class="pill pill-success"
                        >
                            <icon-fa6-solid:check />
                            {{ row.status }}
                        </span>
                        <span
                            v-else-if="row.status === 'queued' || row.status === 'running' && !row.overdue"
                            class="pill pill-info"
                        >
                            <icon-fa6-solid:rotate />
                            {{ row.status }}
                        </span>
                        <span
                            v-else-if="
                                row.status === 'scheduled' ||
                                    row.status === 'queued' ||
                                    row.status === 'running' &&
                                    row.overdue"
                            class="pill pill-warning"
                        >
                            <icon-fa6-solid:info class="align-[-0.1em]" />
                            {{ row.status }}
                        </span>
                        <span
                            v-else-if="row.status === 'inactive'"
                            class="pill"
                        >
                            <icon-fa6-solid:pause />
                            {{ row.status }}
                        </span>
                        <span
                            v-else
                            class="pill pill-error"
                        >
                            <icon-fa6-solid:xmark />
                            {{ row.status }}
                        </span>
                    </template>

                    <template #cell-actions="{ row }">
                        <span
                            class="cursor-pointer opacity-25 hover:opacity-100 tooltip-position-left"
                            data-tooltip="Re-schedule task"
                            @click="onReScheduleTask(row.id)"
                        >
                            <icon-fa6-solid:arrow-rotate-right class="text-base leading-3" />
                        </span>
                    </template>
                </data-table>
            </template>

            <template #panel-queue>
                <data-table
                    :columns="[
                        { key: 'name', name: 'Name', sortable: true },
                        { key: 'size', name: 'Size', sortable: true },
                    ]"
                    :data="shopStore.shop.queueInfo || []"
                />
            </template>

            <template #panel-pagespeed>
                <data-table
                    :columns="[
                        { key: 'createdAt', name: 'Checked At' },
                        { key: 'performance', name: 'Performance' },
                        { key: 'accessibility', name: 'Accessibility' },
                        { key: 'bestPractices', name: 'Best Practices' },
                        { key: 'seo', name: 'SEO' },
                    ]"
                    :data="shopStore.shop.pageSpeed"
                >
                    <template #cell-createdAt="{ row }">
                        <a
                            target="_blank"
                            :href="'https://pagespeed.web.dev/analysis?url=' + shopStore.shop.url"
                        >{{
                            formatDateTime(row.createdAt) }}</a>
                    </template>

                    <template
                        v-for="(cell, cellKey) in {
                            'performance': 'cell-performance',
                            'accessibility': 'cell-accessibility',
                            'bestPractices': 'cell-bestPractices',
                            'seo': 'cell-seo'
                        } as const"
                        #[cell]="{ row, rowIndex }"
                        :key="cellKey"
                    >
                        <span class="mr-2">{{ row[cellKey] }}</span>
                        <template v-if="shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] !== row[cellKey]">
                            <icon-fa6-solid:arrow-right
                                :class="[{
                                    'text-green-400 -rotate-45 dark:text-green-300':
                                        shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] < row[cellKey],
                                    'text-red-600 rotate-45 dark:text-red-400':
                                        shopStore.shop.pageSpeed[(rowIndex + 1)][cellKey] > row[cellKey]
                                }]"
                            />
                        </template>
                        <icon-fa6-solid:minus v-else />
                    </template>
                </data-table>
            </template>

            <template #panel-changelog>
                <data-table
                    :columns="[
                        { key: 'date', name: 'Date', sortable: true },
                    ]"
                    :data="shopStore.shop.changelog"
                >
                    <template #cell-date="{ row }">
                        {{ formatDateTime(row.date) }}
                    </template>

                    <template #cell-actions="{ row }">
                        <span
                            class="cursor-pointer"
                            @click="openShopChangelog(row)"
                        >
                            {{ sumChanges(row) }}
                        </span>
                    </template>
                </data-table>
            </template>
        </tabs>

        <modal
            :show="viewChangelogDialog"
            close-x-mark
            @close="viewChangelogDialog = false"
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

        <modal
            :show="viewShopChangelogDialog"
            close-x-mark
            @close="viewShopChangelogDialog = false"
        >
            <template #title>
                Shop changelog - <span
                    v-if="dialogShopChangelog?.date"
                    class="font-normal"
                >{{
                    formatDateTime(dialogShopChangelog.date) }}</span>
            </template>
            <template #content>
                <template v-if="dialogShopChangelog?.old_shopware_version && dialogShopChangelog?.new_shopware_version">
                    Shopware update from <strong>{{ dialogShopChangelog.old_shopware_version }}</strong> to <strong>{{
                        dialogShopChangelog.new_shopware_version }}</strong>
                </template>

                <div
                    v-if="dialogShopChangelog?.extensions?.length"
                    class="mt-4"
                >
                    <h2 class="text-lg mb-1 font-medium">
                        Shop Plugin Changelog:
                    </h2>
                    <ul class="list-disc">
                        <li
                            v-for="extension in dialogShopChangelog?.extensions"
                            :key="extension.name"
                            class="ml-4 mb-1"
                        >
                            <div>
                                <strong>{{ extension.label }}</strong>
                                <span class="opacity-60">
                                    ({{ extension.name }})
                                </span>
                            </div>
                            {{ extension.state }} <template v-if="extension.state === 'installed' && extension.active">
                                and activated
                            </template>
                            <template v-if="extension.state === 'updated'">
                                from {{ extension.old_version }} to {{ extension.new_version }}
                            </template>
                            <template v-else>
                                {{ extension.new_version }}
                                <template v-if="!extension.new_version">
                                    {{ extension.old_version }}
                                </template>
                            </template>
                        </li>
                    </ul>
                </div>
            </template>
        </modal>

        <modal
            :show="viewUpdateWizardDialog"
            close-x-mark
            @close="viewUpdateWizardDialog = false"
        >
            <template #title>
                <icon-fa6-solid:rotate /> Shopware Plugin Compatibility Check
            </template>
            <template #content>
                <select
                    class="field mb-2"
                    @change="event => loadUpdateWizard((event.target as HTMLSelectElement).value)"
                >
                    <option
                        disabled
                        selected
                    >
                        Select update Version
                    </option>
                    <option
                        v-for="version in shopwareVersions"
                        :key="version"
                    >
                        {{ version }}
                    </option>
                </select>

                <template v-if="loadingUpdateWizard">
                    <div class="text-center">
                        Loading <icon-fa6-solid:rotate class="animate-spin" />
                    </div>
                </template>

                <div
                    v-if="dialogUpdateWizard"
                    :class="{ 'opacity-20': loadingUpdateWizard }"
                >
                    <h2 class="text-lg mb-1 font-medium">
                        Plugin Compatibility
                    </h2>

                    <ul>
                        <li
                            v-for="extension in dialogUpdateWizard"
                            :key="extension.name"
                            class="p-2 odd:bg-gray-100 dark:odd:bg-[#2e2e2e]"
                        >
                            <div class="flex">
                                <div class="mr-2 w-4">
                                    <icon-fa6-regular:circle
                                        v-if="!extension.active"
                                        class="text-base text-gray-400 dark:text-neutral-500"
                                    />
                                    <icon-fa6-solid:circle-info
                                        v-else-if="!extension.compatibility"
                                        class="text-base text-yellow-400 dark:text-yellow-200"
                                    />
                                    <icon-fa6-solid:circle-xmark
                                        v-else-if="extension.compatibility.type == 'red'"
                                        class="text-base text-red-600 dark:text-red-400"
                                    />
                                    <icon-fa6-solid:rotate
                                        v-else-if="extension.compatibility.label === 'Available now'"
                                        class="text-base text-sky-500 dark:text-sky-400"
                                    />
                                    <icon-fa6-solid:circle-check
                                        v-else
                                        class="text-base text-green-400 dark:text-green-300"
                                    />
                                </div>
                                <div>
                                    <strong>{{ extension.label }}</strong>
                                    <span class="opacity-60">({{ extension.name }})</span>
                                    <div v-if="!extension.compatibility">
                                        This plugin is not available in the Store. Please contact the
                                        plugin manufacturer.
                                    </div>
                                    <div v-else>
                                        {{ extension.compatibility.label }}
                                    </div>
                                </div>
                            </div>
                        </li>
                    </ul>
                </div>
            </template>
        </modal>

        <modal
            :show="showShopRefreshModal"
            close-x-mark
            @close="showShopRefreshModal = false"
        >
            <template #icon>
                <icon-fa6-solid:rotate
                    class="h-6 w-6 text-sky-500"
                    aria-hidden="true"
                />
            </template>
            <template #title>
                Refresh {{ shopStore.shop.name }}
            </template>
            <template #content>
                Do you also want to have a new pagespeed test?
            </template>
            <template #footer>
                <button
                    type="button"
                    class="btn w-full sm:ml-3 sm:w-auto"
                    @click="onRefresh(true)"
                >
                    Yes
                </button>
                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
                    @click="onRefresh(false)"
                >
                    No
                </button>
            </template>
        </modal>
    </main-container>
</template>

<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';
import Tabs from '@/components/layout/Tabs.vue';
import Modal from '@/components/layout/Modal.vue';
import RatingStars from '@/components/RatingStars.vue';

import { compareVersions } from 'compare-versions';
import { createNewSortInstance } from 'fast-sort';

import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';

import type { Extension, ExtensionCompatibilitys, ShopChangelog } from '@/types/shop';
import { ref } from 'vue';
import type { Ref } from 'vue';

import { sumChanges } from '@/helpers/changelog';
import { formatDate, formatDateTime } from '@/helpers/formatter';

import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaRocket from '~icons/fa6-solid/rocket';
import FaFileWaverform from '~icons/fa6-solid/file-waveform';
import { trpcClient } from '@/helpers/trpc';

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();

const viewChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);

const viewShopChangelogDialog: Ref<boolean> = ref(false);
const dialogShopChangelog: Ref<ShopChangelog | null> = ref(null);

const viewUpdateWizardDialog: Ref<boolean> = ref(false);
const loadingUpdateWizard: Ref<boolean> = ref(false);
const dialogUpdateWizard: Ref<ExtensionCompatibilitys[] | null> = ref(null);

const showShopRefreshModal: Ref<boolean> = ref(false);
const shopwareVersions: Ref<string[] | null> = ref(null);

const latestShopwareVersion: Ref<string | null> = ref(null);

function openExtensionChangelog(extension: Extension | null) {
    dialogExtension.value = extension;
    viewChangelogDialog.value = true;
}

function openShopChangelog(shopChangelog: ShopChangelog | null) {
    dialogShopChangelog.value = shopChangelog;
    viewShopChangelogDialog.value = true;
}

function openUpdateWizard() {
    dialogUpdateWizard.value = null;
    viewUpdateWizardDialog.value = true;
}

async function loadShop() {
    const organizationId = parseInt(route.params.organizationId as string, 10);
    const shopId = parseInt(route.params.shopId as string, 10);

    await shopStore.loadShop(organizationId, shopId);

    const shopwareVersionsData = await trpcClient.info.getLatestShopwareVersion.query();
    shopwareVersions.value = Object.keys(shopwareVersionsData).reverse()
        .filter((version) => !version.includes('-RC') && compareVersions(shopStore.shop!.shopwareVersion, version) < 0);
    latestShopwareVersion.value = shopwareVersions.value[0];
}

loadShop().then(function() {
    if (shopStore?.shop?.name) {
        document.title = shopStore.shop.name;
    }
});

async function onRefresh(pagespeed: boolean) {
    showShopRefreshModal.value = false;
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.refreshShop(shopStore.shop.organizationId, shopStore.shop.id, pagespeed);
            alertStore.success('Your Shop will refresh soon!');
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function onCacheClear() {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.clearCache(shopStore.shop.organizationId, shopStore.shop.id);
            alertStore.success('Your Shop cache was cleared successfully');
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function onReScheduleTask(taskId: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.reScheduleTask(shopStore.shop.organizationId, shopStore.shop.id, taskId);
            alertStore.success('Task is re-scheduled');
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function loadUpdateWizard(version: string) {
    if (!shopStore.shop || !shopStore.shop.extensions) {
        return;
    }

    loadingUpdateWizard.value = true;

    const body = {
        currentVersion: shopStore.shop!.shopwareVersion,
        futureVersion: version,
        extensions: shopStore.shop.extensions.map((extension) => {
            return {
                name: extension.name,
                version: extension.version,
            };
        }),
    };

    const pluginCompatibilitys = await trpcClient.info.checkExtensionCompatibility.query(body);

    const extensions = JSON.parse(JSON.stringify(shopStore.shop?.extensions));

    for (const extension of extensions) {
        const compatibility = pluginCompatibilitys.find((plugin) => plugin.name === extension.name);
        extension.compatibility = null;
        if (compatibility) {
            extension.compatibility = compatibility.status;
        }
    }

    const naturalSort = createNewSortInstance({
        comparer: new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' }).compare,
    });

    dialogUpdateWizard.value = naturalSort(extensions).by([{ desc: u => u.active }, { asc: u => u.label }]);

    loadingUpdateWizard.value = false;
}

async function ignoreCheck(id: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        shopStore?.shop?.ignores.push(id);

        shopStore.updateShop(shopStore.shop.organizationId, shopStore.shop.id, { ignores: shopStore?.shop?.ignores });
        notificateIgnoreUpdate();
    }
}

async function removeIgnore(id: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        shopStore.shop.ignores = shopStore.shop.ignores.filter((aid: string) => aid !== id);

        shopStore.updateShop(shopStore.shop.organizationId, shopStore.shop.id, { ignores: shopStore?.shop?.ignores });
        notificateIgnoreUpdate();
    }
}

async function notificateIgnoreUpdate() {
    alertStore.info('Ignore state updated. Will effect after next shop update');
}

</script>
