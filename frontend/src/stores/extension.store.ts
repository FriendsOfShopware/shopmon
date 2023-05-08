import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { UserExtension } from '@apiTypes/shop'; 

export const useExtensionStore = defineStore('extension', {
    state: (): { isLoading: boolean, isRefreshing: boolean, extensions: UserExtension[] } => ({
        isLoading: false,
        isRefreshing: false,
        extensions: []
    }),
    actions: {
        
        async loadExtensions() {
            this.isLoading = true;
            this.extensions = await fetchWrapper.get('/account/me/apps') as UserExtension[];
            this.isLoading = false;
        }
    }
})