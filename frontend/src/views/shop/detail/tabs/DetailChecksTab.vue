<template>
    <data-table
        :columns="[
            { key: 'message', name: 'Message' },
        ]"
        :data="shopStore.shop.checks || []"
    >
        <template #cell-message="{ row }">
            <status-icon :status="row.level" />

            <component :is="row.link ? 'a' : 'span'" v-bind="row.link ? {href: row.link, target: '_blank'} : {}">
                {{ row.message }}
                <icon-fa6-solid:up-right-from-square class="icon icon-xs" v-if="row.link"/>
            </component>
        </template>

        <template #cell-actions="{ row }">
            <button
                v-if="shopStore.shop.ignores.includes(row.id)"
                data-tooltip="check is ignored"
                class="tooltip-position-left"
                type="button"
                @click="removeIgnore(row.id)"
            >
                <icon-fa6-solid:eye-slash class="icon icon-error"/>
            </button>

            <button
                v-else
                data-tooltip="check used"
                class="tooltip-position-left"
                type="button"
                @click="ignoreCheck(row.id)"
            >
                <icon-fa6-solid:eye class="icon"/>
            </button>
        </template>
    </data-table>
</template>

<script setup lang="ts">
import { useAlertStore } from '@/stores/alert.store';
import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();
const alertStore = useAlertStore();

async function ignoreCheck(id: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        shopStore?.shop?.ignores.push(id);

        shopStore.updateShop(shopStore.shop.organizationId, shopStore.shop.id, {
            ignores: shopStore?.shop?.ignores,
        });
        notificateIgnoreUpdate();
    }
}

async function removeIgnore(id: string) {
    if (shopStore?.shop?.organizationId && shopStore?.shop?.id) {
        shopStore.shop.ignores = shopStore.shop.ignores.filter(
            (aid: string) => aid !== id,
        );

        shopStore.updateShop(shopStore.shop.organizationId, shopStore.shop.id, {
            ignores: shopStore?.shop?.ignores,
        });
        notificateIgnoreUpdate();
    }
}

async function notificateIgnoreUpdate() {
    alertStore.info('Ignore state updated. Will effect after next shop update');
}
</script>
