<template>
  <header v-if="organization?.name" class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ organization.name }}</h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button as-child>
        <RouterLink
          :to="{ name: 'account.organizations.edit', params: { organizationId: organization.id } }"
        >
          <icon-fa6-solid:pencil class="size-4" aria-hidden="true" />
          {{ $t("organization.editOrganization") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div v-if="organization" class="space-y-6">
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.orgInfo') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <dl class="grid grid-cols-1 md:grid-cols-3 gap-y-2 gap-x-6">
          <div>
            <dt class="text-sm font-medium">{{ $t("organization.orgName") }}</dt>
            <dd class="mt-1 text-sm text-muted-foreground">
              {{ organization.name }}
            </dd>
          </div>

          <div>
            <dt class="text-sm font-medium">{{ $t("common.members") }}</dt>
            <dd class="mt-1 text-sm text-muted-foreground">
              {{ members.length }}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t('common.members') }}</CardTitle>
        <CardAction>
          <Button size="sm" type="button" @click="showAddMemberModal = true">
            <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
            {{ $t("common.add") }}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent class="p-0">
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
            <div class="flex gap-1">
              <Button
                v-if="row.userId !== sessionData?.user?.id && allowedToManageMembers"
                type="button"
                variant="ghost"
                size="icon-sm"
                :title="$t('organization.changeRole')"
                @click="openChangeRoleModal(row as OrganizationMember)"
              >
                <icon-fa6-solid:user-pen aria-hidden="true" class="size-4" />
              </Button>
              <Button
                v-if="row.userId !== sessionData?.user?.id && allowedToManageMembers"
                type="button"
                variant="ghost"
                size="icon-sm"
                :title="$t('organization.unassign')"
                @click="onRemoveMember(row.userId)"
              >
                <icon-fa6-solid:trash aria-hidden="true" class="size-4 text-destructive" />
              </Button>
            </div>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.invitations') }}</CardTitle>
      </CardHeader>
      <CardContent class="p-0">
        <DataTable
          :columns="[
            { key: 'email', name: $t('common.email') },
            { key: 'role', name: $t('common.role') },
            { key: 'status', name: $t('common.status') },
          ]"
          :data="invitations"
        >
          <template #cell-actions="{ row }">
            <Button
              v-if="row.status !== 'canceled'"
              type="button"
              variant="ghost"
              size="icon-sm"
              :title="$t('organization.cancelInvitation')"
              @click="cancelInvitation(row.id)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="size-4 text-destructive" />
            </Button>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.ssoConfig') }}</CardTitle>
        <CardAction>
          <Button size="sm" as-child>
            <RouterLink
              :to="{ name: 'account.organizations.sso', params: { organizationId: organization.id } }"
            >
              <icon-fa6-solid:key class="size-4" aria-hidden="true" />
              {{ $t("organization.manageSso") }}
            </RouterLink>
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div v-if="ssoProviders.length === 0" class="py-8 text-center">
          <p>{{ $t("organization.noSsoProviders") }}</p>
          <p class="mt-2 text-muted-foreground text-sm">
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
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.leaveOrgTitle') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-sm text-muted-foreground mb-4">
          {{ $t("organization.leaveOrgWarning") }}
        </p>

        <Button variant="destructive" @click="leaveOrganization()">
          {{ $t("organization.leaveOrganization") }}
        </Button>
      </CardContent>
    </Card>

    <modal :show="showAddMemberModal" close-x-mark @close="showAddMemberModal = false">
      <template #title> {{ $t("organization.addMember") }} </template>

      <template #content>
        <form id="addMemberForm" class="space-y-4" @submit="onAddMemberSubmit">
          <FormField v-slot="{ componentField }" name="email">
            <FormItem>
              <FormLabel>{{ $t('common.email') }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" autocomplete="email" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="role">
            <FormItem>
              <FormLabel>{{ $t('common.role') }}</FormLabel>
              <FormControl>
                <Select v-bind="componentField">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="member">{{ $t("organization.roleMember") }}</SelectItem>
                    <SelectItem value="admin">{{ $t("organization.roleAdmin") }}</SelectItem>
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
      </template>

      <template #footer>
        <Button variant="outline" type="button" @click="showAddMemberModal = false">
          {{ $t("common.cancel") }}
        </Button>

        <Button
          :disabled="isAddMemberSubmitting"
          type="submit"
          form="addMemberForm"
        >
          <icon-fa6-solid:plus v-if="!isAddMemberSubmitting" class="size-4" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-4" />
          {{ $t("common.add") }}
        </Button>
      </template>
    </modal>

    <modal :show="showChangeRoleModal" close-x-mark @close="showChangeRoleModal = false">
      <template #title> {{ $t("organization.changeMemberRole") }} </template>

      <template #content>
        <form id="changeRoleForm" class="space-y-4" @submit="onChangeRoleSubmit">
          <FormField v-slot="{ componentField }" name="changeRole">
            <FormItem>
              <FormLabel>{{ $t('common.role') }}</FormLabel>
              <FormControl>
                <Select v-bind="componentField">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="member">{{ $t("organization.roleMember") }}</SelectItem>
                    <SelectItem value="admin">{{ $t("organization.roleAdmin") }}</SelectItem>
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
      </template>

      <template #footer>
        <Button
          variant="outline"
          type="button"
          @click="showChangeRoleModal = false"
        >
          {{ $t("common.cancel") }}
        </Button>

        <Button
          :disabled="isChangingRole"
          type="submit"
          form="changeRoleForm"
        >
          <icon-fa6-solid:floppy-disk v-if="!isChangingRole" class="size-4" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-4" />
          {{ $t("common.save") }}
        </Button>
      </template>
    </modal>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import { ref } from "vue";

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
const isAddMemberSubmitting = ref(false);

// Change role modal state
const showChangeRoleModal = ref(false);
const isChangingRole = ref(false);

const selectedMember = ref<OrganizationMember | null>(null);

const addMemberSchema = toTypedSchema(
  z.object({
    email: z.string().email(t("validation.emailInvalid")).min(1, t("validation.emailRequired")),
    role: z.enum(["member", "admin"], { message: t("validation.roleInvalid") }),
  }),
);

const changeRoleSchema = toTypedSchema(
  z.object({
    changeRole: z.enum(["member", "admin"], { message: t("validation.roleInvalid") }),
  }),
);

const { handleSubmit: handleAddMember } = useForm({
  validationSchema: addMemberSchema,
  initialValues: { email: "", role: "member" as const },
});

const { handleSubmit: handleChangeRole, setValues: setChangeRoleValues } = useForm({
  validationSchema: changeRoleSchema,
  initialValues: { changeRole: "member" as const },
});

const onAddMemberSubmit = handleAddMember(async (values) => {
  isAddMemberSubmitting.value = true;
  if (organization.value) {
    try {
      const { error: respError } = await api.POST(
        "/auth/organizations/{organizationId}/invitations",
        {
          params: { path: { organizationId: organization.value.id } },
          body: {
            email: values.email,
            role: values.role,
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
  isAddMemberSubmitting.value = false;
});

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
  setChangeRoleValues({ changeRole: member.role as "member" | "admin" });
  showChangeRoleModal.value = true;
}

const onChangeRoleSubmit = handleChangeRole(async (values) => {
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
          body: { role: values.changeRole },
        },
      );

      if (respError) {
        alert.error(
          (respError as { message?: string }).message ?? t("organization.failedUpdateMemberRole"),
        );
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
});
</script>
