<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { useShopStore } from '@/stores/shop.store';
import { useRoute } from 'vue-router';
import type { ScheduledTask } from '@apiTypes/shop'; 

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
  <Header v-if="shopStore.shop" :title="shopStore.shop.name"/>
  <MainContainer v-if="shopStore.shop">
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">Shop Information</h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6">
        <dl class="grid grid-cols-1 gap-x-4 gap-y-8 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">Shopware Version</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ shopStore.shop.shopware_version }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">Team</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ shopStore.shop.team_name }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">Last Checked At</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ new Date(shopStore.shop.last_scraped_at).toLocaleString()}}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">Environment</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ shopStore.shop.cache_info.environment }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-gray-500">HTTP Cache</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ shopStore.shop.cache_info.httpCache ? 'Enabled' : 'Disabled' }}</dd>
          </div>
        </dl>
      </div>
    </div>

    <div
      class="hidden sm:block"
      aria-hidden="true"
    >
      <div class="py-5">
        <div class="border-t border-gray-200" />
      </div>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Extensions</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">List of all installed extensions</p>
        </div>
        <div class="border-t border-gray-200">
          <div class="inline-block min-w-full align-middle">
            <div class="overflow-hidden shadow-sm ring-1 ring-black ring-opacity-5">
              <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-50">
                  <tr>
                    <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6 lg:pl-8">Name</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Version</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Known Issues</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr v-for="extension in shopStore.shop.extensions" :key="extension.name">
                    <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8">
                      {{ extension.name }}
                      <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-green-600 bg-green-200 uppercase last:mr-0 mr-1" v-if="extension.installed">
                        Installed
                      </span>
                      <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-green-600 bg-green-200 uppercase last:mr-0 mr-1" v-if="extension.active">
                        Active
                      </span>
                    </td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {{ extension.version }}
                      <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1" v-if="extension.latestVersion">
                        Newer version: {{ extension.latestVersion }}
                      </span>
                    </td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      No known issues. <a href="#" class="text-indigo-600 hover:text-indigo-500">Report issue</a>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      class="hidden sm:block"
      aria-hidden="true"
    >
      <div class="py-5">
        <div class="border-t border-gray-200" />
      </div>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Scheduled Tasks</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">List of all Scheduled Tasks</p>
        </div>
        <div class="border-t border-gray-200">
          <div class="inline-block min-w-full align-middle">
            <div class="overflow-hidden shadow-sm ring-1 ring-black ring-opacity-5">
              <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-50">
                  <tr>
                    <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6 lg:pl-8">Name</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Last Executed</th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Next Execution</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr v-for="task in shopStore.shop.scheduled_task" :key="task.name">
                    <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8">
                      {{ task.name }}
                      <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1" v-if="isTaskOverdue(task)">
                        Task overdue
                      </span>
                      <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-green-600 bg-green-200 uppercase last:mr-0 mr-1" v-else>
                        Working
                      </span>
                    </td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {{ new Date(task.lastExecutionTime).toLocaleString() }}
                    </td>
                    <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {{ new Date(task.nextExecutionTime).toLocaleString() }}
                    </td>
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
