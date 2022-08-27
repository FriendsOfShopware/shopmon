<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import Spinner from '@/components/icon/Spinner.vue';
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
  <Form
    v-slot="{ errors, isSubmitting }"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div class="mb-2">
      <Field
        name="email"
        placeholder="Email"
        type="text"
        class="field"
        :class="{ 'is-invalid': errors.email }"
      />
      <div class="text-red-700">
        {{ errors.email }}
      </div>
    </div>
    <div class="">
      <button
        class="btn btn-primary w-full"
        :disabled="isSubmitting"
      >
        <span
          v-if="isSubmitting"
          class="absolute left-0 inset-y-0 flex items-center pl-3"
          :disabled="isSubmitting"
        >
          <Spinner />
        </span>
        Send email
      </button>
      <router-link
        to="login"
        class="text-gray-900 inline-block mt-2 center text-center w-full"
      >
        Cancel
      </router-link>
    </div>
  </Form>
</template>
