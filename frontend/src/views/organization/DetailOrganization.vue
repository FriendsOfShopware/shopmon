<template>
  <div v-if="organization" class="space-y-6">
    <!-- Header -->
    <PageHeader :title="organization.name">
      <Button size="sm" as-child>
        <RouterLink :to="{ name: 'account.organizations.edit' }">
          <icon-fa6-solid:pencil class="mr-1.5 size-3" />
          {{ $t("organization.editOrganization") }}
        </RouterLink>
      </Button>
    </PageHeader>

    <!-- Summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <StatCard :icon="IconUsers" :value="members.length" :label="$t('common.members')" />
      <StatCard
        :icon="IconEnvelope"
        :value="invitations.filter((i) => i.status === 'pending').length"
        :label="$t('organization.invitations')"
      />
      <StatCard :icon="IconShieldHalved" :value="ssoProviders.length" label="SSO providers" />
    </div>

    <!-- Members -->
    <CardSection :icon="IconUsers" :title="$t('common.members')">
      <template v-if="allowedToManageMembers" #action>
        <Button size="sm" @click="showAddMemberModal = true">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("common.add") }}
        </Button>
      </template>
      <div class="space-y-2">
        <div
          v-for="member in members"
          :key="member.id"
          class="flex items-center gap-3 rounded-xl border px-4 py-3"
        >
          <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10">
            <icon-fa6-solid:user class="size-3.5 text-primary" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-medium">{{ member.name }}</div>
            <div class="text-xs text-muted-foreground">{{ member.email }}</div>
          </div>
          <Badge variant="secondary" class="capitalize text-xs">{{ member.role }}</Badge>
          <div
            v-if="member.userId !== sessionData?.user?.id && allowedToManageMembers"
            class="flex shrink-0 items-center gap-1"
          >
            <Button
              variant="ghost"
              size="icon"
              class="size-7"
              :title="$t('organization.changeRole')"
              @click="openChangeRoleModal(member)"
            >
              <icon-fa6-solid:user-pen class="size-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              class="size-7 text-muted-foreground hover:text-destructive"
              :title="$t('organization.unassign')"
              @click="onRemoveMember(member.userId)"
            >
              <icon-fa6-solid:trash class="size-3" />
            </Button>
          </div>
        </div>
      </div>
    </CardSection>

    <!-- Invitations -->
    <CardSection
      v-if="invitations.length > 0"
      :icon="IconEnvelope"
      :title="$t('organization.invitations')"
    >
      <div class="space-y-2">
        <div
          v-for="inv in invitations"
          :key="inv.id"
          class="flex items-center gap-3 rounded-xl border px-4 py-3"
        >
          <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-muted">
            <icon-fa6-solid:envelope class="size-3.5 text-muted-foreground" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-medium">{{ inv.email }}</div>
            <div class="mt-0.5 flex items-center gap-2">
              <Badge variant="secondary" class="capitalize text-xs">{{ inv.role }}</Badge>
              <Badge
                :class="
                  inv.status === 'pending'
                    ? 'bg-warning/10 text-warning border-warning/30'
                    : 'bg-muted text-muted-foreground'
                "
                class="text-xs capitalize"
              >
                {{ inv.status }}
              </Badge>
            </div>
          </div>
          <Button
            v-if="inv.status !== 'canceled'"
            variant="ghost"
            size="icon"
            class="size-7 shrink-0 text-muted-foreground hover:text-destructive"
            :title="$t('organization.cancelInvitation')"
            @click="cancelInvitation(inv.id)"
          >
            <icon-fa6-solid:xmark class="size-3.5" />
          </Button>
        </div>
      </div>
    </CardSection>

    <!-- SSO -->
    <CardSection :icon="IconKey" :title="$t('organization.ssoConfig')">
      <template #action>
        <Button size="sm" variant="ghost" as-child>
          <RouterLink :to="{ name: 'account.organizations.sso' }">
            <icon-fa6-solid:gear class="mr-1.5 size-3" />
            {{ $t("organization.manageSso") }}
          </RouterLink>
        </Button>
      </template>
      <EmptyState
        v-if="ssoProviders.length === 0"
        :icon="IconKey"
        :title="$t('organization.noSsoProviders')"
        :description="$t('organization.ssoConfigureHint')"
        size="sm"
      />
      <div v-else class="space-y-2">
        <div
          v-for="provider in ssoProviders"
          :key="provider.id"
          class="flex items-center gap-3 rounded-xl border px-4 py-3"
        >
          <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
            <icon-fa6-solid:shield-halved class="size-3.5 text-primary" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="font-medium">{{ provider.domain }}</div>
            <div class="text-xs text-muted-foreground truncate">{{ provider.issuer }}</div>
          </div>
        </div>
      </div>
    </CardSection>

    <!-- Leave org -->
    <Card class="border-destructive/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base text-destructive">
          <icon-fa6-solid:right-from-bracket class="size-4" />
          {{ $t("organization.leaveOrgTitle") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("organization.leaveOrgWarning") }}</p>
        <Button variant="destructive" @click="leaveOrganization()">
          <icon-fa6-solid:right-from-bracket class="mr-1.5 size-3.5" />
          {{ $t("organization.leaveOrganization") }}
        </Button>
      </CardContent>
    </Card>

    <!-- Add member dialog -->
    <Dialog
      :open="showAddMemberModal"
      @update:open="(v: boolean) => !v && (showAddMemberModal = false)"
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("organization.addMember") }}</DialogTitle>
        </DialogHeader>
        <form id="addMemberForm" class="space-y-4" @submit="onAddMemberSubmit">
          <FormField v-slot="{ componentField }" name="email">
            <FormItem>
              <FormLabel>{{ $t("common.email") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" autocomplete="email" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
          <FormField v-slot="{ componentField }" name="role">
            <FormItem>
              <FormLabel>{{ $t("common.role") }}</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="member">{{ $t("organization.roleMember") }}</SelectItem>
                  <SelectItem value="admin">{{ $t("organization.roleAdmin") }}</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="showAddMemberModal = false">{{
            $t("common.cancel")
          }}</Button>
          <Button :disabled="isAddMemberSubmitting" type="submit" form="addMemberForm">
            <icon-fa6-solid:plus v-if="!isAddMemberSubmitting" class="mr-1.5 size-3.5" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
            {{ $t("common.add") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Change role dialog -->
    <Dialog
      :open="showChangeRoleModal"
      @update:open="(v: boolean) => !v && (showChangeRoleModal = false)"
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("organization.changeMemberRole") }}</DialogTitle>
        </DialogHeader>
        <form id="changeRoleForm" class="space-y-4" @submit="onChangeRoleSubmit">
          <FormField v-slot="{ componentField }" name="changeRole">
            <FormItem>
              <FormLabel>{{ $t("common.role") }}</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="member">{{ $t("organization.roleMember") }}</SelectItem>
                  <SelectItem value="admin">{{ $t("organization.roleAdmin") }}</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
        <DialogFooter>
          <Button variant="outline" @click="showChangeRoleModal = false">{{
            $t("common.cancel")
          }}</Button>
          <Button :disabled="isChangingRole" type="submit" form="changeRoleForm">
            <icon-fa6-solid:floppy-disk v-if="!isChangingRole" class="mr-1.5 size-3.5" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
            {{ $t("common.save") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import PageHeader from "@/components/PageHeader.vue";
import StatCard from "@/components/StatCard.vue";
import CardSection from "@/components/CardSection.vue";
import EmptyState from "@/components/EmptyState.vue";

import IconUsers from "~icons/fa6-solid/users";
import IconEnvelope from "~icons/fa6-solid/envelope";
import IconShieldHalved from "~icons/fa6-solid/shield-halved";
import IconKey from "~icons/fa6-solid/key";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";

import { ref } from "vue";

const { t } = useI18n();
const { session: sessionData, activeOrganizationId } = useSession();
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
      params: { query: { organizationId: activeOrganizationId.value! } },
    });
    if (!data) {
      alert.error("Failed to load organization");
      return;
    }

    organization.value = data as unknown as OrganizationData;
    members.value = (data as unknown as { members?: OrganizationMember[] }).members ?? [];
    invitations.value = (data as unknown as { invitations?: Invitation[] }).invitations ?? [];

    try {
      const { data: permData } = await api.POST("/auth/has-permission", {
        body: { organizationId: (data as unknown as OrganizationData).id },
      });
      allowedToManageMembers.value = permData?.success ?? false;
    } catch {
      /* ignore */
    }

    if (organization.value?.id) {
      api
        .GET("/organizations/{orgId}/sso-providers", {
          params: { path: { orgId: organization.value.id } },
        })
        .then(({ data }) => {
          if (data) ssoProviders.value = data;
        })
        .catch(() => {});
    }
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

loadOrganization();

const showAddMemberModal = ref(false);
const isAddMemberSubmitting = ref(false);
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
          body: { email: values.email, role: values.role },
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
  if (!organization.value) return;
  try {
    await api.DELETE("/auth/organizations/{organizationId}/members/{userId}", {
      params: { path: { organizationId: organization.value.id, userId } },
    });
    await loadOrganization();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function cancelInvitation(invitationId: string) {
  if (!organization.value) return;
  try {
    await api.POST("/auth/cancel-invitation", { body: { invitationId } });
    await loadOrganization();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
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
            path: { organizationId: organization.value.id, userId: selectedMember.value.userId },
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
