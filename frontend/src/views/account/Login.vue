<script setup>
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { LockClosedIcon } from '@heroicons/vue/solid';
import { Spinner } from '@/components/icon';

import { useAuthStore } from '@/stores';

const schema = Yup.object().shape({
  email: Yup.string().required('Username is required'),
  password: Yup.string().required('Password is required'),
});

async function onSubmit(values) {
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
      <router-link to="register" class="font-medium">register</router-link>
    </p>
  </div>
  <Form
    class="mt-8 space-y-6"
    @submit="onSubmit"
    :validation-schema="schema"
    v-slot="{ errors, isSubmitting }"
  >
    <input type="hidden" name="remember" value="true" />
    <div class="rounded-md shadow-sm -space-y-px">
      <div>
        <label for="email-address" class="sr-only">Email address</label>
        <Field
          id="email-address"
          name="email"
          type="email"
          autocomplete="email"
          required=""
          class="field-no-rounded rounded-t-md"
          placeholder="Email address"
        />
      </div>
      <div>
        <label for="password" class="sr-only">Password</label>
        <Field
          id="password"
          name="password"
          type="password"
          autocomplete="current-password"
          required=""
          class="field-no-rounded rounded-b-md"
          placeholder="Password"
        />
      </div>
    </div>

    <div class="flex items-center justify-between">
      <div class="flex items-center">
        <input
          id="remember-me"
          name="remember-me"
          type="checkbox"
          class="
            h-4
            w-4
            text-sky-500
            focus:ring-sky-500
            border-gray-300
            rounded
          "
        />
        <label for="remember-me" class="ml-2 block text-sm text-gray-900">
          Remember me
        </label>
      </div>

      <div class="text-sm">
        <router-link to="forgot-password" class="font-medium">Forgot your password?</router-link>
      </div>
    </div>

    <div>
      <button type="submit" class="group w-full btn btn-primary">
        <span
          class="absolute left-0 inset-y-0 flex items-center pl-3"
          :disabled="isSubmitting"
        >
          <LockClosedIcon
            class="h-5 w-5 opacity-25 group-hover:opacity-50"
            aria-hidden="true"
            v-if="!isSubmitting"
          />
          <Spinner v-if="isSubmitting" />
        </span>
        Sign in
      </button>
    </div>
  </Form>
</template>
