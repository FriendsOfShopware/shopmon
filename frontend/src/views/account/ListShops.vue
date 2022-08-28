<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { useShopStore } from '@/stores/shop.store';
import { PlusIcon } from '@heroicons/vue/20/solid'
const shopStore = useShopStore();

shopStore.loadShops();

</script>

<template>
  <Header title="My Shops">
    <router-link
      to="/account/shops/new"
      type="button"
      class="btn btn-primary"
    >
      <PlusIcon class="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
      Add Shop
    </router-link>
  </Header>
  <MainContainer v-if="!shopStore.isLoading">
    <div class="text-center" v-if="shopStore.shops.length === 0">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No shops</h3>
      <p class="mt-1 text-sm text-gray-500">Get started by adding your first Shop.</p>
      <div class="mt-6">
        <router-link to="/account/shops/new" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-sky-600 hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500">
          <PlusIcon class="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          New Shop
        </router-link>
      </div>
    </div>

    <div class="px-4 sm:px-6 lg:px-8" v-else>
      <div class="mt-8 flex flex-col">
        <div class="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div class="inline-block min-w-full py-2 align-middle">
            <div class="overflow-hidden shadow-sm ring-1 ring-black ring-opacity-5">
              <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-50">
                  <tr>
                    <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6 lg:pl-8">Name</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Version</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Team</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Last checked at</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr v-for="shop in shopStore.shops" :key="shop.id">
                    <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8">
                      <router-link :to="{ name: 'account.shops.detail', params: { teamId: shop.team_id, shopId: shop.id } }">
                        {{ shop.name }}
                      </router-link>
                    </td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ shop.shopware_version }}</td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ shop.team_name }}</td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ shop.last_scraped_at }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </MainContainer>
</template>
