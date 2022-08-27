<script setup lang="ts">
import { Spinner, Header, MainContainer } from '@/components';
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { storeToRefs } from 'pinia';
import { ref } from 'vue';

import { useAuthStore, useAlertStore } from '@/stores';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { ExclamationTriangleIcon } from '@heroicons/vue/24/outline'

const authStore = useAuthStore();
const alertStore = useAlertStore();

let { user } = storeToRefs(authStore);

const isSubmitting = ref(false);
let showAccountDeletionModal = ref(false)

const schema = Yup.object().shape({
  email: Yup.string().required('Username is required'),
  password: Yup.string()
    .transform((x) => (x === '' ? undefined : x))
    .concat(user ? null : Yup.string().required('Password is required'))
    .min(8, 'Password must be at least 8 characters'),
});

async function onSubmit(values: any) {
  try {
    alertStore.success("User updated");
  } catch (error) {
    alertStore.error(error);
  }
}

async function deleteUser() {
  try {
    await authStore.delete();
  } catch (error) {
    alertStore.error(error);
  }

  showAccountDeletionModal.value = false;
}

</script>

<template>
  <Header title="Profile" />
  <MainContainer>
    <template v-if="!(user?.loading || user?.error)">
      <Form @submit="onSubmit" :validation-schema="schema" :initial-values="user" v-slot="{ errors, isSubmitting }">
        <div class="grid grid-cols-3 gap-4 mb-4">
          <div class="col-span-3">
            <label for="email-address" class="block text-sm font-medium text-gray-700">Email address</label>
            <Field name="email" type="text" class="field" :class="{ 'is-invalid': errors.email }" />
            <div class="invalid-feedback">{{ errors.email }}</div>
          </div>

          <div class="col-span-3">
            <label for="email-address" class="block text-sm font-medium text-gray-700">Password</label>
            <Field name="password" type="password" class="field" :class="{ 'is-invalid': errors.password }" />
            <div class="invalid-feedback">{{ errors.password }}</div>
          </div>
        </div>

        <div class="flex justify-between">          
          <router-link to="/" custom v-slot="{ navigate }">
            <button @click="navigate" @keypress.enter="navigate" role="link" class="btn">Cancel</button>
          </router-link>

          <button class="btn btn-primary w-48" :disabled="isSubmitting">
            <span class="absolute left-0 inset-y-0 flex items-center pl-3" :disabled="isSubmitting" v-if="isSubmitting">
              <Spinner />
            </span>
            Save
          </button>
        </div>
      </Form>

      <div class="bg-white shadow sm:rounded-lg mt-5">
        <div class="px-4 py-5 sm:p-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900">Delete your account</h3>
          <div class="mt-2 max-w-xl text-sm text-gray-500">
            <p>Once you delete your account, you will lose all data associated with it.</p>
          </div>
          <div class="mt-5">
            <button
              @click="showAccountDeletionModal = true" 
              type="button"
             :disabled="isSubmitting" class="inline-flex items-center justify-center px-4 py-2 border border-transparent font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:text-sm">
              <Spinner v-if="isSubmitting" class="mr-2" />
              Delete account
            </button>
          </div>
        </div>
      </div>
    </template>
    <template v-if="user?.loading">
      <div class="text-center m-5">
        <Spinner />
      </div>
    </template>
    <template v-if="user?.error">
      <div class="text-center m-5">
        <div class="text-danger">Error loading user: {{ user.error }}</div>
      </div>
    </template>


      <TransitionRoot as="template" :show="showAccountDeletionModal">
        <Dialog as="div" class="relative z-10" @close="showAccountDeletionModal = false">
          <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
            <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </TransitionChild>

          <div class="fixed z-10 inset-0 overflow-y-auto">
            <div class="flex items-end sm:items-center justify-center min-h-full p-4 text-center sm:p-0">
              <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
                <DialogPanel class="relative bg-white rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:max-w-lg sm:w-full sm:p-6">
                  <div class="sm:flex sm:items-start">
                    <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                      <ExclamationTriangleIcon class="h-6 w-6 text-red-600" aria-hidden="true" />
                    </div>
                    <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                      <DialogTitle as="h3" class="text-lg leading-6 font-medium text-gray-900"> Deactivate account </DialogTitle>
                      <div class="mt-2">
                        <p class="text-sm text-gray-500">Are you sure you want to deactivate your account? All of your data will be permanently removed from our servers forever. This action cannot be undone.</p>
                      </div>
                    </div>
                  </div>
                  <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm" @click="deleteUser">Deactivate</button>
                    <button type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:w-auto sm:text-sm" @click="showAccountDeletionModal = false" ref="cancelButtonRef">Cancel</button>
                  </div>
                </DialogPanel>
              </TransitionChild>
            </div>
          </div>
      </Dialog>
    </TransitionRoot>
  </MainContainer>
</template>
