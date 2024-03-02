<template>
    <header-container title="My Organization">
        <router-link
            to="/account/organizations/new"
            type="button"
            class="group btn btn-primary flex items-center align-middle"
        >
            <icon-fa6-solid:plus
                class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50"
                aria-hidden="true"
            />
            Add Organization
        </router-link>
    </header-container>
    <main-container v-if="user">
        <div
            v-if="organizations && organizations.length === 0"
            class="text-center"
        >
            <svg
                class="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
            >
                <path
                    vector-effect="non-scaling-stroke"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                />
            </svg>
            <h3 class="mt-2 font-medium">
                No Organization
            </h3>
            <p class="mt-1 text-gray-500">
                Get started by adding your first organization.
            </p>
            <div class="mt-6">
                <router-link
                    to="/account/organizations/new"
                    class="btn btn-primary group flex items-center"
                >
                    <icon-fa6-solid:plus
                        class="-ml-1 mr-2 h-4 w-4 opacity-25 group-hover:opacity-50"
                        aria-hidden="true"
                    />
                    Add Organization
                </router-link>
            </div>
        </div>

        <div
            v-else
            class="shadow rounded-md overflow-y-scroll md:overflow-y-hidden dark:shadow-none"
        >
            <data-table
                :columns="[
                    { key: 'name', name: 'Name', sortable: true },
                    { key: 'memberCount', name: 'Members', sortable: true },
                    { key: 'shopCount', name: 'Shops', sortable: true },
                ]"
                :data="organizations || []"
                class="bg-white dark:bg-neutral-800"
            >
                <template #cell-name="{ row }">
                    <router-link :to="{ name: 'account.organizations.detail', params: { organizationId: row.id } }">
                        {{ row.name }}
                    </router-link>
                </template>

                <template #cell-memberCount="{ row }">
                    <icon-fa6-solid:people-group />
                    {{ row.memberCount }}
                </template>

                <template #cell-shopCount="{ row }">
                    <icon-fa6-solid:cart-shopping />
                    {{ row.shopCount }}
                </template>
            </data-table>
        </div>
    </main-container>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useAuthStore } from '@/stores/auth.store';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const organizations = user.value?.organizations;
</script>
