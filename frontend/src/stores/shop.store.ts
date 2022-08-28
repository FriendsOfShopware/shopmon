import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Shop, ShopDetailed } from '@apiTypes/shop'; 

export const useShopStore = defineStore('shop', {
    state: (): {shops: Shop[], isLoading: boolean, shop: ShopDetailed|null} => ({
        isLoading: false,
        shops: [],
        shop: null,
    }),
    actions: {
        async loadShops() {
            this.isLoading = true;
            const shops = await fetchWrapper.get('/account/me/shops') as Shop[];
            this.isLoading = false;

            this.shops = shops;
        },
        async createShop(teamId: number, payload: any) {
            this.isLoading = true;
            await fetchWrapper.post(`/team/${teamId}/shops`, payload);
            this.isLoading = false;

            await this.loadShops();
        },

        async loadShop(teamId: string, shopId: string) {
            this.isLoading = true;
            this.shop = await fetchWrapper.get(`/team/${teamId}/shop/${shopId}`) as ShopDetailed;
            this.isLoading = false;
        }
    }
})