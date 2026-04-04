<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t('settings.title') }}</h1>

    <!-- Profile -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:user class="size-4 text-muted-foreground" />
          {{ $t('settings.profile') }}
        </CardTitle>
        <CardDescription>{{ $t('settings.displayName') }}</CardDescription>
      </CardHeader>
      <CardContent>
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
      </CardContent>
    </Card>

    <!-- Email -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:envelope class="size-4 text-muted-foreground" />
          {{ $t('settings.email') }}
        </CardTitle>
        <CardDescription>{{ $t('settings.changeEmail') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="space-y-4">
          <div>
            <Label for="email">{{ $t("common.emailAddress") }}</Label>
            <Input id="email" v-model="emailAddress" type="email" autocomplete="email" class="mt-1.5" />
          </div>
          <div>
            <Label for="emailCurrentPassword">{{ $t("settings.currentPassword") }}</Label>
            <Input id="emailCurrentPassword" v-model="emailCurrentPassword" type="password" autocomplete="current-password" class="mt-1.5" />
          </div>
        </div>
        <div class="flex justify-end mt-4">
          <Button size="sm" @click="saveEmail">
            <icon-fa6-solid:floppy-disk class="mr-1.5 size-3.5" />
            {{ $t("common.save") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Password -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:lock class="size-4 text-muted-foreground" />
          {{ $t('settings.password') }}
        </CardTitle>
        <CardDescription>{{ $t('settings.changePassword') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="space-y-4">
          <div>
            <Label for="currentPassword">{{ $t("settings.currentPassword") }}</Label>
            <Input id="currentPassword" v-model="currentPassword" type="password" autocomplete="current-password" class="mt-1.5" />
          </div>
          <div>
            <Label for="newPassword">{{ $t("settings.newPassword") }}</Label>
            <div class="relative mt-1.5">
              <Input id="newPassword" v-model="newPassword" :type="newPasswordType" autocomplete="new-password" class="pr-10" />
              <button type="button" class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground opacity-50 hover:opacity-100 transition-opacity" @click="newPasswordType = newPasswordType === 'password' ? 'text' : 'password'">
                <icon-fa6-solid:eye v-if="newPasswordType === 'password'" class="size-3.5" />
                <icon-fa6-solid:eye-slash v-else class="size-3.5" />
              </button>
            </div>
          </div>
          <div>
            <Label for="confirmPassword">{{ $t("settings.confirmNewPassword") }}</Label>
            <div class="relative mt-1.5">
              <Input id="confirmPassword" v-model="confirmPassword" :type="confirmPasswordType" autocomplete="new-password" class="pr-10" />
              <button type="button" class="absolute inset-y-0 right-0 flex items-center pr-3 text-muted-foreground opacity-50 hover:opacity-100 transition-opacity" @click="confirmPasswordType = confirmPasswordType === 'password' ? 'text' : 'password'">
                <icon-fa6-solid:eye v-if="confirmPasswordType === 'password'" class="size-3.5" />
                <icon-fa6-solid:eye-slash v-else class="size-3.5" />
              </button>
            </div>
          </div>
        </div>
        <div class="flex justify-end mt-4">
          <Button size="sm" @click="savePassword">
            <icon-fa6-solid:floppy-disk class="mr-1.5 size-3.5" />
            {{ $t("common.save") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Connected Accounts -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-brands:github class="size-4 text-muted-foreground" />
          {{ $t('settings.connectedAccounts') }}
        </CardTitle>
        <CardDescription>{{ $t('settings.connectedAccountsDesc') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="flex items-center justify-between rounded-xl border px-4 py-3">
          <div class="flex items-center gap-3">
            <div class="flex size-9 items-center justify-center rounded-lg bg-muted">
              <icon-fa6-brands:github class="size-4" />
            </div>
            <div>
              <div class="text-sm font-medium">GitHub</div>
              <div class="text-xs text-muted-foreground">
                {{ connectedProviders.includes('github') ? $t("settings.linked") : $t("settings.notLinked") }}
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
      </CardContent>
    </Card>

    <!-- Passkeys -->
    <Card>
      <CardHeader class="pb-3">
        <div class="flex items-center justify-between">
          <CardTitle class="flex items-center gap-2 text-base">
            <icon-material-symbols:passkey class="size-5 text-muted-foreground" />
            {{ $t('settings.passkeyDevices') }}
          </CardTitle>
          <Button size="sm" @click="showPasskeyCreationModal = true">
            <icon-fa6-solid:plus class="mr-1.5 size-3" />
            {{ $t("settings.addDevice") }}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div v-if="!passkeys?.length" class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-8 text-center">
          <icon-material-symbols:passkey class="size-8 text-muted-foreground" />
          <p class="text-sm text-muted-foreground">No passkeys registered yet.</p>
        </div>
        <div v-else class="space-y-2">
          <div v-for="pk in passkeys" :key="pk.id" class="flex items-center justify-between rounded-xl border px-4 py-3">
            <div>
              <div class="text-sm font-medium">{{ pk.name || 'Unnamed passkey' }}</div>
              <div class="text-xs text-muted-foreground">{{ pk.createdAt }}</div>
            </div>
            <Button variant="ghost" size="icon" class="size-7 text-muted-foreground hover:text-destructive" @click="removePasskey(pk.id)">
              <icon-fa6-solid:trash class="size-3" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Sessions -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:desktop class="size-4 text-muted-foreground" />
          {{ $t('settings.sessions') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="!sessions?.length" class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-8 text-center">
          <icon-fa6-solid:desktop class="size-8 text-muted-foreground" />
          <p class="text-sm text-muted-foreground">No active sessions.</p>
        </div>
        <div v-else class="space-y-2">
          <div v-for="session in sessions" :key="session.id" class="flex items-center justify-between rounded-xl border px-4 py-3">
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-medium">{{ session.userAgent || 'Unknown device' }}</div>
              <div class="text-xs text-muted-foreground">{{ session.createdAt }}</div>
            </div>
            <Badge v-if="session.id === sessionData?.session?.id" variant="secondary" class="mr-2 text-xs">Current</Badge>
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
      </CardContent>
    </Card>

    <!-- Notifications -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:bell class="size-4 text-muted-foreground" />
          {{ $t('settings.notifications') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="!subscribedShops?.length" class="flex flex-col items-center gap-2 rounded-xl border border-dashed py-8 text-center">
          <icon-fa6-regular:bell-slash class="size-8 text-muted-foreground" />
          <p class="font-medium">{{ $t("settings.notSubscribed") }}</p>
          <p class="max-w-sm text-sm text-muted-foreground">{{ $t("settings.notSubscribedHint") }}</p>
        </div>
        <div v-else class="space-y-2">
          <div v-for="env in subscribedShops" :key="env.id" class="flex items-center justify-between rounded-xl border px-4 py-3">
            <div class="text-sm font-medium">{{ env.name }}</div>
            <Button variant="ghost" size="icon" class="size-7 text-muted-foreground hover:text-destructive" :title="$t('settings.unsubscribe')" @click="unsubscribeFromEnvironment(env.id)">
              <icon-fa6-solid:bell-slash class="size-3" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Delete Account -->
    <Card class="border-destructive/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base text-destructive">
          <icon-fa6-solid:triangle-exclamation class="size-4" />
          {{ $t('settings.deleteAccountTitle') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Alert v-if="!canDeleteAccount" variant="destructive" class="mb-4">
          <AlertDescription>{{ $t("settings.deleteAccountOrgWarning") }}</AlertDescription>
        </Alert>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("settings.deleteAccountWarning") }}</p>
        <Button variant="destructive" :disabled="!canDeleteAccount" @click="showAccountDeletionModal = true">
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

    <Dialog :open="showPasskeyCreationModal" @update:open="(v: boolean) => !v && (showPasskeyCreationModal = false)">
      <DialogContent>
        <DialogHeader>
          <div class="flex items-start gap-3">
            <icon-fa6-solid:key class="mt-0.5 size-5 shrink-0 text-primary" />
            <DialogTitle>{{ $t("settings.addPasskeyTitle") }}</DialogTitle>
          </div>
        </DialogHeader>
        <p class="text-muted-foreground">{{ $t("settings.addPasskeyDesc") }}</p>
        <Input v-model="passKeyName" name="name" autocomplete="off" placeholder="e.g. MacBook Pro" />
        <DialogFooter>
          <Button variant="outline" @click="showPasskeyCreationModal = false">{{ $t("common.cancel") }}</Button>
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { startRegistration } from "@simplewebauthn/browser";

const { t } = useI18n();
const { session: sessionData } = useSession();
const alert = useAlert();

const passKeyName = ref("");

interface PasskeyEntry { id: string; name: string | null; createdAt: string }
interface SessionEntry { id: string; expiresAt: string; createdAt: string; userAgent?: string | null; ipAddress?: string | null }

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
  if (profileName.value.length < 5) { alert.error(t("settings.nameMinLength")); return; }
  const { error } = await api.POST("/auth/update-user", { body: { name: profileName.value } });
  if (error) { alert.error((error as { message?: string }).message ?? t("settings.updateNameFailed")); return; }
  alert.success(t("settings.profileUpdated"));
}

const emailAddress = ref(sessionData.value?.user?.email ?? "");
const emailCurrentPassword = ref("");

async function saveEmail() {
  if (emailAddress.value === sessionData.value?.user.email) return;
  if (!emailCurrentPassword.value) { alert.error(t("settings.currentPasswordRequired")); return; }
  const { error } = await api.POST("/auth/change-email", { body: { newEmail: emailAddress.value, currentPassword: emailCurrentPassword.value } });
  if (error) { alert.error((error as { message?: string }).message ?? t("settings.changeEmailFailed")); return; }
  emailCurrentPassword.value = "";
  alert.success(t("settings.emailUpdated"));
}

const currentPassword = ref("");
const newPassword = ref("");
const newPasswordType = ref("password");
const confirmPassword = ref("");
const confirmPasswordType = ref("password");

async function savePassword() {
  if (!newPassword.value) return;
  if (!currentPassword.value) { alert.error(t("settings.currentPasswordRequired")); return; }
  if (newPassword.value.length < 8) { alert.error(t("settings.passwordMinLength")); return; }
  if (newPassword.value !== confirmPassword.value) { alert.error(t("settings.passwordsDoNotMatch")); return; }
  const { error } = await api.POST("/auth/change-password", { body: { currentPassword: currentPassword.value, newPassword: newPassword.value } });
  if (error) { alert.error((error as { message?: string }).message ?? t("settings.changePasswordFailed")); return; }
  currentPassword.value = ""; newPassword.value = ""; confirmPassword.value = "";
  alert.success(t("settings.passwordChanged"));
}

async function loadPasskeys() { try { const { data } = await api.GET("/auth/passkey/list-user-passkeys"); if (data) passkeys.value = data; } catch {} }
async function loadSessions() { try { const { data } = await api.GET("/auth/list-sessions"); if (data) sessions.value = data; } catch {} }
async function loadLinkedAccounts() { try { const { data } = await api.GET("/auth/list-accounts"); if (data && Array.isArray(data)) connectedProviders.value = data.map((a) => a.provider); } catch {} }
async function loadOrganizations() { try { const { data } = await api.GET("/auth/list-organizations"); if (data) organizations.value = data; } catch {} }
async function loadSubscribedEnvironments() { try { const { data } = await api.GET("/account/subscribed-environments"); subscribedShops.value = data ?? null; } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); } }

const showAccountDeletionModal = ref(false);
const showPasskeyCreationModal = ref(false);

async function deleteUser() {
  if (deleteCurrentPassword.value === "") { alert.error(t("settings.passwordRequiredForDelete")); return; }
  try {
    const { error } = await api.POST("/auth/delete-user");
    if (error) { alert.error(t("settings.errorDeleteAccount")); return; }
    alert.success(t("settings.accountDeleted"));
    setTimeout(() => { window.location.reload(); }, 2000);
    showAccountDeletionModal.value = false;
  } catch { alert.error(t("settings.errorDeleteAccount")); }
}

async function createPasskey() {
  if (!passKeyName.value) { alert.error(t("settings.passkeyNameRequired")); return; }
  try {
    const { data: optionsData, error: optionsError } = await api.POST("/auth/passkey/register-options");
    if (optionsError || !optionsData) { alert.error(t("settings.passkeyOptionsFailed")); return; }
    const { options, challengeKey } = optionsData as { options: { publicKey: Parameters<typeof startRegistration>[0]["optionsJSON"] }; challengeKey: string };
    const attestation = await startRegistration({ optionsJSON: options.publicKey });
    const { error: registerError } = await api.POST("/auth/passkey/register", { body: { challengeKey, name: passKeyName.value, ...attestation } as never });
    if (registerError) { alert.error(t("settings.passkeyRegisterFailed")); return; }
    await loadPasskeys();
    showPasskeyCreationModal.value = false;
  } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); }
}

async function removePasskey(id: string) { try { await api.POST("/auth/passkey/delete-passkey", { body: { id } }); await loadPasskeys(); } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); } }
async function removeSession(session: SessionEntry) { try { await api.POST("/auth/revoke-session", { body: { sessionId: session.id } }); await loadSessions(); } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); } }

async function linkSocial(provider: "github") {
  try {
    const { data } = await api.POST("/auth/link-social", { body: { provider, callbackURL: window.location.href } });
    if (data?.url) window.location.href = data.url;
  } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); }
}

async function unlinkSocial(providerId: string) {
  try {
    const { error } = await api.POST("/auth/unlink-account", { body: { providerId } });
    if (!error) { await loadLinkedAccounts(); alert.success(t("settings.unlinkedProvider", { providerId })); }
    else { alert.error(t("settings.errorUnlinkAccount")); }
  } catch (err) { alert.error(err instanceof Error ? err.message : String(err)); }
}

async function unsubscribeFromEnvironment(environmentId: number) {
  try { await api.DELETE("/environments/{environmentId}/subscribe", { params: { path: { environmentId } } }); await loadSubscribedEnvironments(); alert.success(t("settings.unsubscribedEnvironment")); }
  catch (err) { alert.error(err instanceof Error ? err.message : String(err)); }
}

onMounted(() => { loadPasskeys(); loadSessions(); loadLinkedAccounts(); loadOrganizations(); loadSubscribedEnvironments(); });
</script>
