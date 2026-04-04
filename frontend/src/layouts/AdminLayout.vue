<template>
  <div class="admin-layout">
    <div class="background-mask" />

    <nav class="top-bar">
      <div class="top-bar-container container">
        <div class="top-bar-nav">
          <UiButton to="/" class="top-bar-logo">
            <Logo />
          </UiButton>

          <div class="top-bar-nav-links">
            <RouterLink
              to="/admin/dashboard"
              class="nav-link"
              :class="{
                active: $route.name?.toString().startsWith('admin.dashboard'),
              }"
            >
              Dashboard
            </RouterLink>
            <RouterLink
              to="/admin/users"
              class="nav-link"
              :class="{
                active: $route.name?.toString().startsWith('admin.users'),
              }"
            >
              Users
            </RouterLink>
            <RouterLink
              to="/admin/organizations"
              class="nav-link"
              :class="{
                active: $route.name?.toString().startsWith('admin.organizations'),
              }"
            >
              Organizations
            </RouterLink>
            <RouterLink
              to="/admin/environments"
              class="nav-link"
              :class="{
                active: $route.name?.toString().startsWith('admin.environments'),
              }"
            >
              Environments
            </RouterLink>
            <a href="/admin/queues" class="nav-link" target="_blank"> Queues </a>
          </div>
        </div>

        <div class="top-bar-actions">
          <UiButton to="/" variant="primary">
            <icon-fa6-solid:house class="icon" />
            Back to Dashboard
          </UiButton>
        </div>
      </div>
    </nav>

    <ImpersonationBanner />

    <div class="main-container container">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Logo from "@/components/Logo.vue";
import { RouterLink, RouterView } from "vue-router";
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

  .ui-button {
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
