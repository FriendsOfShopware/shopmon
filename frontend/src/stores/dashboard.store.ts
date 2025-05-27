import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useDashboardStore = defineStore('dashboard', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const changelogs = ref<RouterOutput['account']['currentUserChangelogs']>(
        [],
    );

    async function loadChangelogs() {
        isLoading.value = true;
        changelogs.value =
            await trpcClient.account.currentUserChangelogs.query();
        isLoading.value = false;
    }

    return {
        isLoading,
        isRefreshing,
        changelogs,
        loadChangelogs,
    };
});
