<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t("settings.title") }}</h1>

    <!-- Profile -->
    <CardSection
      :icon="IconUser"
      :title="$t('settings.profile')"
      :description="$t('settings.displayName')"
    >
      <div class="space-y-4">
        <div>
          <Label for="name">{{ $t("common.name") }}</Label>
          <Input id="name" v-model="profileName" type="text" autocomplete="name" class="mt-1.5" />
        </div>
      </div>
      <div class="flex justify-end mt-4">
        <Button size="sm" @click="saveProfile">
          <icon-fa6-solid:floppy-disk class="mr-1.5 size-3.5" />
          {{ $t("common.save") }}
        </Button>
      </div>
    </CardSection>

    <!-- Email -->
    <CardSection
      :icon="IconEnvelope"
      :title="$t('settings.email')"
      :description="$t('settings.changeEmail')"
    >
      <div class="space-y-4">
        <div>
          <Label for="email">{{ $t("common.emailAddress") }}</Label>
          <Input
            id="email"
            v-model="emailAddress"
            type="email"
            autocomplete="email"
            class="mt-1.5"
          />
        </div>
        <div>
          <Label for="emailCurrentPassword">{{ $t("settings.currentPassword") }}</Label>
          <Input
            id="emailCurrentPassword"
            v-model="emailCurrentPassword"
            type="password"
            autocomplete="current-password"
            class="mt-1.5"
          />
        </div>
      </div>
      <div class="flex justify-end mt-4">
        <Button size="sm" @click="saveEmail">
          <icon-fa6-solid:floppy-disk class="mr-1.5 size-3.5" />
          {{ $t("common.save") }}
        </Button>
      </div>
    </CardSection>

    <!-- Password -->
    <CardSection
      :icon="IconLock"
      :title="$t('settings.password')"
      :description="$t('settings.changePassword')"
    >
      <div class="space-y-4">
        <div>
          <Label for="currentPassword">{{ $t("settings.currentPassword") }}</Label>
          <div class="mt-1.5">
            <Input
              id="currentPassword"
              v-model="currentPassword"
              type="password"
              autocomplete="current-password"
            />
          </div>
        </div>
        <div>
          <Label for="newPassword">{{ $t("settings.newPassword") }}</Label>
          <div class="mt-1.5">
            <PasswordInput id="newPassword" v-model="newPassword" autocomplete="new-password" />
          </div>
        </div>
        <div>
          <Label for="confirmPassword">{{ $t("settings.confirmNewPassword") }}</Label>
          <div class="mt-1.5">
            <PasswordInput
              id="confirmPassword"
              v-model="confirmPassword"
              autocomplete="new-password"
            />
          </div>
        </div>
      </div>
      <div class="flex justify-end mt-4">
        <Button size="sm" @click="savePassword">
          <icon-fa6-solid:floppy-disk class="mr-1.5 size-3.5" />
          {{ $t("common.save") }}
        </Button>
      </div>
    </CardSection>

    <!-- Connected Accounts -->
    <CardSection
      :icon="IconGithub"
      :title="$t('settings.connectedAccounts')"
      :description="$t('settings.connectedAccountsDesc')"
    >
      <div class="flex items-center justify-between rounded-xl border px-4 py-3">
        <div class="flex items-center gap-3">
          <div class="flex size-9 items-center justify-center rounded-lg bg-muted">
            <icon-fa6-brands:github class="size-4" />
          </div>
          <div>
            <div class="text-sm font-medium">{{ $t("home.openSourceGithub") }}</div>
            <div class="text-xs text-muted-foreground">
              {{
                connectedProviders.includes("github")
                  ? $t("settings.linked")
                  : $t("settings.notLinked")
              }}
            </div>
          </div>
        </div>
        <Button
          v-if="!connectedProviders.includes('github')"
          size="sm"
          variant="outline"
          @click="linkSocial('github')"
        >
          {{ $t("settings.linkGithub") }}
        </Button>
        <Button v-else size="sm" variant="destructive" @click="unlinkSocial('github')">
          {{ $t("settings.unlinkGithub") }}
        </Button>
      </div>
    </CardSection>

    <!-- Passkeys -->
    <CardSection :icon="IconPasskey" :title="$t('settings.passkeyDevices')">
      <template #action>
        <Button size="sm" @click="showPasskeyCreationModal = true">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("settings.addDevice") }}
        </Button>
      </template>
      <EmptyState
        v-if="!passkeys?.length"
        :icon="IconPasskey"
        :title="$t('settings.noPasskeys')"
        size="sm"
      />
      <div v-else class="space-y-2">
        <div
          v-for="pk in passkeys"
          :key="pk.id"
          class="flex items-center justify-between rounded-xl border px-4 py-3"
        >
          <div>
            <div class="text-sm font-medium">{{ pk.name || $t("settings.unnamedPasskey") }}</div>
            <div class="text-xs text-muted-foreground">{{ pk.createdAt }}</div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            class="size-7 text-muted-foreground hover:text-destructive"
            @click="removePasskey(pk.id)"
          >
            <icon-fa6-solid:trash class="size-3" />
          </Button>
        </div>
      </div>
    </CardSection>

    <!-- Sessions -->
    <CardSection :icon="IconDesktop" :title="$t('settings.sessions')">
      <EmptyState
        v-if="!sessions?.length"
        :icon="IconDesktop"
        :title="$t('settings.noSessions')"
        size="sm"
      />
      <div v-else class="space-y-2">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="flex items-center justify-between rounded-xl border px-4 py-3"
        >
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm font-medium">
              {{ session.userAgent || $t("settings.unknownDevice") }}
            </div>
            <div class="text-xs text-muted-foreground">{{ session.createdAt }}</div>
          </div>
          <Badge
            v-if="session.id === sessionData?.session?.id"
            variant="secondary"
            class="mr-2 text-xs"
            >{{ $t("settings.currentSession") }}</Badge
          >
          <Button
            v-if="session.id !== sessionData?.session?.id"
            variant="ghost"
            size="icon"
            class="size-7 shrink-0 text-muted-foreground hover:text-destructive"
            @click="removeSession(session)"
          >
            <icon-fa6-solid:trash class="size-3" />
          </Button>
        </div>
      </div>
    </CardSection>

    <!-- Notifications -->
    <CardSection :icon="IconBell" :title="$t('settings.notifications')">
      <EmptyState
        v-if="!subscribedShops?.length"
        :icon="IconBellSlash"
        :title="$t('settings.notSubscribed')"
        :description="$t('settings.notSubscribedHint')"
        size="sm"
      />
      <div v-else class="space-y-2">
        <div
          v-for="env in subscribedShops"
          :key="env.id"
          class="flex items-center justify-between rounded-xl border px-4 py-3"
        >
          <div class="text-sm font-medium">{{ env.name }}</div>
          <Button
            variant="ghost"
            size="icon"
            class="size-7 text-muted-foreground hover:text-destructive"
            :title="$t('settings.unsubscribe')"
            @click="unsubscribeFromEnvironment(env.id)"
          >
            <icon-fa6-solid:bell-slash class="size-3" />
          </Button>
        </div>
      </div>
    </CardSection>

    <!-- Delete Account -->
    <Card class="border-destructive/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base text-destructive">
          <icon-fa6-solid:triangle-exclamation class="size-4" />
          {{ $t("settings.deleteAccountTitle") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Alert v-if="!canDeleteAccount" variant="destructive" class="mb-4">
          <AlertDescription>{{ $t("settings.deleteAccountOrgWarning") }}</AlertDescription>
        </Alert>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("settings.deleteAccountWarning") }}</p>
        <Button
          variant="destructive"
          :disabled="!canDeleteAccount"
          @click="showAccountDeletionModal = true"
        >
          <icon-fa6-solid:trash class="mr-1.5 size-3.5" />
          {{ $t("settings.deleteAccount") }}
        </Button>
      </CardContent>
    </Card>

    <!-- Modals -->
    <DeleteConfirmationModal
      :show="showAccountDeletionModal"
      :title="$t('settings.deleteAccount')"
      :entity-name="$t('settings.accountEntityName')"
      :require-password="true"
      @close="showAccountDeletionModal = false"
      @confirm="deleteUser"
      @password-change="(password) => (deleteCurrentPassword = password)"
    />

    <Dialog
      :open="showPasskeyCreationModal"
      @update:open="(v: boolean) => !v && (showPasskeyCreationModal = false)"
    >
      <DialogContent>
        <DialogHeader>
          <div class="flex items-start gap-3">
            <icon-fa6-solid:key class="mt-0.5 size-5 shrink-0 text-primary" />
            <DialogTitle>{{ $t("settings.addPasskeyTitle") }}</DialogTitle>
          </div>
        </DialogHeader>
        <p class="text-muted-foreground">{{ $t("settings.addPasskeyDesc") }}</p>
        <Input
          v-model="passKeyName"
          name="name"
          autocomplete="off"
          :placeholder="$t('settings.passkeyPlaceholder')"
        />
        <DialogFooter>
          <Button variant="outline" @click="showPasskeyCreationModal = false">{{
            $t("common.cancel")
          }}</Button>
          <Button @click="createPasskey">{{ $t("common.create") }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import CardSection from "@/components/CardSection.vue";
import PasswordInput from "@/components/PasswordInput.vue";
import EmptyState from "@/components/EmptyState.vue";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";

import IconUser from "~icons/fa6-solid/user";
import IconEnvelope from "~icons/fa6-solid/envelope";
import IconLock from "~icons/fa6-solid/lock";
import IconGithub from "~icons/fa6-brands/github";
import IconPasskey from "~icons/material-symbols/passkey";
import IconDesktop from "~icons/fa6-solid/desktop";
import IconBell from "~icons/fa6-solid/bell";
import IconBellSlash from "~icons/fa6-regular/bell-slash";
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { startRegistration } from "@simplewebauthn/browser";

const { t } = useI18n();
const { session: sessionData } = useSession();
const alert = useAlert();

const passKeyName = ref("");

interface PasskeyEntry {
  id: string;
  name: string | null;
  createdAt: string;
}
interface SessionEntry {
  id: string;
  expiresAt: string;
  createdAt: string;
  userAgent?: string | null;
  ipAddress?: string | null;
}

const passkeys = ref<PasskeyEntry[] | null>([]);
const sessions = ref<SessionEntry[] | null>([]);
const connectedProviders = ref<string[]>([]);
const subscribedShops = ref<components["schemas"]["SubscribedEnvironment"][] | null>(null);
const organizations = ref<{ id: string; name: string }[]>([]);
const deleteCurrentPassword = ref("");

const canDeleteAccount = computed(() => organizations.value.length === 0);

const profileName = ref(sessionData.value?.user?.name ?? "");

async function saveProfile() {
  if (profileName.value === sessionData.value?.user.name) return;
  if (profileName.value.length < 5) {
    alert.error(t("settings.nameMinLength"));
    return;
  }
  const { error } = await api.POST("/auth/update-user", { body: { name: profileName.value } });
  if (error) {
    alert.error((error as { message?: string }).message ?? t("settings.updateNameFailed"));
    return;
  }
  alert.success(t("settings.profileUpdated"));
}

const emailAddress = ref(sessionData.value?.user?.email ?? "");
const emailCurrentPassword = ref("");

async function saveEmail() {
  if (emailAddress.value === sessionData.value?.user.email) return;
  if (!emailCurrentPassword.value) {
    alert.error(t("settings.currentPasswordRequired"));
    return;
  }
  const { error } = await api.POST("/auth/change-email", {
    body: { newEmail: emailAddress.value, currentPassword: emailCurrentPassword.value },
  });
  if (error) {
    alert.error((error as { message?: string }).message ?? t("settings.changeEmailFailed"));
    return;
  }
  emailCurrentPassword.value = "";
  alert.success(t("settings.emailUpdated"));
}

const currentPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");

async function savePassword() {
  if (!newPassword.value) return;
  if (!currentPassword.value) {
    alert.error(t("settings.currentPasswordRequired"));
    return;
  }
  if (newPassword.value.length < 8) {
    alert.error(t("settings.passwordMinLength"));
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    alert.error(t("settings.passwordsDoNotMatch"));
    return;
  }
  const { error } = await api.POST("/auth/change-password", {
    body: { currentPassword: currentPassword.value, newPassword: newPassword.value },
  });
  if (error) {
    alert.error((error as { message?: string }).message ?? t("settings.changePasswordFailed"));
    return;
  }
  currentPassword.value = "";
  newPassword.value = "";
  confirmPassword.value = "";
  alert.success(t("settings.passwordChanged"));
}

async function loadPasskeys() {
  try {
    const { data } = await api.GET("/auth/passkey/list-user-passkeys");
    if (data) passkeys.value = data;
  } catch {}
}
async function loadSessions() {
  try {
    const { data } = await api.GET("/auth/list-sessions");
    if (data) sessions.value = data;
  } catch {}
}
async function loadLinkedAccounts() {
  try {
    const { data } = await api.GET("/auth/list-accounts");
    if (data && Array.isArray(data)) connectedProviders.value = data.map((a) => a.provider);
  } catch {}
}
async function loadOrganizations() {
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) organizations.value = data;
  } catch {}
}
async function loadSubscribedEnvironments() {
  try {
    const { data } = await api.GET("/account/subscribed-environments");
    subscribedShops.value = data ?? null;
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

const showAccountDeletionModal = ref(false);
const showPasskeyCreationModal = ref(false);

async function deleteUser() {
  if (deleteCurrentPassword.value === "") {
    alert.error(t("settings.passwordRequiredForDelete"));
    return;
  }
  try {
    const { error } = await api.POST("/auth/delete-user");
    if (error) {
      alert.error(t("settings.errorDeleteAccount"));
      return;
    }
    alert.success(t("settings.accountDeleted"));
    setTimeout(() => {
      window.location.reload();
    }, 2000);
    showAccountDeletionModal.value = false;
  } catch {
    alert.error(t("settings.errorDeleteAccount"));
  }
}

async function createPasskey() {
  if (!passKeyName.value) {
    alert.error(t("settings.passkeyNameRequired"));
    return;
  }
  try {
    const { data: optionsData, error: optionsError } = await api.POST(
      "/auth/passkey/register-options",
    );
    if (optionsError || !optionsData) {
      alert.error(t("settings.passkeyOptionsFailed"));
      return;
    }
    const { options, challengeKey } = optionsData as {
      options: { publicKey: Parameters<typeof startRegistration>[0]["optionsJSON"] };
      challengeKey: string;
    };
    const attestation = await startRegistration({ optionsJSON: options.publicKey });
    const { error: registerError } = await api.POST("/auth/passkey/register", {
      body: { challengeKey, name: passKeyName.value, ...attestation } as never,
    });
    if (registerError) {
      alert.error(t("settings.passkeyRegisterFailed"));
      return;
    }
    await loadPasskeys();
    showPasskeyCreationModal.value = false;
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function removePasskey(id: string) {
  try {
    await api.POST("/auth/passkey/delete-passkey", { body: { id } });
    await loadPasskeys();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}
async function removeSession(session: SessionEntry) {
  try {
    await api.POST("/auth/revoke-session", { body: { sessionId: session.id } });
    await loadSessions();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function linkSocial(provider: "github") {
  try {
    const { data } = await api.POST("/auth/link-social", {
      body: { provider, callbackURL: window.location.href },
    });
    if (data?.url) window.location.href = data.url;
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unlinkSocial(providerId: string) {
  try {
    const { error } = await api.POST("/auth/unlink-account", { body: { providerId } });
    if (!error) {
      await loadLinkedAccounts();
      alert.success(t("settings.unlinkedProvider", { providerId }));
    } else {
      alert.error(t("settings.errorUnlinkAccount"));
    }
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unsubscribeFromEnvironment(environmentId: number) {
  try {
    await api.DELETE("/environments/{environmentId}/subscribe", {
      params: { path: { environmentId } },
    });
    await loadSubscribedEnvironments();
    alert.success(t("settings.unsubscribedEnvironment"));
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

onMounted(() => {
  loadPasskeys();
  loadSessions();
  loadLinkedAccounts();
  loadOrganizations();
  loadSubscribedEnvironments();
});
</script>
