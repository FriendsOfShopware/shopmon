<script setup lang="ts">
import { storeToRefs } from 'pinia';
import type { User } from '@apiTypes/user';
import type { Shop } from '@apiTypes/shop';
import { useAuthStore } from '@/stores/auth.store';
import { fetchWrapper } from '@/helpers/fetch-wrapper';
import { ref } from 'vue';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';
import Spinner from '@/components/icon/Spinner.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const teams = user.value?.teams.map(team => ({
    ...team,
    membersCount: ref(0),
    shopsCount: ref(0)
}));

async function getMembersCount(teamId) {
  const members = await fetchWrapper.get(`/team/${teamId}/members`) as User[];
  return members.length;
}

async function getShopsCount(teamId) {
  const shops = await fetchWrapper.get(`/team/${teamId}/shops`) as Shop[];
  return shops.length;
}

async function loadCounts() {
  for (const team of teams) {
    team.membersCount.value = await getMembersCount(team.id);
    team.shopsCount.value = await getShopsCount(team.id);
  }
}

loadCounts();
</script>

<template>
  <Header title="My Teams">
    <router-link
      to="/account/shops/new"
      type="button"
      class="group btn btn-primary flex items-center align-middle"
    >
    <icon-fa6-solid:plus
      class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
      aria-hidden="true" 
    />
      Add Team
    </router-link>
  </Header>
  <MainContainer v-if="user">
    <div class="text-center" v-if="teams.length === 0">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      </svg>
      <h3 class="mt-2 font-medium">No Teams</h3>
      <p class="mt-1 text-gray-500">Get started by adding your first Team.</p>
      <div class="mt-6">
        <router-link to="/account/teams/new" class="btn btn-primary group flex items-center">
          <icon-fa6-solid:plus
            class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
            aria-hidden="true" 
          />
          Add Team
        </router-link>
      </div>
    </div>

    <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden" v-else>
      <DataTable
          :labels="{name: {name: 'Name'}, membersCount: {name: 'Members'}, shopsCount: {name: 'Shops'}}"
          :data="teams"
          class="bg-white"
      >
          <template #cell(name)="{ item }">
            <router-link :to="{ name: 'account.teams.detail', params: { teamId: item.id } }">
                {{ item.name }}
              </router-link>
          </template>

          <template #cell(membersCount)="{ item }">
            <span v-if="item.membersCount.value > 0">
              <icon-fa6-solid:people-group />
              {{ item.membersCount }}
            </span>
            <svg
              v-else
              class="animate-spin h-5 w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-25"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          </template>

          <template #cell(shopsCount)="{ item }">
            <span v-if="item.shopsCount.value > 0">
              <icon-fa6-solid:cart-shopping />
              {{ item.shopsCount }}
            </span>
            <svg
              v-else
              class="animate-spin h-5 w-5"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-25"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          </template>
        </DataTable>
    </div>
  </MainContainer>
</template>
