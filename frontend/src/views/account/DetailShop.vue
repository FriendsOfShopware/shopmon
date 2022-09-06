<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';
import type { Extension, ExtensionChangelog, ScheduledTask } from '@apiTypes/shop';
import { ref } from 'vue';
import type { Ref } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const viewChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);

shopStore.loadShop(route.params.teamId as string, route.params.shopId as string);

function openExtensionChangelog(extension: Extension | null) {
  dialogExtension.value = extension;
  viewChangelogDialog.value = true;  
}

async function onRefresh() {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
        try {
            await shopStore.refreshShop(shopStore.shop.team_id, shopStore.shop.id);
            alertStore.success('Your Shop will refresh soon!')
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

</script>

<template>
  <Header v-if="shopStore.shop" :title="shopStore.shop.name">
    <div class="flex gap-2">
      <button class="btn flex items-center" data-tooltip="Refresh shop data" @click="onRefresh"
        :disabled="shopStore.isRefreshing">
        <font-awesome-icon :class="{'animate-spin': shopStore.isRefreshing}" icon="fa-solid fa-rotate" size="lg" />
      </button>
      <router-link
        :to="{ name: 'account.shops.edit', params: { teamId: shopStore.shop.team_id, shopId: shopStore.shop.id } }"
        type="button" class="group btn btn-primary flex items-center">
        <font-awesome-icon icon="fa-solid fa-pencil" class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
          aria-hidden="true" />
        Edit Shop
      </router-link>
    </div>
  </Header>
  <MainContainer v-if="shopStore.shop && shopStore.shop.last_scraped_at">
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          <font-awesome-icon icon="fa-solid fa-circle-info" class="mr-1" />
          Shop Information
        </h3>
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
          <div class="sm:col-span-1">
            <dt class="font-medium">URL</dt>
            <dd class="mt-1 text-sm text-gray-500">
              <a :href="shopStore.shop.url" target="_blank">
                {{ shopStore.shop.url }}
              </a>
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <div class="mb-12 bg-white shadow rounded-md overflow-y-scroll md:overflow-hidden" v-if="shopStore.shop.extensions">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          <font-awesome-icon icon="fa-solid fa-plug" class="mr-1" />
          Extensions
        </h3>
        <p class="mt-1 text-sm text-gray-500">List of all installed extensions</p>
      </div>
      <div class="border-t border-gray-200">
        <table class="min-w-full divide-y-2 divide-gray-300">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="py-3.5 pl-4 pr-3 text-left sm:pl-6 lg:pl-8">Name</th>
              <th scope="col" class="px-3 py-3.5 text-left w-28">Version</th>
              <th scope="col" class="px-3 py-3.5 text-left w-28">Latest</th>
              <th scope="col" class="px-3 py-3.5 text-left w-28">Rating</th>
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
                <span class="text-green-400 mr-1" data-tooltip="Active" v-if="extension.active">
                  <font-awesome-icon size="lg" icon="fa-solid fa-circle-check" />
                </span>
                <span class="text-gray-300 mr-1" data-tooltip="Deactive" v-if="!extension.active">
                  <font-awesome-icon size="lg" icon="fa-solid fa-circle-xmark" />
                </span>
                {{ extension.name }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                {{ extension.version }}
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                <button
                  class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1"
                  v-if="extension.latestVersion && extension.version < extension.latestVersion"
                  @click="openExtensionChangelog(extension)">
                  {{ extension.latestVersion }}
                </button>
              </td>
              <td class="whitespace-nowrap px-3 py-4">
                <template v-if="extension.ratingAverage !== null" v-for="n in 5">
                  <font-awesome-icon icon="fa-regular fa-star" v-if="(extension.ratingAverage / 2) - n < -.5" />
                  <font-awesome-icon icon="fa-regular fa-star-half-stroke"
                    v-else-if="(extension.ratingAverage / 2) - n === -.5" />
                  <font-awesome-icon icon="fa-solid fa-star" v-else />
                </template>
              </td>
              <td class="whitespace-nowrap px-3 py-3.5 pr-4 text-right sm:pr-6 lg:pr-8">
                No known issues. <a href="#">Report issue</a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="mb-12 bg-white shadow rounded-md overflow-y-scroll md:overflow-hidden"
      v-if="shopStore.shop.scheduled_task">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium ">
          <font-awesome-icon icon="fa-solid fa-list-check" class="mr-1" />
          Scheduled Tasks
        </h3>
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
                <span class="text-red-600 mr-1" data-tooltip="Task overdue" v-if="task.overdue">
                  <font-awesome-icon size="lg" icon="fa-solid fa-circle-xmark" />
                </span>
                <span class="text-green-400 mr-1" data-tooltip="Working" v-else>
                  <font-awesome-icon size="lg" icon="fa-solid fa-circle-check" />
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

    <TransitionRoot as="template" :show="viewChangelogDialog">
      <Dialog as="div" class="relative z-10" @close="viewChangelogDialog = false">
        <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100"
          leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div class="fixed z-10 inset-0 overflow-y-auto">
          <div class="flex items-end sm:items-center justify-center min-h-full p-4 text-center sm:p-0">
            <TransitionChild as="template" enter="ease-out duration-300"
              enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200"
              leave-from="opacity-100 translate-y-0 sm:scale-100"
              leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
              <DialogPanel
                class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">

                <button class="absolute top-1 right-2.5 focus:outline-none" @click="viewChangelogDialog = false">
                  <font-awesome-icon aria-hidden="true" size="xl" icon="fa-solid fa-xmark" />
                </button>

                <DialogTitle as="h3" class="text-lg leading-6 font-medium text-gray-900 mb-3">
                  Changelog - <span class="font-normal">{{ dialogExtension?.name }}</span>
                </DialogTitle>

                <ul v-if="dialogExtension?.changelog && dialogExtension.changelog.length > 0">
                  <li class="mb-2" v-for="changeLog in dialogExtension.changelog" :key="changeLog.version">
                    <div class="font-medium mb-1">{{ changeLog.version }} - <span
                        class="text-xs font-normal text-gray-500">{{ new
                        Date(changeLog.creationDate).toLocaleDateString() }}</span></div>
                    <div class="pl-3" v-html="changeLog.text"></div>
                  </li>
                </ul>
                <div class="rounded-md bg-red-50 p-4 border border-red-200" v-else>
                  <div class="flex">
                    <div class="flex-shrink-0">
                      <font-awesome-icon icon="fa-solid fa-circle-xmark" class="h-5 w-5 text-red-600"
                        aria-hidden="true" />
                    </div>
                    <div class="ml-3 text-red-900">
                      No Changelog data provided
                    </div>
                  </div>
                </div>

              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
  </MainContainer>
</template>
