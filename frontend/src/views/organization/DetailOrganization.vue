<template>
    <header-container
        v-if="organization?.data?.name"
        :title="organization.data.name"
        >
        <router-link
            :to="{ name: 'account.organizations.edit', params: { slug: organization.data.slug } }"
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
            <h3 class="panel-title">
                Organization Information
            </h3>

            <dl class="organization-info-list">
                <div class="organization-info-item">
                    <dt>
                        Organization Name
                    </dt>
                    <dd>
                        {{ organization.data.name }}
                    </dd>
                </div>

                <div class="organization-info-item">
                    <dt>
                        Members
                    </dt>
                    <dd>
                        {{ organization.data.members.length }}
                    </dd>
                </div>

                <div class="organization-info-item">
                    <dt>
                        Slug
                    </dt>
                    <dd>
                        {{ organization.data.slug }}
                    </dd>
                </div>
            </dl>
        </div>

        <div class="panel">
            <div class="panel-header">
                <h3>
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
                    { key: 'name', name: 'Name' },
                    { key: 'role', name: 'Role' },
                ]"
                :data="organization.data.members"
            >
                <template #cell-email="{ row }">
                    {{ row.user.email }}
                </template>

                <template #cell-name="{ row }">
                    {{ row.user.name }}
                </template>
                
                <template #cell-role="{ row }">
                    {{ row.role }}
                </template>

                <template #cell-actions="{ row }">
                    <button
                        v-if="row.user.id !== session.data?.user.id && allowedToManageMembers"
                        type="button"
                        class="tooltip tooltip-position-left"
                        data-tooltip="Change Role"
                        @click="openChangeRoleModal(row as OrganizationMember)"
                    >
                        <icon-fa6-solid:user-pen aria-hidden="true" class="icon" />
                    </button>
                    <button
                        v-if="row.user.id !== session.data?.user.id && allowedToManageMembers"
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

        <div class="panel">
            <h3 class="panel-title">
                Invitations
            </h3>

            <DataTable
                :columns="[
                    { key: 'email', name: 'Email' },
                    { key: 'role', name: 'Role' },
                    { key: 'status', name: 'Status' },
                ]"
                :data="organization.data.invitations"
            >
                <template #cell-actions="{ row }">
                    <button
                        v-if="row.status !== 'canceled'"
                        type="button"
                        class="tooltip tooltip-position-left"
                        data-tooltip="Cancel Invitation"
                        @click="cancelInvitation(row.id)"
                    >
                        <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
                    </button>
                </template>
            </DataTable>
        </div>

        <div class="panel">
            <div class="panel-header">
                <h3>
                    SSO Configuration
                </h3>

                <router-link
                    :to="{ name: 'account.organizations.sso', params: { slug: organization.data.slug } }"
                    type="button"
                    class="btn btn-primary"
                >
                    <icon-fa6-solid:key
                        class="icon"
                        aria-hidden="true"
                    />
                    Manage SSO
                </router-link>
            </div>
            
            <div v-if="ssoProviders.length === 0" class="sso-empty">
                <p>No SSO providers configured yet.</p>
                <p class="text-muted">Configure SSO to allow users to login with their corporate identity provider.</p>
            </div>
            
            <DataTable
                v-else
                :columns="[
                    { key: 'domain', name: 'Domain' },
                    { key: 'issuer', name: 'Issuer' },
                ]"
                :data="ssoProviders"
            >
            </DataTable>
        </div>

        <div class="panel">
            <div class="panel-header">
                <h3>
                    Leave Organization
                </h3>
            </div>

            <p class="mb-1">
                If you leave this organization, you will lose access to all resources and data associated with it.
            </p>

            <button
                class="btn btn-danger mt-2"
                @click="leaveOrganization()"
            >
                Leave Organization
            </button>
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
                    :initial-values="{ email: '', role: 'member' }"
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

                    <label for="role">Role</label>
                    <field
                        id="role"
                        as="select"
                        name="role"
                        class="field"
                    >
                        <option value="member">Member</option>
                        <option value="admin">Admin</option>
                    </field>
                    <div class="field-error-message">
                        {{ errors.role }}
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

        <modal
            :show="showChangeRoleModal"
            close-x-mark
            @close="showChangeRoleModal = false"
        >
            <template #title>
                Change Member Role
            </template>

            <template #content>
                <vee-form
                    id="changeRoleForm"
                    v-slot="{ errors }"
                    :validation-schema="schemaChangeRole"
                    :initial-values="{ role: selectedMember?.role || 'member' }"
                    @submit="onChangeRole"
                >
                    <label for="role">Role</label>

                    <field
                        id="role"
                        as="select"
                        name="role"
                        class="field"
                        :class="{ 'has-error': errors.role }"
                    >
                        <option value="member">Member</option>
                        <option value="admin">Admin</option>
                    </field>

                    <div class="field-error-message">
                        {{ errors.role }}
                    </div>
                </vee-form>
            </template>

            <template #footer>
                <button
                    type="reset"
                    class="btn"
                    form="changeRoleForm"
                    @click="showChangeRoleModal = false"
                >
                    Cancel
                </button>

                <button
                    :disabled="isChangingRole"
                    type="submit"
                    class="btn btn-primary"
                    form="changeRoleForm"
                >
                    <icon-fa6-solid:floppy-disk
                        v-if="!isChangingRole"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="icon"
                    />
                    Save
                </button>
            </template>
        </modal>
    </main-container>
</template>

<script setup lang="ts">
import { useAlert } from '@/composables/useAlert';
import { usePermissions } from '@/composables/usePermissions';
import { Field, Form as VeeForm } from 'vee-validate';
import { useRoute } from 'vue-router';

import { authClient } from '@/helpers/auth-client';
import { trpcClient } from '@/helpers/trpc';
import { computed, ref } from 'vue';
import * as Yup from 'yup';

const session = authClient.useSession();

const route = useRoute();
const alert = useAlert();

const organization =
    ref<
        Awaited<ReturnType<typeof authClient.organization.getFullOrganization>>
    >();
const ssoProviders = ref<{ domain: string; issuer: string }[]>([]);

const allowedToManageMembers = usePermissions(
    computed(() => ({
        organizationId: organization.value?.data?.id,
        permissions: {
            member: ['update', 'delete'],
        },
    })),
);

async function leaveOrganization() {
    if (!organization.value) return;

    try {
        await authClient.organization.leave({
            organizationId: organization.value.data.id,
        });

        alert.success('You have left the organization successfully.');
        window.location.href = '/';
    } catch (err) {
        alert.error(err instanceof Error ? err.message : String(err));
    }
}

async function loadOrganization() {
    authClient.organization
        .getFullOrganization({
            query: { organizationSlug: route.params.slug as string },
        })
        .then((org) => {
            organization.value = org;

            // Load SSO providers
            if (organization.value?.data.id) {
                trpcClient.organization.sso.list
                    .query({
                        orgId: organization.value.data.id,
                    })
                    .then((providers) => {
                        ssoProviders.value = providers;
                    })
                    .catch((err) => {
                        alert.error(
                            err instanceof Error ? err.message : String(err),
                        );
                    });
            }
        });
}

loadOrganization();

const showAddMemberModal = ref(false);
const isSubmitting = ref(false);

// Change role modal state
const showChangeRoleModal = ref(false);
const isChangingRole = ref(false);

type OrganizationMember = {
    id: string;
    role: string;
    user: {
        id: string;
        email: string;
        name: string;
    };
};

const selectedMember = ref<OrganizationMember | null>(null);

const schemaMembers = Yup.object().shape({
    email: Yup.string()
        .email('Email address is not valid')
        .required('Email address is required'),
    role: Yup.string()
        .oneOf(['member', 'admin'], 'Role must be either member or admin')
        .required('Role is required'),
});

const schemaChangeRole = Yup.object().shape({
    role: Yup.string()
        .oneOf(['member', 'admin'], 'Role must be either member or admin')
        .required('Role is required'),
});

async function onAddMember(values: Record<string, unknown>) {
    const typedValues = values as Yup.InferType<typeof schemaMembers>;
    isSubmitting.value = true;
    if (organization.value) {
        try {
            await authClient.organization.inviteMember({
                email: typedValues.email,
                role: typedValues.role,
                organizationId: organization.value.data.id,
            });

            showAddMemberModal.value = false;
            await loadOrganization();
        } catch (err) {
            alert.error(err instanceof Error ? err.message : String(err));
        }
    }
    isSubmitting.value = false;
}

async function onRemoveMember(userId: string) {
    if (organization.value) {
        try {
            await authClient.organization.removeMember({
                memberIdOrEmail: userId,
                organizationId: organization.value.data.id,
            });

            await loadOrganization();
        } catch (err) {
            alert.error(err instanceof Error ? err.message : String(err));
        }
    }
}

async function cancelInvitation(invitationId: string) {
    if (organization.value) {
        try {
            await authClient.organization.cancelInvitation({
                invitationId,
            });

            await loadOrganization();
        } catch (err) {
            alert.error(err instanceof Error ? err.message : String(err));
        }
    }
}

function openChangeRoleModal(member: OrganizationMember) {
    selectedMember.value = member;
    showChangeRoleModal.value = true;
}

async function onChangeRole(values: Record<string, unknown>) {
    const typedValues = values as Yup.InferType<typeof schemaChangeRole>;
    isChangingRole.value = true;

    if (organization.value && selectedMember.value) {
        try {
            const resp = await authClient.organization.updateMemberRole({
                memberId: selectedMember.value.id,
                role: typedValues.role,
                organizationId: organization.value.data.id,
            });

            if (resp.error) {
                alert.error(
                    resp.error.message ?? 'Failed to update member role',
                );
                return;
            }

            alert.success('Member role updated successfully');

            showChangeRoleModal.value = false;
            await loadOrganization();
        } catch (err) {
            alert.error(err instanceof Error ? err.message : String(err));
        }
    }

    isChangingRole.value = false;
}
</script>

<style scoped>
.organization-info-list {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;
    gap: 0.5rem 1.5rem;

    @media (min-width: 960px) {
        grid-column: 1 / span 2;
        grid-template-columns: repeat(3, 1fr);
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

.sso-empty {
    padding: 2rem;
    text-align: center;
    
    p {
        margin: 0;
        
        &.text-muted {
            margin-top: 0.5rem;
            color: var(--text-color-muted);
            font-size: 0.875rem;
        }
    }
}
</style>
