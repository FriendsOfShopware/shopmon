<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { useShopStore } from '@/stores/shop.store';
import { useRoute } from 'vue-router';
import type { ScheduledTask } from '@apiTypes/shop';

import {
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/solid';

const route = useRoute();
const shopStore = useShopStore();

shopStore.loadShop(route.params.teamId as string, route.params.shopId as string);

function isTaskOverdue(task: ScheduledTask) {
  // Some timestamps are always UTC
  var currentTime = new Date();
  currentTime.setMinutes(currentTime.getMinutes() + currentTime.getTimezoneOffset());
  const nextExecute = new Date(task.nextExecutionTime);

  return currentTime.getTime() > nextExecute.getTime();
}
</script>

<template>
  <Header v-if="shopStore.shop" :title="shopStore.shop.name" />
  <MainContainer v-if="shopStore.shop">
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">Shop Information</h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8">
        <dl class="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium">Shopware Version</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.shopware_version }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="font-medium">Team</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.team_name }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="font-medium">Last Checked At</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ new Date(shopStore.shop.last_scraped_at).toLocaleString() }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="font-medium">Environment</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.cache_info.environment }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="font-medium">HTTP Cache</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.cache_info.httpCache ? 'Enabled' : 'Disabled' }}
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <div class="mb-12 bg-white shadow rounded-md overflow-y-scroll md:overflow-hidden">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">Extensions</h3>
        <p class="mt-1 text-sm text-gray-500">List of all installed extensions</p>
      </div>
      <div class="border-t border-gray-200">
        <table class="min-w-full divide-y-2 divide-gray-300">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="py-3.5 pl-4 pr-3 text-left sm:pl-6 lg:pl-8">Name</th>
              <th scope="col" class="px-3 py-3.5 text-left w-28">Version</th>
              <th scope="col" class="px-3 py-3.5 text-left w-28">Latest</th>
              <th scope="col" class="px-3 py-3.5 pr-4 text-right sm:pr-6 lg:pr-8">Known Issues</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr class="even:bg-gray-50" v-for="extension in shopStore.shop.extensions" :key="extension.name" :class="[
              extension.installed
                ? 'opacity-100'
                : 'opacity-40',
            ]">
              <td class="whitespace-nowrap py-4 pl-4 pr-3 align-middle sm:pl-6 lg:pl-8">
                <span class="text-green-400" data-tooltip="Active" v-if="extension.active">
                  <CheckCircleIcon class="w-5 inline " />
                </span>
                <span class="text-gray-300" data-tooltip="Deactive" v-if="!extension.active">
                  <XCircleIcon class="inline w-5" />
                </span>
                {{ extension.name }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                {{ extension.version }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                <span
                  class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1"
                  v-if="extension.latestVersion">
                  {{ extension.latestVersion }}
                </span>
              </td>
              <td class="whitespace-nowrap px-3 py-3.5 pr-4 text-right sm:pr-6 lg:pr-8">
                No known issues. <a href="#">Report issue</a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="mb-12 bg-white shadow rounded-md overflow-y-scroll md:overflow-hidden">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium ">Scheduled Tasks</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">List of all Scheduled Tasks</p>
      </div>
      <div class="border-t border-gray-200">
        <table class="min-w-full divide-y-2 divide-gray-300">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="py-3.5 pl-4 pr-3 text-left sm:pl-6 lg:pl-8">Name</th>
              <th scope="col" class="px-3 py-3.5 text-left">Last Executed</th>
              <th scope="col" class="px-3 py-3.5 text-left">Next Execution
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr class="even:bg-gray-50" v-for="task in shopStore.shop.scheduled_task" :key="task.name">
              <td class="whitespace-nowrap py-4 pl-4 pr-3 sm:pl-6 lg:pl-8">
                <span class="text-amber-600" data-tooltip="Task overdue" v-if="isTaskOverdue(task)">
                  <XCircleIcon class="inline w-5" />
                </span>
                <span class="text-green-400" data-tooltip="Working" v-else>
                  <CheckCircleIcon class="w-5 inline" />
                </span>                
                {{ task.name }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                {{ new Date(task.lastExecutionTime).toLocaleString() }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                {{ new Date(task.nextExecutionTime).toLocaleString() }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </MainContainer>
</template>
