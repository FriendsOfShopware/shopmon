<script setup lang="ts">
import * as Yup from 'yup';

import { useAuthStore, useAlertStore } from '@/stores';
import { router } from '@/router';

const schema = Yup.object().shape({
  email: Yup.string().required('Email is required')
});

async function onSubmit(values) {
  const usersStore = useAuthStore();
  const alertStore = useAlertStore();
  try {
    await usersStore.resetPassword(values);
    await router.push('/account/login');
    alertStore.success('Registration successful');
  } catch (error) {
    alertStore.error(error);
  }
}
</script>

<template>
<div>
  <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">Forgot password</h4>
  <p>We will send you a confirmation email. Click on the link in it to change your password.</p>
  </div>
  <Form
    @submit="onSubmit"
    :validation-schema="schema"
    v-slot="{ errors, isSubmitting }"
  >
    <div class="mb-2">
      <Field
        name="email"
        placeholder="Email"
        type="text"
        class="field"
        :class="{ 'is-invalid': errors.email }"
      />
      <div class="text-red-700">{{ errors.email }}</div>
    </div>
    <div class="">
      <button class="btn btn-primary" :disabled="isSubmitting">
        <span
          class="absolute left-0 inset-y-0 flex items-center pl-3"
          :disabled="isSubmitting"
          v-if="isSubmitting"
        >
          <Spinner />
        </span>
        Send email
      </button>
      <router-link
        to="login"
        class="text-gray-900 inline-block mt-2 center text-center w-full"
        >Cancel</router-link
      >
    </div>
  </Form>
</template>
