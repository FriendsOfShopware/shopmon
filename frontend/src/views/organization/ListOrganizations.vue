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
        <template v-if="!organizations.data || organizations.data.length === 0">
            <element-empty title="No Organization" :route="{ name: 'account.organizations.new' }" button="Add Organization">
                Get started by adding your first organization.
            </element-empty>
        </template>

        <div v-else class="panel panel-table">
            <data-table
                :columns="[
                    { key: 'name', name: 'Name', sortable: true },
                    { key: 'slug', name: 'Slug', sortable: true },
                ]"
                :data="organizations.data || []"
                class="bg-white dark:bg-neutral-800"
            >
                <template #cell-name="{ row }">
                    <router-link :to="{ name: 'account.organizations.detail', params: { slug: row.slug } }">
                        {{ row.name }}
                    </router-link>
                </template>

                <template #cell-slug="{ row }">
                    {{ row.slug }}
                </template>
            </data-table>
        </div>
    </main-container>
</template>

<script setup lang="ts">
import ElementEmpty from '@/components/layout/ElementEmpty.vue';
import { authClient } from '@/helpers/auth-client';

const organizations = authClient.useListOrganizations();
</script>
