<template>
    <div class="app-layout">
        <div class="background-mask"></div>
        <top-bar />
        <ImpersonationBanner />
        <Notification />

        <div class="container main-container">
            <sidebar />
            <div class="main-content">
                <router-view />
            </div>
        </div>

        <layout-footer />
    </div>
</template>

<script setup lang="ts">
import ImpersonationBanner from '@/components/ImpersonationBanner.vue';
import Notification from '@/components/Notification.vue';
</script>

<style scoped>
.app-layout {
    min-height: calc(100vh + 1px);
    display: flex;
    flex-direction: column;
}

.main-container {
    padding: 1.5rem 0;

    @media all and (min-width: 1024px) {
        display: grid;
        column-gap: 1rem;
        grid-template-areas:
        "sidebar main-content";
        grid-template-columns: auto 1fr;
    }
}

.main-content {
    grid-area: main-content;
}
</style>
