<script setup lang="ts">
import {storeToRefs} from 'pinia';

import {useAuthStore} from '@/stores/auth.store';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const teams = user.value?.teams.map(team => ({
    ...team,
    initials: team.name.substring(0, 2),
    members: 1,
    bgColor: 'bg-purple-600'
}));
</script>

<template>
  <div v-if="user">
    <Header title="Dashboard" />
    <MainContainer>
      <div>
        <h2 class="text-gray-500 text-sm font-medium">My Team's</h2>
        <ul role="list" class="mt-3 grid grid-cols-1 gap-5 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <li v-for="team in teams" :key="team.id" class="col-span-1 flex shadow-sm rounded-md">
            <div :class="[team.bgColor, 'flex-shrink-0 flex items-center justify-center w-16 text-white text-sm font-medium rounded-l-md']">
              {{ team.initials }}
            </div>
            <div class="flex-1 flex items-center justify-between border-t border-r border-b border-gray-200 bg-white rounded-r-md truncate">
              <div class="flex-1 px-4 py-2 text-sm truncate">
                <router-link :to="{ name: 'account.teams.detail', params: { teamId: team.id } }" class="text-gray-900 font-medium hover:text-gray-600">
                  {{ team.name }}
                </router-link>
                <p class="text-gray-500">{{ team.members }} Members</p>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </MainContainer>
  </div>
</template>
