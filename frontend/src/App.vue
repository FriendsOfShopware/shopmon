<script setup>
import { RouterLink, RouterView } from 'vue-router';

import { Nav, Alert } from '@/components';
import { useAuthStore } from '@/stores';

const authStore = useAuthStore();
</script>

<template>
  <Nav />
  <Alert />
  <Spinner />
  <RouterView />
</template>

<style></style>
