<template>
    <header-container
        v-if="organization"
        :title="organization.name"
    >
        <router-link
            v-if="isOwner"
            :to="{ name: 'account.organizations.edit', params: { organizationId } }"
            type="button"
            class="btn btn-primary"
        >
            <icon-fa6-solid:pencil
                class="icon"
                aria-hidden="true"
            />
            Edit Organization
        </router-link>
    </header-container>

    <main-container v-if="organization">
        <div class="panel organization-info">
            <h3 class="organization-info-heading">
                Organization Information
            </h3>

            <dl class="organization-info-list">
                <div class="organization-info-item">
                    <dt>
                        Organization Name
                    </dt>
                    <dd>
                        {{ organization.name }}
                    </dd>
                </div>

                <div class="organization-info-item">
                    <dt>
                        Members
                    </dt>
                    <dd>
                        {{ organization.memberCount }}
                    </dd>
                </div>

                <div class="organization-info-item">
                    <dt>
                        Shops
                    </dt>
                    <dd>
                        {{ organization.shopCount }}
                    </dd>
                </div>
            </dl>
        </div>

        <div class="panel panel-table">
            <div class="organization-members-header">
                <h3 class="organization-members-heading">
                    Members
                </h3>

                <button
                    class="btn btn-primary"
                    type="button"
                    @click="showAddMemberModal = true"
                >
                    <icon-fa6-solid:plus
                        class="icon"
                        aria-hidden="true"
                    />
                    Add
                </button>
            </div>

            <DataTable
                :columns="[
                    { key: 'email', name: 'Email' },
                    { key: 'displayName', name: 'Name' },
                ]"
                :data="organizationMembers"
            >
                <template #cell-displayName="{ row }">
                    {{ row.displayName }}
                    <template v-if="row.id === organization?.ownerId">
                        (Owner)
                    </template>
                </template>

                <template #cell-actions="{ row }">
                    <button
                        v-if="isOwner && row.id !== organization?.ownerId"
                        type="button"
                        class="tooltip tooltip-position-left"
                        data-tooltip="Unassign"
                        @click="onRemoveMember(row.id)"
                    >
                        <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
                    </button>
                </template>
            </DataTable>
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
                    <label for="email">Email</label>

                    <field
                        id="email"
                        type="text"
                        name="email"
                        autocomplete="email"
                        class="field"
                        :class="{ 'has-error': errors.email }"
                    />
                    <div class="field-error-message">
                        {{ errors.email }}
                    </div>
                </vee-form>
            </template>

            <template #footer>
                <button
                    type="reset"
                    class="btn"
                    form="addMemberForm"
                    @click="showAddMemberModal = false"
                >
                    Cancel
                </button>

                <button
                    :disabled="isSubmitting"
                    type="submit"
                    class="btn btn-primary"
                    form="addMemberForm"
                >
                    <icon-fa6-solid:plus
                        v-if="!isSubmitting"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="icon"
                    />
                    Add
                </button>
            </template>
        </modal>
    </main-container>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { Field, Form as VeeForm } from 'vee-validate';
import { useRoute, useRouter } from 'vue-router';

import { authClient } from '@/helpers/auth-client';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { computed, ref } from 'vue';
import * as Yup from 'yup';

const route = useRoute();
const router = useRouter();
const { error } = useAlert();
const session = authClient.useSession();

const organizationId = Number.parseInt(
    route.params.organizationId as string,
    10,
);

const organization = ref<
    RouterOutput['organization']['listSingleOrganization'] | null
>(null);
const organizationMembers = ref<
    RouterOutput['organization']['listMembers'] | null
>(null);

trpcClient.organization.listSingleOrganization
    .query({ orgId: organizationId })
    .then((org) => {
        organization.value = org;
    });

const isOwner = computed(
    () => organization.value?.ownerId === session.value.data?.user.id,
);

const showAddMemberModal = ref(false);
const isSubmitting = ref(false);

const schemaMembers = Yup.object().shape({
    email: Yup.string()
        .email('Email address is not valid')
        .required('Email address is required'),
});

async function loadOrganizationMembers() {
    organizationMembers.value = await trpcClient.organization.listMembers.query(
        {
            orgId: organizationId,
        },
    );
}

loadOrganizationMembers();

async function onAddMember(values: Yup.InferType<typeof schemaMembers>) {
    isSubmitting.value = true;
    if (organization.value) {
        try {
            await trpcClient.organization.addMember.mutate({
                orgId: organization.value.id,
                email: values.email,
            });

            await loadOrganizationMembers();

            showAddMemberModal.value = false;
            await router.push({
                name: 'account.organizations.detail',
                params: {
                    organizationId: organization.value.id,
                },
            });
        } catch (err) {
            error(err instanceof Error ? err.message : String(err));
        }
    }
    isSubmitting.value = false;
}

async function onRemoveMember(userId: string) {
    if (organization.value) {
        try {
            await trpcClient.organization.removeMember.mutate({
                orgId: organization.value.id,
                userId,
            });

            await loadOrganizationMembers();

            await router.push({
                name: 'account.organizations.detail',
                params: {
                    organizationId: organization.value.id,
                },
            });
        } catch (err) {
            error(err instanceof Error ? err.message : String(err));
        }
    }
}
</script>

<style scoped>
.organization-info {
    padding: 0;
    margin-bottom: 3rem;

    &-heading {
        padding: 1.25rem 1rem;
        font-size: 1.125rem;
        font-weight: 500;

        @media (min-width: 1024px) {
            padding-left: 2rem;
            padding-right: 2rem;
        }
    }
}

.organization-info-list {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;
    gap: 0.5rem 1.5rem;
    border-top: 1px solid var(--panel-border-color);
    padding: 1.25rem 1rem;

    @media (min-width: 960px) {
        grid-column: 1 / span 2;
        grid-template-columns: repeat(3, 1fr);
    }
    @media (min-width: 1024px) {
        padding-left: 2rem;
        padding-right: 2rem;
    }
}

.organization-info-item {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;

    @media (min-width: 960px) {
        grid-column: span 1;
    }

    dt {
        font-size: 0.875rem;
        font-weight: 500;
    }

    dd {
        margin-top: 0.25rem;
        font-size: 0.875rem;
        color: var(--text-color-muted);
    }
}

.organization-members {
    &-header {
        display: flex;
        padding: 1.5rem 2rem;
        justify-content: space-between;
        align-items: center;
        border-bottom: 1px solid var(--panel-border-color);

        @media (min-width: 1024px) {
            padding-left: 2rem;
            padding-right: 2rem;
        }
    }

    &-heading {
        font-size: 1.25rem;
        font-weight: 500;
    }
}
</style>
