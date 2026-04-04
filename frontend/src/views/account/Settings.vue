<template>
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold tracking-tight">{{ $t('settings.title') }}</h1>
  </div>

  <div class="space-y-6">
    <!-- Profile -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.profile') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('settings.displayName') }}</p>
        </div>

        <div class="space-y-4 mt-6">
          <div>
            <Label for="name">{{ $t("common.name") }}</Label>
            <Input id="name" v-model="profileName" type="text" autocomplete="name" class="mt-1" />
          </div>
        </div>

        <div class="flex justify-end mt-6">
          <Button @click="saveProfile">
            <icon-fa6-solid:floppy-disk class="size-3.5" aria-hidden="true" />
            {{ $t("common.save") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Email -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.email') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('settings.changeEmail') }}</p>
        </div>

        <div class="space-y-4 mt-6">
          <div>
            <Label for="email">{{ $t("common.emailAddress") }}</Label>
            <Input
              id="email"
              v-model="emailAddress"
              type="email"
              autocomplete="email"
              class="mt-1"
            />
          </div>

          <div>
            <Label for="emailCurrentPassword">{{ $t("settings.currentPassword") }}</Label>
            <Input
              id="emailCurrentPassword"
              v-model="emailCurrentPassword"
              type="password"
              autocomplete="current-password"
              class="mt-1"
            />
          </div>
        </div>

        <div class="flex justify-end mt-6">
          <Button @click="saveEmail">
            <icon-fa6-solid:floppy-disk class="size-3.5" aria-hidden="true" />
            {{ $t("common.save") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Password -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.password') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('settings.changePassword') }}</p>
        </div>

        <div class="space-y-4 mt-6">
          <div>
            <Label for="currentPassword">{{ $t("settings.currentPassword") }}</Label>
            <Input
              id="currentPassword"
              v-model="currentPassword"
              type="password"
              autocomplete="current-password"
              class="mt-1"
            />
          </div>

          <div>
            <Label for="newPassword">{{ $t("settings.newPassword") }}</Label>
            <div class="relative mt-1">
              <Input
                id="newPassword"
                v-model="newPassword"
                :type="newPasswordType"
                autocomplete="new-password"
                class="pr-10"
              />
              <div
                class="absolute right-0 top-0 bottom-0 flex items-center pr-3 cursor-pointer opacity-40 hover:opacity-100 transition-opacity"
                @click="newPasswordType = newPasswordType === 'password' ? 'text' : 'password'"
              >
                <icon-fa6-solid:eye v-if="newPasswordType === 'password'" class="size-4" />
                <icon-fa6-solid:eye-slash v-else class="size-4" />
              </div>
            </div>
          </div>

          <div>
            <Label for="confirmPassword">{{ $t("settings.confirmNewPassword") }}</Label>
            <div class="relative mt-1">
              <Input
                id="confirmPassword"
                v-model="confirmPassword"
                :type="confirmPasswordType"
                autocomplete="new-password"
                class="pr-10"
              />
              <div
                class="absolute right-0 top-0 bottom-0 flex items-center pr-3 cursor-pointer opacity-40 hover:opacity-100 transition-opacity"
                @click="confirmPasswordType = confirmPasswordType === 'password' ? 'text' : 'password'"
              >
                <icon-fa6-solid:eye v-if="confirmPasswordType === 'password'" class="size-4" />
                <icon-fa6-solid:eye-slash v-else class="size-4" />
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end mt-6">
          <Button @click="savePassword">
            <icon-fa6-solid:floppy-disk class="size-3.5" aria-hidden="true" />
            {{ $t("common.save") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Connected Accounts -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.connectedAccounts') }}</h3>
          <p class="text-sm text-muted-foreground">{{ $t('settings.connectedAccountsDesc') }}</p>
        </div>

        <div class="space-y-4 mt-6">
          <p>{{ $t("settings.githubLoginText") }}</p>

          <Button
            v-if="!connectedProviders.includes('github')"
            @click="linkSocial('github')"
          >
            <icon-fa6-brands:github class="size-4" aria-hidden="true" />
            {{ $t("settings.linkGithub") }}
          </Button>

          <Button v-else variant="destructive" @click="unlinkSocial('github')">
            <icon-fa6-brands:github class="size-4" aria-hidden="true" />
            {{ $t("settings.unlinkGithub") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Passkey Devices -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.passkeyDevices') }}</h3>
        </div>

        <div class="mt-6">
          <data-table
            v-if="passkeys"
            :columns="[
              { key: 'name', name: $t('common.name'), sortable: true },
              { key: 'createdAt', name: $t('common.createdAt'), sortable: true },
            ]"
            :data="passkeys"
          >
            <template #cell-actions="{ row }">
              <Button
                variant="ghost"
                size="icon-sm"
                :title="$t('common.delete')"
                @click="removePasskey(row.id)"
              >
                <icon-fa6-solid:trash aria-hidden="true" class="size-3.5 text-destructive" />
              </Button>
            </template>
          </data-table>
        </div>

        <div class="flex justify-end mt-6">
          <Button @click="showPasskeyCreationModal = true">
            <icon-material-symbols:passkey class="size-5" aria-hidden="true" />
            {{ $t("settings.addDevice") }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Sessions -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.sessions') }}</h3>
        </div>

        <div class="mt-6">
          <data-table
            v-if="sessions && sessions.length"
            :columns="[
              { key: 'userAgent', name: $t('settings.userAgent'), sortable: true },
              { key: 'createdAt', name: $t('common.createdAt'), sortable: true },
            ]"
            :data="sessions"
          >
            <template #cell-actions="{ row }">
              <Button
                v-if="row.id !== sessionData?.session?.id"
                variant="ghost"
                size="icon-sm"
                :title="$t('common.delete')"
                @click="removeSession(row)"
              >
                <icon-fa6-solid:trash aria-hidden="true" class="size-3.5 text-destructive" />
              </Button>
            </template>
          </data-table>
        </div>
      </CardContent>
    </Card>

    <!-- Notifications -->
    <Card>
      <CardContent>
        <div class="border-b pb-1 mb-4">
          <h3 class="text-lg font-medium">{{ $t('settings.notifications') }}</h3>
        </div>

        <div class="mt-6">
          <div v-if="!subscribedShops || subscribedShops.length === 0" class="text-center py-12">
            <icon-fa6-regular:bell-slash class="size-12 mx-auto text-muted-foreground opacity-50 mb-4" />
            <p class="text-lg font-medium mb-2">{{ $t("settings.notSubscribed") }}</p>
            <p class="text-sm text-muted-foreground max-w-md mx-auto">
              {{ $t("settings.notSubscribedHint") }}
            </p>
          </div>

          <data-table
            v-else
            :columns="[{ key: 'name', name: $t('settings.environmentName'), sortable: true }]"
            :data="subscribedShops"
          >
            <template #cell-actions="{ row }">
              <Button
                variant="ghost"
                size="icon-sm"
                :title="$t('settings.unsubscribe')"
                @click="unsubscribeFromEnvironment(row.id)"
              >
                <icon-fa6-solid:bell-slash aria-hidden="true" class="size-3.5 text-destructive" />
              </Button>
            </template>
          </data-table>
        </div>
      </CardContent>
    </Card>

    <!-- Delete Account -->
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('settings.deleteAccountTitle') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <Alert v-if="!canDeleteAccount" variant="destructive" class="mb-4">
          <AlertDescription>{{ $t("settings.deleteAccountOrgWarning") }}</AlertDescription>
        </Alert>

        <p class="mb-4">{{ $t("settings.deleteAccountWarning") }}</p>

        <Button
          variant="destructive"
          :disabled="!canDeleteAccount"
          @click="showAccountDeletionModal = true"
        >
          <icon-fa6-solid:trash class="size-3.5" />
          {{ $t("settings.deleteAccount") }}
        </Button>
      </CardContent>
    </Card>

    <delete-confirmation-modal
      :show="showAccountDeletionModal"
      :title="$t('settings.deleteAccount')"
      :entity-name="$t('settings.accountEntityName')"
      :require-password="true"
      @close="showAccountDeletionModal = false"
      @confirm="deleteUser"
      @password-change="(password) => (deleteCurrentPassword = password)"
    />

    <Modal :show="showPasskeyCreationModal" @close="showPasskeyCreationModal = false">
      <template #title> {{ $t("settings.addPasskeyTitle") }} </template>

      <template #icon>
        <icon-fa6-solid:key class="size-5 text-primary" aria-hidden="true" />
      </template>

      <template #content>
        {{ $t("settings.addPasskeyDesc") }}
        <Input v-model="passKeyName" name="name" autocomplete="off" class="mt-2" />
      </template>

      <template #footer>
        <Button @click="createPasskey">
          {{ $t("common.create") }}
        </Button>

        <Button
          ref="cancelButtonRef"
          variant="ghost"
          @click="showPasskeyCreationModal = false"
        >
          {{ $t("common.cancel") }}
        </Button>
      </template>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

const canDeleteAccount = computed(() => {
  return organizations.value.length === 0;
});

// Profile
const profileName = ref(sessionData.value?.user?.name ?? "");

async function saveProfile() {
  if (profileName.value === sessionData.value?.user.name) {
    return;
  }

  if (profileName.value.length < 5) {
    alert.error(t("settings.nameMinLength"));
    return;
  }

  const { error } = await api.POST("/auth/update-user", {
    body: { name: profileName.value },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? t("settings.updateNameFailed"));
    return;
  }

  alert.success(t("settings.profileUpdated"));
}

// Email
const emailAddress = ref(sessionData.value?.user?.email ?? "");
const emailCurrentPassword = ref("");

async function saveEmail() {
  if (emailAddress.value === sessionData.value?.user.email) {
    return;
  }

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

// Password
const currentPassword = ref("");
const newPassword = ref("");
const newPasswordType = ref("password");
const confirmPassword = ref("");
const confirmPasswordType = ref("password");

async function savePassword() {
  if (!newPassword.value) {
    return;
  }

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

// Data loading
async function loadPasskeys() {
  try {
    const { data } = await api.GET("/auth/passkey/list-user-passkeys");
    if (data) {
      passkeys.value = data;
    }
  } catch {
    // silently ignore
  }
}

async function loadSessions() {
  try {
    const { data } = await api.GET("/auth/list-sessions");
    if (data) {
      sessions.value = data;
    }
  } catch {
    // silently ignore
  }
}

async function loadLinkedAccounts() {
  try {
    const { data } = await api.GET("/auth/list-accounts");
    if (data && Array.isArray(data)) {
      connectedProviders.value = data.map((account) => account.provider);
    }
  } catch {
    // silently ignore
  }
}

async function loadOrganizations() {
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) {
      organizations.value = data;
    }
  } catch {
    // silently ignore
  }
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
    // Get registration options from server
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

    // Run WebAuthn browser API
    const attestation = await startRegistration({ optionsJSON: options.publicKey });

    // Send attestation to server
    const { error: registerError } = await api.POST("/auth/passkey/register", {
      body: {
        challengeKey,
        name: passKeyName.value,
        ...attestation,
      } as never,
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
    await api.POST("/auth/passkey/delete-passkey", {
      body: { id },
    });
    await loadPasskeys();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function removeSession(session: SessionEntry) {
  try {
    await api.POST("/auth/revoke-session", {
      body: { sessionId: session.id },
    });
    await loadSessions();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function linkSocial(provider: "github") {
  try {
    const { data } = await api.POST("/auth/link-social", {
      body: {
        provider,
        callbackURL: window.location.href,
      },
    });

    if (data?.url) {
      window.location.href = data.url;
    }
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unlinkSocial(providerId: string) {
  try {
    const { error } = await api.POST("/auth/unlink-account", {
      body: { providerId },
    });

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
