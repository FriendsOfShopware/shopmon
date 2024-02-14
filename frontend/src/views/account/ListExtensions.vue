<script setup lang="ts">
import { ref } from 'vue'
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

import { useExtensionStore } from '@/stores/extension.store';

const extensionStore = useExtensionStore();

const term = ref('');

extensionStore.loadExtensions();

</script>
  
  <template>
  <Header title="My Apps">
  </Header>
  
  <MainContainer v-if="!extensionStore.isLoading">
    <div class="text-center" v-if="extensionStore.extensions.length === 0">
      <svg class="mx-auto h-12 w-12 text-gray-400 dark:text-neutral-500" fill="none" viewBox="0 0 24 24"
        stroke="currentColor" aria-hidden="true">
        <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      </svg>
      <h3 class="mt-2 font-medium">No Apps</h3>
      <p class="mt-1 text-gray-500 dark:text-neutral-500">Get started by adding your first Shop.</p>
      <div class="mt-6">
        <router-link to="/account/shops/new" class="btn btn-primary group flex items-center">
          <icon-fa6-solid:plus class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50" aria-hidden="true" />
          Add Shop
        </router-link>
      </div>
    </div>

    <template v-else>
      <input v-model="term" class="field field-search mb-3" placeholder="Search ..." />

      <div class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden bg-white dark:bg-neutral-800 dark:shadow-none">
        <DataTable
            :labels="{label: {name: 'Name', sortable: true}, shop: {name: 'Shop'}, version: {name: 'Version'}, latestVersion: {name: 'Latest'}, ratingAverage: {name: 'Rating', sortable: true}, issue: {name: 'Known Issue', class: 'px-3 text-right'}}"
            :data="extensionStore.extensions"
            :default-sorting="{by: 'label'}"
            :term="term">
            <template #cell(label)="{ item }">            
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
            </template>

            <template #cell(shop)="{ item }">
              <div class="mb-1" v-for="shop in item.shops">
                <router-link :to="{ name: 'account.shops.detail', params: { organizationId: shop.organizationId, shopId: shop.id } }">
                  <span class="leading-5 text-gray-400 mr-1 text-base dark:text-neutral-500" data-tooltip="Not installed" v-if="!shop.installed">
                    <icon-fa6-regular:circle />
                  </span>
                  <span class="leading-5 text-green-400 mr-1 text-base dark:text-green-300" data-tooltip="Active" v-else-if="shop.active">
                    <icon-fa6-solid:circle-check />
                  </span>
                  <span class="leading-5 text-gray-300 mr-1 text-base dark:text-neutral-500" data-tooltip="Inactive" v-else>
                    <icon-fa6-solid:circle-xmark />
                  </span>
                  {{ shop.name }}
                </router-link>
              </div>
            </template>

            <template #cell(version)="{ item }">
              <div class="mb-1" v-for="shop in item.shops">
                {{ shop.version }}
                <span v-if="item.latestVersion && shop.version < item.latestVersion" data-tooltip="Update available">
                  <icon-fa6-solid:rotate class="ml-1 text-base text-amber-600 dark:text-amber-400" />
                </span>
              </div>
            </template>

            <template #cell(ratingAverage)="{ item }">
              <RatingStars :rating="item.ratingAverage" />
            </template>

            <template #cell(installedAt)="{ item }">
              <template v-if="item.installedAt">
                {{ formatDateTime(item.installedAt) }}
              </template>
            </template>

            <template #cell(issue)="{ item }">
              No known issues. <a href="#">Report issue</a>
            </template>
          </DataTable>
      </div>
    </template>
    
  </MainContainer>
</template>
