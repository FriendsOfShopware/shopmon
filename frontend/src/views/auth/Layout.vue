<script setup lang="ts">

import LoginContainer from '@/components/layout/LoginContainer.vue';
import { useAuthStore } from '@/stores/auth.store';

const authStore = useAuthStore();

</script>

<template>
  <LoginContainer v-if="!authStore.user">
    <router-view />
  </LoginContainer>
  <template v-else>
    <router-view />
  </template>
</template>
