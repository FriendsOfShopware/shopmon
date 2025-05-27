<template>
    <login-container v-if="!authStore.user">
        <router-view />
    </login-container>
    <template v-else>
        <router-view />
    </template>
</template>

<script setup lang="ts">
import LoginContainer from '@/components/layout/LoginContainer.vue';
import { useAuthStore } from '@/stores/auth.store';

const authStore = useAuthStore();
</script>
