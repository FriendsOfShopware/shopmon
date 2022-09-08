import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Shop, ShopDetailed } from '@apiTypes/shop'; 

export const useShopStore = defineStore('shop', {
    state: (): { shops: Shop[], isLoading: boolean, isRefreshing: boolean, shop: ShopDetailed|null } => ({
        isLoading: false,
        isRefreshing: false,
        shops: [],
        shop: null,
    }),
    actions: {
        async createShop(teamId: number, payload: any) {
            this.isLoading = true;
            await fetchWrapper.post(`/team/${teamId}/shops`, payload);
            this.isLoading = false;

            await this.loadShops();
        },

        async loadShops() {
            this.isLoading = true;
            const shops = await fetchWrapper.get('/account/me/shops') as Shop[];
            this.isLoading = false;

            this.shops = shops;
        },

        async loadShop(teamId: number, shopId: number) {
            this.isLoading = true;
            this.shop = await fetchWrapper.get(`/team/${teamId}/shop/${shopId}`) as ShopDetailed;
            this.isLoading = false;
        },
        
        async updateShop(teamId: number, id: number, payload: any) {
            await fetchWrapper.patch(`/team/${teamId}/shop/${id}`, payload);

            await this.loadShop(teamId, id);
        },

        async refreshShop(teamId: number, id: number) {
            this.isRefreshing = true;
            await fetchWrapper.post(`/team/${teamId}/shop/${id}/refresh`);
            this.isRefreshing = false;
        },

        async delete(teamId: number, shopId: number) {
            await fetchWrapper.delete(`/team/${teamId}/shop/${shopId}`);
        }
    }
})