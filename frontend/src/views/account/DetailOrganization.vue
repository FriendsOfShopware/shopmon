<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useRoute, useRouter } from 'vue-router';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

import { ref } from 'vue';
import { useForm } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import * as z from 'zod';
import { trpcClient } from '@/helpers/trpc';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const alertStore = useAlertStore();
const { user } = storeToRefs(authStore);

const organizationId = parseInt(route.params.organizationId as string, 10);
const organization = await trpcClient.organization.get.query(organizationId);
const members = await trpcClient.organization.listMembers.query({ orgId: organizationId });

const isOwner = organization?.ownerId === user.value!.id;

const showAddMemberModal = ref(false);


const { values, handleSubmit, errors, isSubmitting  } = useForm({
    validationSchema: toTypedSchema(z.object({
        email: z.string().email('Email address is not valid'),
    })),
});

const submit = handleSubmit(async (values) => {
    try {
        await trpcClient.organization.addMember.mutate({ orgId: organizationId, email: values.email });
        showAddMemberModal.value = false;
        await router.push({
            name: 'account.organizations.detail',
            params: {
                organizationId: organizationId,
            },
        });
    } catch (error: any) {
        alertStore.error(error);
    }
});

async function onRemoveMember(userId: number) {
    if (organization) {
        try {
            await trpcClient.organization.removeMember.mutate({ orgId: organization.id, userId: userId});
            await router.push({
                name: 'account.organizations.detail',
                params: {
                    organizationId: organization.id,
                },
            });
        } catch (error: any) {
            alertStore.error(error);
        }
    }
}
</script>

<template>
    <header-container
        v-if="organization"
        :title="organization.name"
    >
        <div class="flex gap-2">
            <router-link
                v-if="isOwner"
                :to="{ name: 'account.organizations.edit', params: { organizationId } }"
                type="button"
                class="group btn btn-primary flex items-center"
            >
                <icon-fa6-solid:pencil
                    class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
                    aria-hidden="true"
                />
                Edit Organization
            </router-link>
        </div>
    </header-container>

    <main-container v-if="organization">
        <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg dark:shadow-none dark:bg-neutral-800">
            <div class="py-5 px-4 sm:px-6 lg:px-8">
                <h3 class="text-lg leading-6 font-medium">
                    Organization Information
                </h3>
            </div>
            <div class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8 dark:border-neutral-700">
                <dl class="grid grid-cols-1 gap-6 sm:grid-cols-2">
                    <div class="sm:col-span-1">
                        <dt class="text-sm font-medium">
                            Organization Name
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ organization.name }}
                        </dd>
                    </div>

                    <div class="sm:col-span-1">
                        <dt class="font-medium">
                            Member Count
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ organization.memberCount }}
                        </dd>
                    </div>
                    <div class="sm:col-span-1">
                        <dt class="font-medium">
                            Shop Count
                        </dt>
                        <dd class="mt-1 text-sm text-gray-500">
                            {{ organization.shopCount }}
                        </dd>
                    </div>
                </dl>
            </div>
        </div>
        <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg dark:bg-neutral-800 dark-shadow-none">
            <div class="py-5 px-4 sm:px-6 lg:px-8 flex justify-between items-center">
                <h3 class="text-lg leading-6 font-medium">
                    Members
                </h3>
                <button
                    type="button"
                    class="btn btn-primary"
                    @click="showAddMemberModal = true"
                >
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <icon-fa6-solid:plus
                            class="h-5 w-5"
                            aria-hidden="true"
                        />
                    </span>
                    Add
                </button>
            </div>
            <div class="border-t border-gray-200 dark:border-neutral-700">
                <DataTable
                    :labels="{
                        email: { name: 'Email' },
                        displayName: { name: 'Display Name' },
                        actions: { name: '', class: 'text-right' }
                    }"
                    :data="members"
                >
                    <template #cell(displayName)="{ item }">
                        {{ item.displayName }} <template v-if="isOwner">
                            (Owner)
                        </template>
                    </template>
                    <template #cell(actions)="{ item }">
                        <button
                            v-if="!isOwner"
                            type="button"
                            class="tooltip-position-left text-red-600 opacity-50 dark:text-red-400 hover:opacity-100"
                            data-tooltip="Unassign"
                            @click="onRemoveMember(item.id)"
                        >
                            <icon-fa6-solid:trash aria-hidden="true" />
                        </button>
                    </template>
                </DataTable>
            </div>
        </div>

        <Modal
            :show="showAddMemberModal"
            close-x-mark
            @close="showAddMemberModal = false"
        >
            <template #title>
                Add member
            </template>
            <template #content>
                <form
                    id="addMemberForm"
                    @submit="submit"
                >
                    <label
                        for="email"
                        class="block text-sm font-medium mb-1"
                    > Email </label>
                    <input
                        id="email"
                        v-model="values.email"
                        type="text"
                        name="email"
                        autocomplete="email"
                        class="field"
                        :class="{ 'is-invalid': errors.email }"
                    >
                    <div class="text-red-700">
                        {{ errors.email }}
                    </div>
                </form>
            </template>
            <template #footer>
                <button
                    type="reset"
                    class="btn w-full sm:ml-3 sm:w-auto"
                    form="addMemberForm"
                    @click="showAddMemberModal = false"
                >
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <icon-fa6-solid:xmark
                            class="h-5 w-5"
                            aria-hidden="true"
                        />
                    </span>
                    Cancel
                </button>
                <button
                    :disabled="isSubmitting"
                    type="submit"
                    class="btn btn-primary w-full sm:ml-3 sm:w-auto"
                    form="addMemberForm"
                >
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <icon-fa6-solid:plus
                            v-if="!isSubmitting"
                            class="h-5 w-5"
                            aria-hidden="true"
                        />
                        <icon-line-md:loading-twotone-loop
                            v-else
                            class="w-5 h-5"
                        />
                    </span>
                    Add
                </button>
            </template>
        </Modal>
    </main-container>
</template>
