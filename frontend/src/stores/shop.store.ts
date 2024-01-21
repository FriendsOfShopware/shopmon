import { defineStore } from 'pinia';
import { trpcClient, RouterOutput, RouterInput } from '@/helpers/trpc';

export const useShopStore = defineStore('shop', {
    state: (): { 
        shops: RouterOutput['account']['currentUserShops'],
        isLoading: boolean,
        isRefreshing: boolean,
        isCacheClearing: boolean,
        isReSchedulingTask: boolean,
        shop: RouterOutput['organization']['shop']['get'] | null
    } => ({
        isLoading: false,
        isRefreshing: false,
        isCacheClearing: false,
        isReSchedulingTask: false,
        shops: [],
        shop: null,
    }),
    actions: {
        async createShop(payload: RouterInput['organization']['shop']['create']) {
            this.isLoading = true;
            payload.shopUrl = payload.shopUrl.replace(/\/+$/, '');
            await trpcClient.organization.shop.create.mutate(payload);
            this.isLoading = false;

            await this.loadShops();
        },

        async loadShops() {
            this.isLoading = true;
            const shops = await trpcClient.account.currentUserShops.query();
            this.isLoading = false;

            const enrichedShops = this.setShopsInitials(shops);

            this.shops = enrichedShops;
        },

        async loadShop(orgId: number, shopId: number) {
            this.isLoading = true;
            this.shop = await trpcClient.organization.shop.get.query({ orgId, shopId });
            this.isLoading = false;
        },

        async updateShop(orgId: number, shopId: number,
            payload: { name?: string, ignores?: string[], shopUrl?: string, clientId?: string, clientSecret?: string
            }) {
            if (payload.shopUrl) {
                payload.shopUrl = payload.shopUrl.replace(/\/+$/, '');
            }
            await trpcClient.organization.shop.update.mutate({ orgId, shopId, ...payload });

            await this.loadShop(orgId, shopId);
        },

        async refreshShop(orgId: number, id: number, pageSpeed: boolean) {
            this.isRefreshing = true;
            await trpcClient.organization.shop.refreshShop.mutate({ orgId, shopId: id, pageSpeed });
            this.isRefreshing = false;
        },

        async clearCache(orgId: number, shopId: number) {
            this.isCacheClearing = true;
            await trpcClient.organization.shop.clearShopCache.mutate({ orgId, shopId });
            this.isCacheClearing = false;
        },

        async reScheduleTask(orgId: number, shopId: number, taskId: string) {
            this.isReSchedulingTask = true;
            await trpcClient.organization.shop.rescheduleTask.mutate({ orgId, shopId, taskId });
            await this.loadShop(orgId, shopId);
            this.isReSchedulingTask = false;
        },

        async delete(orgId: number, shopId: number) {
            await trpcClient.organization.shop.delete.mutate({ orgId, shopId });
        },

        setShopsInitials(shops: RouterOutput['account']['currentUserShops']) {
            return shops?.map(shop => ({
                ...shop,
                initials: this.getShopInitias(shop.name),
            }));
        },

        getShopInitias(name: string) {
            const initials = name.split(/\s/).slice(0, 2).reduce((response, word) => response += word.slice(0, 1), '');

            if (initials && initials.length > 1) {
                return initials.toUpperCase();
            }

            return name.slice(0, 2).toUpperCase();
        },
    },
});
