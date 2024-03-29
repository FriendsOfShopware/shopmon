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
                <div class="grid grid-cols-6 gap-6">
                    <div class="col-span-6">
                        <label
                            for="currentPassword"
                            class="block text-sm font-medium mb-1"
                        >Current Password*</label>
                        <field
                            id="currentPassword"
                            type="password"
                            name="currentPassword"
                            autocomplete="current-password"
                            class="field"
                            :class="{ 'is-invalid': errors.currentPassword }"
                        />
                        <div class="text-red-700">
                            {{ errors.currentPassword }}
                        </div>
                    </div>

                    <div class="col-span-6">
                        <label
                            for="displayName"
                            class="block text-sm font-medium mb-1"
                        >displayName</label>
                        <field
                            id="displayName"
                            type="text"
                            name="displayName"
                            autocomplete="name"
                            class="field"
                            :class="{ 'is-invalid': errors.displayName }"
                        />
                        <div class="text-red-700">
                            {{ errors.displayName }}
                        </div>
                    </div>

                    <div class="col-span-6">
                        <label
                            for="email"
                            class="block text-sm font-medium mb-1"
                        >Email address</label>
                        <field
                            id="email"
                            type="text"
                            name="email"
                            autocomplete="email"
                            class="field"
                            :class="{ 'is-invalid': errors.email }"
                        />
                        <div class="text-red-700">
                            {{ errors.email }}
                        </div>
                    </div>

                    <div class="col-span-6">
                        <label
                            for="newPassword"
                            class="block text-sm font-medium mb-1"
                        >New Password</label>
                        <field
                            id="newPassword"
                            type="password"
                            name="newPassword"
                            autocomplete="new-password"
                            class="field"
                            :class="{ 'is-invalid': errors.newPassword }"
                        />
                    </div>
                </div>
                <div class="text-right flex justify-end">
                    <button
                        :disabled="isSubmitting"
                        type="submit"
                        class="btn btn-primary flex items-center group"
                    >
                        <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                            <icon-fa6-solid:floppy-disk
                                v-if="!isSubmitting"
                                class="h-5 w-5"
                                aria-hidden="true"
                            />
                            <icon-line-md:loading-twotone-loop
                                v-else
                                class="w-5 h-5"
                            />
                        </span>
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
                        class="tooltip-position-left text-red-600 opacity-50 dark:text-red-400 hover:opacity-100"
                        data-tooltip="Delete"
                        @click="removePasskey(row.id)"
                    >
                        <icon-fa6-solid:trash aria-hidden="true" />
                    </button>
                </template>
            </data-table>

            <button
                type="button"
                class="btn btn-primary"
                @click="createPasskey"
            >
                Add a new Device
            </button>
        </form-group>

        <form-group title="Deleting your Account">
            <form
                action="#"
                method="POST"
            >
                <p>
                    Once you delete your account, you will lose all data associated with it.
                    All owning organization will be also deleted with all shops associated.
                </p>

                <div class="mt-5">
                    <button
                        type="button"
                        class="btn btn-danger group flex items-center"
                        @click="showAccountDeletionModal = true"
                    >
                        <icon-fa6-solid:trash class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50" />
                        Delete account
                    </button>
                </div>
            </form>
        </form-group>

        <Modal
            :show="showAccountDeletionModal"
            @close="showAccountDeletionModal = false"
        >
            <template #icon>
                <icon-fa6-solid:triangle-exclamation
                    class="h-6 w-6 text-red-600 dark:text-red-400"
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
                    class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                    @click="deleteUser"
                >
                    Deactivate
                </button>
                <button
                    ref="cancelButtonRef"
                    type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
                    @click="showAccountDeletionModal = false"
                >
                    Cancel
                </button>
            </template>
        </Modal>
    </main-container>
</template>

<script setup lang="ts">
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

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
