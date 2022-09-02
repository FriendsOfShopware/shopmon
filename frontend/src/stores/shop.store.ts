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
        async refreshShops() {
            if(!this.shop) {
                return;
            }

            const shops = await fetchWrapper.get('/account/me/shops') as Shop[];

            sessionStorage.setItem('shops', JSON.stringify(shops));
        },

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
        },

        async updateShop(teamId: number, id: number, payload: any) {
            await fetchWrapper.patch(`/team/${teamId}/shop/${id}`, payload);

            await this.refreshShops();
        },

        async delete(teamId: number, shopId: number) {
            await fetchWrapper.delete(`/team/${teamId}/shop/${shopId}`);
        }
    }
})