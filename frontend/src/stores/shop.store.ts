import {
    type RouterInput,
    type RouterOutput,
    trpcClient,
} from '@/helpers/trpc';
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useShopStore = defineStore('shop', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const isCacheClearing = ref(false);
    const isReSchedulingTask = ref(false);
    const shop = ref<RouterOutput['organization']['shop']['get'] | null>(null);

    async function createShop(
        payload: RouterInput['organization']['shop']['create'],
    ) {
        isLoading.value = true;
        payload.shopUrl = payload.shopUrl.replace(/\/+$/, '');
        await trpcClient.organization.shop.create.mutate(payload);
        isLoading.value = false;
    }

    async function loadShop(orgId: number, shopId: number) {
        isLoading.value = true;
        shop.value = await trpcClient.organization.shop.get.query({
            orgId,
            shopId,
        });
        isLoading.value = false;
    }

    async function updateShop(
        orgId: number,
        shopId: number,
        payload: {
            name?: string;
            ignores?: string[];
            shopUrl?: string;
            clientId?: string;
            clientSecret?: string;
        },
    ) {
        if (payload.shopUrl) {
            payload.shopUrl = payload.shopUrl.replace(/\/+$/, '');
        }
        await trpcClient.organization.shop.update.mutate({
            orgId,
            shopId,
            ...payload,
        });
        await loadShop(orgId, shopId);
    }

    async function refreshShop(
        orgId: number,
        shopId: number,
        pageSpeed = false,
    ) {
        isRefreshing.value = true;
        await trpcClient.organization.shop.refreshShop.mutate({
            orgId,
            shopId,
            pageSpeed,
        });
        isRefreshing.value = false;

        await loadShop(orgId, shopId);
    }

    async function clearCache(orgId: number, shopId: number) {
        isCacheClearing.value = true;
        await trpcClient.organization.shop.clearShopCache.mutate({
            orgId,
            shopId,
        });
        isCacheClearing.value = false;

        await loadShop(orgId, shopId);
    }

    async function reScheduleTask(
        orgId: number,
        shopId: number,
        taskId: string,
    ) {
        isReSchedulingTask.value = true;
        await trpcClient.organization.shop.rescheduleTask.mutate({
            orgId,
            shopId,
            taskId,
        });
        isReSchedulingTask.value = false;

        await loadShop(orgId, shopId);
    }

    async function deleteShop(orgId: number, shopId: number) {
        await trpcClient.organization.shop.delete.mutate({ orgId, shopId });
    }

    return {
        isLoading,
        isRefreshing,
        isCacheClearing,
        isReSchedulingTask,
        shop,
        createShop,
        loadShop,
        updateShop,
        refreshShop,
        clearCache,
        reScheduleTask,
        deleteShop,
    };
});
