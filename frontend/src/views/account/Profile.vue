<script setup>
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { useRoute } from 'vue-router';
import { storeToRefs } from 'pinia';

import { useAuthStore, useAlertStore } from '@/stores';
import { router } from '@/router';

import { Spinner, Header, MainContainer } from '@/components';

const authStore = useAuthStore();
const alertStore = useAlertStore();

let { user } = storeToRefs(authStore);

const schema = Yup.object().shape({
  email: Yup.string().required('Username is required'),
  password: Yup.string()
    .transform((x) => (x === '' ? undefined : x))
    // password optional in edit mode
    .concat(user ? null : Yup.string().required('Password is required'))
    .min(6, 'Password must be at least 6 characters'),
});

async function onSubmit(values) {
  try {
    let message;
    if (user) {
      await usersStore.update(user.value.id, values);
      message = 'User updated';
    } else {
      await usersStore.register(values);
      message = 'User added';
    }
    await router.push('/users');
    alertStore.success(message);
  } catch (error) {
    alertStore.error(error);
  }
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
  </MainContainer>
</template>
