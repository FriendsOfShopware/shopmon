<template>
    <div class="admin-layout">
        <div class="background-mask"></div>

        <nav class="top-bar">
            <div class="top-bar-container container">
                <div class="top-bar-nav">
                    <RouterLink to="/" class="top-bar-logo">
                        <Logo/>
                    </RouterLink>

                    <div class="top-bar-nav-links">
                        <RouterLink
                            to="/admin/users"
                            class="nav-link"
                            :class="{
                                'active': $route.name?.toString().startsWith('admin.users')
                            }"
                        >
                            Users
                        </RouterLink>
                    </div>
                </div>

                <div class="top-bar-actions">
                    <RouterLink
                        to="/"
                        class="btn btn-primary"
                    >
                        <icon-fa6-solid:house class="icon" />
                        Back to Dashboard
                    </RouterLink>
                </div>
            </div>
        </nav>

        <ImpersonationBanner />

        <div class="main-container container">
            <RouterView/>
        </div>
    </div>
</template>

<script setup lang="ts">
import ImpersonationBanner from '@/components/ImpersonationBanner.vue';
import Logo from '@/components/Logo.vue';
import {RouterLink, RouterView, useRoute} from 'vue-router';

const $route = useRoute();
</script>

<style scoped>
.top-bar-nav {
    display: flex;
    align-items: center;
}

.top-bar-nav-links {
    display: none;
    margin-left: 1.5rem;
    gap: 1rem;

    @media (min-width: 640px) {
        display: flex;
    }
}

.nav-link {
    display: inline-flex;
    align-items: center;
    padding: 0.5rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #e2e8f0;
    transition: all 0.2s;

    &.active {
        background-color: #0284c780;
        color: #ffffff;
    }
    &:hover {
        background-color: #0284c7;
        color: #ffffff;
    }
}

.top-bar-actions {
    display: flex;
    align-items: center;

    .btn {
        background-color: #0284c780;

        &:hover {
            background-color: #0284c7;
        }
    }
}

.main-container {
    margin: 1.5rem auto;
}
</style>
