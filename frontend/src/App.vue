<script setup>
import { RouterView } from 'vue-router';

import { Nav, Alert } from '@/components';
import { useAuthStore } from '@/stores';

const authStore = useAuthStore();
</script>

<template>
  <Nav v-if="authStore.user" />
  <Alert />
  <RouterView />
</template>

<style></style>
