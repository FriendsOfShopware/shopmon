<template>
  <div class="admin-layout">
    <nav class="admin-nav">
      <div class="admin-nav-container">
        <div class="admin-nav-main">
          <RouterLink to="/" class="admin-nav-logo">
            <Logo />
          </RouterLink>
          <div class="admin-nav-links">
            <RouterLink
              to="/admin/users"
              class="admin-nav-link"
              :class="{
                'active': $route.name?.toString().startsWith('admin.users')
              }"
            >
              Users
            </RouterLink>
          </div>
        </div>
        <div class="admin-nav-actions">
          <RouterLink
            to="/"
            class="admin-nav-back"
          >
            Back to Dashboard
          </RouterLink>
        </div>
      </div>
    </nav>
    
    <ImpersonationBanner />

    <main class="admin-main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import ImpersonationBanner from '@/components/ImpersonationBanner.vue';
import Logo from '@/components/Logo.vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';

const $route = useRoute();
</script>

<style>
.admin-layout {
  min-height: 100vh;
  background-color: var(--background-color);
}

.admin-nav {
  background-color: var(--admin-nav-background, #1e293b);
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
}

.admin-nav-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 4rem;
}

.admin-nav-main {
  display: flex;
  align-items: center;
}

.admin-nav-logo {
  display: flex;
  align-items: center;
  padding: 0 0.5rem;
  color: #ffffff;
  
  svg {
    height: 2.5rem;
    width: auto;
  }
}

.admin-nav-links {
  display: none;
  margin-left: 1.5rem;
  gap: 1rem;
  
  @media (min-width: 640px) {
    display: flex;
  }
}

.admin-nav-link {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
  transition: all 0.2s;
  
  &:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: #ffffff;
  }
  
  &.active {
    background-color: rgba(255, 255, 255, 0.2);
    color: #ffffff;
  }
}

.admin-nav-actions {
  display: flex;
  align-items: center;
}

.admin-nav-back {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #e2e8f0;
  transition: all 0.2s;
  
  &:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: #ffffff;
  }
}

.admin-main {
  padding: 2rem 1rem;
  max-width: 1280px;
  margin: 0 auto;
}

.admin-nav-main img {
  width: 32px;
}

/* Dark mode adjustments */
:root[data-theme="dark"] {
  .admin-nav {
    background-color: #0f172a;
  }
}
</style>
