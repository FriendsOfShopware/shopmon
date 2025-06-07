<template>
    <login-container>
        <Notification />
        <router-view />
        <footer class="login-footer">
            <div class="footer-links">
                <router-link :to="{ name: 'privacy.unauthenticated' }" class="footer-link">
                    Privacy
                </router-link>
                <span class="footer-separator">|</span>
                <router-link :to="{ name: 'imprint.unauthenticated' }" class="footer-link">
                    Legal Notice
                </router-link>
            </div>
        </footer>
    </login-container>
</template>

<script setup lang="ts">
import Notification from '@/components/Notification.vue';
</script>

<style scoped>
.login-footer {
    position: fixed;
    bottom: 1rem;
    left: 50%;
    transform: translateX(-50%);
    z-index: 10;
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
