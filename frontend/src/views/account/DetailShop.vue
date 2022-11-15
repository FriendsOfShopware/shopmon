<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';
import Tabs from '@/components/layout/Tabs.vue';
import Modal from '@/components/layout/Modal.vue';
import { compareVersions } from 'compare-versions';
import { sort } from 'fast-sort';

import { useShopStore } from '@/stores/shop.store';
import { useAlertStore } from '@/stores/alert.store';
import { useRoute } from 'vue-router';

import type { Extension, ExtensionDiff, ShopChangelog, ShopwareVersion } from '@apiTypes/shop';
import { ref } from 'vue';
import type { Ref } from 'vue';

import { fetchWrapper } from '@/helpers/fetch-wrapper';

import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaRocket from '~icons/fa6-solid/rocket';
import FaFileWaverform from '~icons/fa6-solid/file-waveform';

const route = useRoute();
const shopStore = useShopStore();
const alertStore = useAlertStore();

const viewChangelogDialog: Ref<boolean> = ref(false);
const dialogExtension: Ref<Extension | null> = ref(null);

const viewShopChangelogDialog: Ref<boolean> = ref(false);
const dialogShopChangelog: Ref<ShopChangelog | null> = ref(null);

const viewUpdateWizardDialog: Ref<boolean> = ref(false);
const loadingUpdateWizard: Ref<boolean> = ref(false);
const dialogUpdateWizard: Ref<ShopChangelog | null> = ref(null);

const showShopRefreshModal: Ref<boolean> = ref(false);
const shopwareVersions: Ref<ShopwareVersion[] | null> = ref(null);

const latestShopwareVersion: Ref<string|null> = ref(null);

function openExtensionChangelog(extension: Extension | null) {
  dialogExtension.value = extension;
  viewChangelogDialog.value = true;
}

function openShopChangelog(shopChangelog: ShopChangelog | null) {
  dialogShopChangelog.value = shopChangelog;
  viewShopChangelogDialog.value = true;
}

function openUpdateWizard() {
  dialogUpdateWizard.value = null;
  viewUpdateWizardDialog.value = true;
}

async function loadShop() {
  const teamId = parseInt(route.params.teamId as string, 10);
  const shopId = parseInt(route.params.shopId as string, 10);

  await shopStore.loadShop(teamId, shopId);
  const shopwareVersionsData = await fetchWrapper.get('/info/latest-shopware-version') as ShopwareVersion[];
  shopwareVersions.value = shopwareVersionsData.filter((entry: ShopwareVersion) => compareVersions(shopStore.shop.shopware_version, entry.version) < 0);
  latestShopwareVersion.value = shopwareVersions.value[0]?.version;
}

loadShop().then(function() {
  if ( shopStore?.shop?.name ) {
    document.title = shopStore.shop.name;
  }
});

async function onRefresh(pagespeed: boolean) {
  showShopRefreshModal.value = false;
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
        try {
            await shopStore.refreshShop(shopStore.shop.team_id, shopStore.shop.id, pagespeed);
            alertStore.success('Your Shop will refresh soon!');
        } catch (e: any) {
            alertStore.error(e);
        }
    }
}

async function onCacheClear() {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
    try {
      await shopStore.clearCache(shopStore.shop.team_id, shopStore.shop.id);
      alertStore.success('Your Shop cache was cleared successfully');
    } catch (e: any) {
      alertStore.error(e);
    }
  }
}

async function onReScheduleTask(taskId: string) {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
    try {
      await shopStore.reScheduleTask(shopStore.shop.team_id, shopStore.shop.id, taskId);
      alertStore.success('Task is re-scheduled');
    } catch (e: any) {
      alertStore.error(e);
    }
  }
}

async function loadUpdateWizard(version: string) {
  loadingUpdateWizard.value = true;

  const body = {
    "currentVersion": shopStore.shop?.shopware_version,
    "futureVersion": version,
    "plugins": shopStore.shop?.extensions
  }

  dialogUpdateWizard.value = await fetchWrapper.post('/info/check-extension-compatibility', body);

  loadingUpdateWizard.value = false;
}

async function ignoreCheck(id: string) {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
    shopStore?.shop?.ignores.push(id);
  
    shopStore.updateShop(shopStore.shop.team_id, shopStore.shop.id, {ignores: shopStore?.shop?.ignores});
    notificateIgnoreUpdate();
  }
}

async function removeIgnore(id: string) {
  if ( shopStore?.shop?.team_id && shopStore?.shop?.id ) {
    shopStore.shop.ignores = shopStore.shop.ignores.filter((aid: string) => aid !== id);
  
    shopStore.updateShop(shopStore.shop.team_id, shopStore.shop.id, {ignores: shopStore?.shop?.ignores});
    notificateIgnoreUpdate();
  }
}

async function notificateIgnoreUpdate() {
  alertStore.info('Ignore state updated. Will effect after next shop update');
}

function sumChanges(changes: ShopChangelog) {
  const messages: string[] = [];

  if (changes.old_shopware_version && changes.new_shopware_version) {
    messages.push(
      `Shopware Update from ${changes.old_shopware_version} to ${changes.new_shopware_version}`
    );
  }

  const stateCounts: Record<string, number> = {};
  for (const extension of changes.extensions) {
    if (stateCounts[extension.state] !== undefined) {
      stateCounts[extension.state] = stateCounts[extension.state] + 1;
    } else {
      stateCounts[extension.state] = 1;
    }
  }

  for (const [state, count] of Object.entries(stateCounts)) {
    messages.push(
      `${state} ${count} extension` + (count > 1 ? 's' : '')
    )
  }  

  return messages.join(', ');
}

</script>

<template>
  <Header v-if="shopStore.shop" :title="shopStore.shop.name">
    <div class="flex gap-2">
      <button class="btn" data-tooltip="Clear shop cache" @click="onCacheClear"
              :disabled="shopStore.isCacheClearing">
        <icon-ic:baseline-cleaning-services :class="{'animate-pulse': shopStore.isCacheClearing}" class="w-4 h-4" />
      </button>
      <button class="btn" data-tooltip="Refresh shop data" @click="showShopRefreshModal = true"
        :disabled="shopStore.isRefreshing">
        <icon-fa6-solid:rotate :class="{'animate-spin': shopStore.isRefreshing}" />
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
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg dark:bg-neutral-800 dark:shadow-none">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          <icon-fa6-solid:circle-xmark class="text-red-600 mr-1 dark:text-red-400 " v-if="shopStore.shop.status == 'red'" />
          <icon-fa6-solid:circle-info class="text-yellow-400 mr-1 dark:text-yellow-200" v-else-if="shopStore.shop.status === 'yellow'" />
          <icon-fa6-solid:circle-check class="text-green-400 mr-1 dark:text-green-300" v-else />
          Shop Information
        </h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8 dark:border-neutral-700 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3">
        <dl class="grid grid-cols-1 auto-rows-min gap-6 md:col-span-2 md:grid-cols-2">
          <div class="md:col-span-1">
            <dt class="text-sm font-medium">Shopware Version</dt>
            <dd class="mt-1 text-sm text-gray-500">
              {{ shopStore.shop.shopware_version }}
              <template v-if="latestShopwareVersion && latestShopwareVersion != shopStore.shop.shopware_version">
                <a class="badge badge-warning"  
                    :href="'https://github.com/shopware/platform/releases/tag/v' + latestShopwareVersion" target="_blank" >
                    {{ latestShopwareVersion }}
                </a>
                <button @click="openUpdateWizard" class="ml-2 badge badge-info">
                  <icon-fa6-solid:rotate/>
                  Update Wizard
                </button>
              </template>                
            </dd>
          </div>
          <div class="md:col-span-1">
            <dt class="font-medium">Team</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.team_name }}</dd>
          </div>
          <div class="md:col-span-1">
            <dt class="font-medium">Last Checked At</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ new Date(shopStore.shop.last_scraped_at).toLocaleString() }}</dd>
          </div>
          <div class="md:col-span-1">
            <dt class="font-medium">Environment</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.cache_info.environment }}</dd>
          </div>
          <div class="md:col-span-1">
            <dt class="font-medium">HTTP Cache</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ shopStore.shop.cache_info.httpCache ? 'Enabled' : 'Disabled' }}
            </dd>
          </div>
          <div class="md:col-span-1">
            <dt class="font-medium">URL</dt>
            <dd class="mt-1 text-sm text-gray-500">
              <a :href="shopStore.shop.url" data-tooltip="Go to storefront" target="_blank">
                <icon-fa6-solid:store /> Storefront
              </a>
              &nbsp;/&nbsp;
              <a :href="shopStore.shop.url + '/admin'" data-tooltip="Go to shopware admin" target="_blank">
                <icon-fa6-solid:shield-halved /> Admin
              </a>
            </dd>
          </div>
        </dl>
        <div class="mt-6 sm:mt-0 sm:col-span-1 sm:row-span-full sm:col-start-2 sm:row-start-1 md:col-start-3 md:row-span-full md:row-start-1">
          <img :src="`/api/team/${shopStore.shop.shop_image}`" 
                class="h-[200px] sm:h-[400px] md:h-[200px]" 
                v-if="shopStore.shop.shop_image">
          <icon-fa6-solid:image v-else class="text-gray-100 text-9xl" />
        </div>
      </div>
    </div>

    <Tabs :labels="{
      checks: {title: 'Checks', count: shopStore.shop.checks.length, icon: FaCircleCheck}, 
      extensions: {title: 'Extensions', count: shopStore.shop.extensions.length, icon: FaPlug}, 
      tasks: {title: 'Scheduled Tasks', count: shopStore.shop.scheduled_task.length, icon: FaListCheck}, 
      queue: {title: 'Queue', count: shopStore.shop.queue_info.length, icon: FaCircleCheck},
      pagespeed: {title: 'Pagespeed', count: shopStore.shop.pagespeed.length, icon: FaRocket},
      changelog: {title: 'Changelog', count: shopStore.shop.changelog.length, icon: FaFileWaverform}
    }">

      <template #panel(checks)="{ label }">
        <DataTable
          :labels="{message: {name: 'Message'}, actions: {name: 'Ignore', class: 'px-3 text-right'}}"
          :data="shopStore.shop.checks">
          <template #cell(message)="{ item }">
            <icon-fa6-solid:circle-xmark class="text-red-600 mr-2 text-base dark:text-red-400 " v-if="item.level == 'red'" />
            <icon-fa6-solid:circle-info class="text-yellow-400 mr-2 text-base dark:text-yellow-200 " v-else-if="item.level === 'yellow'" />
            <icon-fa6-solid:circle-check  class="text-green-400 mr-2 text-base dark:text-green-300" v-else />
            <a :href="item.link" target="_blank" v-if="item.link">
              {{ item.message }} 
              <icon-fa6-solid:up-right-from-square  class="text-xs"/>
            </a>
            <template v-else>
              {{ item.message }} 
            </template>
          </template>
          <template #cell(actions)="{ item }">
            <button @click="removeIgnore(item.id)" v-if="shopStore.shop.ignores.includes(item.id)" data-tooltip="check is ignored" class="text-red-600 opacity-25 tooltip-position-left dark:text-red-400 group-hover:opacity-100">
              <icon-fa6-solid:eye-slash />
            </button>
            <button @click="ignoreCheck(item.id)" v-else data-tooltip="check used" class="opacity-25 tooltip-position-left group-hover:opacity-100">
              <icon-fa6-solid:eye />
            </button>
          </template>
        </DataTable>
      </template> 

      <template #panel(extensions)="{ label }">
        <DataTable
          :labels="{name: {name: 'Name', sortable: true}, version: {name: 'Version'}, latest: {name: 'Latest'}, ratingAverage: {name: 'Rating', sortable: true}, installedAt: {name: 'Installed at', sortable: true}, issue: {name: 'Known Issue', class: 'px-3 text-right'}}"
          :data="shopStore.shop.extensions">
          <template #cell(name)="{ item }">
            <div class="flex items-start">
              <span class="leading-5 text-gray-400 mr-2 text-base dark:text-neutral-500" data-tooltip="Not installed" v-if="!item.installed">
                <icon-fa6-regular:circle />
              </span>
              <span class="leading-5 text-green-400 mr-2 text-base dark:text-green-300" data-tooltip="Active" v-else-if="item.active">
                <icon-fa6-solid:circle-check />
              </span>
              <span class="leading-5 text-gray-300 mr-2 text-base dark:text-neutral-500" data-tooltip="Inactive" v-else>
                <icon-fa6-solid:circle-xmark />
              </span>
              <div v-if="item.storeLink">
                <a :href="item.storeLink" target="_blank">
                  <div class="font-bold whitespace-normal" v-if="item.label">{{ item.label }}</div>
                  <div class="dark:opacity-50">{{ item.name }}</div>
                </a>
              </div>
              <div v-else>
                <div class="font-bold whitespace-normal" v-if="item.label">{{ item.label }}</div>
                <div class="dark:opacity-50">{{ item.name }}</div>
              </div>
            </div>
            
            
          </template>

          <template #cell(latest)="{ item }">
            <button
              class="badge badge-warning"
              v-if="item.latestVersion && item.version < item.latestVersion" @click="openExtensionChangelog(item)">
              {{ item.latestVersion }}
            </button>
          </template>

          <template #cell(ratingAverage)="{ item }">
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
          :labels="{name: {name: 'Name', sortable: true}, interval: {name: 'Interval'}, lastExecutionTime: {name: 'Last Executed', sortable: true}, nextExecutionTime: {name: 'Next Execution', sortable: true}, status: {name: 'Status'}, actions: {name: '', class: 'px-3 text-right w-16'}}"
          :data="shopStore.shop.scheduled_task"
          :default-sorting="{by: 'nextExecutionTime'}">          

          <template #cell(lastExecutionTime)="{ item }">
            {{ new Date(item.lastExecutionTime).toLocaleString() }}
          </template>

          <template #cell(nextExecutionTime)="{ item }">
            {{ new Date(item.nextExecutionTime).toLocaleString() }}
          </template>

          <template #cell(status)="{ item }">
            <span class="pill pill-success" v-if="item.status === 'scheduled' && !item.overdue" >
              <icon-fa6-solid:check />
              {{ item.status }}
            </span>
            <span class="pill pill-info" v-else-if="item.status === 'queued' || item.status === 'running' && !item.overdue" >
                <icon-fa6-solid:rotate />
              {{ item.status }}
            </span>
            <span class="pill pill-warning" v-else-if="item.status === 'scheduled' || item.status === 'queued' || item.status === 'running' && item.overdue">
                <icon-fa6-solid:info class="align-[-0.1em]" />
              {{ item.status }}
            </span>
            <span class="pill" v-else-if="item.status === 'inactive'">
                <icon-fa6-solid:pause />
              {{ item.status }}
            </span>
            <span class="pill pill-error" v-else>
                <icon-fa6-solid:xmark />
              {{ item.status }}
            </span>
          </template>

          <template #cell(actions)="{ item }">
            <span class="cursor-pointer opacity-25 hover:opacity-100 tooltip-position-left"
                  data-tooltip="Re-schedule task" 
                  @click="onReScheduleTask(item.id)"
            >
              <icon-fa6-solid:arrow-rotate-right class="text-base leading-3" />
            </span>
            
          </template>
        </DataTable>
      </template>

      <template #panel(queue)="{ label }">
        <DataTable
          :labels="{name: {name: 'Name', sortable: true}, size: {name: 'Size', sortable: true}}"
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

          <template v-slot:cell="{ item, data, itemKey }" v-for="(cell, cellKey) in {'performance': 'cell(performance)', 'accessibility': 'cell(accessibility)', 'bestpractices': 'cell(bestpractices)', 'seo': 'cell(seo)'}">
            <span class="mr-2">{{ item[cellKey] }}</span>
            <template v-if="data[(itemKey + 1)] && data[(itemKey + 1)][cellKey] !== item[cellKey]">
              <icon-fa6-solid:arrow-right 
                :class="[{
                  'text-green-400 -rotate-45 dark:text-green-300': data[(itemKey + 1)][cellKey] < item[cellKey],
                  'text-red-600 rotate-45 dark:text-red-400': data[(itemKey + 1)][cellKey] > item[cellKey]
                }]"
              />
            </template>
            <icon-fa6-solid:minus v-else />
          </template>

        </DataTable>
      </template> 

      <template #panel(changelog)="{ label }">
        <DataTable
          :labels="{date: {name: 'Date', sortable: true}, changes: {name: 'Changes'}}"
          :data="shopStore.shop.changelog">

          <template #cell(date)="{ item }">
            {{ new Date(item.date).toLocaleString() }}
          </template>

          <template #cell(changes)="{ item }">
            <span @click="openShopChangelog(item)" class="cursor-pointer">{{ sumChanges(item) }}</span>
          </template>

        </DataTable>
      </template>

    </Tabs>

    <Modal :show="viewChangelogDialog" :closeXMark="true" @close="viewChangelogDialog = false">
      <template #title>Changelog - <span class="font-normal">{{ dialogExtension?.name }}</span></template>
      <template #content>
        <ul v-if="dialogExtension?.changelog && dialogExtension.changelog.length > 0">
          <li class="mb-2" v-for="changeLog in dialogExtension.changelog" :key="changeLog.version">
            <div class="font-medium mb-1">
              <span data-tooltip="not compatible with your version" v-if="!changeLog.isCompatible">
                <icon-fa6-solid:circle-info class="text-yellow-400 text-base dark:text-yellow-200" />
              </span>
              {{ changeLog.version }} - 
              <span class="text-xs font-normal text-gray-500">
                {{ new Date(changeLog.creationDate).toLocaleDateString() }}
              </span>
            </div>
            <div class="pl-3" v-html="changeLog.text"></div>
          </li>
        </ul>
        <div class="rounded-md bg-red-50 p-4 border border-red-200" v-else>
          <div class="flex">
            <div class="flex-shrink-0">
              <icon-fa6-solid:circle-xmark class="h-5 w-5 text-red-600 dark:text-red-400 "
                aria-hidden="true" />
            </div>
            <div class="ml-3 text-red-900">
              No Changelog data provided
            </div>
          </div>
        </div>
      </template>
    </Modal>

    <Modal :show="viewShopChangelogDialog" :closeXMark="true" @close="viewShopChangelogDialog = false">
      <template #title>Shop changelog - <span class="font-normal" v-if="dialogShopChangelog?.date">{{ new Date(dialogShopChangelog.date).toLocaleString() }}</span></template>
      <template #content>
        <template v-if="dialogShopChangelog?.old_shopware_version && dialogShopChangelog?.new_shopware_version">
          Shopware update from <strong>{{ dialogShopChangelog.old_shopware_version }}</strong> to <strong>{{ dialogShopChangelog.new_shopware_version }}</strong>
        </template>

        <div class="mt-4" v-if="dialogShopChangelog?.extensions?.length">
          <h2 class="text-lg mb-1 font-medium">Shop Plugin Changelog:</h2>
          <ul class="list-disc">
            <li class="ml-4 mb-1" v-for="extension in dialogShopChangelog?.extensions" :key="extension.name">
              <strong>{{ extension.label }}</strong> <span class="opacity-60">({{ extension.name }})</span> {{ extension.state }}
              <template>
                {{ extension.old_version }} {{ extension.new_version }}
              </template>
              <template v-if="extension.old_version && extension.new_version">
                from {{ extension.old_version }} to {{ extension.new_version }}
              </template>
            </li>
          </ul>
        </div>
      </template>
    </Modal>

    <Modal :show="viewUpdateWizardDialog" :closeXMark="true" @close="viewUpdateWizardDialog = false">
      <template #title><icon-fa6-solid:rotate /> Shopware Update Wizard</template>
      <template #content>
        <select class="field mb-2" @change="event => loadUpdateWizard((event.target as HTMLSelectElement).value)">
          <option disabled selected>Select update Version</option>
          <option v-for="data in shopwareVersions" :key="data.version">{{ data.version }}</option>
        </select>

        <template v-if="loadingUpdateWizard">
          <div class="text-center">
            Loading <icon-fa6-solid:rotate class="animate-spin" />
          </div>          
        </template>
        
        <div v-if="dialogUpdateWizard" :class="{'opacity-20': loadingUpdateWizard}">
          <h2 class="text-lg mb-1 font-medium">Plugin Compatibility</h2>

          <ul>
            <li class="p-2 odd:bg-gray-100 dark:odd:bg-[#2e2e2e]" v-for="extension in sort(shopStore.shop.extensions).by([{desc: u => u.active}, {asc: u => u.label}])" :key="extension.name">
              <template v-if="!dialogUpdateWizard.find((plugin: Extension) => plugin.name === extension.name)">
                <div class="flex">
                  <div class="mr-2 w-4">
                    <icon-fa6-regular:circle class="text-base text-gray-400 dark:text-neutral-500" v-if="!extension.active" />
                    <icon-fa6-solid:circle-info class="text-base text-yellow-400 dark:text-yellow-200" v-else />
                  </div>                  
                  <div>
                    <strong>{{ extension.label }}</strong> <span class="opacity-60">({{ extension.name }})</span>
                    <div >This plugin is not available in the Store. Please contact the plugin manufacturer.</div>
                  </div>
                </div>
              </template>
              <template v-else v-for="compatibility in dialogUpdateWizard.filter((plugin: Extension) => plugin.name === extension.name)">
                <div class="flex">
                  <div class="mr-2 w-4">
                    <icon-fa6-regular:circle class="text-base text-gray-400 dark:text-neutral-500" v-if="!extension.active" />
                    <icon-fa6-solid:circle-xmark class="text-base text-red-600 dark:text-red-400" v-else-if="compatibility.status.type == 'red'" />
                    <icon-fa6-solid:circle-check  class="text-base text-green-400 dark:text-green-300" v-else />
                  </div>    
                  <div>
                    <strong>{{ extension.label }}</strong> <span class="opacity-60">({{ extension.name }})</span>
                    <div >{{ compatibility.status.label }}</div>
                  </div>
                </div>              
              </template>
            </li>
          </ul>
        </div>
        
      </template>
    </Modal>

    <Modal :show="showShopRefreshModal" :closeXMark="true" @close="showShopRefreshModal = false">
        <template #icon><icon-fa6-solid:rotate class="h-6 w-6 text-sky-500" aria-hidden="true" /></template>
        <template #title>Refresh {{ shopStore.shop.name }}</template>
        <template #content>                
          Do you also want to have a new pagespeed test?
        </template>
        <template #footer>
            <button type="button" class="btn w-full sm:ml-3 sm:w-auto"
                @click="onRefresh(true)">
                Yes
            </button>
            <button ref="cancelButtonRef" type="button"
                class="btn w-full mt-3 sm:w-auto sm:mt-0"
                @click="onRefresh(false)">
                No
            </button>
        </template>
    </Modal>

  </MainContainer>
</template>
