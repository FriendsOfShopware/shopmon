import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Changelogs } from '@apiTypes/dashboard'; 

export const useDashboardStore = defineStore('dashboard', {
    state: (): { isLoading: boolean, isRefreshing: boolean, changelogs: Changelogs[] } => ({
        isLoading: false,
        isRefreshing: false,
        changelogs: []
    }),
    actions: {
        
        async loadChangelogs() {
            this.isLoading = true;
            this.changelogs = await fetchWrapper.get('/account/me/changelogs') as Changelogs[];
            this.isLoading = false;
        }
    }
})