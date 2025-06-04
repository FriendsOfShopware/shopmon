<template>
    <header-container title="My Extensions" />

    <main-container v-if="extensions">
        <template v-if="extensions.length === 0">
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
                    :data="extensions"
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
                                        slug: shop.organizationSlug,
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
                            <span class="extension-version" :data-tooltip="shop.version">{{ shop.version }}</span>
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
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { ref } from 'vue';

const term = ref('');
const extensions = ref<RouterOutput['account']['currentUserExtensions']>([]);

trpcClient.account.currentUserExtensions.query().then((data) => {
    extensions.value = data;
});

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

.extension-version {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
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
