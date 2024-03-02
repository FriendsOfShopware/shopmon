<template>
    <button
        class="absolute top-2 right-2 text-gray-800 h-8 w-8 opacity-80 hover:opacity-100 focus:outline-none dark:text-white"
        type="button"
        @click="darkModeStore.toggleDarkMode"
    >
        <icon-fa6-regular:moon
            v-if="darkModeStore.darkMode"
            class="w-5 h-5"
        />
        <icon-fa6-regular:sun
            v-else
            class="w-5 h-5"
        />
    </button>

    <div class="min-h-full flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div class="max-w-md w-full space-y-8">
            <logo class="mx-auto h-36 w-auto" />
            <slot />
        </div>
    </div>
</template>

<script setup lang="ts">
import Logo from '@/components/Logo.vue';
import { useDarkModeStore } from '@/stores/darkMode.store';

const darkModeStore = useDarkModeStore();
</script>
