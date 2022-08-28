<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';
import Spinner from '@/components/icon/Spinner.vue';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

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
    <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" :initial-values="shops" @submit="onSubmit">
      <FormGroup title="Shop information" subTitle="">
        <div class="sm:col-span-6">
          <label for="Name" class="block text-sm font-medium text-gray-700 mb-1"> Name </label>
          <Field type="text" name="name" id="name" autocomplete="username" class="field"
            v-bind:class="{ 'is-invalid': errors.name }" />
          <div class="text-red-700">
            {{ errors.name }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="teamId" class="block text-sm font-medium text-gray-700 mb-1"> Team </label>
          <Field as="select" id="teamId" name="teamId" class="field">
            <option v-for="team in authStore.user.teams" :value="team.id" :key="team.id">{{ team.name }}</option>
          </Field>
          <div class="text-red-700">
            {{ errors.teamId }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="shop_url" class="block text-sm font-medium text-gray-700 mb-1"> URL </label>
          <Field type="text" name="shop_url" id="shop_url" autocomplete="url" class="field"
            v-bind:class="{ 'is-invalid': errors.shop_url }" />
          <div class="text-red-700">
            {{ errors.shop_url }}
          </div>
        </div>

      </FormGroup>

      <FormGroup title="Integration" info="The Shopware Integration must be created on an User to be able access the Extension Manager">
        <div class="sm:col-span-6">
          <label for="client_id" class="block text-sm font-medium text-gray-700 mb-1"> Client-ID </label>
          <Field type="text" name="client_id" id="client_id" class="field"
            v-bind:class="{ 'is-invalid': errors.client_id }" />
          <div class="text-red-700">
            {{ errors.client_id }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="client_secret" class="block text-sm font-medium text-gray-700 mb-1"> Client-Secret </label>
          <Field type="text" name="client_secret" id="client_secret" class="field"
            v-bind:class="{ 'is-invalid': errors.client_secret }" />
          <div class="text-red-700">
            {{ errors.client_secret }}
          </div>
        </div>
      </FormGroup>

      <div class="flex justify-end">
        <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
          <Spinner v-if="isSubmitting" class="mr-2" />
          Save
        </button>
      </div>
    </Form>
  </MainContainer>
</template>
