<template>
    <header-container
        v-if="organization"
        :title="'Edit ' + organization.name"
    >
        <router-link
            :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
            type="button"
            class="group btn"
        >
            Cancel
        </router-link>
    </header-container>
    <main-container v-if="organization && authStore.user">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="organization"
            @submit="onSaveOrganization"
        >
            <form-group
                title="Organization Information"
                sub-title=""
            >
                <div class="sm:col-span-6">
                    <label
                        for="Name"
                        class="block text-sm font-medium mb-1"
                    > Name </label>
                    <field
                        id="name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :class="{ 'is-invalid': errors.name }"
                    />
                    <div class="text-red-700">
                        {{ errors.name }}
                    </div>
                </div>
            </form-group>

            <div class="flex justify-end pb-12">
                <button
                    type="submit"
                    class="btn btn-primary"
                >
                    <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                        <icon-fa6-solid:floppy-disk
                            v-if="!isSubmitting"
                            class="h-5 w-5"
                            aria-hidden="true"
                        />
                        <icon-line-md:loading-twotone-loop
                            v-else
                            class="w-5 h-5"
                        />
                    </span>
                    Save
                </button>
            </div>
        </vee-form>

        <form-group :title="'Deleting organization ' + organization.name">
            <form
                action="#"
                method="POST"
            >
                <p>Once you delete your organization, you will lose all data associated with it. </p>

                <div class="mt-5">
                    <button
                        type="button"
                        class="btn btn-danger group flex items-center"
                        @click="showOrganizationDeletionModal = true"
                    >
                        <icon-fa6-solid:trash class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                        Delete organization
                    </button>
                </div>
            </form>
        </form-group>

        <Modal
            :show="showOrganizationDeletionModal"
            @close="showOrganizationDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="h-6 w-6 text-red-600 dark:text-red-400"
                    aria-hidden="true"
                />
            </template>
            <template #title>
                Delete organization
            </template>
            <template #content>
                Are you sure you want to delete your organization?
                All your data will be permanently deleted from our servers. This action cannot be undone.
            </template>
            <template #footer>
                <button
                    type="button"
                    class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                    @click="deleteOrganization"
                >
                    Delete
                </button>
                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
                    @click="showOrganizationDeletionModal = false"
                >
                    Cancel
                </button>
            </template>
        </Modal>
    </main-container>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';

import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useOrganizationStore } from '@/stores/organization.store';

import { Field, Form as VeeForm } from 'vee-validate';
import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';
import { ref } from 'vue';

const authStore = useAuthStore();
const organizationStore = useOrganizationStore();
const alertStore = useAlertStore();
const router = useRouter();
const route = useRoute();

const { user } = storeToRefs(authStore);

const organizationId = parseInt(route.params.organizationId as string, 10);
const organization = user.value?.organizations.find(organization => organization.id == organizationId);

const showOrganizationDeletionModal = ref(false);

const schema = Yup.object().shape({
    name: Yup.string().required('Name of organization is required'),
});

async function onSaveOrganization(values: Yup.InferType<typeof schema>) {
    if (organization) {
        try {
            await organizationStore.updateOrganization(organization.id, values.name);
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

async function deleteOrganization() {
    if (organization) {
        try {
            await organizationStore.delete(organization.id);

            await router.push({ name: 'account.organizations.list' });
        } catch (error: any) {
            alertStore.error(error);
        }
    }
}

</script>
