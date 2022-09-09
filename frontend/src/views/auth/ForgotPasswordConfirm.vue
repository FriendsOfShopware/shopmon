<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import Spinner from '@/components/icon/Spinner.vue';
import { useAuthStore } from '@/stores/auth.store';
import { useAlertStore } from '@/stores/alert.store';

const schema = Yup.object().shape({
    password: Yup.string()
        .required('Password is required')
        .min(8, 'Password must be at least 8 characters')
})

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter();
const alertStore = useAlertStore();

async function onSubmit(values: any): Promise<void> {  
    try {
        await authStore.confirmResetPassword(route.params.token as string, values.password);
        alertStore.success('Password has been resetted. You will be redirected to login page in 2 seconds.');

        setTimeout(() => {
            router.push('/account/login');
        }, 2000);
    } catch (error: Error) {
        alertStore.error(error.message);
    }
}

let isLoading = ref(true);
let tokenFound = ref(false);

onMounted(async () => {
    try {
        await authStore.resetAvailable(route.params.token as string);
        tokenFound.value = true;
    } catch (e) {
        tokenFound.value = false
    } finally {
        isLoading.value = false;
    }
})

function goToResend() {
    router.push('/account/forgot-password');
}
</script>

<template>
  <div>
    <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
      Change Password
    </h4>
  </div>
  <div
    v-if="isLoading"
    class="rounded-md bg-blue-50 p-4 border border-sky-200"
  >
    <div class="flex">
      envelope
        icon="fa-solid fa-circle-info"
        class="h-5 w-5 text-sky-400"
        aria-hidden="true"
      />
      <div class="ml-3 flex-1 md:flex md:justify-between">
        <p class="text-sm text-sky-700">
          Loading...
        </p>
      </div>
    </div>
  </div>

  <div v-else>
    <Form
      v-if="tokenFound"
      v-slot="{ errors, isSubmitting }"
      :validation-schema="schema"
      @submit="onSubmit"
    >
      <div class="mb-2">
        <Field
          name="password"
          placeholder="Password"
          type="password"
          class="field"
          :class="{ 'is-invalid': errors.password }"
        />
        <div class="text-red-700">
          {{ errors.password }}
        </div>
      </div>
      <div class="">
        <button
          class="w-full group btn btn-primary"
          :disabled="isSubmitting"
        >
          <span class="absolute left-0 inset-y-0 flex items-center pl-3 opacity-25 group-hover:opacity-50">
          <icon-fa6-solid:key
            class="h-5 w-5" 
            aria-hidden="true"
            v-if="!isSubmitting" 
          />
          <Spinner v-else />
        </span>
          Change Password
        </button>
        <router-link
          to="login"
          class="text-gray-900 inline-block mt-2 center text-center w-full"
        >
          Cancel
        </router-link>
      </div>
    </Form>
    <div v-else>
      <div class="rounded-md bg-red-50 p-4 border border-red-200">
        <div class="flex">
          <div class="flex-shrink-0">
            <icon-fa6-solid:circle-xmark
              class="h-5 w-5 text-red-600"
              aria-hidden="true"
            />
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-900">
              Invalid Token
            </h3>
            <div class="mt-2 text-sm text-red-800">
              <p>It looks like your Token is expired.</p>
            </div>
          </div>
        </div>
      </div>
      <button
        type="button"
        class="mt-4 btn btn-primary w-full"
        @click="goToResend"
      >
        <span class="absolute left-0 inset-y-0 flex items-center pl-3 opacity-25 group-hover:opacity-50">
           <icon-fa6-solid:envelope
            class="h-5 w-5" 
            aria-hidden="true"
          />
        </span>
        Resend mail
      </button>
    </div>
  </div>
</template>
