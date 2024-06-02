<template>
    <div v-if="user && !shopStore.isLoading">
        <header-container title="Dashboard" />
        <main-container>
            <h2 class="section-title">
                <icon-fa6-solid:shop />
                My Shops's
            </h2>

            <ul class="dashboard-grid-container">
                <li v-for="shop in shopStore.shops" :key="shop.id" class="dashboard-grid-item">
                    <router-link 
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                organizationId: shop.organizationId,
                                shopId: shop.id
                            }
                        }" class="item-link"
                    >
                        <div class="item-logo">
                            <img v-if="shop.favicon" :src="shop.favicon" alt="Shop Logo" class="item-logo-img">
                        </div>

                        <div class="item-info">
                            <div class="item-name">
                                {{ shop.name }}
                            </div>

                            <div class="item-content">
                                {{ shop.organizationName }}<br>
                                {{ shop.shopwareVersion }}
                            </div>

                            <div class="item-state">
                                <status-icon :status="shop.status" />
                            </div>
                        </div>
                    </router-link>
                </li>
            </ul>

            <h2 class="section-title">
                <icon-fa6-solid:people-group />
                My Organizations
            </h2>
            <ul class="dashboard-grid-container">
                <li v-for="organization in organizations" :key="organization.id" class="dashboard-grid-item item-item">
                    <router-link
                        :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
                        class="item-link"
                    >
                        <div class="item-logo">
                            <icon-fa6-solid:people-group class="item-logo-icon" />
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
                </li>
            </ul>

            <template v-if="dashboardStore.changelogs.length > 0">
                <h2 class="section-title">
                    <icon-fa6-solid:file-waveform />
                    Last Changes
                </h2>
                
                <div class="panel panel-table">
                    <data-table
                        :columns="[
                            { key: 'shopName', name: 'Shop', sortable: true },
                            { key: 'extensions', name: 'Changes', sortable: true },
                            { key: 'date', name: 'Date', sortable: true, sortPath: 'date' }
                        ]"
                        :data="dashboardStore.changelogs"
                    >
                        <template #cell-shopName="{ row }">
                            <router-link
                                :to="{
                                    name: 'account.shops.detail',
                                    params: {
                                        organizationId: row.shopOrganizationId,
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
            </template>
        </main-container>
    </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';
import { useDashboardStore } from '@/stores/dashboard.store';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

import { sumChanges } from '@/helpers/changelog';
import { formatDateTime } from '@/helpers/formatter';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const organizations = user.value?.organizations.map(organization => ({
    ...organization,
    initials: organization.name.substring(0, 2),
}));

const shopStore = useShopStore();
shopStore.loadShops();

const dashboardStore = useDashboardStore();
dashboardStore.loadChangelogs();
</script>

<style scoped>
.section-title {
    color: var(--section-title-color);
    font-size: 1.125rem;
    line-height: 1.75rem;
    font-weight: 500;
    padding-bottom: 0.25rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid var(--section-title-border-color);
}

.dashboard-grid {
    &-container {
        display: grid;
        grid-template-columns: repeat(1, minmax(0, 1fr));
        gap: 1.25rem;
        margin-bottom: 2.5rem;

        @media (min-width: 640px) {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }

        @media (min-width: 1024px) {
            grid-template-columns: repeat(4, minmax(0, 1fr));
        }
    }

    &-item {
        position: relative;
        box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
        border-radius: 0.375rem;
        background-color: var(--item-background);

        .dark & {
            box-shadow: none;
        }

        &:hover {
            background-color: var(--item-hover-background);
        }

        .item-state {
            position: absolute;
            right: 0.5rem;
            bottom: 0.25rem;
        }
    }
}

.item {
    &-link {
        display: flex;
        text-decoration: none;
    }

    &-logo {
        flex-shrink: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        width: 4rem;
        border-top-left-radius: 0.375rem;
        border-bottom-left-radius: 0.375rem;
        background-color: var(--item-sub-background);

        &-img {
            width: 1.25rem;
            height: 1.25rem;
        }

        &-icon {
            font-size: 1.25rem;
            color: var(--item-icon-color);
        }
    }

    &-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: center;
        padding: 0.5rem 1rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    &-name {
        font-weight: 500;
        color: var(--item-title-color);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    &-content {
        color: #6b7280;
        font-size: 0.875rem;
        line-height: 1.25rem;
    }
}
</style>
