<script setup lang="ts">
import Header from '@/components/layout/Header.vue'
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';
import Spinner from '@/components/icon/Spinner.vue';

import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const authStore = useAuthStore();
const alertStore = useAlertStore();

const user = {
  username: authStore.user?.username,
  email: authStore.user?.email,
}

const showAccountDeletionModal = ref(false)

const schema = Yup.object().shape({
  currentPassword: Yup.string().required('Current password is required'),
  email: Yup.string().email(),
  username: Yup.string().min(5, 'Username must be at least 5 characters'),
  newPassword: Yup.string()
    .transform((x) => (x === '' ? undefined : x))
    .min(8, 'Password must be at least 8 characters'),
});

async function onSubmit(values: any) {
  try {
    await authStore.updateProfile(values);
    alertStore.success("User updated");
  } catch (error: Error) {
    alertStore.error(error);
  }
}

async function deleteUser() {
  try {
    await authStore.delete();
  } catch (error: Error) {
    alertStore.error(error);
  }

  showAccountDeletionModal.value = false;
}
</script>

<template>
  <Header title="Settings" />
  <MainContainer>
    <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" :initial-values="user" @submit="onSubmit">
      <FormGroup title="Account" subTitle="Manage Your Account">
        <div class="grid grid-cols-6 gap-6">
          <div class="col-span-6">
            <label for="currentPassword" class="block text-sm font-medium text-gray-700 mb-1">Current Password*</label>
            <Field id="currentPassword" type="password" name="currentPassword" autocomplete="current-password"
              class="field"
              :class="{ 'is-invalid': errors.currentPassword }" />
            <div class="text-red-700">
              {{ errors.currentPassword }}
            </div>
          </div>

          <div class="col-span-6">
            <label for="username" class="block text-sm font-medium text-gray-700 mb-1">Username</label>
            <Field id="username" type="text" name="username" autocomplete="username"
              class="field"
              :class="{ 'is-invalid': errors.username }" />
            <div class="text-red-700">
              {{ errors.username }}
            </div>
          </div>

          <div class="col-span-6">
            <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email address</label>
            <Field id="email" type="text" name="email" autocomplete="email"
              class="field"
              :class="{ 'is-invalid': errors.email }" />
            <div class="text-red-700">
              {{ errors.email }}
            </div>
          </div>

          <div class="col-span-6">
            <label for="newPassword" class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
            <Field id="newPassword" type="password" name="newPassword" autocomplete="new-password"
              class="field"
              :class="{ 'is-invalid': errors.newPassword }" />
          </div>
        </div>
        <div class="text-right flex justify-end">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary flex items-center group">
            <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
              <icon-fa6-solid:floppy-disk
                class="h-5 w-5" 
                aria-hidden="true" 
                v-if="!isSubmitting" 
              />
              <Spinner v-else />
            </span>
            Save
          </button>
        </div>
      </FormGroup>
    </Form>

    <FormGroup title="Deleting your Account">
      <form action="#" method="POST">

        <p>Once you delete your account, you will lose all data associated with it. All owning Teams will be also
          deleted with all shops associated.</p>

        <div class="mt-5">
          <button type="button"
            class="btn btn-danger group flex items-center"
            @click="showAccountDeletionModal = true">
            <icon-fa6-solid:trash 
              class="w-4 h-4 -ml-1 mr-2 opacity-25 group-hover:opacity-50"
            />
            Delete account
          </button>
        </div>
      </form>
    </FormGroup>


    <TransitionRoot as="template" :show="showAccountDeletionModal">
      <Dialog as="div" class="relative z-10" @close="showAccountDeletionModal = false">
        <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100"
          leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div class="fixed z-10 inset-0 overflow-y-auto">
          <div class="flex items-end sm:items-center justify-center min-h-full p-4 text-center sm:p-0">
            <TransitionChild as="template" enter="ease-out duration-300"
              enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200"
              leave-from="opacity-100 translate-y-0 sm:scale-100"
              leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
              <DialogPanel
                class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
                <div class="sm:flex sm:items-start">
                    <icon-fa6-solid:triangle-exclamation 
                      class="h-6 w-6 text-red-600" 
                      aria-hidden="true" 
                    />
                  <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <DialogTitle as="h3" class="text-lg leading-6 font-medium text-gray-900">
                      Deactivate account
                    </DialogTitle>
                    <div class="mt-2">
                      <p class="text-sm text-gray-500">
                        Are you sure you want to deactivate your account? All of your data will be permanently removed
                        from our servers forever. This action cannot be undone.
                      </p>
                    </div>
                  </div>
                </div>
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                  <button type="button"
                    class="btn btn-danger w-full sm:ml-3 sm:w-auto"
                    @click="deleteUser">
                    Deactivate
                  </button>
                  <button ref="cancelButtonRef" type="button"
                    class="btn w-full mt-3 sm:w-auto sm:mt-0"
                    @click="showAccountDeletionModal = false">
                    Cancel
                  </button>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
  </MainContainer>
</template>
