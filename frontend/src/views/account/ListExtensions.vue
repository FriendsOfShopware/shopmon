<template>
    <header-container title="My Extensions" />

    <main-container v-if="!extensionStore.isLoading">
        <template v-if="extensionStore.extensions.length === 0">
            <element-empty title="No Extensions" button="Add Shop" :route="{ name: 'account.shops.new' }">
                Get started by adding your first Shop.
            </element-empty>
        </template>

        <template v-else>
            <input
                v-model="term"
                class="field field-search"
                placeholder="Search ..."
            >

            <div class="panel panel-table">
                <data-table
                    :columns="[
                        { key: 'label', name: 'Name', class: 'extension-label', sortable: true, searchable: true},
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
                        <component :is="row.storeLink ? 'a' : 'span'" v-bind="row.storeLink ? {href: row.storeLink, target: '_blank'} : {}">
                            <div class="extension-name">{{ row.label }}</div>
                            <span class="extension-technical-name">{{ row.name }}</span>
                        </component>
                    </template>

                    <template #cell-shops="{ row }">
                        <div
                            v-for="(shop, rowIndex) in row.shops"
                            :key="rowIndex"
                            class="shops-row"
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
                                <status-icon :status="getExtensionState(row)" :tooltip="true" />
                                {{ shop.name }}
                            </router-link>
                        </div>
                    </template>

                    <template #cell-version="{ row }">
                        <div
                            v-for="(shop, rowIndex) in row.shops"
                            :key="rowIndex"
                            class="shops-row"
                        >
                            {{ shop.version }}
                            <span
                                v-if="row.latestVersion && shop.version < row.latestVersion"
                                data-tooltip="Update available"
                            >
                                <icon-fa6-solid:rotate class="icon icon-warning icon-update" />
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
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import { formatDateTime } from '@/helpers/formatter';
import type { RouterOutput } from '@/helpers/trpc';
import { useExtensionStore } from '@/stores/extension.store';
import { ref } from 'vue';

const extensionStore = useExtensionStore();
const term = ref('');

extensionStore.loadExtensions();

function getExtensionState(
    extension: RouterOutput['account']['currentUserExtensions'][number],
) {
    if (!extension.installed) {
        return 'not installed';
    }
    if (extension.active) {
        return 'active';
    }

    return 'inactive';
}
</script>

<style scoped>
.field-search {
    margin-bottom: 0.75rem;
}

.extension-label {
    .extension-name {
        font-weight: bold;
        white-space: normal;
    }

    .extension-technical-name {
        font-weight: normal;
        opacity: .6;
    }
}

.shops-row {
    white-space: nowrap;
    line-height: 1.2rem;

    &:not(:last-child) {
        margin-bottom: .35rem;
    }

    .icon-update {
        vertical-align: -.1em;
        margin-left: .15rem;
    }
}
</style>

<style>
.shops-row {
    .icon-status {
        vertical-align: -.2em;
    }
}
</style>
