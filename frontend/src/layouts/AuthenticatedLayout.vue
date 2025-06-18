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
    display: flex;
    gap: 20px;
    align-items: flex-start;
}

.main-content {
    flex: 1 1 0%;
    min-width: 0;
}

.background-mask {
    mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1440 320' preserveAspectRatio='none'%3E%3Cpath fill='%230099ff' fill-opacity='1' d='M0,192L80,197.3C160,203,320,213,480,202.7C640,192,800,160,960,170.7C1120,181,1280,235,1360,261.3L1440,288L1440,0L1360,0C1280,0,1120,0,960,0C800,0,640,0,480,0C320,0,160,0,80,0L0,0Z'%3E%3C/path%3E%3C/svg%3E");
    mask-size: 100% 111%;
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 290px;
    z-index: -1;
    background-color: var(--primary-color);

    &:before {
        content: '';
        display: block;
        background: linear-gradient(90deg, var(--primary-color) -0.48%, #0284c7 57.18%, #0284c7 60.83%, var(--primary-color) 97.96%);
        width: 100%;
        height: 100%;
        max-width: var(--container-width);
        margin: 0 auto;
    }
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
