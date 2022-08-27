<script setup>
import { RouterView } from 'vue-router';

import { Nav, Alert } from '@/components';
import { useAuthStore } from '@/stores';
import { Spinner } from '@/components/icon';

const authStore = useAuthStore();
</script>

<template>
  <Nav />
  <Alert />
  <Spinner />
  <RouterView />
</template>

<style></style>
