<template>
    <header-container title="Settings" />
    <main-container>
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="user"
            class="panel"
            @submit="onSubmit"
        >
            <form-group
                title="Account"
                sub-title="Manage Your Account"
            >
                <PasswordField
                    name="currentPassword"
                    label="Current Password"
                    :error="errors.currentPassword"
                />

                <div>
                    <label for="name">Name</label>
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
                    <label for="email">Email address</label>
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
                    label="New Password"
                    :error="errors.newPassword"
                />

                <p>Login using your GitHub account</p>

                <button
                    v-if="!connectedProviders.includes('github')"
                        type="button"
                        class="btn btn-primary"
                        @click="linkSocial('github')"
                    >
                    <icon-fa6-brands:github
                        class="icon"
                        aria-hidden="true"
                    />
                    Link GitHub
                </button>
            
                <button
                    v-else
                    type="button"
                    class="btn btn-danger"
                    @click="unlinkSocial('github')"
                >
                    <icon-fa6-brands:github
                        class="icon"
                        aria-hidden="true"
                    />
                    Unlink from GitHub
                </button>
            </form-group>

            <div class="form-submit">
                <button
                    :disabled="isSubmitting"
                    type="submit"
                    class="btn btn-primary"
                >
                    <icon-fa6-solid:floppy-disk
                        v-if="!isSubmitting"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-line-md:loading-twotone-loop
                        v-else
                        class="icon"
                    />
                    Save
                </button>
            </div>
        </vee-form>

        <div class="panel">
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
                            class="tooltip-position-left"
                            data-tooltip="Delete"
                            @click="removePasskey(row.id)"
                        >
                            <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
                        </button>
                    </template>
                </data-table>
            </form-group>

            <div class="form-submit">
                <button
                    type="button"
                    class="btn btn-primary"
                    @click="showPasskeyCreationModal = true"
                >
                    <icon-material-symbols:passkey
                        class="icon icon-passkey"
                        aria-hidden="true"
                    />
                    Add a new Device
                </button>
            </div>
        </div>

        <form-group title="Sessions" class="form-group-table panel">
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
                        v-if="row.token !== session.data?.session.token"
                        type="button"
                        class="tooltip-position-left"
                        data-tooltip="Delete"
                        @click="removeSession(row)"
                    >
                        <icon-fa6-solid:trash aria-hidden="true" class="icon icon-error" />
                    </button>
                </template>
            </data-table>
        </form-group>

        <form-group title="Notifications" class="form-group-table panel">
            <div v-if="!subscribedShops || subscribedShops.length === 0" class="empty-state">
                <icon-fa6-regular:bell-slash class="empty-state-icon" />
                <p class="empty-state-text">
                    You are not subscribed to any shop notifications.
                </p>

                <p class="empty-state-subtext">
                    Visit a shop's detail page and click the watch button to receive notifications about changes.
                </p>
            </div>

            <data-table
                v-else
                :columns="[
                    { key: 'name', name: 'Shop Name', sortable: true },
                    { key: 'organizationName', name: 'Organization', sortable: true },
                    { key: 'shopwareVersion', name: 'Version', sortable: true },
                ]"
                :data="subscribedShops"
            >
                <template #cell-name="{ row }">
                    <router-link
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                slug: row.organizationSlug,
                                shopId: row.id
                            }
                        }"
                        class="link"
                    >
                        {{ row.name }}
                    </router-link>
                </template>

                <template #cell-actions="{ row }">
                    <button
                        type="button"
                        class="tooltip-position-left"
                        data-tooltip="Unsubscribe"
                        @click="unsubscribeFromShop(row.id)"
                    >
                        <icon-fa6-solid:bell-slash aria-hidden="true" class="icon icon-error" />
                    </button>
                </template>
            </data-table>
        </form-group>

        <form-group title="Deleting your Account" class="panel">
                <Alert v-if="!canDeleteAccount" type="error">
                    To delete your account, you must delete first all organizations or leave them.
                </Alert>

                <p>
                    Once you delete your account, you will lose all data associated with it.
                </p>

                <button
                    type="button"
                    class="btn btn-danger"
                    :disabled="!canDeleteAccount"
                    @click="showAccountDeletionModal = true"
                >
                    <icon-fa6-solid:trash class="icon icon-trash" />
                    Delete account
                </button>
        </form-group>

        <Modal
            :show="showAccountDeletionModal"
            @close="showAccountDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="icon icon-error"
                    aria-hidden="true"
                />
            </template>

            <template #title>
                Delete account
            </template>

            <template #content>
                <p>
                    Are you sure you want to delete your account? All of your data will be permanently removed from our servers forever. 
                </p>

                <p class="mb-1">
                    <strong>This action cannot be undone.</strong>
                </p>

                <div>
                    <label for="deleteCurrentPassword">Current Password</label>
                    <input
                        id="deleteCurrentPassword"
                        v-model="deleteCurrentPassword"
                        type="password"
                        class="field"
                        autocomplete="off"
                    />
                </div>

            </template>
            
            <template #footer>
                <button
                    type="button"
                    class="btn btn-danger"
                    @click="deleteUser"
                >
                    Delete
                </button>

                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn btn-cancel"
                    @click="showAccountDeletionModal = false"
                >
                    Cancel
                </button>
            </template>
        </Modal>

        <Modal :show="showPasskeyCreationModal"
            @close="showPasskeyCreationModal = false">
            <template #title>
                Add a Passkey Device
            </template>

            <template #icon>
                <icon-fa6-solid:key
                    class="icon icon-info"
                    aria-hidden="true"
                />
            </template>

            <template #content>
                Please give a name to your new Passkey Device.
                <field
                        v-model="passKeyName"
                        type="text"
                        name="name"
                        autocomplete="off"
                        class="field"
                    />
            </template>

            <template #footer>
                <button
                    type="button"
                    class="btn btn-primary"
                    @click="createPasskey"
                >
                    Create
                </button>

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
import type { Passkey } from 'better-auth/plugins/passkey';
import { Field, Form as VeeForm } from 'vee-validate';
import { computed, ref } from 'vue';
import * as Yup from 'yup';

import { useAlert } from '@/composables/useAlert';
import { authClient } from '@/helpers/auth-client';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import type { Session } from 'better-auth/types';

const session = authClient.useSession();
const orgs = authClient.useListOrganizations();

const alert = useAlert();

const passKeyName = ref('');

const passkeys = ref<Passkey[] | null>([]);
const sessions = ref<Session[] | null>([]);
const connectedProviders = ref<string[]>([]);
const subscribedShops = ref<RouterOutput['account']['subscribedShops'] | null>(
    null,
);

const deleteCurrentPassword = ref('');

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
    name: session.value.data?.user?.name ?? '',
    email: session.value.data?.user?.email ?? '',
    currentPassword: '',
    newPassword: '',
};

async function loadLinkedAccounts() {
    authClient.listAccounts().then((data) => {
        if (data.data) {
            connectedProviders.value = data.data?.map(
                (account) => account.provider,
            );
        }
    });
}
loadLinkedAccounts();

async function loadSubscribedShops() {
    try {
        subscribedShops.value =
            await trpcClient.account.subscribedShops.query();
    } catch (err) {
        alert.error(err instanceof Error ? err.message : String(err));
    }
}
loadSubscribedShops();

const showAccountDeletionModal = ref(false);
const showPasskeyCreationModal = ref(false);

const schema = Yup.object().shape({
    currentPassword: Yup.string().required('Current password is required'),
    email: Yup.string().email().required(),
    name: Yup.string().min(5, 'Name must be at least 5 characters'),
    newPassword: Yup.string()
        .transform((x) => (x === '' ? undefined : x))
        .min(8, 'Password must be at least 8 characters'),
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
    if (deleteCurrentPassword.value === '') {
        alert.error('Please provide your current password to delete your account.');
        return;
    }

    const resp = await authClient.deleteUser({
        password: deleteCurrentPassword.value,
    });

    if (resp.error) {
        alert.error(resp.error.message ?? 'An error occurred while deleting your account.');
        return;
    }

    alert.success('Your account has been successfully deleted.');
    setTimeout(() => {
        window.location.reload();
    }, 2000);
    showAccountDeletionModal.value = false;
}

async function createPasskey() {
    if (!passKeyName.value) {
        alert.error('Please provide a name for the Passkey Device.');
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

async function linkSocial(provider: 'github') {
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
        alert.success(`Successfully unlinked your account from ${providerId}`);
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
        alert.success('Successfully unsubscribed from shop notifications');
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
