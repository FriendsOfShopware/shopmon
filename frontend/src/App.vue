<template>
    <router-view />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';
import { useDarkMode } from './composables/useDarkMode';

useDarkMode();
</script>
