import { defineStore } from 'pinia';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

export const useExtensionStore = defineStore('extension', {
    state: (): {
        isLoading: boolean,
        isRefreshing: boolean,
        extensions: RouterOutput['account']['currentUserApps']
    } => ({
        isLoading: false,
        isRefreshing: false,
        extensions: [],
    }),
    actions: {

        async loadExtensions() {
            this.isLoading = true;
            this.extensions = await trpcClient.account.currentUserApps.query();
            this.isLoading = false;
        },
    },
});
