<template>
    <nav-bar v-if="authStore.user" />
    <Notification />
    <router-view />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';

import { useAuthStore } from '@/stores/auth.store';
import { useNotificationStore } from './stores/notification.store';
import { useDarkModeStore } from './stores/darkMode.store';


const authStore = useAuthStore();
const notificationStore = useNotificationStore();
const darkModeStore = useDarkModeStore();

if (authStore.isAuthenticated) {
    authStore.refreshUser();
    notificationStore.loadNotifications();
}

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    darkModeStore.setDarkMode(e.matches);
});

darkModeStore.updateDarkModeClass();
</script>
