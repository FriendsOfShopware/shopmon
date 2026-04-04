<template>
  <header-container :title="$t('settings.title')" />
  <main-container>
    <Panel>
      <form-group :title="$t('settings.profile')" :sub-title="$t('settings.displayName')">
        <div>
          <label for="name">{{ $t("common.name") }}</label>
          <BaseInput id="name" v-model="profileName" type="text" autocomplete="name" />
        </div>
      </form-group>

      <div class="form-submit">
        <UiButton type="button" variant="primary" @click="saveProfile">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          {{ $t("common.save") }}
        </UiButton>
      </div>
    </Panel>

    <Panel>
      <form-group :title="$t('settings.email')" :sub-title="$t('settings.changeEmail')">
        <div>
          <label for="email">{{ $t("common.emailAddress") }}</label>
          <input
            id="email"
            v-model="emailAddress"
            type="email"
            autocomplete="email"
            class="field"
          />
        </div>

        <div>
          <label for="emailCurrentPassword">{{ $t("settings.currentPassword") }}</label>
          <input
            id="emailCurrentPassword"
            v-model="emailCurrentPassword"
            type="password"
            autocomplete="current-password"
            class="field"
          />
        </div>
      </form-group>

      <div class="form-submit">
        <UiButton type="button" variant="primary" @click="saveEmail">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          {{ $t("common.save") }}
        </UiButton>
      </div>
    </Panel>

    <Panel>
      <form-group :title="$t('settings.password')" :sub-title="$t('settings.changePassword')">
        <div>
          <label for="currentPassword">{{ $t("settings.currentPassword") }}</label>
          <input
            id="currentPassword"
            v-model="currentPassword"
            type="password"
            autocomplete="current-password"
            class="field"
          />
        </div>

        <div>
          <label for="newPassword">{{ $t("settings.newPassword") }}</label>
          <div class="password-wrapper">
            <input
              id="newPassword"
              v-model="newPassword"
              :type="newPasswordType"
              autocomplete="new-password"
              class="field field-password"
            />
            <div
              class="password-toggle"
              @click="newPasswordType = newPasswordType === 'password' ? 'text' : 'password'"
            >
              <icon-fa6-solid:eye v-if="newPasswordType === 'password'" class="icon" />
              <icon-fa6-solid:eye-slash v-else class="icon" />
            </div>
          </div>
        </div>

        <div>
          <label for="confirmPassword">{{ $t("settings.confirmNewPassword") }}</label>
          <div class="password-wrapper">
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              :type="confirmPasswordType"
              autocomplete="new-password"
              class="field field-password"
            />
            <div
              class="password-toggle"
              @click="
                confirmPasswordType = confirmPasswordType === 'password' ? 'text' : 'password'
              "
            >
              <icon-fa6-solid:eye v-if="confirmPasswordType === 'password'" class="icon" />
              <icon-fa6-solid:eye-slash v-else class="icon" />
            </div>
          </div>
        </div>
      </form-group>

      <div class="form-submit">
        <UiButton type="button" variant="primary" @click="savePassword">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          {{ $t("common.save") }}
        </UiButton>
      </div>
    </Panel>

    <Panel>
      <form-group
        :title="$t('settings.connectedAccounts')"
        :sub-title="$t('settings.connectedAccountsDesc')"
      >
        <p>{{ $t("settings.githubLoginText") }}</p>

        <UiButton
          v-if="!connectedProviders.includes('github')"
          type="button"
          variant="primary"
          @click="linkSocial('github')"
        >
          <icon-fa6-brands:github class="icon" aria-hidden="true" />
          {{ $t("settings.linkGithub") }}
        </UiButton>

        <UiButton v-else type="button" variant="destructive" @click="unlinkSocial('github')">
          <icon-fa6-brands:github class="icon" aria-hidden="true" />
          {{ $t("settings.unlinkGithub") }}
        </UiButton>
      </form-group>
    </Panel>

    <Panel>
      <form-group :title="$t('settings.passkeyDevices')" class="form-group-table">
        <data-table
          v-if="passkeys"
          :columns="[
            { key: 'name', name: $t('common.name'), sortable: true },
            { key: 'createdAt', name: $t('common.createdAt'), sortable: true },
          ]"
          :data="passkeys"
        >
          <template #cell-actions="{ row }">
            <UiButton
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('common.delete')"
              @click="removePasskey(row.id)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
            </UiButton>
          </template>
        </data-table>
      </form-group>

      <div class="form-submit">
        <UiButton type="button" variant="primary" @click="showPasskeyCreationModal = true">
          <icon-material-symbols:passkey class="icon icon-passkey" aria-hidden="true" />
          {{ $t("settings.addDevice") }}
        </UiButton>
      </div>
    </Panel>

    <Panel>
      <form-group :title="$t('settings.sessions')" class="form-group-table">
        <data-table
          v-if="sessions && sessions.length"
          :columns="[
            { key: 'userAgent', name: $t('settings.userAgent'), sortable: true },
            { key: 'createdAt', name: $t('common.createdAt'), sortable: true },
          ]"
          :data="sessions"
        >
          <template #cell-actions="{ row }">
            <UiButton
              v-if="row.id !== sessionData?.session?.id"
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('common.delete')"
              @click="removeSession(row)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
            </UiButton>
          </template>
        </data-table>
      </form-group>
    </Panel>

    <Panel>
      <form-group :title="$t('settings.notifications')" class="form-group-table">
        <div v-if="!subscribedShops || subscribedShops.length === 0" class="empty-state">
          <icon-fa6-regular:bell-slash class="empty-state-icon" />
          <p class="empty-state-text">{{ $t("settings.notSubscribed") }}</p>

          <p class="empty-state-subtext">
            {{ $t("settings.notSubscribedHint") }}
          </p>
        </div>

        <data-table
          v-else
          :columns="[{ key: 'name', name: $t('settings.environmentName'), sortable: true }]"
          :data="subscribedShops"
        >
          <template #cell-actions="{ row }">
            <button
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('settings.unsubscribe')"
              @click="unsubscribeFromEnvironment(row.id)"
            >
              <icon-fa6-solid:bell-slash aria-hidden="true" class="icon icon-error" />
            </button>
          </template>
        </data-table>
      </form-group>
    </Panel>

    <Panel :title="$t('settings.deleteAccountTitle')">
      <Banner v-if="!canDeleteAccount" variant="error">
        {{ $t("settings.deleteAccountOrgWarning") }}
      </Banner>

      <p>{{ $t("settings.deleteAccountWarning") }}</p>

      <UiButton
        type="button"
        variant="destructive"
        :disabled="!canDeleteAccount"
        @click="showAccountDeletionModal = true"
      >
        <icon-fa6-solid:trash class="icon icon-trash" />
        {{ $t("settings.deleteAccount") }}
      </UiButton>
    </Panel>

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
        <icon-fa6-solid:key class="icon icon-info" aria-hidden="true" />
      </template>

      <template #content>
        {{ $t("settings.addPasskeyDesc") }}
        <BaseInput v-model="passKeyName" name="name" autocomplete="off" />
      </template>

      <template #footer>
        <UiButton type="button" variant="primary" @click="createPasskey">
          {{ $t("common.create") }}
        </UiButton>

        <UiButton
          ref="cancelButtonRef"
          type="button"
          variant="ghost"
          @click="showPasskeyCreationModal = false"
        >
          {{ $t("common.cancel") }}
        </UiButton>
      </template>
    </Modal>
  </main-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

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

<style scoped>
.empty-state {
  text-align: center;
  padding: 3rem 1.5rem;
}

.empty-state-icon {
  font-size: 3rem;
  color: var(--text-color-muted);
  opacity: 0.5;
  margin-bottom: 1rem;
}

.empty-state-text {
  font-size: 1.125rem;
  color: var(--text-color);
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.empty-state-subtext {
  font-size: 0.875rem;
  color: var(--text-color-muted);
  max-width: 28rem;
  margin: 0 auto;
}

.password-wrapper {
  position: relative;
}

.password-toggle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  padding-right: 0.75rem;
  cursor: pointer;
  z-index: 10;
  opacity: 0.4;
  transition: opacity 0.4s;

  &:hover {
    opacity: 1;
  }
}
</style>
