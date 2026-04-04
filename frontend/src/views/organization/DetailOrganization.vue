<template>
  <header-container v-if="organization?.name" :title="organization.name">
    <router-link
      :to="{ name: 'account.organizations.edit', params: { organizationId: organization.id } }"
      type="button"
      class="btn btn-primary"
    >
      <icon-fa6-solid:pencil class="icon" aria-hidden="true" />
      {{ $t("organization.editOrganization") }}
    </router-link>
  </header-container>

  <main-container v-if="organization">
    <Panel :title="$t('organization.orgInfo')" class="organization-info">
      <dl class="organization-info-list">
        <div class="organization-info-item">
          <dt>{{ $t("organization.orgName") }}</dt>
          <dd>
            {{ organization.name }}
          </dd>
        </div>

        <div class="organization-info-item">
          <dt>{{ $t("common.members") }}</dt>
          <dd>
            {{ members.length }}
          </dd>
        </div>
      </dl>
    </Panel>

    <Panel :title="$t('common.members')">
      <template #action>
        <button class="btn btn-sm btn-primary" type="button" @click="showAddMemberModal = true">
          <icon-fa6-solid:plus class="icon" aria-hidden="true" />
          {{ $t("common.add") }}
        </button>
      </template>

      <DataTable
        :columns="[
          { key: 'email', name: $t('common.email') },
          { key: 'name', name: $t('common.name') },
          { key: 'role', name: $t('common.role') },
        ]"
        :data="members"
      >
        <template #cell-email="{ row }">
          {{ row.email }}
        </template>

        <template #cell-name="{ row }">
          {{ row.name }}
        </template>

        <template #cell-role="{ row }">
          {{ row.role }}
        </template>

        <template #cell-actions="{ row }">
          <button
            v-if="row.userId !== sessionData?.user?.id && allowedToManageMembers"
            type="button"
            class="tooltip tooltip-top-left"
            :data-tooltip="$t('organization.changeRole')"
            @click="openChangeRoleModal(row as OrganizationMember)"
          >
            <icon-fa6-solid:user-pen aria-hidden="true" class="icon" />
          </button>
          <button
            v-if="row.userId !== sessionData?.user?.id && allowedToManageMembers"
            type="button"
            class="tooltip tooltip-top-left"
            :data-tooltip="$t('organization.unassign')"
            @click="onRemoveMember(row.userId)"
          >
            <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
          </button>
        </template>
      </DataTable>
    </Panel>

    <Panel :title="$t('organization.invitations')">
      <DataTable
        :columns="[
          { key: 'email', name: $t('common.email') },
          { key: 'role', name: $t('common.role') },
          { key: 'status', name: $t('common.status') },
        ]"
        :data="invitations"
      >
        <template #cell-actions="{ row }">
          <button
            v-if="row.status !== 'canceled'"
            type="button"
            class="tooltip tooltip-top-left"
            :data-tooltip="$t('organization.cancelInvitation')"
            @click="cancelInvitation(row.id)"
          >
            <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
          </button>
        </template>
      </DataTable>
    </Panel>

    <Panel :title="$t('organization.ssoConfig')">
      <template #action>
        <router-link
          :to="{ name: 'account.organizations.sso', params: { organizationId: organization.id } }"
          type="button"
          class="btn btn-sm btn-primary"
        >
          <icon-fa6-solid:key class="icon" aria-hidden="true" />
          {{ $t("organization.manageSso") }}
        </router-link>
      </template>

      <div v-if="ssoProviders.length === 0" class="sso-empty">
        <p>{{ $t("organization.noSsoProviders") }}</p>
        <p class="text-muted">
          {{ $t("organization.ssoConfigureHint") }}
        </p>
      </div>

      <DataTable
        v-else
        :columns="[
          { key: 'domain', name: $t('organization.domain') },
          { key: 'issuer', name: $t('organization.issuer') },
        ]"
        :data="ssoProviders"
      />
    </Panel>

    <Panel :title="$t('organization.leaveOrgTitle')">
      <p class="mb-1">
        {{ $t("organization.leaveOrgWarning") }}
      </p>

      <button class="btn btn-danger mt-2" @click="leaveOrganization()">
        {{ $t("organization.leaveOrganization") }}
      </button>
    </Panel>

    <modal :show="showAddMemberModal" close-x-mark @close="showAddMemberModal = false">
      <template #title> {{ $t("organization.addMember") }} </template>

      <template #content>
        <vee-form
          id="addMemberForm"
          v-slot="{ errors }"
          :validation-schema="schemaMembers"
          :initial-values="{ email: '', role: 'member' }"
          @submit="onAddMember"
        >
          <InputField name="email" :label="$t('common.email')" autocomplete="email" :error="errors.email" />

          <SelectField name="role" :label="$t('common.role')" :error="errors.role">
            <option value="member">{{ $t("organization.roleMember") }}</option>
            <option value="admin">{{ $t("organization.roleAdmin") }}</option>
          </SelectField>
        </vee-form>
      </template>

      <template #footer>
        <button type="reset" class="btn" form="addMemberForm" @click="showAddMemberModal = false">
          {{ $t("common.cancel") }}
        </button>

        <button :disabled="isSubmitting" type="submit" class="btn btn-primary" form="addMemberForm">
          <icon-fa6-solid:plus v-if="!isSubmitting" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t("common.add") }}
        </button>
      </template>
    </modal>

    <modal :show="showChangeRoleModal" close-x-mark @close="showChangeRoleModal = false">
      <template #title> {{ $t("organization.changeMemberRole") }} </template>

      <template #content>
        <vee-form
          id="changeRoleForm"
          v-slot="{ errors }"
          :validation-schema="schemaChangeRole"
          :initial-values="{ role: selectedMember?.role || 'member' }"
          @submit="onChangeRole"
        >
          <SelectField name="role" :label="$t('common.role')" :error="errors.role">
            <option value="member">{{ $t("organization.roleMember") }}</option>
            <option value="admin">{{ $t("organization.roleAdmin") }}</option>
          </SelectField>
        </vee-form>
      </template>

      <template #footer>
        <button type="reset" class="btn" form="changeRoleForm" @click="showChangeRoleModal = false">
          {{ $t("common.cancel") }}
        </button>

        <button
          :disabled="isChangingRole"
          type="submit"
          class="btn btn-primary"
          form="changeRoleForm"
        >
          <icon-fa6-solid:floppy-disk v-if="!isChangingRole" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t("common.save") }}
        </button>
      </template>
    </modal>
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import { ref } from "vue";
import * as Yup from "yup";

const { t } = useI18n();
const { session: sessionData } = useSession();

const route = useRoute();
const alert = useAlert();

interface OrganizationData {
  id: string;
  name: string;
}

interface OrganizationMember {
  id: string;
  userId: string;
  role: string;
  name: string;
  email: string;
}

interface Invitation {
  id: string;
  email: string;
  role: string;
  status: string;
}

const organization = ref<OrganizationData | null>(null);
const members = ref<OrganizationMember[]>([]);
const invitations = ref<Invitation[]>([]);
const ssoProviders = ref<components["schemas"]["SsoProvider"][]>([]);
const allowedToManageMembers = ref(false);

async function leaveOrganization() {
  if (!organization.value) return;

  try {
    const { error: respError } = await api.POST("/auth/organizations/{organizationId}/leave", {
      params: { path: { organizationId: organization.value.id } },
    });

    if (respError) {
      alert.error((respError as { message?: string }).message ?? t("organization.failedLeaveOrg"));
      return;
    }

    alert.success(t("organization.leftOrg"));
    window.location.href = "/";
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function loadOrganization() {
  try {
    const { data } = await api.GET("/auth/get-full-organization", {
      params: { query: { organizationId: route.params.organizationId as string } },
    });

    if (!data) {
      alert.error("Failed to load organization");
      return;
    }

    organization.value = data as unknown as OrganizationData;
    members.value = (data as unknown as { members?: OrganizationMember[] }).members ?? [];
    invitations.value = (data as unknown as { invitations?: Invitation[] }).invitations ?? [];

    // Check if user can manage members
    try {
      const { data: permData } = await api.POST("/auth/has-permission", {
        body: {
          organizationId: (data as unknown as OrganizationData).id,
        },
      });
      allowedToManageMembers.value = permData?.success ?? false;
    } catch {
      // silently ignore permission check failure
    }

    // Load SSO providers
    if (organization.value?.id) {
      api
        .GET("/organizations/{orgId}/sso-providers", {
          params: { path: { orgId: organization.value.id } },
        })
        .then(({ data }) => {
          if (data) ssoProviders.value = data;
        })
        .catch((err) => {
          alert.error(err instanceof Error ? err.message : String(err));
        });
    }
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

loadOrganization();

const showAddMemberModal = ref(false);
const isSubmitting = ref(false);

// Change role modal state
const showChangeRoleModal = ref(false);
const isChangingRole = ref(false);

const selectedMember = ref<OrganizationMember | null>(null);

const schemaMembers = Yup.object().shape({
  email: Yup.string().email(t("validation.emailInvalid")).required(t("validation.emailRequired")),
  role: Yup.string()
    .oneOf(["member", "admin"], t("validation.roleInvalid"))
    .required(t("validation.roleRequired")),
});

const schemaChangeRole = Yup.object().shape({
  role: Yup.string()
    .oneOf(["member", "admin"], t("validation.roleInvalid"))
    .required(t("validation.roleRequired")),
});

async function onAddMember(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schemaMembers>;
  isSubmitting.value = true;
  if (organization.value) {
    try {
      const { error: respError } = await api.POST(
        "/auth/organizations/{organizationId}/invitations",
        {
          params: { path: { organizationId: organization.value.id } },
          body: {
            email: typedValues.email,
            role: typedValues.role,
          },
        },
      );

      if (respError) {
        alert.error((respError as { message?: string }).message ?? "Failed to invite member");
      } else {
        showAddMemberModal.value = false;
        await loadOrganization();
      }
    } catch (err) {
      alert.error(err instanceof Error ? err.message : String(err));
    }
  }
  isSubmitting.value = false;
}

async function onRemoveMember(userId: string) {
  if (organization.value) {
    try {
      await api.DELETE("/auth/organizations/{organizationId}/members/{userId}", {
        params: {
          path: {
            organizationId: organization.value.id,
            userId,
          },
        },
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
      await api.POST("/auth/cancel-invitation", {
        body: { invitationId },
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
      const { error: respError } = await api.PATCH(
        "/auth/organizations/{organizationId}/members/{userId}",
        {
          params: {
            path: {
              organizationId: organization.value.id,
              userId: selectedMember.value.userId,
            },
          },
          body: { role: typedValues.role },
        },
      );

      if (respError) {
        alert.error((respError as { message?: string }).message ?? t("organization.failedUpdateMemberRole"));
        return;
      }

      alert.success(t("organization.memberRoleUpdated"));

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

.organization-members-header {
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

.organization-members-heading {
  font-size: 1.25rem;
  font-weight: 500;
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
