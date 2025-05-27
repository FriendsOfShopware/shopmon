<template>
    <header-container
        v-if="organization"
        :title="'Edit ' + organization.name"
    >
        <router-link
            :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
            type="button"
            class="btn"
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
            <form-group title="Organization Information">
                <div>
                    <label for="Name">Name</label>
                    <field
                        id="name"
                        type="text"
                        name="name"
                        autocomplete="name"
                        class="field"
                        :class="{ 'has-error': errors.name }"
                    />
                    <div class="field-error-message">
                        {{ errors.name }}
                    </div>
                </div>
            </form-group>

            <div class="form-submit">
                <button
                    type="submit"
                    class="btn btn-primary"
                >
                        <icon-fa6-solid:floppy-disk
                            v-if="!isSubmitting"
                            class="icon"
                            aria-hidden="true"
                        />
                        <icon-line-md:loading-twotone-loop
                            v-else
                            class="icon"
                        />
                    Save
                </button>
            </div>
        </vee-form>

        <form-group :title="'Deleting organization ' + organization.name">
            <p>Once you delete your organization, you will lose all data associated with it. </p>

            <button
                type="button"
                class="btn btn-danger"
                @click="showOrganizationDeletionModal = true"
            >
                <icon-fa6-solid:trash class="icon" />
                Delete organization
            </button>
        </form-group>

        <Modal
            :show="showOrganizationDeletionModal"
            @close="showOrganizationDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="icon icon-error"
                    aria-hidden="true"
                />
            </template>

            <template #title>
                Delete organization
            </template>

            <template #content>
                Are you sure you want to delete your organization? All of your data will
                be permanently removed
                from our servers forever. This action cannot be undone.
            </template>

            <template #footer>
                <button
                    type="button"
                    class="btn btn-danger"
                    @click="deleteOrganization"
                >
                    Delete
                </button>

                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn"
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

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useOrganizationStore } from '@/stores/organization.store';

import { Field, Form as VeeForm } from 'vee-validate';
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';

const authStore = useAuthStore();
const organizationStore = useOrganizationStore();
const alertStore = useAlertStore();
const router = useRouter();
const route = useRoute();

const { user } = storeToRefs(authStore);

const organizationId = Number.parseInt(
    route.params.organizationId as string,
    10,
);
const organization = user.value?.organizations.find(
    (organization) => organization.id === organizationId,
);

const showOrganizationDeletionModal = ref(false);

const schema = Yup.object().shape({
    name: Yup.string().required('Name of organization is required'),
});

async function onSaveOrganization(values: Yup.InferType<typeof schema>) {
    if (organization) {
        try {
            await organizationStore.updateOrganization(
                organization.id,
                values.name,
            );
            await router.push({
                name: 'account.organizations.detail',
                params: {
                    organizationId: organization.id,
                },
            });
        } catch (error) {
            alertStore.error(
                error instanceof Error ? error.message : String(error),
            );
        }
    }
}

async function deleteOrganization() {
    if (organization) {
        try {
            await organizationStore.delete(organization.id);

            await router.push({ name: 'account.organizations.list' });
        } catch (error) {
            alertStore.error(
                error instanceof Error ? error.message : String(error),
            );
        }
    }
}
</script>
