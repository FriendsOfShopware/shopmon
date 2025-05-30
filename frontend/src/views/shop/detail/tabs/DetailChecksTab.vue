<template>
    <data-table
        :columns="[
            { key: 'message', name: 'Message' },
        ]"
        :data="shop.checks || []"
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
                v-if="shop.ignores.includes(row.id)"
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
import { useAlert } from '@/composables/useAlert';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';

const { shop } = defineProps<{
    shop: RouterOutput['organization']['shop']['get'];
}>();

const { info } = useAlert();

async function ignoreCheck(id: string) {
    const updatedIgnores = [...shop.ignores, id];
    
    await trpcClient.organization.shop.update.mutate({
        orgId: shop.organizationId,
        shopId: shop.id,
        ignores: updatedIgnores,
    });
    
    // Update local shop data
    shop.ignores = updatedIgnores;
    notificateIgnoreUpdate();
}

async function removeIgnore(id: string) {
    const updatedIgnores = shop.ignores.filter((aid: string) => aid !== id);
    
    await trpcClient.organization.shop.update.mutate({
        orgId: shop.organizationId,
        shopId: shop.id,
        ignores: updatedIgnores,
    });
    
    // Update local shop data
    shop.ignores = updatedIgnores;
    notificateIgnoreUpdate();
}

async function notificateIgnoreUpdate() {
    info('Ignore state updated. Will effect after next shop update');
}
</script>
