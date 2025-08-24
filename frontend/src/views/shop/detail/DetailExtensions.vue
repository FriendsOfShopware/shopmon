<template>
    <div class="panel panel-table">
        <data-table
        v-if="shop"
        :columns="[
            { key: 'label', name: 'Name', sortable: true },
            { key: 'version', name: 'Version', class: 'extension-version-column' },
            { key: 'latestVersion', name: 'Latest', class: 'extension-version-column' },
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
            <span :data-tooltip="row.version.length > 6 ? row.version : ''">{{ row.version.replace(/(.{6})..+/, "$1&hellip;") }}</span>
            <span
                v-if="row.latestVersion && row.version < row.latestVersion"
                data-tooltip="Update available"
                class="extension-update-available"
                @click="openExtensionChangelog(row as Extension)"
            >
                <icon-fa6-solid:rotate
                    class="icon icon-warning"
                />
            </span>
        </template>

        <template #cell-latestVersion="{ row }">
            <span v-if="row.latestVersion" :data-tooltip="row.latestVersion.length > 6 ? row.latestVersion : ''">{{ row.latestVersion.replace(/(.{6})..+/, "$1&hellip;") }}</span>
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

    <!-- Extension Changelog Modal -->
    <extension-changelog
        :show="viewExtensionChangelogDialog"
        :extension="dialogExtension"
        @close="closeExtensionChangelog"
    />
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from '@/helpers/formatter';
import type { RouterOutput } from '@/helpers/trpc';
import { useShopDetail } from '@/composables/useShopDetail';
import { useExtensionChangelogModal } from '@/composables/useExtensionChangelogModal';
import ExtensionChangelog from '@/components/modal/ExtensionChangelog.vue';

type Extension = RouterOutput['account']['currentUserExtensions'][number];

const {
    shop,
} = useShopDetail();

const {
    viewExtensionChangelogDialog,
    dialogExtension,
    openExtensionChangelog,
    closeExtensionChangelog,
} = useExtensionChangelogModal();

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

<style>
.extension-version-column {
    white-space: nowrap;
}
</style>

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

.extension-update-available {
    margin-left: .4rem;
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

.extension-update-available {
    cursor: pointer;
    
    &:hover {
        opacity: 0.7;
    }
}
</style>
