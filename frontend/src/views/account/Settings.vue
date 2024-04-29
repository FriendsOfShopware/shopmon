<template>
    <header title="Settings" />
    <main-container>
        <vee-form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="user"
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
                    <label for="displayName">displayName</label>
                    <field
                        id="displayName"
                        type="text"
                        name="displayName"
                        autocomplete="name"
                        class="field"
                        :class="{ 'has-error': errors.displayName }"
                    />
                    <div class="field-error-message">
                        {{ errors.displayName }}
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
            </form-group>
        </vee-form>

        <form-group title="Passkey Devices">
            <data-table
                v-if="authStore.passkeys"
                :columns="[
                    { key: 'name', name: 'Name', sortable: true },
                    { key: 'createdAt', name: 'Created At', sortable: true },
                ]"
                :data="authStore.passkeys"
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

            <button
                type="button"
                class="btn btn-primary"
                @click="createPasskey"
            >
                <icon-material-symbols:passkey
                    v-if="!isAuthenticated"
                    class="icon icon-passkey"
                    aria-hidden="true"
                />
                Add a new Device
            </button>
        </form-group>

        <form-group title="Deleting your Account">
                <p>
                    Once you delete your account, you will lose all data associated with it.
                    All owning organization will be also deleted with all shops associated.
                </p>

                <button
                    type="button"
                    class="btn btn-danger"
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
                Deactivate account
            </template>

            <template #content>
                Are you sure you want to deactivate your account? All of your data will be permanently removed
                from our servers forever. This action cannot be undone.
            </template>
            
            <template #footer>
                <button
                    type="button"
                    class="btn btn-danger"
                    @click="deleteUser"
                >
                    Deactivate
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
    </main-container>
</template>

<script setup lang="ts">
import { Form as VeeForm, Field } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';
import { trpcClient } from '@/helpers/trpc';
import { client } from '@passwordless-id/webauthn';

const authStore = useAuthStore();
const alertStore = useAlertStore();

authStore.loadPasskeys();

const user = {
    displayName: authStore.user?.displayName,
    email: authStore.user?.email,
};

const showAccountDeletionModal = ref(false);

const schema = Yup.object().shape({
    currentPassword: Yup.string().required('Current password is required'),
    email: Yup.string().email(),
    displayName: Yup.string().min(5, 'Display name must be at least 5 characters'),
    newPassword: Yup.string()
        .transform((x) => (x === '' ? undefined : x))
        .min(8, 'Password must be at least 8 characters'),
});

async function onSubmit(values: any) {
    try {
        await authStore.updateProfile(values);
        alertStore.success('User updated');
    } catch (error: any) {
        alertStore.error(error);
    }
}

async function deleteUser() {
    try {
        await authStore.delete();
    } catch (error: any) {
        alertStore.error(error);
    }

    showAccountDeletionModal.value = false;
}

async function createPasskey() {
    const challenge = await trpcClient.auth.passkey.challenge.mutate();

    const registration = await client.register(authStore.user?.email!, challenge, {
        authenticatorType: 'auto',
        userVerification: 'required',
        timeout: 60000,
        attestation: false,
    });

    await trpcClient.auth.passkey.registerDevice.mutate(registration);
    await authStore.loadPasskeys();
}

async function removePasskey(id: string) {
    await trpcClient.auth.passkey.removeDevice.mutate(id);
    await authStore.loadPasskeys();
}
</script>
