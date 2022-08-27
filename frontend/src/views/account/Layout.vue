<script setup>

import { LoginContainer, MainContainer } from '@/components'
import { useAuthStore } from '@/stores';

const authStore = useAuthStore();

</script>

<template>
  <LoginContainer v-if="!authStore.user">
    <router-view></router-view>
  </LoginContainer>
  <MainContainer v-else>
    <router-view></router-view>
  </MainContainer>
</template>
