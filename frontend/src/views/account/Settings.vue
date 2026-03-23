<template>
  <header-container title="Settings" />
  <main-container>
    <Panel>
      <form-group title="Profile" sub-title="Your display name">
        <div>
          <label for="name">Name</label>
          <input
            id="name"
            v-model="profileName"
            type="text"
            autocomplete="name"
            class="field"
          />
        </div>
      </form-group>

      <div class="form-submit">
        <button type="button" class="btn btn-primary" @click="saveProfile">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          Save
        </button>
      </div>
    </Panel>

    <Panel>
      <form-group title="Email" sub-title="Change your email address">
        <div>
          <label for="email">Email address</label>
          <input
            id="email"
            v-model="emailAddress"
            type="email"
            autocomplete="email"
            class="field"
          />
        </div>

        <div>
          <label for="emailCurrentPassword">Current Password</label>
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
        <button type="button" class="btn btn-primary" @click="saveEmail">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          Save
        </button>
      </div>
    </Panel>

    <Panel>
      <form-group title="Password" sub-title="Change your password">
        <div>
          <label for="currentPassword">Current Password</label>
          <input
            id="currentPassword"
            v-model="currentPassword"
            type="password"
            autocomplete="current-password"
            class="field"
          />
        </div>

        <div>
          <label for="newPassword">New Password</label>
          <div class="password-wrapper">
            <input
              id="newPassword"
              v-model="newPassword"
              :type="newPasswordType"
              autocomplete="new-password"
              class="field field-password"
            />
            <div class="password-toggle" @click="newPasswordType = newPasswordType === 'password' ? 'text' : 'password'">
              <icon-fa6-solid:eye v-if="newPasswordType === 'password'" class="icon" />
              <icon-fa6-solid:eye-slash v-else class="icon" />
            </div>
          </div>
        </div>

        <div>
          <label for="confirmPassword">Confirm New Password</label>
          <div class="password-wrapper">
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              :type="confirmPasswordType"
              autocomplete="new-password"
              class="field field-password"
            />
            <div class="password-toggle" @click="confirmPasswordType = confirmPasswordType === 'password' ? 'text' : 'password'">
              <icon-fa6-solid:eye v-if="confirmPasswordType === 'password'" class="icon" />
              <icon-fa6-solid:eye-slash v-else class="icon" />
            </div>
          </div>
        </div>
      </form-group>

      <div class="form-submit">
        <button type="button" class="btn btn-primary" @click="savePassword">
          <icon-fa6-solid:floppy-disk class="icon" aria-hidden="true" />
          Save
        </button>
      </div>
    </Panel>

    <Panel>
      <form-group title="Connected Accounts" sub-title="Link or unlink external login providers">
        <p>Login using your GitHub account</p>

        <button
          v-if="!connectedProviders.includes('github')"
          type="button"
          class="btn btn-primary"
          @click="linkSocial('github')"
        >
          <icon-fa6-brands:github class="icon" aria-hidden="true" />
          Link GitHub
        </button>

        <button v-else type="button" class="btn btn-danger" @click="unlinkSocial('github')">
          <icon-fa6-brands:github class="icon" aria-hidden="true" />
          Unlink from GitHub
        </button>
      </form-group>
    </Panel>

    <Panel>
      <form-group title="Passkey Devices" class="form-group-table">
        <data-table
          v-if="passkeys"
          :columns="[
            { key: 'name', name: 'Name', sortable: true },
            { key: 'createdAt', name: 'Created At', sortable: true },
          ]"
          :data="passkeys"
        >
          <template #cell-actions="{ row }">
            <button
              type="button"
              class="tooltip-top-left"
              data-tooltip="Delete"
              @click="removePasskey(row.id)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
            </button>
          </template>
        </data-table>
      </form-group>

      <div class="form-submit">
        <button type="button" class="btn btn-primary" @click="showPasskeyCreationModal = true">
          <icon-material-symbols:passkey class="icon icon-passkey" aria-hidden="true" />
          Add a new Device
        </button>
      </div>
    </Panel>

    <Panel>
      <form-group title="Sessions" class="form-group-table">
        <data-table
          v-if="sessions && sessions.length"
          :columns="[
            { key: 'userAgent', name: 'User Agent', sortable: true },
            { key: 'createdAt', name: 'Created At', sortable: true },
          ]"
          :data="sessions"
        >
          <template #cell-actions="{ row }">
            <button
              v-if="row.id !== sessionData?.session?.id"
              type="button"
              class="tooltip-top-left"
              data-tooltip="Delete"
              @click="removeSession(row)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
            </button>
          </template>
        </data-table>
      </form-group>
    </Panel>

    <Panel>
      <form-group title="Notifications" class="form-group-table">
        <div v-if="!subscribedShops || subscribedShops.length === 0" class="empty-state">
          <icon-fa6-regular:bell-slash class="empty-state-icon" />
          <p class="empty-state-text">You are not subscribed to any shop notifications.</p>

          <p class="empty-state-subtext">
            Visit a shop's detail page and click the watch button to receive notifications about
            changes.
          </p>
        </div>

        <data-table
          v-else
          :columns="[{ key: 'name', name: 'Shop Name', sortable: true }]"
          :data="subscribedShops"
        >
          <template #cell-actions="{ row }">
            <button
              type="button"
              class="tooltip-top-left"
              data-tooltip="Unsubscribe"
              @click="unsubscribeFromShop(row.id)"
            >
              <icon-fa6-solid:bell-slash aria-hidden="true" class="icon icon-error" />
            </button>
          </template>
        </data-table>
      </form-group>
    </Panel>

    <Panel title="Deleting your Account">
      <Alert v-if="!canDeleteAccount" type="error">
        To delete your account, you must delete first all organizations or leave them.
      </Alert>

      <p>Once you delete your account, you will lose all data associated with it.</p>

      <button
        type="button"
        class="btn btn-danger"
        :disabled="!canDeleteAccount"
        @click="showAccountDeletionModal = true"
      >
        <icon-fa6-solid:trash class="icon icon-trash" />
        Delete account
      </button>
    </Panel>

    <delete-confirmation-modal
      :show="showAccountDeletionModal"
      title="Delete account"
      entity-name="your account"
      :require-password="true"
      @close="showAccountDeletionModal = false"
      @confirm="deleteUser"
      @password-change="(password) => (deleteCurrentPassword = password)"
    />

    <Modal :show="showPasskeyCreationModal" @close="showPasskeyCreationModal = false">
      <template #title> Add a Passkey Device </template>

      <template #icon>
        <icon-fa6-solid:key class="icon icon-info" aria-hidden="true" />
      </template>

      <template #content>
        Please give a name to your new Passkey Device.
        <input v-model="passKeyName" type="text" autocomplete="off" class="field" />
      </template>

      <template #footer>
        <button type="button" class="btn btn-primary" @click="createPasskey">Create</button>

        <button
          ref="cancelButtonRef"
          type="button"
          class="btn btn-cancel"
          @click="showPasskeyCreationModal = false"
        >
          Cancel
        </button>
      </template>
    </Modal>

  </main-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { startRegistration } from "@simplewebauthn/browser";

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
const subscribedShops = ref<components["schemas"]["SubscribedShop"][] | null>(null);
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
    alert.error("Name must be at least 5 characters");
    return;
  }

  const { error } = await api.POST("/auth/update-user", {
    body: { name: profileName.value },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? "Failed to update name");
    return;
  }

  alert.success("Profile updated successfully");
}

// Email
const emailAddress = ref(sessionData.value?.user?.email ?? "");
const emailCurrentPassword = ref("");

async function saveEmail() {
  if (emailAddress.value === sessionData.value?.user.email) {
    return;
  }

  if (!emailCurrentPassword.value) {
    alert.error("Please enter your current password");
    return;
  }

  const { error } = await api.POST("/auth/change-email", {
    body: { newEmail: emailAddress.value, currentPassword: emailCurrentPassword.value },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? "Failed to change email");
    return;
  }

  emailCurrentPassword.value = "";
  alert.success("Email updated successfully");
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
    alert.error("Please enter your current password");
    return;
  }

  if (newPassword.value.length < 8) {
    alert.error("Password must be at least 8 characters");
    return;
  }

  if (newPassword.value !== confirmPassword.value) {
    alert.error("Passwords do not match");
    return;
  }

  const { error } = await api.POST("/auth/change-password", {
    body: { currentPassword: currentPassword.value, newPassword: newPassword.value },
  });

  if (error) {
    alert.error((error as { message?: string }).message ?? "Failed to change password");
    return;
  }

  currentPassword.value = "";
  newPassword.value = "";
  confirmPassword.value = "";
  alert.success("Password changed successfully");
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

async function loadSubscribedShops() {
  try {
    const { data } = await api.GET("/account/subscribed-shops");
    subscribedShops.value = data ?? null;
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

const showAccountDeletionModal = ref(false);
const showPasskeyCreationModal = ref(false);

async function deleteUser() {
  if (deleteCurrentPassword.value === "") {
    alert.error("Please provide your current password to delete your account.");
    return;
  }

  try {
    const { error } = await api.POST("/auth/delete-user");

    if (error) {
      alert.error("An error occurred while deleting your account.");
      return;
    }

    alert.success("Your account has been successfully deleted.");
    setTimeout(() => {
      window.location.reload();
    }, 2000);
    showAccountDeletionModal.value = false;
  } catch {
    alert.error("An error occurred while deleting your account.");
  }
}

async function createPasskey() {
  if (!passKeyName.value) {
    alert.error("Please provide a name for the Passkey Device.");
    return;
  }

  try {
    // Get registration options from server
    const { data: optionsData, error: optionsError } = await api.POST(
      "/auth/passkey/register-options",
    );

    if (optionsError || !optionsData) {
      alert.error("Failed to get passkey registration options");
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
      alert.error("Failed to register passkey");
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
      alert.success(`Successfully unlinked your account from ${providerId}`);
    } else {
      alert.error("Failed to unlink account");
    }
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unsubscribeFromShop(shopId: number) {
  try {
    await api.DELETE("/shops/{shopId}/subscribe", {
      params: { path: { shopId } },
    });
    await loadSubscribedShops();
    alert.success("Successfully unsubscribed from shop notifications");
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

onMounted(() => {
  loadPasskeys();
  loadSessions();
  loadLinkedAccounts();
  loadOrganizations();
  loadSubscribedShops();
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
