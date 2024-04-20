<template>
    <data-table
        :columns="[
            { key: 'date', name: 'Date', sortable: true },
        ]"
        :data="shopStore.shop.changelog"
    >
        <template #cell-date="{ row }">
            {{ formatDateTime(row.date) }}
        </template>

        <template #cell-actions="{ row }">
            <span class="cursor-pointer" @click="openShopChangelog(row)">
                {{ sumChanges(row) }}
            </span>
        </template>
    </data-table>

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
</template>

<script setup lang="ts">
import { formatDateTime } from "@/helpers/formatter";
import { sumChanges } from "@/helpers/changelog";
import type { ShopChangelog } from "@/types/shop";
import { useShopStore } from '@/stores/shop.store';
import { ref, Ref } from "vue";

const shopStore = useShopStore();

const viewShopChangelogDialog: Ref<boolean> = ref(false);
const dialogShopChangelog: Ref<ShopChangelog | null> = ref(null);

function openShopChangelog(shopChangelog: ShopChangelog | null) {
    dialogShopChangelog.value = shopChangelog;
    viewShopChangelogDialog.value = true;
}
</script>
