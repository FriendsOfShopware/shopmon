<template>
    <div class="panel panel-table">
        <data-table
            v-if="shop"
            :columns="[
            { key: 'message', name: 'Message' },
        ]"
            :data="shop.checks || []"
        >
            <template #cell-message="{ row }">
                <status-icon :status="row.level" />

                <component :is="row.link ? 'a' : 'span'" v-bind="row.link ? {href: row.link, target: '_blank'} : {}">
                    {{ row.message }}
                    <icon-fa6-solid:up-right-from-square v-if="row.link" class="icon icon-xs" />
                </component>
            </template>

            <template #cell-actions="{ row }">
                <button
                    v-if="shop.ignores.includes(row.id)"
                    data-tooltip="check is ignored"
                    class="tooltip-position-left"
                    type="button"
                    @click="removeIgnore(row.id)"
                >
                    <icon-fa6-solid:eye-slash class="icon icon-error" />
                </button>

                <button
                    v-else
                    data-tooltip="check used"
                    class="tooltip-position-left"
                    type="button"
                    @click="ignoreCheck(row.id)"
                >
                    <icon-fa6-solid:eye class="icon" />
                </button>
            </template>
        </data-table>
    </div>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { trpcClient } from '@/helpers/trpc';
import { useShopDetail } from "@/composables/useShopDetail";

const {
    shop,
} = useShopDetail();

const { info } = useAlert();

async function ignoreCheck(id: string) {
    if (!shop.value) return;

    const updatedIgnores = [...shop.value.ignores, id];

    await trpcClient.organization.shop.update.mutate({
        orgId: shop.value.organizationId,
        shopId: shop.value.id,
        ignores: updatedIgnores,
    });

    // Update shop data
    shop.value = { ...shop.value, ignores: updatedIgnores };
    notificateIgnoreUpdate();
}

async function removeIgnore(id: string) {
    if (!shop.value) return;

    const updatedIgnores = shop.value.ignores.filter((aid: string) => aid !== id);

    await trpcClient.organization.shop.update.mutate({
        orgId: shop.value.organizationId,
        shopId: shop.value.id,
        ignores: updatedIgnores,
    });

    // Update shop data
    shop.value = { ...shop.value, ignores: updatedIgnores };
    notificateIgnoreUpdate();
}

async function notificateIgnoreUpdate() {
    info('Ignore state updated. Will effect after next shop update');
}
</script>
