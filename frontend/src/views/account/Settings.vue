<template>
  <header-container :title="$t('settings.title')" />
  <main-container>
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="user"
        @submit="onSubmit"
      >
        <form-group :title="$t('settings.account')" :sub-title="$t('settings.manageAccount')">
          <PasswordField
            name="currentPassword"
            :label="$t('settings.currentPassword')"
            :error="errors.currentPassword"
          />

          <div>
            <label for="name">{{ $t("common.name") }}</label>
            <field
              id="name"
              type="text"
              name="name"
              autocomplete="name"
              class="field"
              :class="{ 'has-error': errors.name }"
            />
            <div class="field-error-message">
              {{ errors.name }}
            </div>
          </div>

          <div>
            <label for="email">{{ $t("common.emailAddress") }}</label>
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
          </div>

          <PasswordField
            name="newPassword"
            :label="$t('settings.newPassword')"
            :error="errors.newPassword"
          />

          <p>{{ $t("settings.githubLoginText") }}</p>

          <button
            v-if="!connectedProviders.includes('github')"
            type="button"
            class="btn btn-primary"
            @click="linkSocial('github')"
          >
            <icon-fa6-brands:github class="icon" aria-hidden="true" />
            {{ $t("settings.linkGithub") }}
          </button>

          <button v-else type="button" class="btn btn-danger" @click="unlinkSocial('github')">
            <icon-fa6-brands:github class="icon" aria-hidden="true" />
            {{ $t("settings.unlinkGithub") }}
          </button>
        </form-group>

        <div class="form-submit">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </button>
        </div>
      </vee-form>
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
            <button
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('common.delete')"
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
          {{ $t("settings.addDevice") }}
        </button>
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
            <button
              v-if="row.token !== session.data?.session.token"
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('common.delete')"
              @click="removeSession(row)"
            >
              <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
            </button>
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
          :columns="[
            { key: 'name', name: $t('settings.shopName'), sortable: true },
            { key: 'organizationName', name: $t('settings.organization'), sortable: true },
            { key: 'shopwareVersion', name: $t('common.version'), sortable: true },
          ]"
          :data="subscribedShops"
        >
          <template #cell-name="{ row }">
            <router-link
              :to="{
                name: 'account.shops.detail',
                params: {
                  slug: row.organizationSlug,
                  shopId: row.id,
                },
              }"
              class="link"
            >
              {{ row.name }}
            </router-link>
          </template>

          <template #cell-actions="{ row }">
            <button
              type="button"
              class="tooltip-top-left"
              :data-tooltip="$t('settings.unsubscribe')"
              @click="unsubscribeFromShop(row.id)"
            >
              <icon-fa6-solid:bell-slash aria-hidden="true" class="icon icon-error" />
            </button>
          </template>
        </data-table>
      </form-group>
    </Panel>

    <Panel :title="$t('settings.deleteAccountTitle')">
      <Alert v-if="!canDeleteAccount" type="error">
        {{ $t("settings.deleteAccountOrgWarning") }}
      </Alert>

      <p>{{ $t("settings.deleteAccountWarning") }}</p>

      <button
        type="button"
        class="btn btn-danger"
        :disabled="!canDeleteAccount"
        @click="showAccountDeletionModal = true"
      >
        <icon-fa6-solid:trash class="icon icon-trash" />
        {{ $t("settings.deleteAccount") }}
      </button>
    </Panel>

    <delete-confirmation-modal
      :show="showAccountDeletionModal"
      :title="$t('settings.deleteAccount')"
      entity-name="your account"
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
        <field v-model="passKeyName" type="text" name="name" autocomplete="off" class="field" />
      </template>

      <template #footer>
        <button type="button" class="btn btn-primary" @click="createPasskey">
          {{ $t("common.create") }}
        </button>

        <button
          ref="cancelButtonRef"
          type="button"
          class="btn btn-cancel"
          @click="showPasskeyCreationModal = false"
        >
          {{ $t("common.cancel") }}
        </button>
      </template>
    </Modal>
  </main-container>
</template>

<script setup lang="ts">
import type { Passkey } from "better-auth/plugins/passkey";
import { Field, Form as VeeForm } from "vee-validate";
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import * as Yup from "yup";

import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";
import { type RouterOutput, trpcClient } from "@/helpers/trpc";
import type { Session } from "better-auth/types";

const { t } = useI18n();

const session = authClient.useSession();
const orgs = authClient.useListOrganizations();

const alert = useAlert();

const passKeyName = ref("");

const passkeys = ref<Passkey[] | null>([]);
const sessions = ref<Session[] | null>([]);
const connectedProviders = ref<string[]>([]);
const subscribedShops = ref<RouterOutput["account"]["subscribedShops"] | null>(null);

const deleteCurrentPassword = ref("");

const canDeleteAccount = computed(() => {
  return orgs.value?.data?.length === 0;
});

authClient.passkey.listUserPasskeys().then((data) => {
  passkeys.value = data.data;
});

authClient.listSessions().then((data) => {
  sessions.value = data.data;
});

const user = {
  name: session.value.data?.user?.name ?? "",
  email: session.value.data?.user?.email ?? "",
  currentPassword: "",
  newPassword: "",
};

async function loadLinkedAccounts() {
  authClient.listAccounts().then((data) => {
    if (data.data) {
      connectedProviders.value = data.data?.map((account) => account.provider);
    }
  });
}
loadLinkedAccounts();

async function loadSubscribedShops() {
  try {
    subscribedShops.value = await trpcClient.account.subscribedShops.query();
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}
loadSubscribedShops();

const showAccountDeletionModal = ref(false);
const showPasskeyCreationModal = ref(false);

const schema = Yup.object().shape({
  currentPassword: Yup.string().required(t("settings.currentPasswordRequired")),
  email: Yup.string().email().required(),
  name: Yup.string().min(5, t("validation.nameMinLength")),
  newPassword: Yup.string()
    .transform((x) => (x === "" ? undefined : x))
    .min(8, t("validation.passwordMinLength")),
});

async function onSubmit(values: Record<string, unknown>) {
  await authClient.changeEmail({
    newEmail: values.email as string,
  });

  if (values.displayName) {
    await authClient.updateUser({
      name: values.displayName as string,
    });
  }

  if (values.newPassword) {
    await authClient.changePassword({
      currentPassword: values.currentPassword as string,
      newPassword: values.newPassword as string,
      revokeOtherSessions: true,
    });
  }
}

async function deleteUser() {
  if (deleteCurrentPassword.value === "") {
    alert.error(t("settings.passwordRequiredForDelete"));
    return;
  }

  const resp = await authClient.deleteUser({
    password: deleteCurrentPassword.value,
  });

  if (resp.error) {
    alert.error(resp.error.message ?? t("settings.errorDeleteAccount"));
    return;
  }

  alert.success(t("settings.accountDeleted"));
  setTimeout(() => {
    window.location.reload();
  }, 2000);
  showAccountDeletionModal.value = false;
}

async function createPasskey() {
  if (!passKeyName.value) {
    alert.error(t("settings.passkeyNameRequired"));
    return;
  }

  await authClient.passkey.addPasskey({ name: passKeyName.value });
  authClient.passkey.listUserPasskeys().then((data) => {
    passkeys.value = data.data;
  });
  showPasskeyCreationModal.value = false;
}

async function removePasskey(id: string) {
  await authClient.passkey.deletePasskey({ id });
  authClient.passkey.listUserPasskeys().then((data) => {
    passkeys.value = data.data;
  });
}

async function removeSession(session: Session) {
  await authClient.revokeSession({ token: session.token });
  authClient.listSessions().then((data) => {
    sessions.value = data.data;
  });
}

async function linkSocial(provider: "github") {
  try {
    await authClient.linkSocial({
      provider,
      callbackURL: window.location.href,
    });
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unlinkSocial(providerId: string) {
  try {
    await authClient.unlinkAccount({ providerId });
    await loadLinkedAccounts();
    alert.success(t("settings.unlinkedProvider", { providerId }));
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}

async function unsubscribeFromShop(shopId: number) {
  try {
    await trpcClient.organization.shop.unsubscribeFromNotifications.mutate({
      shopId,
    });
    await loadSubscribedShops();
    alert.success(t("settings.unsubscribedShop"));
  } catch (err) {
    alert.error(err instanceof Error ? err.message : String(err));
  }
}
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
</style>
