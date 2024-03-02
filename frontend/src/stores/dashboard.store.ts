import { defineStore } from 'pinia';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

export const useDashboardStore = defineStore('dashboard', {
    state: (): {
        isLoading: boolean,
        isRefreshing: boolean,
        changelogs: RouterOutput['account']['currentUserChangelogs']
    } => ({
        isLoading: false,
        isRefreshing: false,
        changelogs: [],
    }),
    actions: {

        async loadChangelogs() {
            this.isLoading = true;
            this.changelogs = await trpcClient.account.currentUserChangelogs.query();
            this.isLoading = false;
        },
    },
});
