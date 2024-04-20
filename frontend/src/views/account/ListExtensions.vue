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
                class="field field-search"
                placeholder="Search ..."
            >

            <div
                class="panel panel-table"
            >
                <data-table
                    :columns="[
                        { key: 'label', name: 'Name', class: 'app-label', sortable: true, searchable: true},
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
                            <div class="app-name">{{ row.label }}</div>
                            <span class="app-technical-name">{{ row.name }}</span>
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
                                <span
                                    v-if="!shop.installed"
                                    data-tooltip="Not installed"
                                >
                                    <icon-fa6-regular:circle class="icon-muted icon-state" />
                                </span>
                                <span
                                    v-else-if="shop.active"
                                    data-tooltip="Active"
                                >
                                    <icon-fa6-solid:circle-check class="icon-success icon-state"/>
                                </span>
                                <span
                                    v-else
                                    data-tooltip="Inactive"
                                >
                                    <icon-fa6-solid:circle-xmark class="icon-muted icon-state"/>
                                </span>
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
import { ref } from 'vue';
import { useExtensionStore } from '@/stores/extension.store';
import { formatDateTime } from '@/helpers/formatter';

const extensionStore = useExtensionStore();
const term = ref('');

extensionStore.loadExtensions();
</script>

<style scoped>
.field-search {
    margin-bottom: 0.75rem;
}

.app-label {
    .app-name {
        font-weight: bold;
        white-space: normal;
    }

    .app-technical-name {
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

    .icon-state {
        font-size: 1rem;
        margin-right: .25rem;
    }

    .icon-update {
        vertical-align: -.1em;
        margin-left: .15rem;
    }
}
</style>
