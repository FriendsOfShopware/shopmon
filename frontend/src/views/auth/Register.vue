<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { router } from '@/router';

import Spinner from '@/components/icon/Spinner.vue';
import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
    username: Yup.string().required('Username is required'),
    email: Yup.string().required('Email is required'),
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters'),
});

async function onSubmit(values: any) {
    const authStore = useAuthStore();
    const alertStore = useAlertStore();
    try {
        await authStore.register(values);
        await router.push('/account/login');
        alertStore.success('Registration successful. Please check your mailbox and confirm your email address.');
    } catch (error: any) {
        alertStore.error(error);
    }
}
</script>

<template>
  <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
    Register
  </h4>
  <Form
    v-slot="{ errors, isSubmitting }"
    :validation-schema="schema"
    @submit="onSubmit"
  >
    <div class="mb-2">
      <Field
        name="username"
        placeholder="Username"
        type="text"
        class="field"
        :class="{ 'is-invalid': errors.username }"
      />
      <div class="text-red-700">
        {{ errors.username }}
      </div>
    </div>
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
    <div class="mb-2">
      <Field
        name="password"
        type="password"
        placeholder="Password"
        class="field"
        :class="{ 'is-invalid': errors.password }"
      />
      <div class="text-red-700">
        {{ errors.password }}
      </div>
    </div>
    <div class="">
      <button
        class="btn btn-primary w-full group"
        :disabled="isSubmitting"
      >
        <span class="absolute left-0 inset-y-0 flex items-center pl-3 opacity-25 group-hover:opacity-50">
           <font-awesome-icon 
            class="h-5 w-5" 
            aria-hidden="true"
            icon="fa-solid fa-user-plus" 
            v-if="!isSubmitting" 
          />
          <Spinner v-else />
        </span>
        Register
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
