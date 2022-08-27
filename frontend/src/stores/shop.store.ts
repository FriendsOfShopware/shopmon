import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Shop } from '@apiTypes/shop'; 

export const useShopStore = defineStore('shop', {
    state: (): {shops: Shop[], isLoading: boolean} => ({
        isLoading: false,
        shops: [],
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
        }
    }
})