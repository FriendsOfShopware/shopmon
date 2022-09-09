<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

import { useShopStore } from '@/stores/shop.store';

const shopStore = useShopStore();

shopStore.loadShops();

</script>

<template>
  <Header title="My Shops">
    <router-link
      to="/account/shops/new"
      type="button"
      class="group btn btn-primary flex items-center align-middle"
    >
    <icon-fa6-solid:plus 
        class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
        aria-hidden="true" 
      />
      Add Shop
    </router-link>
  </Header>
  <MainContainer v-if="!shopStore.isLoading">
    <div class="text-center" v-if="shopStore.shops.length === 0">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      </svg>
      <h3 class="mt-2 font-medium">No shops</h3>
      <p class="mt-1 text-gray-500">Get started by adding your first Shop.</p>
      <div class="mt-6">
        <router-link to="/account/shops/new" class="btn btn-primary group flex items-center">
          <icon-fa6-solid:plus
            class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
            aria-hidden="true" 
          />
          Add Shop
        </router-link>
      </div>
    </div>    

    <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden bg-white" v-else>
      <DataTable
        :labels="{favicon: {name: '', classOverride: true, class: 'w-11 py-3.5 px-3'}, name: {name: 'Name'}, url: {name: 'URL'}, shopware_version: {name: 'Version'}, team_name: {name: 'Team'}, checked: {name: 'last checked at'}}"
        :data="shopStore.shops">
        <template #cell(favicon)="{ item }">
          <img :src="item.favicon" class="inline-block w-5 h-5 mr-2 align-middle" v-if="item.favicon" />
        </template>

        <template #cell(name)="{ item }">
          <icon-fa6-solid:circle-xmark class="text-red-600 mr-2 text-base" v-if="item.status == 'red'" />
          <icon-fa6-solid:circle-info class="text-yellow-400 mr-2 text-base" v-else-if="item.status === 'yellow'" />
          <icon-fa6-solid:circle-check class="text-green-400 mr-2 text-base" v-else />             
          <router-link :to="{ name: 'account.shops.detail', params: { teamId: item.team_id, shopId: item.id } }">
            {{ item.name }}
          </router-link>
        </template>

        <template #cell(url)="{ item }">
          <a :href="item.url" :data-tooltip="item.url" target="_blank">
            <icon-fa6-solid:up-right-from-square />
          </a>
          &nbsp;
          <a :href="item.url + '/admin'" data-tooltip="Go to Shop Admin" target="_blank">
            <icon-fa6-solid:right-to-bracket />
          </a>
        </template>

        <template #cell(checked)="{ item }">
          {{ new Date(item.last_scraped_at).toLocaleString() }}
        </template>
      </DataTable>
    </div>
  </MainContainer>
</template>
