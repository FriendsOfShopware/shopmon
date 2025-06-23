<template>
    <div v-if="shops">
        <header-container title="Dashboard"/>
        <div class="panel">
            <h2 class="panel-title">
                <icon-fa6-solid:shop/>
                My Shops
            </h2>

            <div class="item-grid">
                <div v-for="shop in shops" :key="shop.id" class="item">
                    <router-link
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                slug: shop.organizationSlug,
                                shopId: shop.id
                            }
                        }" class="item-link item-wrapper"
                    >
                        <div class="item-logo">
                            <img v-if="shop.favicon" :src="shop.favicon" alt="Shop Logo" class="item-logo-img">
                        </div>

                        <div class="item-info">
                            <div class="item-name">
                                {{ shop.name }}
                            </div>

                            <div class="item-content">
                                {{ shop.projectName }}<br>
                                {{ shop.shopwareVersion }}
                            </div>

                            <div class="item-state">
                                <status-icon :status="shop.status"/>
                            </div>
                        </div>
                    </router-link>
                </div>
            </div>
        </div>

        <div class="panel">
            <h2 class="panel-title">
                <icon-fa6-solid:building/>
                My Organizations
            </h2>

            <div class="item-grid">
                <div v-for="organization in organizations" :key="organization.id" class="item">
                    <router-link
                        :to="{ name: 'account.organizations.detail', params: { slug: organization.slug } }"
                        class="item-link item-wrapper"
                    >
                        <div class="item-logo">
                            <icon-fa6-solid:building class="item-logo-icon"/>
                        </div>

                        <div class="item-info">
                            <div class="item-name">
                                {{ organization.name }}
                            </div>

                            <div class="item-content">
                                {{ organization.memberCount }} Members, {{ organization.shopCount }} Shops
                            </div>
                        </div>
                    </router-link>
                </div>
            </div>
        </div>

        <div v-if="changelogs.length > 0" class="panel">
            <h2 class="panel-title">
                <icon-fa6-solid:file-waveform/>
                Last Changes
            </h2>

            <data-table
                :columns="[
                { key: 'shopName', name: 'Shop', sortable: true },
                { key: 'extensions', name: 'Changes', sortable: true },
                { key: 'date', name: 'Date', sortable: true, sortPath: 'date' }
            ]"
                :data="changelogs"
            >
                <template #cell-shopName="{ row }">
                    <router-link
                        :to="{
                        name: 'account.shops.detail',
                        params: {
                            slug: row.organizationSlug,
                            shopId: row.shopId
                        }
                    }"
                    >
                        {{ row.shopName }}
                    </router-link>
                </template>

                <template #cell-extensions="{ row }">
                    {{ sumChanges(row) }}
                </template>

                <template #cell-date="{ row }">
                    {{ formatDateTime(row.date) }}
                </template>
            </data-table>
        </div>

    </div>
</template>

<script setup lang="ts">
import {sumChanges} from '@/helpers/changelog';
import {formatDateTime} from '@/helpers/formatter';
import {type RouterOutput, trpcClient} from '@/helpers/trpc';
import {ref} from 'vue';

const organizations = ref<RouterOutput['account']['listOrganizations']>();
trpcClient.account.listOrganizations.query().then((data) => {
    organizations.value = data;
});

const shops = ref<RouterOutput['account']['currentUserShops']>([]);

trpcClient.account.currentUserShops.query().then((data) => {
    shops.value = data;
});

const changelogs = ref<RouterOutput['account']['currentUserChangelogs']>([]);

trpcClient.account.currentUserChangelogs.query().then((data) => {
    changelogs.value = data;
});
</script>

<style scoped>

</style>
