<template>
    <div class="panel panel-table">
        <data-table
        v-if="shop"
        :columns="[
            { key: 'label', name: 'Name', sortable: true },
            { key: 'version', name: 'Version' },
            { key: 'latestVersion', name: 'Latest' },
            { key: 'ratingAverage', name: 'Rating', sortable: true },
            { key: 'installedAt', name: 'Installed at', sortable: true },
        ]"
        :data="shop?.extensions || []"
        :default-sort="{ key: 'label', desc: false }"
    >
        <template #cell-actions-header>
            Known Issues
        </template>

        <template #cell-label="{ row }">
            <div class="extension-label">
                <status-icon :status="getExtensionState(row)" :tooltip="true" />

                <component :is="row.storeLink ? 'a' : 'span'" v-bind="row.storeLink ? {href: row.storeLink, target: '_blank'} : {}">
                    <div class="extension-name">{{ row.label }}</div>
                    <span class="extension-technical-name">{{ row.name }}</span>
                </component>
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
                    class="icon icon-warning"
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
    </div>

    <modal
        :show="viewExtensionChangelogDialog"
        close-x-mark
        @close="viewExtensionChangelogDialog = false"
    >
        <template #title>
            Changelog - <span class="extension-changelog-name">{{ dialogExtension?.name }}</span>
        </template>

        <template #content>
            <ul v-if="dialogExtension?.changelog?.length > 0" class="extension-changelog">
                <li
                    v-for="changeLog in dialogExtension.changelog"
                    :key="changeLog.version"
                    class="extension-changelog-item"
                >
                    <div class="extension-changelog-title">
                        <span
                            v-if="!changeLog.isCompatible"
                            data-tooltip="not compatible with your version"
                        >
                            <icon-fa6-solid:circle-info class="icon icon-warning" />
                        </span>

                        {{ changeLog.version }} -
                        <span class="extension-changelog-date">
                            {{ formatDate(changeLog.creationDate) }}
                        </span>
                    </div>

                    <!-- eslint-disable vue/no-v-html -->
                    <div
                        class="extension-changelog-content"
                        v-html="changeLog.text"
                    />
                    <!-- eslint-enable vue/no-v-html -->
                </li>
            </ul>

            <alert v-else type="error">
                No Changelog data provided
            </alert>
        </template>
    </modal>
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from '@/helpers/formatter';
import type { RouterOutput } from '@/helpers/trpc';
import { useShopDetail } from '@/composables/useShopDetail';

import { type Ref, ref } from 'vue';
type Extension = RouterOutput['account']['currentUserExtensions'][number];

const {
    shop,
} = useShopDetail();

const viewExtensionChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);

function openExtensionChangelog(extension: Extension | null) {
    dialogExtension.value = extension;
    viewExtensionChangelogDialog.value = true;
}

function getExtensionState(extension: Extension) {
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
.extension-label {
    display: flex;
    justify-content: flex-start;
}

.extension-name {
    font-weight: bold;
    white-space: normal;
}

.extension-technical-name,
.extension-changelog-name {
    font-weight: normal;
    color: var(--text-color-muted);
}

.extension-changelog {
    &-item {
        margin-bottom: .75rem;
    }

    &-title {
        font-weight: 600;
        margin-bottom: .25rem;
    }

    &-date {
        color: var(--text-color-muted);
        font-weight: normal;
    }

    &-content {
        padding-left: 1.5rem;
    }

}
</style>
