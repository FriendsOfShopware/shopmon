<script setup lang="ts">

import { useAuthStore } from '@/stores/auth.store';
</script>

<script lang="ts">
export default {
    data() {
        return {
            isLoading: true,
            success: false,
        }
    },
    async created() {
        const authStore = useAuthStore();

        try {
            await authStore.confirmMail(this.$route.params.token as string);
            this.success = true;
        } catch (e) {
            this.success = false;
        } finally {
            this.isLoading = false;
        }
    },
    methods: {
        goToLogin() {
            this.$router.push('/account/login');
        },
    }
}
</script>

<template>
  <h4 class="mb-6 text-center text-3xl tracking-tight font-bold">
    Confirming your Account Registration
  </h4>
  <div
    v-if="isLoading"
    class="rounded-md bg-blue-50 p-4 border border-sky-200"
  >
    <div class="flex">
      <icon-fa6-solid:circle-info
        class="h-5 w-5 text-sky-500"
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
    <div
      v-if="success"
      class="rounded-md bg-green-50 p-4 border border-green-200"
    >
      <div class="flex">
        <icon-fa6-solid:circle-check
          class="h-5 w-5 text-green-500"
          aria-hidden="true"
        />
        <div class="ml-3 flex-1 md:flex md:justify-between">
          <p class="text-sm text-green-700">
            Your email address has been confirmed.
            <router-link
              to="/account/login"
            >
              Login
            </router-link>
          </p>
        </div>
      </div>
    </div>

    <div
      v-else
      class="rounded-md bg-red-50 p-4 border border-red-200"
    >
      <div class="flex">
        <icon-fa6-solid:circle-xmark
          class="h-5 w-5 text-red-600"
          aria-hidden="true"
        />
        <div class="ml-3 flex-1 md:flex md:justify-between">
          <p class="text-sm text-red-900">
            The given token has been expired
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
