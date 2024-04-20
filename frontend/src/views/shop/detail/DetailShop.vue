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
                    <status-icon :status="shopStore.shop.status" />
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
            :labels="
                [{
                     key: 'checks',
                     title: 'Checks',
                     count: shopStore.shop.checks?.length ?? 0,
                     icon: FaCircleCheck,
                 },
                 {
                     key: 'extensions',
                     title: 'Extensions',
                     count: shopStore.shop.extensions?.length ?? 0,
                     icon: FaPlug
                 },
                 {
                     key: 'tasks',
                     title: 'Scheduled Tasks',
                     count: shopStore.shop.scheduledTask?.length ?? 0,
                     icon: FaListCheck
                 },
                 {
                     key: 'queue',
                     title: 'Queue',
                     count: shopStore.shop.queueInfo?.length ?? 0
                     , icon: FaCircleCheck },
                 {
                     key: 'pagespeed',
                     title: 'Pagespeed',
                     count: shopStore.shop.pageSpeed?.length ?? 0
                     ,icon: FaRocket },
                 {
                     key: 'changelog',
                     title: 'Changelog',
                     count: shopStore.shop.changelog?.length ?? 0,
                     icon: FaFileWaverform
                 }]"
        >
            <template #panel-checks>
                <detail-checks-tab />
            </template>

            <template #panel-extensions>
                <detail-extensions-tab />
            </template>

            <template #panel-tasks>
                <detail-scheduled-tasks-tab />
            </template>

            <template #panel-queue>
                <detail-queue-tab />
            </template>

            <template #panel-pagespeed>
                <detail-pagespeed-tab />
            </template>

            <template #panel-changelog>
                <detail-changelog-tab />
            </template>
        </tabs>

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
import { compareVersions } from 'compare-versions';
import { createNewSortInstance } from 'fast-sort';

import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';
import { ref, Ref } from 'vue';

import type { Extension, ExtensionCompatibilitys } from '@/types/shop';

import { formatDate, formatDateTime } from '@/helpers/formatter';

import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaRocket from '~icons/fa6-solid/rocket';
import FaFileWaverform from '~icons/fa6-solid/file-waveform';
import { trpcClient } from '@/helpers/trpc';

import DetailChecksTab from "@/views/shop/detail/tabs/DetailChecksTab.vue";
import DetailExtensionsTab from "@/views/shop/detail/tabs/DetailExtensionsTab.vue";
import DetailScheduledTasksTab from "@/views/shop/detail/tabs/DetailScheduledTasksTab.vue";
import DetailQueueTab from "@/views/shop/detail/tabs/DetailQueueTab.vue";
import DetailPagespeedTab from "@/views/shop/detail/tabs/DetailPagespeedTab.vue";
import DetailChangelogTab from "@/views/shop/detail/tabs/DetailChangelogTab.vue";

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();

const viewUpdateWizardDialog: Ref<boolean> = ref(false);
const loadingUpdateWizard: Ref<boolean> = ref(false);
const dialogUpdateWizard: Ref<ExtensionCompatibilitys[] | null> = ref(null);

const showShopRefreshModal: Ref<boolean> = ref(false);
const shopwareVersions: Ref<string[] | null> = ref(null);

const latestShopwareVersion: Ref<string | null> = ref(null);

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
</script>
