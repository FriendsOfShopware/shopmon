<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const teams = user.value?.teams.map(team => ({
    ...team,
    initials: team.name.substring(0, 2),
    members: 1,
    bgColor: 'bg-purple-600',
    shops: 1
}));
</script>

<template>
  <Header title="My Teams">
    <router-link
      to="/account/shops/new"
      type="button"
      class="group btn btn-primary flex items-center align-middle"
    >
      <font-awesome-icon 
        icon="fa-solid fa-plus" 
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
          <font-awesome-icon 
            icon="fa-solid fa-plus" 
            class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
            aria-hidden="true" 
          />
          Add Team
        </router-link>
      </div>
    </div>

    <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden" v-else>
      <DataTable
          :labels="{name: {name: 'Name'}, members: {name: 'Members'}, shops: {name: 'Shops'}}"
          :data="teams"
          class="bg-white"
      >
          <template #cell(name)="{ item }">
            <router-link :to="{ name: 'account.teams.detail', params: { teamId: item.id } }">
                {{ item.name }}
              </router-link>
          </template>

          <template #cell(members)="{ item }">
            <font-awesome-icon 
                icon="fa-solid fa-people-group"
            />
            {{ item.members }}
          </template>

          <template #cell(shops)="{ item }">
            <font-awesome-icon 
                icon="fa-solid fa-cart-shopping"
            />
            {{ item.shops }}
          </template>
        </DataTable>
    </div>
  </MainContainer>
</template>
