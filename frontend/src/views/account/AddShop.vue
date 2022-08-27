<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import Spinner from '@/components/icon/Spinner.vue';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';
import { InformationCircleIcon } from '@heroicons/vue/20/solid'
import { Form, Field } from 'vee-validate';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const authStore = useAuthStore();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const router = useRouter();

const schema = Yup.object().shape({
    name: Yup.string().required('Shop name is required'),
    shop_url: Yup.string().required('Shop name is required').url(),
    teamId: Yup.number().required('Team is required'),
    client_id: Yup.string().required('Client ID is required'),
    client_secret: Yup.string().required('Client Secret is required'),
});

const shops = {
  teamId: authStore.user?.teams[0]?.id,
}

async function onSubmit(values: any) {
  try {
    await shopStore.createShop(values.teamId, values);

    router.push('/account/shops');
  } catch (e: any) {
    alertStore.error(e);
  }
}

</script>

<template>
  <Header title="New Shop" />
  <MainContainer v-if="authStore.user">
    <Form
      v-slot="{ errors, isSubmitting }" 
      :validation-schema="schema"
      :initial-values="shops"
      @submit="onSubmit"
      class="space-y-8 divide-y divide-gray-200"
    >
      <div class="space-y-8 divide-y divide-gray-200">
        <div>
          <div class="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
            <div class="sm:col-span-6">
              <label for="Name" class="block text-sm font-medium text-gray-700"> Name </label>
              <div class="mt-1 flex rounded-md shadow-sm">
                <Field
                  type="text" 
                  name="name" 
                  id="name" 
                  autocomplete="username" 
                  class="flex-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full min-w-0 rounded-none rounded-r-md sm:text-sm border-gray-300"
                  v-bind:class="{ 'is-invalid': errors.name }"
                  />
              </div>
              <div class="text-red-700">
                {{ errors.name }}
              </div>
            </div>

            <div class="sm:col-span-6">
              <label for="teamId" class="block text-sm font-medium text-gray-700"> Team </label>
              <div class="mt-1">
                <Field as="select" id="teamId" name="teamId" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md">
                  <option v-for="team in authStore.user.teams" :value="team.id" :key="team.id">{{ team.name }}</option>
                </Field>
                <div class="text-red-700">
                  {{ errors.teamId }}
                </div>
              </div>
            </div>

            <div class="sm:col-span-6">
              <label for="shop_url" class="block text-sm font-medium text-gray-700"> URL </label>
              <div class="mt-1 flex rounded-md shadow-sm">
                <Field 
                  type="text" 
                  name="shop_url" 
                  id="shop_url" 
                  autocomplete="url" 
                  class="flex-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full min-w-0 rounded-none rounded-r-md sm:text-sm border-gray-300"
                  v-bind:class="{ 'is-invalid': errors.shop_url }"
                />
              </div>
              <div class="text-red-700">
                {{ errors.shop_url }}
              </div>
            </div>
          </div>
        </div>

        <div class="pt-8">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">Integration</h3>
            <div class="rounded-md bg-blue-50 p-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <InformationCircleIcon class="h-5 w-5 text-blue-400" aria-hidden="true" />
                </div>
                <div class="ml-3 flex-1 md:flex md:justify-between">
                  <p class="text-sm text-blue-700">The Shopware Integration must be created on an User to be able access the Extension Manager</p>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
            <div class="sm:col-span-6">
              <label for="client_id" class="block text-sm font-medium text-gray-700"> Client-ID </label>
              <div class="mt-1 flex rounded-md shadow-sm">
                <Field 
                  type="text" 
                  name="client_id" 
                  id="client_id" 
                  class="flex-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full min-w-0 rounded-none rounded-r-md sm:text-sm border-gray-300"
                  v-bind:class="{ 'is-invalid': errors.client_id }"
                />
              </div>
              <div class="text-red-700">
                  {{ errors.client_id }}
                </div>
            </div>

            <div class="sm:col-span-6">
              <label for="client_secret" class="block text-sm font-medium text-gray-700"> Client-Secret </label>
              <div class="mt-1 flex rounded-md shadow-sm">
                <Field 
                  type="text" 
                  name="client_secret" 
                  id="client_secret" 
                  class="flex-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full min-w-0 rounded-none rounded-r-md sm:text-sm border-gray-300"
                  v-bind:class="{ 'is-invalid': errors.client_secret }" 
                />
              </div>
              <div class="text-red-700">
                {{ errors.client_secret }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="pt-5">
        <div class="flex justify-end">
          <button 
            :disabled="isSubmitting"
            type="submit" 
            class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
          <Spinner
            v-if="isSubmitting"
            class="mr-2"
          />
          Save
        </button>
        </div>
      </div>
    </Form>
  </MainContainer>
</template>
