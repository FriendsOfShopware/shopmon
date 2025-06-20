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

        <footer class="app-footer">
            <div class="container">
                <div class="footer-links">
                    <router-link :to="{ name: 'privacy.authenticated' }" class="footer-link">
                        Privacy
                    </router-link>
                    <span class="footer-separator">|</span>
                    <router-link :to="{ name: 'imprint.authenticated' }" class="footer-link">
                        Legal Notice
                    </router-link>
                </div>
            </div>
        </footer>
    </div>
</template>

<script setup lang="ts">
import ImpersonationBanner from '@/components/ImpersonationBanner.vue';
import Notification from '@/components/Notification.vue';
</script>

<style scoped>
.app-layout {
    min-height: 100vh;
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

.app-footer {
    margin-top: auto;
    padding: 2rem 0;
    border-top: 1px solid var(--panel-border-color);
    background-color: var(--panel-background);
}

.footer-links {
    text-align: center;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
}

.footer-link {
    color: var(--text-color-muted);
    text-decoration: none;
    font-size: 0.875rem;
    transition: color 0.2s ease;
}

.footer-link:hover {
    color: var(--primary-color);
}

.footer-separator {
    color: var(--text-color-muted);
    font-size: 0.875rem;
}
</style>
