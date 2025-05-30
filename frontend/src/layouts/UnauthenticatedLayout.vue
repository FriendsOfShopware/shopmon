<template>
    <login-container>
        <Notification />
        <router-view />
    </login-container>
</template>

<script setup lang="ts">
import Notification from '@/components/Notification.vue';
</script>
