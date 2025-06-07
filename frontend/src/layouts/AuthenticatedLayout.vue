<template>
    <div class="app-layout">
        <nav-bar />
        <ImpersonationBanner />
        <Notification />
        <main class="app-main">
            <router-view />
        </main>
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

.app-main {
    flex: 1;
}

.app-footer {
    margin-top: auto;
    padding: 2rem 0;
    border-top: 1px solid var(--border-color, #e5e7eb);
    background-color: var(--footer-background, #f9fafb);
}

.footer-links {
    text-align: center;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
}

.footer-link {
    color: var(--text-color-muted, #6b7280);
    text-decoration: none;
    font-size: 0.875rem;
    transition: color 0.2s ease;
}

.footer-link:hover {
    color: var(--primary-color, #3b82f6);
}

.footer-separator {
    color: var(--text-color-muted, #6b7280);
    font-size: 0.875rem;
}
</style>
