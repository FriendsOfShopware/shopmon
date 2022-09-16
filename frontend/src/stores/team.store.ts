import {fetchWrapper} from "@/helpers/fetch-wrapper";
import {defineStore} from "pinia";
import {useAuthStore} from '@/stores/auth.store';
import type {Team} from '@apiTypes/team';
import type {User} from '@apiTypes/user';

const authStore = useAuthStore();

export const useTeamStore = defineStore('team', {
    state: (): { isLoading: boolean, isRefreshing: boolean, team: Team|null, members: User[] } => ({
        isLoading: false,
        isRefreshing: false,
        team: null,
        members: []
    }),
    actions: {
        async loadMembers(teamId: number) {
            this.isLoading = true;
            this.members = await fetchWrapper.get(`/team/${teamId}/members`) as User[];
            this.isLoading = false;
        },

        async delete(teamId: number) {
            await fetchWrapper.delete(`/team/${teamId}`);
            await authStore.refreshUser();
        },

        async addMember(teamId: number, payload: any) {
            await fetchWrapper.post(`/team/${teamId}/members`, payload);
            await authStore.refreshUser();
            await this.loadMembers(teamId);
        },

        async removeMember(teamId: number, memberId: number) {
            await fetchWrapper.delete(`/team/${teamId}/members/${memberId}`);
            await authStore.refreshUser();
            await this.loadMembers(teamId);
        },

        async createTeam(payload: any) {
            await fetchWrapper.post(`/team`, payload);
            await authStore.refreshUser();
        },

        async updateTeam(teamId: number, payload: any) {
            await fetchWrapper.patch(`/team/${teamId}`, payload);
            await authStore.refreshUser();
        }
    }
})