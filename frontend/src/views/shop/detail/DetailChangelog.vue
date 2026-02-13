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

    <!-- Changelog Modal -->
    <shop-changelog 
        :show="viewShopChangelogDialog"
        :changelog="dialogShopChangelog"
        @close="closeShopChangelog"
    />
</template>

<script setup lang="ts">
import { sumChanges } from '@/helpers/changelog';
import { formatDateTime } from '@/helpers/formatter';
import { useShopDetail } from '@/composables/useShopDetail';
import { useShopChangelogModal } from '@/composables/useShopChangelogModal';
import ShopChangelog from '@/components/modal/ShopChangelog.vue';

const { shop } = useShopDetail();
const { viewShopChangelogDialog, dialogShopChangelog, openShopChangelog, closeShopChangelog } = useShopChangelogModal();
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
</style>
