<template>
    <header-container title="My Organization">
        <router-link
            :to="{ name: 'account.organizations.new' }"
            type="button"
            class="btn btn-primary"
        >
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Add Organization
        </router-link>
    </header-container>

    <main-container>
        <template v-if="organizations && organizations.length === 0">
            <element-empty title="No Organization" :route="{ name: 'account.organizations.new' }" button="Add Organization">
                Get started by adding your first organization.
            </element-empty>
        </template>

        <div v-else class="panel panel-table">
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
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { ref } from 'vue';

const organizations = ref<RouterOutput['account']['listOrganizations']>();
trpcClient.account.listOrganizations.query().then((data) => {
    organizations.value = data;
});
</script>
