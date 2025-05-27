<template>
    <header-container
        v-if="shopStore.shop"
        :title="shopStore.shop.name"
    >
        <button
            class="btn icon-only"
            data-tooltip="Clear shop cache"
            :disabled="shopStore.isCacheClearing"
            type="button"
            @click="onCacheClear"
        >
            <icon-ic:baseline-cleaning-services
                :class="{ 'animate-pulse': shopStore.isCacheClearing }"
                class="icon"
            />
        </button>
        <button
            class="btn icon-only"
            data-tooltip="Refresh shop data"
            :disabled="shopStore.isRefreshing"
            type="button"
            @click="showShopRefreshModal = true"
        >
            <icon-fa6-solid:rotate :class="{ 'icon': true, 'animate-spin': shopStore.isRefreshing }" />
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
            class="btn btn-primary"
        >
            <icon-fa6-solid:pencil class="icon" aria-hidden="true" />
            Edit Shop
        </router-link>
    </header-container>

    <main-container v-if="shopStore.shop && shopStore.shop.lastScrapedAt">
        <div class="panel shop-info">
            <h3 class="shop-info-heading">
                <status-icon :status="shopStore.shop.status" />
                Shop Information
            </h3>
            <div class="shop-info-grid">
                <dl class="shop-info-list">
                    <div class="shop-info-item">
                        <dt>
                            Shopware Version
                        </dt>
                        <dd>
                            {{ shopStore.shop.shopwareVersion }}
                            <template v-if="latestShopwareVersion && latestShopwareVersion != shopStore.shop.shopwareVersion">
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
                    <div class="shop-info-item">
                        <dt>
                            Last Shop Update
                        </dt>
                        <dd>
                            <template v-if="shopStore.shop.lastChangelog && shopStore.shop.lastChangelog.date">
                                {{ formatDate(shopStore.shop.lastChangelog.date) }}
                            </template>
                            <template v-else>
                                never
                            </template>
                        </dd>
                    </div>
                    <div class="shop-info-item">
                        <dt>
                            Organization
                        </dt>
                        <dd>
                            {{ shopStore.shop.organizationName }}
                        </dd>
                    </div>
                    <div class="shop-info-item">
                        <dt>
                            Last Checked At
                        </dt>
                        <dd>
                            {{ formatDateTime(shopStore.shop.lastScrapedAt) }}
                        </dd>
                    </div>
                    <div class="shop-info-item">
                        <dt>
                            Environment
                        </dt>
                        <dd>
                            {{ shopStore.shop.cacheInfo?.environment }}
                        </dd>
                    </div>
                    <div class="shop-info-item">
                        <dt>
                            HTTP Cache
                        </dt>
                        <dd>
                            {{ shopStore.shop.cacheInfo?.httpCache ? 'Enabled' : 'Disabled' }}
                        </dd>
                    </div>
                    <div class="shop-info-item">
                        <dt>
                            URL
                        </dt>
                        <dd>
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
                <div class="shop-image-container">
                    <img
                        v-if="shopStore.shop.shopImage"
                        :src="`/${shopStore.shop.shopImage}`"
                        class="shop-image"
                    >
                    <icon-fa6-solid:image
                        v-else
                        class="placeholder-image"
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
            class="update-wizard"
            :show="viewUpdateWizardDialog"
            close-x-mark
            @close="viewUpdateWizardDialog = false"
        >
            <template #title>
                <icon-fa6-solid:rotate /> Shopware Plugin Compatibility Check
            </template>

            <template #content>
                <select
                    class="field"
                    @change="event => loadUpdateWizard((event.target as HTMLSelectElement).value)"
                >
                    <option disabled selected>
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
                    <div class="update-wizard-loader">
                        Loading <icon-fa6-solid:rotate class="animate-spin" />
                    </div>
                </template>

                <div
                    v-if="dialogUpdateWizard"
                    :class="{ 'update-wizard-refresh': loadingUpdateWizard }"
                >
                    <h2 class="update-wizard-plugins-heading">
                        Plugin Compatibility
                    </h2>

                    <ul>
                        <li
                            v-for="extension in dialogUpdateWizard"
                            :key="extension.name"
                            class="update-wizard-plugin"
                        >
                            <div class="mr-2 w-4">
                                <icon-fa6-regular:circle
                                    v-if="!extension.active"
                                    class="icon icon-muted"
                                />
                                <icon-fa6-solid:circle-info
                                    v-else-if="!extension.compatibility"
                                    class="icon icon-warning"
                                />
                                <icon-fa6-solid:circle-xmark
                                    v-else-if="extension.compatibility.type == 'red'"
                                    class="icon icon-error"
                                />
                                <icon-fa6-solid:rotate
                                    v-else-if="extension.compatibility.label === 'Available now'"
                                    class="icon icon-info"
                                />
                                <icon-fa6-solid:circle-check
                                    v-else
                                    class="icon icon-success"
                                />
                            </div>

                            <div>
                                <component
                                    :is="extension.storeLink ? 'a' : 'span'"
                                    v-bind="extension.storeLink ? {href: extension.storeLink, target: '_blank'} : {}"
                                >
                                    <strong>{{ extension.label }}</strong>
                                </component>
                                <span class="update-wizard-plugin-technical-name"> ({{ extension.name }})</span>

                                <div v-if="!extension.compatibility || !extension.storeLink">
                                    This plugin is not available in the Store. Please contact the
                                    plugin manufacturer.
                                </div>
                                <div v-else>
                                    {{ extension.compatibility.label }}
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
                    class="icon icon-info"
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
                    class="btn btn-primary"
                    @click="onRefresh(true)"
                >
                    Yes
                </button>

                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn btn-danger"
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

import { useAlertStore } from '@/stores/alert.store';
import { useShopStore } from '@/stores/shop.store';
import { type Ref, ref } from 'vue';
import { useRoute } from 'vue-router';

import type { Extension, ExtensionCompatibilitys } from '@/types/shop';

import { formatDate, formatDateTime } from '@/helpers/formatter';

import { trpcClient } from '@/helpers/trpc';
import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaFileWaverform from '~icons/fa6-solid/file-waveform';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaRocket from '~icons/fa6-solid/rocket';

import DetailChangelogTab from '@/views/shop/detail/tabs/DetailChangelogTab.vue';
import DetailChecksTab from '@/views/shop/detail/tabs/DetailChecksTab.vue';
import DetailExtensionsTab from '@/views/shop/detail/tabs/DetailExtensionsTab.vue';
import DetailPagespeedTab from '@/views/shop/detail/tabs/DetailPagespeedTab.vue';
import DetailQueueTab from '@/views/shop/detail/tabs/DetailQueueTab.vue';
import DetailScheduledTasksTab from '@/views/shop/detail/tabs/DetailScheduledTasksTab.vue';

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();

const viewUpdateWizardDialog: Ref<boolean> = ref(false);
const loadingUpdateWizard: Ref<boolean> = ref(false);
const dialogUpdateWizard: Ref<ExtensionCompatibilitys[] | null> = ref(null);

const showShopRefreshModal: Ref<boolean> = ref(false);
const shopwareVersions: Ref<string[] | null> = ref(null);

const latestShopwareVersion: Ref<string | null> = ref(null);

async function loadShop() {
    const organizationId = Number.parseInt(
        route.params.organizationId as string,
        10,
    );
    const shopId = Number.parseInt(route.params.shopId as string, 10);

    await shopStore.loadShop(organizationId, shopId);

    const shopwareVersionsData =
        await trpcClient.info.getLatestShopwareVersion.query();
    shopwareVersions.value = Object.keys(shopwareVersionsData)
        .reverse()
        .filter(
            (version) =>
                !version.includes('-RC') &&
                compareVersions(shopStore.shop?.shopwareVersion, version) < 0,
        );
    latestShopwareVersion.value = shopwareVersions.value[0];
}

loadShop().then(() => {
    if (shopStore?.shop?.name) {
        document.title = shopStore.shop.name;
    }
});

async function onRefresh(pagespeed: boolean) {
    showShopRefreshModal.value = false;
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.refreshShop(
                shopStore.shop.organizationId,
                shopStore.shop.id,
                pagespeed,
            );
            alertStore.success('Your Shop will refresh soon!');
        } catch (e) {
            alertStore.error(e instanceof Error ? e.message : String(e));
        }
    }
}

async function onCacheClear() {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        try {
            await shopStore.clearCache(
                shopStore.shop.organizationId,
                shopStore.shop.id,
            );
            alertStore.success('Your Shop cache was cleared successfully');
        } catch (e) {
            alertStore.error(e instanceof Error ? e.message : String(e));
        }
    }
}

function openUpdateWizard() {
    dialogUpdateWizard.value = null;
    viewUpdateWizardDialog.value = true;
}

async function loadUpdateWizard(version: string) {
    if (!shopStore.shop || !shopStore.shop.extensions) {
        return;
    }

    loadingUpdateWizard.value = true;

    const body = {
        currentVersion: shopStore.shop?.shopwareVersion,
        futureVersion: version,
        extensions: shopStore.shop.extensions.map((extension) => {
            return {
                name: extension.name,
                version: extension.version,
            };
        }),
    };

    const pluginCompatibility =
        await trpcClient.info.checkExtensionCompatibility.query(body);

    const extensions = JSON.parse(JSON.stringify(shopStore.shop?.extensions));

    for (const extension of extensions) {
        const compatibility = pluginCompatibility.find(
            (plugin) => plugin.name === extension.name,
        );
        extension.compatibility = null;
        if (compatibility) {
            extension.compatibility = compatibility.status;
        }
    }

    const naturalSort = createNewSortInstance({
        comparer: new Intl.Collator(undefined, {
            numeric: true,
            sensitivity: 'base',
        }).compare,
    });

    dialogUpdateWizard.value = naturalSort(extensions).by([
        { desc: (u) => u.active },
        { asc: (u) => u.label },
    ]);

    loadingUpdateWizard.value = false;
}
</script>

<style scoped>
.shop-info {
    padding: 0;
    margin-bottom: 3rem;

    &-heading {
        padding: 1.25rem 1rem;
        font-size: 1.125rem;
        font-weight: 500;

        @media (min-width: 1024px) {
            padding-left: 2rem;
            padding-right: 2rem;
        }

        .icon {
            margin-right: 0.25rem;
        }
    }
}

.shop-info-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    border-top: 1px solid var(--panel-border-color);
    padding: 1.25rem 1rem;

    @media (min-width: 640px) {
        grid-template-columns: repeat(2, 1fr);
    }

    @media (min-width: 960px) {
        grid-template-columns: repeat(3, 1fr);
    }

    @media (min-width: 1024px) {
        padding-left: 2rem;
        padding-right: 2rem;
    }
}

.shop-info-list {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;
    gap: 0.5rem 1.5rem;

    @media (min-width: 960px) {
        grid-column: 1 / span 2;
        grid-template-columns: repeat(2, 1fr);
    }
}

.shop-info-item {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;

    @media (min-width: 960px) {
        grid-column: span 1;
    }
    
    dt {
        font-size: 0.875rem;
        font-weight: 500;
    }

    dd {
        margin-top: 0.25rem;
        font-size: 0.875rem;
        color: var(--text-color-muted);
    }
}

.shop-image-container {
    margin-top: 1.5rem;
    display: flex;
    justify-content: center;

    @media (min-width: 640px) {
        grid-column: 2 / span 1;
        grid-row: 1 / span 1;
        margin-top: 0;
    }

    @media (min-width: 960px) {
        grid-column: 3 / span 1;
        grid-row: 1 / span 1;
    }
}

.shop-image {
    height: 6.25rem;

    @media (min-width: 640px) {
        height: 25rem;
    }

    @media (min-width: 960px) {
        height: 12.5rem;
    }
}

.placeholder-image {
    color: #e5e7eb;
    font-size: 9rem;
}

.update-wizard {
    .field {
        margin-bottom: .75rem;
    }

    &-loader {
        text-align:center;
    }

    &-refresh {
        opacity: .2;
    }

    &-plugins {
        &-heading {
            font-size: 1.125rem;
            font-weight: 500;
            margin-bottom: .5rem;
        }
    }

    &-plugin {
        background-color: var(--item-background);
        padding: .5rem;
        display: flex;

        &:nth-child(odd) {
            background-color: var(--item-odd-background);
        }

        &:hover {
            background-color: var(--item-hover-background);
        }

        &-technical-name {
            opacity: .6;
        }
    }
}
</style>
