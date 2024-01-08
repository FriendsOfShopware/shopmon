import { defineStore } from "pinia";
import { useAuthStore } from '@/stores/auth.store';
import { trpcClient, RouterOutput } from "@/helpers/trpc";

const authStore = useAuthStore();

export const useTeamStore = defineStore('team', {
    state: (): { isLoading: boolean, isRefreshing: boolean, members: RouterOutput['organization']['listMembers'] } => ({
        isLoading: false,
        isRefreshing: false,
        members: []
    }),
    actions: {
        async loadMembers(orgId: number) {
            this.isLoading = true;
            this.members = await trpcClient.organization.listMembers.query({ orgId });
            this.isLoading = false;
        },

        async delete(orgId: number) {
            await trpcClient.organization.delete.mutate({ orgId });
            await authStore.refreshUser();
        },

        async addMember(orgId: number, email: string) {
            await trpcClient.organization.addMember.mutate({ orgId, email });
            await authStore.refreshUser();
            await this.loadMembers(orgId);
        },

        async removeMember(orgId: number, userId: number) {
            await trpcClient.organization.removeMember.mutate({ orgId, userId });
            await authStore.refreshUser();
            await this.loadMembers(orgId);
        },

        async createTeam(name: string) {
            await trpcClient.organization.create.mutate(name);
            await authStore.refreshUser();
        },

        async updateTeam(orgId: number, name: string) {
            await trpcClient.organization.update.mutate({ orgId, name });
            await authStore.refreshUser();
        }
    }
})
