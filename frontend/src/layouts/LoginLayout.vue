<template>
    <login-container>
        <Notification />
        <router-view />
    </login-container>

    <layout-footer />
</template>

<script setup lang="ts">
import Notification from '@/components/Notification.vue';
</script>

<style scoped>

</style>
