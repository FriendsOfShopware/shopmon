<script setup lang="ts">
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';
import { ref } from 'vue';

import { useAuthStore } from '@/stores/auth.store';
import { useRouter } from 'vue-router';
import { useAlertStore } from '@/stores/alert.store';

const router = useRouter();
const authStore = useAuthStore();

const passwordType = ref('password')

const schema = Yup.object().shape({
  email: Yup.string().required('Email is required').email(),
  password: Yup.string().required('Password is required'),
});

async function onSubmit(values: any) {
  const { email, password } = values;
  try {
    await authStore.login(email, password);

    // redirect to previous url or default to home page
    router.push(authStore.returnUrl || '/');
  } catch (e) {
    const alertStore = useAlertStore();
    alertStore.error(e);
  }
}
</script>

<template>
  <div v-if="!authStore.isAuthenticated">
    <div>
      <h2 class="text-center text-3xl tracking-tight font-bold">
        Sign in to your account
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600 dark:text-neutral-500">
        Or
        {{ ' ' }}
        <router-link to="register" class="font-medium">
          register
        </router-link>
      </p>
    </div>
    <Form v-slot="{ errors, isSubmitting }" class="mt-8 space-y-6" :validation-schema="schema" @submit="onSubmit">
      
      <div class="-space-y-px">
        <div v-if="errors" class="pb-2 text-red-500">
          <div v-for="error in errors" key="error">
            {{ error }}
          </div>
        </div>

        <div>
          <label for="email-address" class="sr-only">Email address</label>
          <Field id="email-address" name="email" type="email" autocomplete="email" required=""
            class="field-no-rounded rounded-t-md" placeholder="Email address" :class="{ 'has-error': errors.email }" />
        </div>
        
        <div class="relative">
          <label for="password" class="sr-only">Password</label>
          <Field id="password" name="password" :type="passwordType" autocomplete="current-password" required=""
          class="field-no-rounded field-password rounded-b-md" placeholder="Password" :class="{ 'has-error': errors.password }" />
          <div class="absolute right-0 inset-y-0 flex items-center pr-3 cursor-pointer z-10">
            <icon-fa6-solid:eye class="w-[18px]" v-if="passwordType == 'password'" @click="passwordType = 'text'" />
            <icon-fa6-solid:eye-slash class="w-[18px]" v-else @click="passwordType = 'password'" />
          </div>      
        </div>
      </div>      

      <div>
        <router-link to="/account/forgot-password" class="font-medium">
        Forgot your password?
      </router-link>
      </div>      

      <div>
        <button type="submit" class="group w-full btn btn-primary" :disabled="isSubmitting">
          <span class="absolute left-0 inset-y-0 flex items-center pl-3 opacity-25 group-hover:opacity-50">
            <icon-fa6-solid:key
              class="h-5 w-5" 
              aria-hidden="true"
              v-if="!isSubmitting" 
            />
            <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
          </span>
          Sign in
        </button>
      </div>
    </Form>
  </div>
</template>
