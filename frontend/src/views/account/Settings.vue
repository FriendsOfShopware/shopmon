<script setup lang="ts">

import Spinner from '@/components/icon/Spinner.vue'
import Header from '@/components/layout/Header.vue'
import MainContainer from '@/components/layout/MainContainer.vue';
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline'
import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const authStore = useAuthStore();
const alertStore = useAlertStore();

const user = {
    currentPassword: '',
    newPassword: '',
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
        .concat(Yup.string().required('Password is required'))
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
    <div>
      <div class="md:grid md:grid-cols-3 md:gap-6">
        <div class="md:col-span-1">
          <div class="px-4 sm:px-0">
            <h3 class="text-lg font-medium leading-6 text-gray-900">
              Account
            </h3>
            <p class="mt-1 text-sm text-gray-600">
              Manage your Account
            </p>
          </div>
        </div>
        <div class="mt-5 md:mt-0 md:col-span-2">
          <Form
            v-slot="{ errors, isSubmitting }"
            :validation-schema="schema"
            :initial-values="user"
            @submit="onSubmit"
          >
            <div class="shadow sm:rounded-md sm:overflow-hidden">
              <div class="px-4 py-5 bg-white space-y-6 sm:p-6">
                <div class="grid grid-cols-6 gap-6">
                  <div class="col-span-6">
                    <label
                      for="currentPassword"
                      class="block text-sm font-medium text-gray-700"
                    >Current Password*</label>
                    <Field 
                      id="currentPassword" 
                      type="password" 
                      name="currentPassword" 
                      autocomplete="current-password" 
                      class="mt-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
                      :class="{ 'is-invalid': errors.currentPassword }"
                    />
                    <div class="text-red-700">
                      {{ errors.currentPassword }}
                    </div>
                  </div>

                  <div class="col-span-6">
                    <label
                      for="username"
                      class="block text-sm font-medium text-gray-700"
                    >Username</label>
                    <Field 
                      id="username" 
                      type="text" 
                      name="username" 
                      autocomplete="username" 
                      class="mt-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
                      :class="{ 'is-invalid': errors.username }"
                    />
                    <div class="text-red-700">
                      {{ errors.username }}
                    </div>
                  </div>

                  <div class="col-span-6">
                    <label
                      for="email"
                      class="block text-sm font-medium text-gray-700"
                    >Email address</label>
                    <Field 
                      id="email" 
                      type="text" 
                      name="email" 
                      autocomplete="email" 
                      class="mt-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
                      :class="{ 'is-invalid': errors.email }"
                    />
                    <div class="text-red-700">
                      {{ errors.email }}
                    </div>
                  </div>

                  <div class="col-span-6">
                    <label
                      for="newPassword"
                      class="block text-sm font-medium text-gray-700"
                    >New Password</label>
                    <Field 
                      id="newPassword"
                      type="password" 
                      name="newPassword" 
                      autocomplete="new-password" 
                      class="mt-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
                      :class="{ 'is-invalid': errors.newPassword }"
                    />
                  </div>
                </div>
              </div>
              <div class="px-4 py-3 bg-gray-50 text-right sm:px-6">
                <button
                  :disabled="isSubmitting"
                  type="submit" 
                  class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-sky-500 hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500"
                >
                  Save
                </button>
              </div>
            </div>
          </Form>
        </div>
      </div>
    </div>

    <div
      class="hidden sm:block"
      aria-hidden="true"
    >
      <div class="py-5">
        <div class="border-t border-gray-200" />
      </div>
    </div>

    <div class="mt-10 sm:mt-0">
      <div class="md:grid md:grid-cols-3 md:gap-6">
        <div class="md:col-span-1">
          <div class="px-4 sm:px-0">
            <h3 class="text-lg font-medium leading-6 text-gray-900">
              Deleting your Account
            </h3>
          </div>
        </div>
        <div class="mt-5 md:mt-0 md:col-span-2">
          <form
            action="#"
            method="POST"
          >
            <div class="shadow overflow-hidden sm:rounded-md">
              <div class="px-4 py-5 bg-white space-y-6 sm:p-6">
                <div class="mt-2 max-w-xl text-sm text-gray-500">
                  <p>Once you delete your account, you will lose all data associated with it. All owning Teams will be also deleted with all shops associated.</p>
                </div>
                <div class="mt-5">
                  <button
                    type="button"
                    class="inline-flex items-center justify-center px-4 py-2 border border-transparent font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:text-sm"
                    @click="showAccountDeletionModal = true"
                  >
                    <Spinner
                      class="mr-2"
                    />
                    Delete account
                  </button>
                </div>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>

    <TransitionRoot
      as="template"
      :show="showAccountDeletionModal"
    >
      <Dialog
        as="div"
        class="relative z-10"
        @close="showAccountDeletionModal = false"
      >
        <TransitionChild
          as="template"
          enter="ease-out duration-300"
          enter-from="opacity-0"
          enter-to="opacity-100"
          leave="ease-in duration-200"
          leave-from="opacity-100"
          leave-to="opacity-0"
        >
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>

        <div class="fixed z-10 inset-0 overflow-y-auto">
          <div class="flex items-end sm:items-center justify-center min-h-full p-4 text-center sm:p-0">
            <TransitionChild
              as="template"
              enter="ease-out duration-300"
              enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enter-to="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-200"
              leave-from="opacity-100 translate-y-0 sm:scale-100"
              leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <DialogPanel class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
                <div class="sm:flex sm:items-start">
                  <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                    <ExclamationTriangleIcon
                      class="h-6 w-6 text-red-600"
                      aria-hidden="true"
                    />
                  </div>
                  <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                    <DialogTitle
                      as="h3"
                      class="text-lg leading-6 font-medium text-gray-900"
                    >
                      Deactivate account
                    </DialogTitle>
                    <div class="mt-2">
                      <p class="text-sm text-gray-500">
                        Are you sure you want to deactivate your account? All of your data will be permanently removed from our servers forever. This action cannot be undone.
                      </p>
                    </div>
                  </div>
                </div>
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                  <button
                    type="button"
                    class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm"
                    @click="deleteUser"
                  >
                    Deactivate
                  </button>
                  <button
                    ref="cancelButtonRef"
                    type="button"
                    class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:w-auto sm:text-sm"
                    @click="showAccountDeletionModal = false"
                  >
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
