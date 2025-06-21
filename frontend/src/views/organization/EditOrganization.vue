<template>
    <header-container
        v-if="organization?.data"
        :title="'Edit ' + organization.data.name"
    >
        <router-link
            :to="{ name: 'account.organizations.detail', params: { slug: organization.data.slug } }"
            type="button"
            class="btn"
        >
            Cancel
        </router-link>
    </header-container>

    <main-container v-if="organization?.data">
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="organization.data"
            class="panel"
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

                <div>
                    <label for="slug">Slug</label>
                    <field
                        id="slug"
                        type="text"
                        name="slug"
                        autocomplete="slug"
                        class="field"
                        :class="{ 'has-error': errors.slug }"
                    />
                    <div class="field-error-message">
                        {{ errors.slug }}
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

        <form-group v-if="canDeleteOrganization" :title="'Deleting organization ' + organization.data.name" class="panel">
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
import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';

import { Field, Form as VeeForm } from 'vee-validate';
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import * as Yup from 'yup';

const { error } = useAlert();
const router = useRouter();
const route = useRoute();

const organization =
    ref<
        Awaited<ReturnType<typeof authClient.organization.getFullOrganization>>
    >();
const canDeleteOrganization = ref<boolean>(false);

authClient.organization
    .getFullOrganization({
        query: { organizationSlug: route.params.slug as string },
    })
    .then((org) => {
        organization.value = org;
        authClient.organization
            .hasPermission({
                organizationId: org.data?.id,
                permissions: {
                    organization: ['delete'],
                },
            })
            .then((hasPermission) => {
                canDeleteOrganization.value =
                    hasPermission.data?.success ?? false;
            });
    });

const showOrganizationDeletionModal = ref(false);

const schema = Yup.object().shape({
    name: Yup.string().required('Name of organization is required'),
    slug: Yup.string()
        .required('Slug for organization is required')
        .matches(
            /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
            'Slug must be lowercase and can only contain letters, numbers, and hyphens',
        ),
});

async function onSaveOrganization(values: Record<string, unknown>) {
    const typedValues = values as Yup.InferType<typeof schema>;
    if (organization.value) {
        try {
            await authClient.organization.update({
                organizationId: organization.value.data.id,
                data: {
                    name: typedValues.name,
                    slug: typedValues.slug,
                },
            });

            await router.push({
                name: 'account.organizations.detail',
                params: {
                    slug: typedValues.slug,
                },
            });
        } catch (err) {
            error(err instanceof Error ? err.message : String(err));
        }
    }
}

async function deleteOrganization() {
    if (organization.value) {
        try {
            const resp = await authClient.organization.delete({
                organizationId: organization.value.data.id,
            });

            if (resp.error) {
                error(resp.error.message ?? 'Failed to delete organization');
                return;
            }

            await router.push({ name: 'account.organizations.list' });
        } catch (err) {
            error(err instanceof Error ? err.message : String(err));
        }
    }
}
</script>
