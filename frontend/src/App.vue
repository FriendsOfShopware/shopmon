<script setup lang="ts">
import { RouterView } from 'vue-router';

import Nav from '@/components/layout/Nav.vue';
import Alert from '@/components/Alert.vue';

import { useAuthStore } from '@/stores/auth.store';
import { useNotificationStore } from './stores/notification.store';

const authStore = useAuthStore();
const notificationStore = useNotificationStore()

if (authStore.isAuthenticated) {
    authStore.refreshUser();
    
    const authToken = authStore.access_token;
    
    if(authToken) {
      notificationStore.connect(authToken);
      notificationStore.loadNotifications();
    }
}

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
  e.matches ? document.documentElement.classList.add('dark') : document.documentElement.classList.remove('dark');
});


if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
  document.documentElement.classList.add('dark');
}
</script>

<template>
  <Nav v-if="authStore.user" />
  <Alert />
  <RouterView />
</template>
