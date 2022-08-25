<script setup>
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { useUsersStore, useAlertStore } from '@/stores';
import { router } from '@/router';

import { Spinner } from '@/components/icon';

const schema = Yup.object().shape({
  username: Yup.string().required('Username is required'),
  email: Yup.string().required('Email is required'),
  password: Yup.string()
    .required('Password is required')
    .min(6, 'Password must be at least 6 characters'),
});

async function onSubmit(values) {
  const usersStore = useUsersStore();
  const alertStore = useAlertStore();
  try {
    await usersStore.register(values);
    await router.push('/account/login');
    alertStore.success('Registration successful');
  } catch (error) {
    alertStore.error(error);
  }
}
</script>

<template>
  <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">Register</h4>
  <Form
    @submit="onSubmit"
    :validation-schema="schema"
    v-slot="{ errors, isSubmitting }"
  >
    <div class="mb-2">
      <Field
        name="username"
        placeholder="Username"
        type="text"
        class="field"
        :class="{ 'is-invalid': errors.username }"
      />
      <div class="text-red-700">{{ errors.username }}</div>
    </div>
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
    <div class="mb-2">
      <Field
        name="password"
        type="password"
        placeholder="Password"
        class="field"
        :class="{ 'is-invalid': errors.password }"
      />
      <div class="text-red-700">{{ errors.password }}</div>
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
        Register
      </button>
      <router-link
        to="login"
        class="text-gray-900 inline-block mt-2 center text-center w-full"
        >Cancel</router-link
      >
    </div>
  </Form>
</template>
