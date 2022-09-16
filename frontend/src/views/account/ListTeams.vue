<script setup lang="ts">
import {storeToRefs} from 'pinia';
import {useAuthStore} from '@/stores/auth.store';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const teams = user.value?.teams;
</script>

<template>
  <Header title="My Teams">
    <router-link
      to="/account/teams/new"
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
    <div class="text-center" v-if="teams && teams.length === 0">
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

    <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden dark:shadow-none" v-else>
      <DataTable
          :labels="{name: {name: 'Name'}, memberCount: {name: 'Members'}, shopCount: {name: 'Shops'}}"
          :data="teams"
          class="bg-white dark:bg-neutral-800"
      >
          <template #cell(name)="{ item }">
            <router-link :to="{ name: 'account.teams.detail', params: { teamId: item.id } }">
                {{ item.name }}
              </router-link>
          </template>

          <template #cell(memberCount)="{ item }">
            <icon-fa6-solid:people-group />
            {{ item.memberCount }}
          </template>

          <template #cell(shopsCount)="{ item }">
            <icon-fa6-solid:cart-shopping />
            {{ item.shopCount }}
          </template>
        </DataTable>
    </div>
  </MainContainer>
</template>
