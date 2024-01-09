<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useShopStore } from '@/stores/shop.store';

import { Form, Field } from 'vee-validate';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';
import {  RouterInput } from "@/helpers/trpc";

const authStore = useAuthStore();
const shopStore = useShopStore();
const alertStore = useAlertStore();
const router = useRouter();

const schema = Yup.object().shape({
  name: Yup.string().required('Shop name is required'),
  shopUrl: Yup.string().required('Shop URL is required').url(),
  orgId: Yup.number().required('Organization is required'),
  client_id: Yup.string().required('Client ID is required'),
  client_secret: Yup.string().required('Client Secret is required'),
});

const shops = {
  orgId: authStore.user?.organizations[0]?.id,
}

async function onSubmit(values: RouterInput['organization']['shop']['create']) {
  try {
    await shopStore.createShop(values);

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
          <label for="Name" class="block text-sm font-medium mb-1"> Name </label>
          <Field type="text" name="name" id="name" class="field"
            v-bind:class="{ 'is-invalid': errors.name }" />
          <div class="text-red-700">
            {{ errors.name }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="orgId" class="block text-sm font-medium mb-1"> Organization </label>
          <Field as="select" id="orgId" name="orgId" class="field">
            <option v-for="organization in authStore.user.organizations" :value="organization.id" :key="organization.id">
              {{ organization.name }}
            </option>
          </Field>
          <div class="text-red-700">
            {{ errors.orgId }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="shopUrl" class="block text-sm font-medium mb-1"> URL </label>
          <Field type="text" name="shopUrl" id="shopUrl" autocomplete="url" class="field"
            v-bind:class="{ 'is-invalid': errors.shop_url }" />
          <div class="text-red-700">
            {{ errors.shopUrl }}
          </div>
        </div>

      </FormGroup>

      <FormGroup title="Integration"
        info="<p>The created integration must have access to following <a href='https://github.com/FriendsOfShopware/shopmon/blob/main/app/manifest.xml#L18'>permissions</a></p>">
        <div class="sm:col-span-6">
          <label for="clientId" class="block text-sm font-medium mb-1"> Client-ID </label>
          <Field type="text" name="clientId" id="clientId" class="field"
            v-bind:class="{ 'is-invalid': errors.clientId }" />
          <div class="text-red-700">
            {{ errors.clientId }}
          </div>
        </div>

        <div class="sm:col-span-6">
          <label for="clientSecret" class="block text-sm font-medium mb-1"> Client-Secret </label>
          <Field type="text" name="clientSecret" id="clientSecret" class="field"
            v-bind:class="{ 'is-invalid': errors.clientSecret }" />
          <div class="text-red-700">
            {{ errors.clientSecret }}
          </div>
        </div>
      </FormGroup>

      <div class="flex justify-end group">
        <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
          <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
            <icon-fa6-solid:floppy-disk class="h-5 w-5" aria-hidden="true" v-if="!isSubmitting" />
            <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
          </span>
          Save
        </button>
      </div>
    </Form>
  </MainContainer>
</template>
