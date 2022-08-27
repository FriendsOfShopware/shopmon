<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { LockClosedIcon } from '@heroicons/vue/24/solid'
import  Spinner from '@/components/icon/Spinner.vue';

import { useAuthStore } from '@/stores/auth.store';

const schema = Yup.object().shape({
    email: Yup.string().required('Email is required'),
    password: Yup.string().required('Password is required'),
});

async function onSubmit(values: any) {
    const authStore = useAuthStore();
    const { email, password } = values;
    await authStore.login(email, password);
}
</script>

<template>
  <div>
    <h2 class="text-center text-3xl tracking-tight font-bold text-gray-900">
      Sign in to your account
    </h2>
    <p class="mt-2 text-center text-sm text-gray-600">
      Or
      {{ ' ' }}
      <router-link
        to="register"
        class="font-medium"
      >
        register
      </router-link>
    </p>
  </div>
  <Form
    v-slot="{ errors, isSubmitting }"
    class="mt-8 space-y-6"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div class="rounded-md shadow-sm -space-y-px">
      <div>
        <label
          for="email-address"
          class="sr-only"
        >Email address</label>
        <Field
          id="email-address"
          name="email"
          type="email"
          autocomplete="email"
          required=""
          class="field-no-rounded rounded-t-md"
          placeholder="Email address"
          :class="{ 'is-invalid': errors.email }"
        />
      </div>
      <div>
        <label
          for="password"
          class="sr-only"
        >Password</label>
        <Field
          id="password"
          name="password"
          type="password"
          autocomplete="current-password"
          required=""
          class="field-no-rounded rounded-b-md"
          placeholder="Password"
          :class="{ 'is-invalid': errors.password }"
        />
      </div>
    </div>

    <div class="flex items-end">
      <div class="text-sm flex">
        <router-link
          to="/account/forgot-password"
          class="font-medium"
        >
          Forgot your password?
        </router-link>
      </div>
    </div>

    <div>
      <button
        type="submit"
        class="group w-full btn btn-primary"
      >
        <span
          class="absolute left-0 inset-y-0 flex items-center pl-3"
          :disabled="isSubmitting"
        >
          <LockClosedIcon
            v-if="!isSubmitting"
            class="h-5 w-5 opacity-25 group-hover:opacity-50"
            aria-hidden="true"
          />
          <Spinner v-if="isSubmitting" />
        </span>
        Sign in
      </button>
    </div>
  </Form>
</template>