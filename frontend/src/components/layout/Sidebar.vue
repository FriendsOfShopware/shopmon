<template>
  <aside
    class="app-sidebar"
    :class="[currentViewClass, { 'sidebar-collapsed': isCollapsed }]"
  >
    <div class="app-sidebar-header">
      <button type="button" class="app-sidebar-search" @click="$router.push({ name: 'account.dashboard' })">
        <icon-fa6-solid:magnifying-glass class="app-sidebar-search-icon" />
        <span class="app-sidebar-search-label">Quick search...</span>
      </button>
    </div>

    <div class="app-sidebar-content">
      <div class="app-sidebar-group">
        <div class="app-sidebar-group-label">Overview</div>
        <nav class="app-sidebar-menu">
          <router-link
            v-for="item in navigation"
            :key="item.route"
            :to="{ name: item.route }"
            :class="[
              'app-sidebar-menu-button',
              { 'app-sidebar-menu-button-active': isActive(item, $route) },
            ]"
          >
            <component
              :is="$router.resolve({ name: item.route }).meta.icon"
              v-if="$router.resolve({ name: item.route }).meta.icon"
              class="app-sidebar-menu-icon"
            />
            <span class="app-sidebar-menu-label">{{
              $t($router.resolve({ name: item.route }).meta.titleKey ?? "")
            }}</span>
          </router-link>
        </nav>
      </div>

      <hr class="app-sidebar-separator" />

      <div v-if="environments.length > 0" class="app-sidebar-group">
        <div class="app-sidebar-group-label">Environments</div>
        <nav class="app-sidebar-menu">
          <router-link
            v-for="env in environments"
            :key="env.id"
            :to="{
              name: 'account.environments.detail',
              params: {
                organizationId: env.organizationId,
                environmentId: env.id,
              },
            }"
            active-class=""
            class="app-sidebar-menu-button app-sidebar-menu-button-env"
          >
            <div class="app-sidebar-menu-icon app-sidebar-menu-icon-env">
              <img v-if="env.favicon" :src="env.favicon" alt="" class="app-sidebar-favicon" />
              <icon-fa6-solid:earth-americas v-else class="app-sidebar-placeholder-icon" />
            </div>
            <span class="app-sidebar-menu-label">{{ env.name }}</span>
            <span class="app-sidebar-menu-trailing">
              <status-icon :status="env.status" />
            </span>
          </router-link>
        </nav>
      </div>
    </div>

    <div class="app-sidebar-footer">
      <router-link :to="{ name: 'account.settings' }" class="app-sidebar-menu-button">
        <icon-fa6-solid:gear class="app-sidebar-menu-icon" />
        <span class="app-sidebar-menu-label">Manage account</span>
      </router-link>
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { RouteLocationNormalizedLoaded } from "vue-router";

import { computed, ref, watchEffect } from "vue";
import { useRoute } from "vue-router";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";

const { environments } = useAccountEnvironments();

const route = useRoute();
const isCollapsed = ref(false);

const currentViewClass = computed(() => {
  const routeName = route.name as string;
  if (!routeName) return "";
  return "is-" + routeName.replace(/\./g, "-");
});

watchEffect(() => {
  const name = route.name as string;
  isCollapsed.value = !!name && name.includes("environments.detail");
});

const navigation = [
  { route: "account.dashboard" },
  { route: "account.shop.list", active: "shop" },
  { route: "account.extension.list" },
  { route: "account.organizations.list", active: "organizations" },
];

function isActive(item: { route: string; active?: string }, $route: RouteLocationNormalizedLoaded) {
  if (item.route === $route.name) return true;
  return !!(
    $route.name &&
    typeof $route.name === "string" &&
    item.active &&
    $route.name.match(item.active)
  );
}
</script>

<style>
.app-sidebar {
  position: relative;
  flex-direction: column;
  height: 100%;
  flex-shrink: 0;
  flex-grow: 0;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  background-color: var(--panel-background);
  color: var(--sidebar-nav-link-color);
  border-right: 1px solid var(--panel-border-color);
  transition:
    width 250ms cubic-bezier(0.77, 0, 0.175, 1),
    transform 200ms ease,
    opacity 200ms ease;
  will-change: width, transform;
  z-index: 50;

  @media (min-width: 768px) {
    display: flex;
    width: 16rem;
  }

  @media (max-width: 767px) {
    display: flex;
    position: fixed;
    inset-y: 0;
    left: 0;
    width: 16rem;
    transform: translateX(-100%);
    opacity: 0;
    pointer-events: none;
    box-shadow: 0 10px 30px rgba(15, 23, 42, 0.18);
  }

  &.sidebar-mobile-open {
    transform: translateX(0);
    opacity: 1;
    pointer-events: auto;
  }

  &.sidebar-collapsed {
    @media (min-width: 768px) {
      width: 3rem;
    }
  }
}

.app-sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid var(--panel-border-color);
  overflow: hidden;

  .sidebar-collapsed & {
    border-bottom: 0;
  }
}

.app-sidebar-search {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  background-color: var(--control-background);
  box-shadow: inset 0 0 0 1px var(--panel-border-color);
  color: var(--text-color-muted);
  cursor: pointer;
  transition:
    background-color 150ms ease,
    box-shadow 250ms cubic-bezier(0.77, 0, 0.175, 1),
    padding 250ms cubic-bezier(0.77, 0, 0.175, 1);
  flex-shrink: 0;
  text-align: left;

  &:hover {
    background-color: var(--control-hover-background);
  }

  &:focus-visible {
    box-shadow: inset 0 0 0 1px var(--field-focus-ring-color);
  }

  .sidebar-collapsed & {
    padding-left: 0.5rem;
    padding-right: 0.5rem;
    justify-content: center;
    box-shadow: none;
  }
}

.app-sidebar-search-icon {
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
  color: var(--text-color-muted);
}

.app-sidebar-search-label {
  flex: 1;
  min-width: 0;
  font-size: 0.875rem;
  line-height: 1.25rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  .sidebar-collapsed & {
    display: none;
  }
}

.app-sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.5rem;

  .sidebar-collapsed & {
    gap: 0;
    padding-top: 0;
    padding-bottom: 0;
    overflow-x: hidden;
  }
}

.app-sidebar-group {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;

  .sidebar-collapsed & {
    gap: 0;
  }
}

.app-sidebar-group-label {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1;
  color: var(--text-color-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  user-select: none;

  .sidebar-collapsed & {
    display: none;
  }
}

.app-sidebar-menu {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;

  .sidebar-collapsed & {
    gap: 0;
  }
}

.app-sidebar-menu-button {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  gap: 0.5rem;
  border-radius: 0.75rem;
  min-height: 34px;
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--sidebar-nav-link-color);
  text-decoration: none;
  cursor: pointer;
  outline: none;
  transition:
    background-color 150ms ease,
    color 150ms ease,
    padding 250ms cubic-bezier(0.77, 0, 0.175, 1);

  &:hover {
    background-color: var(--button-ghost-hover-background);
    color: var(--text-color);
  }

  &:focus-visible {
    box-shadow: inset 0 0 0 1px var(--primary-color);
  }

  &.app-sidebar-menu-button-active,
  &.router-link-active {
    background-color: var(--button-ghost-hover-background);
    color: var(--text-color);
    font-weight: 500;
  }

  .sidebar-collapsed & {
    padding-left: 0.5rem;
    padding-right: 0.5rem;
    justify-content: center;
  }
}

.app-sidebar-menu-icon {
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--menu-icon-color);

  .app-sidebar-menu-button-active &,
  .router-link-active & {
    color: var(--primary-color);
  }

  .app-sidebar-menu-button:hover & {
    color: var(--text-color);
  }
}

.app-sidebar-menu-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left;

  .sidebar-collapsed & {
    display: none;
  }
}

.app-sidebar-menu-trailing {
  margin-left: auto;
  flex-shrink: 0;
  display: flex;
  align-items: center;

  .sidebar-collapsed & {
    display: none;
  }
}

.app-sidebar-menu-button-env {
  min-height: 36px;
}

.app-sidebar-menu-icon-env {
  width: 1rem;
  height: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-sidebar-favicon {
  width: 1rem;
  height: 1rem;
  border-radius: 0.25rem;
  object-fit: contain;
}

.app-sidebar-placeholder-icon {
  width: 0.875rem;
  height: 0.875rem;
  opacity: 0.4;
}

.app-sidebar-separator {
  margin: 0.5rem 0.5rem;
  height: 1px;
  min-height: 1px;
  border: none;
  background-color: var(--panel-border-color);
  flex-shrink: 0;

  .sidebar-collapsed & {
    display: none;
  }
}

.app-sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.5rem;
  border-top: 1px solid var(--panel-border-color);
  flex-shrink: 0;

  .sidebar-collapsed & {
    border-top: none;
    padding-top: 0.25rem;
  }

  .app-sidebar-menu-button {
    color: var(--text-color-muted);

    &:hover {
      color: var(--text-color);
      background-color: var(--button-ghost-hover-background);
    }
  }
}
</style>
