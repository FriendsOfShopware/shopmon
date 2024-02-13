<script setup lang="ts">
import Logo from '@/components/layout/Logo.vue';
import { useDarkModeStore } from '@/stores/darkMode.store';

const darkModeStore = useDarkModeStore();
</script>

<template>
  <button
    class="absolute top-2 right-2 text-gray-800 h-8 w-8 opacity-80 hover:opacity-100 focus:outline-none dark:text-white"
    @click="darkModeStore.toggleDarkMode">
    <icon-fa6-regular:moon class="w-5 h-5" v-if="darkModeStore.darkMode" />
    <icon-fa6-regular:sun class="w-5 h-5" v-else />
  </button>

  <div class="min-h-full flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <Logo class="mx-auto h-36 w-auto" />
      <slot />
    </div>
  </div>
</template>
