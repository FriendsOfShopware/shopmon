<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';
import Tabs from '@/components/layout/Tabs.vue';

import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';

import type { Extension } from '@apiTypes/shop';
import { ref } from 'vue';
import type { Ref } from 'vue';

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { fetchWrapper } from '@/helpers/fetch-wrapper';

import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaRocket from '~icons/fa6-solid/rocket';

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const viewChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);
const latestShopwareVersion: Ref<string|null> = ref(null);

function openExtensionChangelog(extension: Extension | null) {
  dialogExtension.value = extension;
  viewChangelogDialog.value = true;  
}

async function loadShop() {
  const teamId = parseInt(route.params.teamId as string, 10);
  const shopId = parseInt(route.params.shopId as string, 10);

  shopStore.loadShop(teamId, shopId);
  latestShopwareVersion.value = await fetchWrapper.get('/info/latest-shopware-version') as string;
}

loadShop();

async function onRefresh() {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
        try {
            await shopStore.refreshShop(shopStore.shop.team_id, shopStore.shop.id);
            alertStore.success('Your Shop will refresh soon!');
            setTimeout(() => {
              loadShop();
            }, 5000);
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
        <icon-fa6-solid:rotate :class="{'animate-spin': shopStore.isRefreshing}" size="lg" />
      </button>

      <router-link
        :to="{ name: 'account.shops.edit', params: { teamId: shopStore.shop.team_id, shopId: shopStore.shop.id } }"
        type="button" class="group btn btn-primary flex items-center">
        <icon-fa6-solid:pencil class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
          aria-hidden="true" />
        Edit Shop
      </router-link>
    </div>
  </Header>

  <MainContainer v-if="shopStore.shop && shopStore.shop.last_scraped_at">
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          <icon-fa6-solid:circle-xmark class="text-red-600 mr-1" v-if="shopStore.shop.status == 'red'" />
          <icon-fa6-solid:circle-info class="text-yellow-400 mr-1" v-else-if="shopStore.shop.status === 'yellow'" />
          <icon-fa6-solid:circle-check class="text-green-400 mr-1" v-else />
          Shop Information
        </h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8">
        <dl class="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium">Shopware Version</dt>
            <dd class="mt-1 text-sm text-gray-500">
              {{ shopStore.shop.shopware_version }}
              <span class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1" v-if="latestShopwareVersion && latestShopwareVersion != shopStore.shop.shopware_version">{{ latestShopwareVersion }}</span>
            </dd>
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
              &nbsp;/&nbsp;
              <a :href="shopStore.shop.url + '/admin'" data-tooltip="Shopware admin URL" target="_blank">
                admin
              </a>
            </dd>
          </div>
        </dl>
      </div>
    </div>

    <Tabs :labels="{
      checks: {title: 'Checks', count: shopStore.shop.checks.length, icon: FaCircleCheck}, 
      extensions: {title: 'Extensions', count: shopStore.shop.extensions.length, icon: FaPlug}, 
      tasks: {title: 'Scheduled Tasks', count: shopStore.shop.scheduled_task.length, icon: FaListCheck}, 
      queue: {title: 'Queue', count: shopStore.shop.queue_info.length, icon: FaCircleCheck},
      pagespeed: {title: 'Pagespeed', count: shopStore.shop.pagespeed.length, icon: FaRocket}
    }">

      <template #panel(checks)="{ label }">
        <DataTable
          :labels="{message: {name: 'Message'}}"
          :data="shopStore.shop.checks">
          <template #cell(message)="{ item }">
            <icon-fa6-solid:circle-xmark class="text-red-600 mr-1 text-base" v-if="item.level == 'red'" />
            <icon-fa6-solid:circle-info class="text-yellow-400 mr-1 text-base" v-else-if="item.level === 'yellow'" />
            <icon-fa6-solid:circle-check  class="text-green-400 mr-1 text-base" v-else />
            <a :href="item.link" target="_blank" v-if="item.link">
              {{ item.message }} 
              <icon-fa6-solid:up-right-from-square  class="text-xs"/>
            </a>
            <template v-else>
              {{ item.message }} 
            </template>
          </template>
        </DataTable>
      </template> 

      <template #panel(extensions)="{ label }">
        <DataTable
          :labels="{name: {name: 'Name'}, version: {name: 'Version'}, latest: {name: 'Latest'}, rating: {name: 'Rating'}, storeLink: {name: 'Store'}, installedAt: {name: 'Installed at'}, issue: {name: 'Known Issue', class: 'px-3 text-right'}}"
          :data="shopStore.shop.extensions">
          <template #cell(name)="{ item }">
            <icon-fa6-regular:circle class="text-gray-400 mr-1 text-base" data-tooltip="Not installed" v-if="!item.installed" />
            <icon-fa6-solid:circle-check class="text-green-400 mr-1 text-base" data-tooltip="Active" v-else-if="item.active" />
            <icon-fa6-solid:circle-xmark class="text-gray-300 mr-1 text-base" data-tooltip="Deactive" v-else />
            {{ item.name }}
          </template>

          <template #cell(latest)="{ item }">
            <button
              class="text-xs font-semibold inline-block py-1 px-2 uppercase rounded text-amber-600 bg-amber-200 uppercase last:mr-0 mr-1"
              v-if="item.latestVersion && item.version < item.latestVersion" @click="openExtensionChangelog(item)">
              {{ item.latestVersion }}
            </button>
          </template>

          <template #cell(rating)="{ item }">
            <template v-if="item.ratingAverage !== null" v-for="n in 5">
              <icon-fa6-regular:star v-if="(item.ratingAverage / 2) - n < -.5" />
              <icon-fa6-solid:star-half-stroke
                v-else-if="(item.ratingAverage / 2) - n === -.5" />
              <icon-fa6-solid:star v-else />
            </template>
          </template>

          <template #cell(installedAt)="{ item }">
            <template v-if="item.installedAt">
              {{ new Date(item.installedAt).toLocaleString() }}
            </template>
          </template>

          <template #cell(storeLink)="{ item }">
            <a :href="item.storeLink" :data-tooltip="item.storeLink" target="_blank" v-if="item.storeLink">
              <icon-fa6-solid:up-right-from-square/>
              </a>
          </template>

          <template #cell(issue)="{ item }">
            No known issues. <a href="#">Report issue</a>
          </template>
        </DataTable>
      </template>

      <template #panel(tasks)="{ label }">
        <DataTable
          :labels="{name: {name: 'Name'}, last: {name: 'Last Executed'}, next: {name: 'Next Execution'}}"
          :data="shopStore.shop.scheduled_task">
          <template #cell(name)="{ item }">
            <icon-fa6-solid:circle-xmark class="text-red-600 mr-1 text-base" data-tooltip="Task overdue" v-if="item.overdue"/>
            <icon-fa6-solid:circle-check class="text-green-400 mr-1 text-base" data-tooltip="Working" v-else/>
            {{ item.name }}
          </template>

          <template #cell(last)="{ item }">
            {{ new Date(item.lastExecutionTime).toLocaleString() }}
          </template>

          <template #cell(next)="{ item }">
            {{ new Date(item.nextExecutionTime).toLocaleString() }}
          </template>
        </DataTable>
      </template>

      <template #panel(queue)="{ label }">
        <DataTable
          :labels="{name: {name: 'Name'}, size: {name: 'Size'}}"
          :data="shopStore.shop.queue_info">
        </DataTable>
      </template>

      <template #panel(pagespeed)="{ label }">
        <DataTable
          :labels="{created: {name: 'Checked At'}, performance: {name: 'Performance'}, accessibility: {name: 'Accessibility'}, bestpractices: {name: 'Best Practices'}, seo: {name: 'SEO'}}"
          :data="shopStore.shop.pagespeed">
          <template #cell(created)="{ item }">
            <a target="_blank" :href="'https://web.dev/measure/?url='+ shopStore.shop.url">{{ new Date(item.created_at).toLocaleString() }}</a>
          </template>
        </DataTable>
      </template> 

    </Tabs>

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

                <button class="absolute top-1 right-2.5 focus:outline-none text-base" @click="viewChangelogDialog = false">
                  <icon-fa6-solid:xmark aria-hidden="true" size="xl"/>
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
                      <icon-fa6-solid:circle-xmark class="h-5 w-5 text-red-600"
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
