<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';
import { useTeamStore } from '@/stores/team.store';
import { useRoute } from 'vue-router';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

const route = useRoute();
const authStore = useAuthStore();
const teamStore = useTeamStore();
const { user } = storeToRefs(authStore);

const teamId = parseInt(route.params.teamId as string, 10);
const team = user.value?.teams.find(team => team.id == teamId);

teamStore.loadMembers(teamId);
</script>

<template>
  <Header v-if="team" :title="team.name">
    <div class="flex gap-2">
      <router-link
        v-if="team.is_owner"
        :to="{ name: 'account.teams.edit', params: { teamId } }"
        type="button" class="group btn btn-primary flex items-center">
        <icon-fa6-solid:plus class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
          aria-hidden="true" />
        Edit Team
      </router-link>
    </div>
  </Header>

  <MainContainer>
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          Members
        </h3>
      </div>
      <DataTable
          :labels="{email: {name: 'Email'}, username: {name: 'Username'}}"
          :data="teamStore.members">
        </DataTable>
    </div>
  </MainContainer>
</template>
