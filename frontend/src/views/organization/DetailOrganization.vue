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
                    class="btn btn-primary"
                    type="button"
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
                    :columns="[
                        { key: 'email', name: 'Email' },
                        { key: 'displayName', name: 'Display Name' },
                    ]"
                    :data="organizationStore.members"
                >
                    <template #cell-displayName="{ row }">
                        {{ row.displayName }}
                        <template v-if="isOwner">
                            (Owner)
                        </template>
                    </template>
                    <template #cell-actions="{ row }">
                        <button
                            v-if="!isOwner"
                            type="button"
                            class="tooltip-position-left text-red-600 opacity-50 dark:text-red-400 hover:opacity-100"
                            data-tooltip="Unassign"
                            @click="onRemoveMember(row.id)"
                        >
                            <icon-fa6-solid:trash aria-hidden="true" />
                        </button>
                    </template>
                </DataTable>
            </div>
        </div>

        <modal
            :show="showAddMemberModal"
            close-x-mark
            @close="showAddMemberModal = false"
        >
            <template #title>
                Add member
            </template>
            <template #content>
                <vee-form
                    id="addMemberForm"
                    v-slot="{ errors }"
                    :validation-schema="schemaMembers"
                    @submit="onAddMember"
                >
                    <label
                        for="email"
                        class="block text-sm font-medium mb-1"
                    > Email </label>
                    <field
                        id="email"
                        type="text"
                        name="email"
                        autocomplete="email"
                        class="field"
                        :class="{ 'is-invalid': errors.emailAddress }"
                    />
                    <div class="text-red-700">
                        {{ errors.email }}
                    </div>
                </vee-form>
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
        </modal>
    </main-container>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useOrganizationStore } from '@/stores/organization.store';
import { useRoute, useRouter } from 'vue-router';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { Field, Form as VeeForm } from 'vee-validate';

import { ref } from 'vue';
import * as Yup from 'yup';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const organizationStore = useOrganizationStore();
const alertStore = useAlertStore();
const { user } = storeToRefs(authStore);

const organizationId = parseInt(route.params.organizationId as string, 10);
const organization = user.value?.organizations.find(organization => organization.id == organizationId);
const isOwner = organization?.ownerId === user.value!.id;

const showAddMemberModal = ref(false);
const isSubmitting = ref(false);

organizationStore.loadMembers(organizationId);

const schemaMembers = Yup.object().shape({
    email: Yup.string().email('Email address is not valid').required('Email address is required'),
});

async function onAddMember(values: Yup.InferType<typeof schemaMembers>) {
    isSubmitting.value = true;
    if (organization) {
        try {
            await organizationStore.addMember(organization.id, values.email);
            showAddMemberModal.value = false;
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
    isSubmitting.value = false;
}

async function onRemoveMember(userId: number) {
    if (organization) {
        try {
            await organizationStore.removeMember(organization.id, userId);
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
