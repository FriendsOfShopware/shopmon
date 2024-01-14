<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
  email: Yup.string().required('Email is required')
})

async function onSubmit(values: any): Promise<void> {
  const usersStore = useAuthStore();
  const alertStore = useAlertStore();

  try {
    await usersStore.resetPassword(values.email);
    alertStore.success('Password reset email sent');
  } catch (error: Error) {
    alertStore.error(error.message);
  }
}
</script>

<template>
  <div>
    <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
      Forgot password
    </h4>
    <p>We will send you a confirmation email. Click on the link in it to change your password.</p>
  </div>

  <Form class="space-y-4" v-slot="{ errors, isSubmitting }" :validation-schema="schema" @submit="onSubmit">
    <div>
      <Field name="email" placeholder="Email address" type="text" class="field" :class="{ 'is-invalid': errors.email }" />
      <div class="text-red-700">
        {{ errors.email }}
      </div>
    </div>

    <button class="btn btn-primary w-full group" :disabled="isSubmitting">
      <span class="relative -ml-7 mr-2 opacity-40 group-hover:opacity-60">
        <icon-fa6-solid:envelope class="h-5 w-5" aria-hidden="true" v-if="!isSubmitting" />
        <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
      </span>
      Send email
    </button>

    <router-link to="login" class="inline-block center text-center w-full">
      Cancel
    </router-link>
  </Form>
</template>
