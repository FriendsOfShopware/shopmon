<script setup>
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores';

import { Header, MainContainer } from '@/components';

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
