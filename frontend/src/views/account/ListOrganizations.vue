<script setup lang="ts">
import { trpcClient } from '@/helpers/trpc';

import HeaderContainer from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import DataTable from '@/components/layout/DataTable.vue';

const organizations = await trpcClient.account.currentUserOrganizations.query();
</script>

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
    <main-container>
        <div
            v-if="organizations.length === 0"
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
                :labels="{name: {name: 'Name'}, memberCount: {name: 'Members'}, shopCount: {name: 'Shops'}}"
                :data="organizations || []"
                class="bg-white dark:bg-neutral-800"
            >
                <template #cell(name)="{ item }">
                    <router-link :to="{ name: 'account.organizations.detail', params: { organizationId: item.id } }">
                        {{ item.name }}
                    </router-link>
                </template>

                <template #cell(memberCount)="{ item }">
                    <icon-fa6-solid:people-group />
                    {{ item.memberCount }}
                </template>

                <template #cell(shopsCount)="{ item }">
                    <icon-fa6-solid:cart-shopping />
                    {{ item.shopCount }}
                </template>
            </data-table>
        </div>
    </main-container>
</template>
