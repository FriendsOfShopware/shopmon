<template>
    <router-view />
</template>

<script setup lang="ts">
import { useDarkMode } from './composables/useDarkMode';

useDarkMode();
</script>
