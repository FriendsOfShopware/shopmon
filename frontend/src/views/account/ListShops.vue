<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
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
      <font-awesome-icon 
        icon="fa-solid fa-plus" 
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
          <font-awesome-icon 
            icon="fa-solid fa-plus" 
            class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" 
            aria-hidden="true" 
          />
          Add Shop
        </router-link>
      </div>
    </div>

    <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden" v-else>
      <table class="min-w-full divide-y-2 divide-gray-200 bg-white">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="w-11 py-3.5 px-1 text-left"></th>
            <th scope="col" class="px-3 py-3.5 text-left">Name</th>
            <th scope="col" class="px-3 py-3.5 text-left">URL</th>
            <th scope="col" class="px-3 py-3.5 text-left">Version</th>
            <th scope="col" class="px-3 py-3.5 text-left">Team</th>
            <th scope="col" class="px-3 py-3.5 text-left">Last checked at</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr class="even:bg-gray-50" v-for="shop in shopStore.shops" :key="shop.id">
            <td class="whitespace-nowrap py-2 px-3">
              <img :src="shop.favicon" class="inline-block w-5 h-5 mr-2 align-middle" v-if="shop.favicon" />
            </td>
            <td class="whitespace-nowrap px-3 py-4">
              <font-awesome-icon 
                icon="fa-solid fa-circle"
                size="sm"
                class="mr-2"
                :class="[
                  {'text-green-400': shop.status === 'green'},
                  {'text-yellow-400': shop.status === 'yellow'},
                  {'text-red-600': shop.status === 'red'},
                ]"
              />              
              <router-link :to="{ name: 'account.shops.detail', params: { teamId: shop.team_id, shopId: shop.id } }">
                {{ shop.name }}
              </router-link>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-gray-500">
              <a :href="shop.url" :data-tooltip="shop.url" target="_blank">
                <font-awesome-icon 
                  icon="fa-solid fa-up-right-from-square"
                />
              </a>
              &nbsp;
              <a :href="shop.url + '/admin'" data-tooltip="Go to Shop Admin" target="_blank">
                <font-awesome-icon 
                  icon="fa-solid fa-right-to-bracket"
                />
              </a>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-gray-500">{{ shop.shopware_version }}</td>            
            <td class="whitespace-nowrap px-3 py-4 text-gray-500">{{ shop.team_name }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-gray-500">{{ new Date(shop.last_scraped_at).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </MainContainer>
</template>
