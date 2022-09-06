<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import Spinner from '@/components/icon/Spinner.vue';

import { useAuthStore } from '@/stores/auth.store';
import { useRouter } from 'vue-router';
import { useAlertStore } from '@/stores/alert.store';

const router = useRouter();

const passwordType = ref('password')

const schema = Yup.object().shape({
  email: Yup.string().required('Email is required'),
  password: Yup.string().required('Password is required'),
});

async function onSubmit(values: any) {
  const authStore = useAuthStore();
  const { email, password } = values;
  try {
    await authStore.login(email, password);

    // redirect to previous url or default to home page
    router.push(authStore.returnUrl || '/');
  } catch (e) {
    const alertStore = useAlertStore();
    alertStore.error(error);
  }
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
      <router-link to="register" class="font-medium">
        register
      </router-link>
    </p>
  </div>
  <Form v-slot="{ errors, isSubmitting }" class="mt-8 space-y-6" :validation-schema="schema" @submit="onSubmit">
    <div class="rounded-md shadow-sm -space-y-px">
      <div>
        <label for="email-address" class="sr-only">Email address</label>
        <Field id="email-address" name="email" type="email" autocomplete="email" required=""
          class="field-no-rounded rounded-t-md" placeholder="Email address" :class="{ 'is-invalid': errors.email }" />
      </div>
      <div>
        <label for="password" class="sr-only">Password</label>
        <div class="relative">
          <Field id="password" name="password" :type="passwordType" autocomplete="current-password" required=""
          class="field-no-rounded field-password rounded-b-md" placeholder="Password" :class="{ 'is-invalid': errors.password }" />
          <div class="absolute right-0 inset-y-0 flex items-center pr-3 cursor-pointer z-10">
            <font-awesome-icon icon="fa-solid fa-eye" class="w-[18px]" v-if="passwordType == 'password'" @click="passwordType = 'text'" />
            <font-awesome-icon icon="fa-solid fa-eye-slash" class="w-[18px]" v-else @click="passwordType = 'password'" />
          </div>
        </div>        
      </div>
    </div>

    <div class="flex items-end">
      <div class="text-sm flex">
        <router-link to="/account/forgot-password" class="font-medium">
          Forgot your password?
        </router-link>
      </div>
    </div>

    <div>
      <button type="submit" class="group w-full btn btn-primary" :disabled="isSubmitting">
        <span class="absolute left-0 inset-y-0 flex items-center pl-3 opacity-25 group-hover:opacity-50">
          <font-awesome-icon 
            class="h-5 w-5" 
            aria-hidden="true"
            icon="fa-solid fa-key" 
            v-if="!isSubmitting" 
          />
          <Spinner v-else />
        </span>
        Sign in
      </button>
    </div>
  </Form>
</template>
