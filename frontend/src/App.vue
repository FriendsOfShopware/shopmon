<script setup lang="ts">
import { RouterView } from 'vue-router';

import Nav from '@/components/layout/Nav.vue';
import Alert from '@/components/Alert.vue';

import { useAuthStore } from '@/stores/auth.store';

const authStore = useAuthStore();

if (authStore.isAuthenticated) {
    authStore.refreshUser();
}
</script>

<template>
  <Nav v-if="authStore.user" />
  <Alert />
  <RouterView />
</template>
