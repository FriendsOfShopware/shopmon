import { fetchWrapper } from "@/helpers/fetch-wrapper";
import { defineStore } from "pinia";
import type { Team } from '@apiTypes/team';
import type { User } from '@apiTypes/user';

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
    }
})