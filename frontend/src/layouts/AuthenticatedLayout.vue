<template>
    <nav-bar />
    <ImpersonationBanner />
    <Notification />
    <router-view />
</template>

<script setup lang="ts">
import ImpersonationBanner from '@/components/ImpersonationBanner.vue';
import Notification from '@/components/Notification.vue';
</script>

