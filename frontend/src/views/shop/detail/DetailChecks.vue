<template>
    <div class="panel panel-table">
        <data-table
            v-if="shop"
            :columns="[
            { key: 'message', name: 'Message' },
        ]"
            :data="sortedChecks"
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
                    class="tooltip-top-left"
                    type="button"
                    @click="removeIgnore(row.id)"
                >
                    <icon-fa6-solid:eye-slash class="icon icon-error" />
                </button>

                <button
                    v-else
                    data-tooltip="check used"
                    class="tooltip-top-left"
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
import { computed } from 'vue';

const { shop } = useShopDetail();

// Sort checks by status: red first, then yellow, then green
const sortedChecks = computed(() => {
    if (!shop.value?.checks) return [];
    return [...shop.value.checks].sort((a, b) => {
        if (a.level === 'red' && b.level !== 'red') return -1;
        if (a.level !== 'red' && b.level === 'red') return 1;
        if (a.level === 'yellow' && b.level === 'green') return -1;
        if (a.level === 'green' && b.level === 'yellow') return 1;
        return 0;
    });
});

const { info } = useAlert();

async function ignoreCheck(id: string) {
    if (!shop.value) return;

    const updatedIgnores = [...shop.value.ignores, id];

    await trpcClient.organization.shop.update.mutate({
        orgId: shop.value.organizationId,
        shopId: shop.value.id,
        projectId: shop.value.projectId,
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
        projectId: shop.value.projectId,
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
