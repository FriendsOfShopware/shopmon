<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);
</script>

<template>
    <div v-if="user">
        <Header title="Dashboard"></Header>
        <MainContainer>
            Das ist mein Inhalt
        </MainContainer>
    </div>
</template>
