import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { useAuthStore } from '@/stores/auth.store';
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useOrganizationStore = defineStore('organization', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const members = ref<RouterOutput['organization']['listMembers']>([]);

    async function loadMembers(orgId: number) {
        isLoading.value = true;
        members.value = await trpcClient.organization.listMembers.query({
            orgId,
        });
        isLoading.value = false;
    }

    async function deleteOrganization(orgId: number) {
        await trpcClient.organization.delete.mutate({ orgId });
        const authStore = useAuthStore();
        await authStore.refreshUser();
    }

    async function addMember(orgId: number, email: string) {
        await trpcClient.organization.addMember.mutate({ orgId, email });
        const authStore = useAuthStore();
        await authStore.refreshUser();
        await loadMembers(orgId);
    }

    async function removeMember(orgId: number, userId: string) {
        await trpcClient.organization.removeMember.mutate({ orgId, userId });
        const authStore = useAuthStore();
        await authStore.refreshUser();
        await loadMembers(orgId);
    }

    async function createOrganization(name: string) {
        await trpcClient.organization.create.mutate(name);
        const authStore = useAuthStore();
        await authStore.refreshUser();
    }

    async function updateOrganization(orgId: number, name: string) {
        await trpcClient.organization.update.mutate({ orgId, name });
        const authStore = useAuthStore();
        await authStore.refreshUser();
    }

    return {
        isLoading,
        isRefreshing,
        members,
        loadMembers,
        deleteOrganization,
        addMember,
        removeMember,
        createOrganization,
        updateOrganization,
    };
});
