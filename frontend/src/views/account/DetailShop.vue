<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';

import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';
import Tabs from '@/components/layout/Tabs.vue';
import Modal from '@/components/layout/Modal.vue';
import RatingStars from '@/components/layout/RatingStars.vue';

import { compareVersions } from 'compare-versions';
import { createNewSortInstance } from 'fast-sort';

import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';

import type { ShopChangelog } from '@/types/shop';
import { ref } from 'vue';

import { sumChanges } from '@/helpers/changelog';
import { formatDate, formatDateTime } from '@/helpers/formatter';

import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaRocket from '~icons/fa6-solid/rocket';
import FaFileWaverform from '~icons/fa6-solid/file-waveform';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

type Extension = RouterOutput['account']['currentUserApps'][0];
type ExtensionCompatibilitys = RouterOutput['info']['checkExtensionCompatibility'][0];

const route = useRoute();
const alertStore = useAlertStore();

const viewChangelogDialog = ref(false);
const dialogExtension = ref<Extension | null>(null);

const viewShopChangelogDialog = ref(false);
const dialogShopChangelog = ref<ShopChangelog | null>(null);

const viewUpdateWizardDialog = ref(false);
const loadingUpdateWizard = ref(false);
const dialogUpdateWizard = ref<ExtensionCompatibilitys[] | null>(null);

const showShopRefreshModal = ref(false);
const shopwareVersions = ref<string[] | null>(null);

const latestShopwareVersion = ref<string | null>(null);

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

const organizationId = parseInt(route.params.organizationId as string, 10);
const shopId = parseInt(route.params.shopId as string, 10);

const shop = ref(await trpcClient.organization.shop.get.query({ orgId: organizationId, shopId }));

const shopwareVersionsData = await trpcClient.info.getLatestShopwareVersion.query();
shopwareVersions.value = Object.keys(shopwareVersionsData).reverse()
    .filter((version) => compareVersions(shop.value.shopware_version, version) < 0);
latestShopwareVersion.value = shopwareVersions.value[0];

document.title = shop.value.name;


const isRefreshing = ref(false);
async function onRefresh(pageSpeed: boolean) {
    showShopRefreshModal.value = false;
    isRefreshing.value = true;
    try {
        await trpcClient.organization.shop.refreshShop.mutate({ 
            orgId: organizationId,
            shopId, 
            pageSpeed,
        });
        alertStore.success('Your Shop will refresh soon!');
    } catch (e: any) {
        alertStore.error(e);
    } finally {
        isRefreshing.value = false;
    }
}

const isCacheClearing = ref(false);
async function onCacheClear() {
    isCacheClearing.value = true;
    try {
        await trpcClient.organization.shop.clearShopCache.mutate({ orgId: organizationId, shopId });
        alertStore.success('Your Shop cache was cleared successfully');
    } catch (e: any) {
        alertStore.error(e);
    } finally {
        isCacheClearing.value = false;
    }
}

const isReSchedulingTask = ref(false);
async function onReScheduleTask(taskId: string) {
    isReSchedulingTask.value = true;
    try {
        await trpcClient.organization.shop.rescheduleTask.mutate({ orgId: organizationId, shopId, taskId });
        shop.value = await trpcClient.organization.shop.get.query({ orgId: organizationId, shopId });

        alertStore.success('Task is re-scheduled');
    } catch (e: any) {
        alertStore.error(e);
    } finally {
        isReSchedulingTask.value = false;
    }
}

async function loadUpdateWizard(version: string) {
    loadingUpdateWizard.value = true;

    const body = {
        currentVersion: shop.value.shopware_version,
        futureVersion: version,
        extensions: shop.value.extensions?.map(extension => ({
            name: extension.name,
            version: extension.version,
        })) || [],
    };

    const pluginCompatibilitys = await trpcClient.info.checkExtensionCompatibility.query(body);

    const extensions = JSON.parse(JSON.stringify(shop.value.extensions));

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

async function updateShop(
    orgId: number,
    shopId: number, 
    payload: {
     name?: string,
     ignores?: string[],
     shopUrl?: string,
     clientId?: string,
     clientSecret?: string
    }) {
    if (payload.shopUrl) {
        payload.shopUrl = payload.shopUrl.replace(/\/+$/, '');
    }
    await trpcClient.organization.shop.update.mutate({ orgId, shopId, ...payload });
    shop.value = await trpcClient.organization.shop.get.query({ orgId, shopId });
}


async function ignoreCheck(id: string) {
    shop.value?.ignores.push(id);

    await updateShop(organizationId, shopId, { ignores: shop.value.ignores || '' });
    notificateIgnoreUpdate();
}

async function removeIgnore(id: string) {
    shop.value.ignores = shop.value.ignores.filter((aid: string) => aid !== id);

    await updateShop(organizationId, shopId, { ignores: shop.value?.ignores || '' });
    notificateIgnoreUpdate();
}

async function notificateIgnoreUpdate() {
    alertStore.info('Ignore state updated. Will effect after next shop update');
}

</script>

<template>
    <header-container
        :title="shop.name"
    >
        <div class="flex gap-2">
            <button
                type="button"
                class="btn"
                data-tooltip="Clear shop cache"
                :disabled="isCacheClearing"
                @click="onCacheClear"
            >
                <icon-ic:baseline-cleaning-services
                    :class="{ 'animate-pulse': isCacheClearing }"
                    class="w-4 h-4"
                />
            </button>
            <button
                class="btn"
                type="button"
                data-tooltip="Refresh shop data"
                :disabled="isRefreshing"
                @click="showShopRefreshModal = true"
            >
                <icon-fa6-solid:rotate :class="{ 'animate-spin': isRefreshing }" />
            </button>

            <router-link
                :to="{ 
                    name: 'account.shops.edit',
                    params: { 
                        organizationId: shop.organizationId,
                        shopId: shop.id
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

    <main-container v-if="shop.lastScrapedAt">
        <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg dark:bg-neutral-800 dark:shadow-none">
            <div class="py-5 px-4 sm:px-6 lg:px-8">
                <h3 class="text-lg leading-6 font-medium">
                    <icon-fa6-solid:circle-xmark
                        v-if="shop.status == 'red'"
                        class="text-red-600 mr-1 dark:text-red-400 "
                    />
                    <icon-fa6-solid:circle-info
                        v-else-if="shop.status === 'yellow'"
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
                            {{ shop.shopware_version }}
                            <template
                                v-if="latestShopwareVersion &&
                                    latestShopwareVersion != shop.shopware_version"
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
                            <template v-if="shop.lastUpdated">
                                {{ formatDate(shop.lastUpdated) }}
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
                            {{ shop.organizationName }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Last Checked At
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ formatDateTime(shop.lastScrapedAt) }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            Environment
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shop.cacheInfo?.environment }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            HTTP Cache
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ shop.cacheInfo?.httpCache ? 'Enabled' : 'Disabled' }}
                        </dd>
                    </div>
                    <div class="md:col-span-1">
                        <dt class="font-medium">
                            URL
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            <a
                                :href="shop.url"
                                data-tooltip="Go to storefront"
                                target="_blank"
                            >
                                <icon-fa6-solid:store /> Storefront
                            </a>
                            &nbsp;/&nbsp;
                            <a
                                :href="shop.url + '/admin'"
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
                        v-if="shop.shopImage"
                        :src="`/api/organization/${shop.shopImage}`"
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
            :labels="{
                checks: { title: 'Checks', count: shop.checks?.length ?? 0, icon: FaCircleCheck },
                extensions: { title: 'Extensions', count: shop.extensions?.length ?? 0, icon: FaPlug },
                tasks: { title: 'Scheduled Tasks', count: shop.scheduledTask?.length ?? 0, icon: FaListCheck },
                queue: { title: 'Queue', count: shop.queueInfo?.length ?? 0, icon: FaCircleCheck },
                pagespeed: { title: 'Pagespeed', count: shop.pageSpeed.length, icon: FaRocket },
                changelog: { title: 'Changelog', count: shop.changelog.length, icon: FaFileWaverform }
            }"
        >
            <template #panel(checks)="{ label }">
                <data-table
                    :labels="{ message: { name: 'Message' }, actions: { name: 'Ignore', class: 'px-3 text-right' } }"
                    :data="shop.checks"
                >
                    <template #cell(message)="{ item }">
                        <icon-fa6-solid:circle-xmark
                            v-if="item.level == 'red'"
                            class="text-red-600 mr-2 text-base dark:text-red-400 "
                        />
                        <icon-fa6-solid:circle-info
                            v-else-if="item.level === 'yellow'"
                            class="text-yellow-400 mr-2 text-base dark:text-yellow-200 "
                        />
                        <icon-fa6-solid:circle-check
                            v-else
                            class="text-green-400 mr-2 text-base dark:text-green-300"
                        />
                        <a
                            v-if="item.link"
                            :href="item.link"
                            target="_blank"
                        >
                            {{ item.message }}
                            <icon-fa6-solid:up-right-from-square class="text-xs" />
                        </a>
                        <template v-else>
                            {{ item.message }}
                        </template>
                    </template>
                    <template #cell(actions)="{ item }">
                        <button
                            v-if="shop.ignores.includes(item.id)"
                            data-tooltip="check is ignored"
                            class="text-red-600 opacity-25 tooltip-position-left dark:text-red-400 group-hover:opacity-100"
                            type="button"
                            @click="removeIgnore(item.id)"
                        >
                            <icon-fa6-solid:eye-slash />
                        </button>
                        <button
                            v-else
                            data-tooltip="check used"
                            class="opacity-25 tooltip-position-left group-hover:opacity-100"
                            type="button"
                            @click="ignoreCheck(item.id)"
                        >
                            <icon-fa6-solid:eye />
                        </button>
                    </template>
                </data-table>
            </template>

            <template #panel(extensions)="{ label }">
                <data-table
                    :labels="{ 
                        label: { name: 'Name', sortable: true },
                        version: { name: 'Version' },
                        latestVersion: { name: 'Latest' },
                        ratingAverage: { name: 'Rating', sortable: true },
                        installedAt: { name: 'Installed at', sortable: true },
                        issue: { name: 'Known Issue', class: 'px-3 text-right' }
                    }"
                    :data="shop.extensions"
                    :default-sorting="{ by: 'label' }"
                >
                    <template #cell(label)="{ item }">
                        <div class="flex items-start">
                            <span
                                v-if="!item.installed"
                                class="leading-5 text-gray-400 mr-2 text-base dark:text-neutral-500"
                                data-tooltip="Not installed"
                            >
                                <icon-fa6-regular:circle />
                            </span>
                            <span
                                v-else-if="item.active"
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
                            <div v-if="item.storeLink">
                                <a
                                    :href="item.storeLink"
                                    target="_blank"
                                >
                                    <div
                                        v-if="item.label"
                                        class="font-bold whitespace-normal"
                                    >{{ item.label }}</div>
                                    <div class="dark:opacity-50">{{ item.name }}</div>
                                </a>
                            </div>
                            <div v-else>
                                <div
                                    v-if="item.label"
                                    class="font-bold whitespace-normal"
                                >
                                    {{ item.label }}
                                </div>
                                <div class="dark:opacity-50">
                                    {{ item.name }}
                                </div>
                            </div>
                        </div>
                    </template>

                    <template #cell(version)="{ item }">
                        {{ item.version }}
                        <span
                            v-if="item.latestVersion && item.version < item.latestVersion"
                            data-tooltip="Update available"
                            @click="openExtensionChangelog(item as Extension)"
                        >
                            <icon-fa6-solid:rotate
                                class="ml-1 text-base text-amber-600 dark:text-amber-400 cursor-pointer"
                            />
                        </span>
                    </template>

                    <template #cell(ratingAverage)="{ item }">
                        <rating-stars :rating="item.ratingAverage" />
                    </template>

                    <template #cell(installedAt)="{ item }">
                        <template v-if="item.installedAt">
                            {{ formatDateTime(item.installedAt) }}
                        </template>
                    </template>

                    <template #cell(issue)="{ item }">
                        No known issues. <a href="#">Report issue</a>
                    </template>
                </data-table>
            </template>

            <template #panel(tasks)="{ label }">
                <data-table
                    :labels="{ 
                        name: { name: 'Name', sortable: true },
                        interval: { name: 'Interval' },
                        lastExecutionTime: { name: 'Last Executed', sortable: true },
                        nextExecutionTime: { name: 'Next Execution', sortable: true },
                        status: { name: 'Status' },
                        actions: { name: '', class: 'px-3 text-right w-16' }
                    }"
                    :data="shop.scheduledTask"
                    :default-sorting="{ by: 'nextExecutionTime' }"
                >
                    <template #cell(lastExecutionTime)="{ item }">
                        {{ formatDateTime(item.lastExecutionTime) }}
                    </template>

                    <template #cell(nextExecutionTime)="{ item }">
                        {{ formatDateTime(item.nextExecutionTime) }}
                    </template>

                    <template #cell(status)="{ item }">
                        <span
                            v-if="item.status === 'scheduled' && !item.overdue"
                            class="pill pill-success"
                        >
                            <icon-fa6-solid:check />
                            {{ item.status }}
                        </span>
                        <span
                            v-else-if="item.status === 'queued' || item.status === 'running' && !item.overdue"
                            class="pill pill-info"
                        >
                            <icon-fa6-solid:rotate />
                            {{ item.status }}
                        </span>
                        <span
                            v-else-if="item.status === 'scheduled' ||
                                item.status === 'running' && item.overdue ||
                                item.status === 'queued' 
                            "
                            class="pill pill-warning"
                        >
                            <icon-fa6-solid:info class="align-[-0.1em]" />
                            {{ item.status }}
                        </span>
                        <span
                            v-else-if="item.status === 'inactive'"
                            class="pill"
                        >
                            <icon-fa6-solid:pause />
                            {{ item.status }}
                        </span>
                        <span
                            v-else
                            class="pill pill-error"
                        >
                            <icon-fa6-solid:xmark />
                            {{ item.status }}
                        </span>
                    </template>

                    <template #cell(actions)="{ item }">
                        <span
                            class="cursor-pointer opacity-25 hover:opacity-100 tooltip-position-left"
                            data-tooltip="Re-schedule task"
                            @click="onReScheduleTask(item.id)"
                        >
                            <icon-fa6-solid:arrow-rotate-right class="text-base leading-3" />
                        </span>
                    </template>
                </data-table>
            </template>

            <template #panel(queue)="{ label }">
                <data-table
                    :labels="{ name: { name: 'Name', sortable: true }, size: { name: 'Size', sortable: true } }"
                    :data="shop.queueInfo"
                />
            </template>

            <template #panel(pagespeed)="{ label }">
                <data-table
                    :labels="{ 
                        created: { name: 'Checked At' },
                        performance: { name: 'Performance' },
                        accessibility: { name: 'Accessibility' },
                        bestpractices: { name: 'Best Practices' },
                        seo: { name: 'SEO' } 
                    }"
                    :data="shop.pageSpeed"
                >
                    <template #cell(created)="{ item }">
                        <a
                            target="_blank"
                            :href="'https://pagespeed.web.dev/analysis?url=' + shop.url"
                        >{{
                            formatDateTime(item.created_at) }}</a>
                    </template>

                    <template
                        v-for="(cell, cellKey) in {
                            'performance': 'cell(performance)',
                            'accessibility': 'cell(accessibility)',
                            'bestpractices': 'cell(bestpractices)',
                            'seo': 'cell(seo)'
                        }"
                        #[cell]="{ item, data, itemKey }"
                        :key="cell"
                    >
                        <span class="mr-2">{{ item[cellKey] }}</span>
                        <template v-if="data[(itemKey + 1)] && data[(itemKey + 1)][cellKey] !== item[cellKey]">
                            <icon-fa6-solid:arrow-right
                                :class="[{
                                    'text-green-400 -rotate-45 dark:text-green-300': 
                                        data[(itemKey + 1)][cellKey] < item[cellKey],
                                    'text-red-600 rotate-45 dark:text-red-400':
                                        data[(itemKey + 1)][cellKey] > item[cellKey]
                                }]"
                            />
                        </template>
                        <icon-fa6-solid:minus v-else />
                    </template>
                </data-table>
            </template>

            <template #panel(changelog)="{ label }">
                <data-table
                    :labels="{ date: { name: 'Date', sortable: true }, changes: { name: 'Changes' } }"
                    :data="shop.changelog"
                >
                    <template #cell(date)="{ item }">
                        {{ formatDateTime(item.date) }}
                    </template>

                    <template #cell(changes)="{ item }">
                        <span
                            class="cursor-pointer"
                            @click="openShopChangelog(item)"
                        >{{ sumChanges(item) }}</span>
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
                        <div
                            class="pl-3"
                            v-html="changeLog.text"
                        />
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
                Refresh {{ shop.name }}
            </template>
            <template #content>
                Do you also want to run a new pagespeed test?
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
