<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const letterColors = {
  a: '#00B7FF', b: '#004DFF', c: '#00FFFF', d: '#826400', e: '#580041', f: '#ec4899', g: '#00FF00', h: '#C500FF', i: '#B4FFD7', j: '#FFCA00', k:'#969600', l: '#B4A2FF', m: '#C20078',
  n: '#1e3a8a', o: '#FF8B00', p: '#FFC8FF', q: '#666666', r: '#dc2626', s: '#CCCCCC', t: '#009E8F', u: '#D7A870', v: '#8200FF', w: '#960000', x: '#BBFF00', y: '#FFFF00', z: '#006F00'  
} as {[key: string]: string};

function getContrastYIQ(hexcolor: string) {
	// If a leading # is provided, remove it
	if (hexcolor.slice(0, 1) === '#') {
		hexcolor = hexcolor.slice(1);
	}

	// If a three-character hexcode, make six-character
	if (hexcolor.length === 3) {
		hexcolor = hexcolor.split('').map(function (hex) {
			return hex + hex;
		}).join('');
	}

	// Convert to RGB value
	var r = parseInt(hexcolor.substr(0,2),16);
	var g = parseInt(hexcolor.substr(2,2),16);
	var b = parseInt(hexcolor.substr(4,2),16);

	// Get YIQ ratio
	var yiq = ((r * 299) + (g * 587) + (b * 114)) / 1000;

	// Check contrast
	return (yiq >= 128) ? 'black' : 'white';
}

function getLetterColor(name: string) {
  return letterColors[name.substring(0, 1).toLowerCase()];
}

const teams = user.value?.teams.map(team => ({
    ...team,
    initials: team.name.substring(0, 2)
}));

const shopStore = useShopStore();
shopStore.loadShops();
</script>

<template>
  <div v-if="user && !shopStore.isLoading">
    <Header title="Dashboard" />
    <MainContainer>

      <h2 class="text-gray-500 text-sm font-medium">My Shops's</h2>
      <ul role="list" class="mt-3 grid grid-cols-1 gap-5 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-6">
        <li v-for="shop in shopStore.shops" :key="shop.id" class="col-span-1 shadow-sm rounded-md dark:shadow-none group">
          <router-link :to="{ name: 'account.shops.detail', params: { teamId: shop.team_id, shopId: shop.id } }" class="flex">
            <div
              class="flex-shrink-0 flex items-center justify-center w-16 text-sm font-medium rounded-l-md text-gray-900"
              :style="`background-color: ${getLetterColor(shop.initials)}`">
              <span :class="[{'text-white': getContrastYIQ(getLetterColor(shop.initials)) === 'white' }]">
                {{ shop.initials }}
              </span>
            </div>
            <div class="flex-1 items-center justify-between bg-white rounded-r-md px-4 py-2 truncate dark:bg-neutral-800 group-hover:bg-sky-50 dark:group-hover:bg-[#2a2b2f]">
              <div class="text-gray-900 font-medium truncate dark:text-neutral-400">{{ shop.name }}</div>
              <div class="text-gray-500">
                {{ shop.shopware_version }}
              </div>
            </div>
          </router-link>

        </li>
      </ul>

      <h2 class="text-gray-500 text-sm font-medium">My Team's</h2>
      <ul role="list" class="mt-3 grid grid-cols-1 gap-5 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <li v-for="team in teams" :key="team.id" class="col-span-1 shadow-sm rounded-md dark:shadow-none group">
          <router-link :to="{ name: 'account.teams.detail', params: { teamId: team.id } }" class="flex">
            <div
              class="flex-shrink-0 flex items-center justify-center w-16 text-sm font-medium rounded-l-md uppercase text-gray-900"
              :style="`background-color: ${getLetterColor(team.name)}`">
              <span :class="[{'text-white': getContrastYIQ(getLetterColor(team.name)) === 'white' }]">
                {{ team.name.substring(0, 2) }}
              </span>
            </div>
            <div class="flex-1 items-center justify-between bg-white rounded-r-md px-4 py-2 truncate dark:bg-neutral-800 group-hover:bg-sky-50 dark:group-hover:bg-[#2a2b2f]">
              <div class="text-gray-900 font-medium truncate dark:text-neutral-400">{{ team.name }}</div>
              <div class="text-gray-500">
                {{ team.memberCount }} Members, {{ team.shopCount }} Shops
              </div>
            </div>
          </router-link>

        </li>
      </ul>
    </MainContainer>
  </div>
</template>
