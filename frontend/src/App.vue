<template>
    <router-view />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';

import { useDarkModeStore } from './stores/darkMode.store';

const darkModeStore = useDarkModeStore();

window
    .matchMedia('(prefers-color-scheme: dark)')
    .addEventListener('change', (e) => {
        darkModeStore.setDarkMode(e.matches);
    });

darkModeStore.updateDarkModeClass();
</script>
