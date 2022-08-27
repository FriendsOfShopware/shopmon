<script setup>
import { RouterView } from 'vue-router';

import { Nav, Alert } from '@/components';
import { useAuthStore } from '@/stores';

const authStore = useAuthStore();
</script>

<template>
  <Nav />
  <Alert />
  <RouterView />
</template>

<style></style>
