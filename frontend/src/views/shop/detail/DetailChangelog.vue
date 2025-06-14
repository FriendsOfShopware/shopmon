<template>
    <div class="panel panel-table">
        <data-table
        v-if="shop"
        :columns="[
            { key: 'date', name: 'Date', class: 'changelog-date', sortable: true },
            { key: 'log', name: 'Log', sortable: false },
        ]"
        :data="shop.changelog"
    >
        <template #cell-date="{ row }">
            {{ formatDateTime(row.date) }}
        </template>

        <template #cell-log="{ row }">
            <span class="modal-open-changelog" @click="openShopChangelog(row)">
                {{ sumChanges(row) }}
            </span>
        </template>
    </data-table>
    </div>

    <modal
        :show="viewShopChangelogDialog"
        close-x-mark
        @close="viewShopChangelogDialog = false"
    >
        <template #title>
            Shop changelog - <span v-if="dialogShopChangelog?.date" class="modal-changelog-title-date">{{ formatDateTime(dialogShopChangelog.date) }}</span>
        </template>

        <template #content>
            <template v-if="dialogShopChangelog?.oldShopwareVersion && dialogShopChangelog?.newShopwareVersion">
                Shopware update from <strong>{{ dialogShopChangelog.oldShopwareVersion }}</strong> to <strong>{{ dialogShopChangelog.newShopwareVersion }}</strong>
            </template>

            <div v-if="dialogShopChangelog?.extensions?.length">
                <h2 class="modal-changelog-subtitle">
                    Shop Plugin Changelog:
                </h2>

                <ul class="modal-changelog-logs">
                    <li
                        v-for="extension in dialogShopChangelog?.extensions"
                        :key="extension.name"
                    >
                        <div class="modal-changelog-extension-name">
                            {{ extension.label }} <span>({{ extension.name }})</span>
                        </div>

                        {{ extension.state }}
                        <template v-if="extension.state === 'installed' && extension.active">
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
</template>

<script setup lang="ts">
import { sumChanges } from '@/helpers/changelog';
import { formatDateTime } from '@/helpers/formatter';
import type { ShopChangelog } from '@/types/shop';
import { type Ref, ref } from 'vue';
import { useShopDetail } from '@/composables/useShopDetail';

const {
    shop,
} = useShopDetail();

const viewShopChangelogDialog: Ref<boolean> = ref(false);
const dialogShopChangelog: Ref<ShopChangelog | null> = ref(null);

function openShopChangelog(shopChangelog: ShopChangelog | null) {
    dialogShopChangelog.value = shopChangelog;
    viewShopChangelogDialog.value = true;
}
</script>

<style>
.changelog-date {
    width: 200px;
}
</style>

<style scoped>
.modal-open-changelog {
    cursor: pointer;
}

.modal-changelog {
    &-title-date {
        font-weight: normal;
        color: var(--text-color-muted);
    }

    &-subtitle {
        font-size: 1.1rem;
        margin-bottom: .5rem;
        margin-top: .5rem;
    }

    &-logs {
        list-style: disc;

        li {
            margin-left: 1rem;
            margin-bottom: .5rem;
        }
    }

    &-extension-name {
        font-weight: 500;

        span {
            font-weight: normal;
            color: var(--text-color-muted);
        }

    }
}
</style>
