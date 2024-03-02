<template>
    <nav-bar v-if="authStore.user" />
    <alert />
    <router-view />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';

import NavBar from '@/components/layout/NavBar.vue';
import Alert from '@/components/Alert.vue';

import { useAuthStore } from '@/stores/auth.store';
import { useNotificationStore } from './stores/notification.store';
import { useDarkModeStore } from './stores/darkMode.store';


const authStore = useAuthStore();
const notificationStore = useNotificationStore();
const darkModeStore = useDarkModeStore();

if (authStore.isAuthenticated) {
    authStore.refreshUser();

    const authToken = authStore.access_token;

    if (authToken) {
        notificationStore.connect(authToken);
        notificationStore.loadNotifications();
    }
}

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    darkModeStore.setDarkMode(e.matches);
});

darkModeStore.updateDarkModeClass();
</script>
