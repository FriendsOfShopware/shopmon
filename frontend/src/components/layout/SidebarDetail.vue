<template>
  <div v-if="props.environment" class="detail-sidebar-shell">
    <UiButton class="detail-sidebar-mobile-trigger" block @click="toggleMobileOpen">
      <icon-fa6-solid:bars-staggered class="icon" />
      {{ environment?.organizationName + " / " + environment?.name }}
      {{ currentRouteTitle && route.name !== "account.environments.detail" ? " / " : "" }}
      {{ currentRouteTitle }}
    </UiButton>

    <div v-if="isMobileOpen" class="detail-sidebar-overlay" @click="isMobileOpen = false" />

    <aside :class="['detail-sidebar', { 'mobile-open': isMobileOpen }]">
      <div class="detail-sidebar-header">
        <div class="detail-sidebar-brand">
          <img
            v-if="environment?.environmentImage"
            :src="`${environment.environmentImage}`"
            class="detail-sidebar-brand-image"
          />
          <icon-fa6-solid:image v-else class="detail-sidebar-brand-fallback" />
        </div>

        <div class="detail-sidebar-meta">
          <div class="detail-sidebar-organization">{{ environment.organizationName }}</div>
          <div class="detail-sidebar-title-row">
            <h4>{{ environment?.name }}</h4>
            <status-icon :status="environment?.status ?? ''" />
          </div>
          <div class="detail-sidebar-subtitle">{{ environmentHost }}</div>
        </div>
      </div>

      <div class="detail-sidebar-content">
        <div class="detail-sidebar-group">
          <div class="detail-sidebar-group-label">Navigation</div>
          <nav class="detail-sidebar-menu">
            <router-link
              v-for="item in detailNavigation"
              :key="item.name"
              :to="{
                name: item.route,
                params: {
                  organizationId: route.params.organizationId,
                  environmentId: route.params.environmentId,
                },
              }"
              :class="{
                'detail-sidebar-menu-button': true,
                active: isRouteActive(item.route),
              }"
              active-class=""
              exact-active-class=""
            >
              <component
                :is="$router.resolve({ name: item.route }).meta.icon"
                v-if="$router.resolve({ name: item.route }).meta.icon"
                class="detail-sidebar-menu-icon"
              />
              <span class="detail-sidebar-menu-label">{{
                $t($router.resolve({ name: item.route }).meta.titleKey ?? "")
              }}</span>
              <span v-if="item.count !== undefined" class="detail-sidebar-menu-count">
                {{ item.count }}
              </span>
            </router-link>
          </nav>
        </div>
      </div>

      <div class="detail-sidebar-footer">
        <a
          :href="environment?.url"
          class="detail-sidebar-menu-button"
          target="_blank"
          rel="noopener noreferrer"
        >
          <icon-fa6-solid:store class="detail-sidebar-menu-icon" />
          <span class="detail-sidebar-menu-label">Storefront</span>
        </a>

        <a
          :href="(environment?.url ?? '') + '/admin'"
          class="detail-sidebar-menu-button"
          target="_blank"
          rel="noopener noreferrer"
        >
          <icon-fa6-solid:user-gear class="detail-sidebar-menu-icon" />
          <span class="detail-sidebar-menu-label">Admin</span>
        </a>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import type { components } from "@/types/api";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import StatusIcon from "@/components/StatusIcon.vue";

const route = useRoute();
const { t } = useI18n();

const isMobileOpen = ref(false);

const props = defineProps<{
  environment: components["schemas"]["EnvironmentDetail"] | null;
}>();

const currentRouteTitle = computed(() => {
  const titleKey = route.meta.titleKey;
  return typeof titleKey === "string" ? t(titleKey) : "";
});

const environmentHost = computed(() => {
  if (!props.environment?.url) {
    return "No URL configured";
  }

  try {
    return new URL(props.environment.url).host;
  } catch {
    return props.environment.url;
  }
});

const isRouteActive = (routeName: string) => {
  if (route.name === routeName) return true;

  if (
    routeName === "account.environments.detail.deployments" &&
    route.name === "account.environments.detail.deployment"
  ) {
    return true;
  }

  return false;
};

const detailNavigation = computed(() => [
  {
    name: "Environment Information",
    route: "account.environments.detail",
  },
  {
    name: "Checks",
    route: "account.environments.detail.checks",
    count: props.environment?.checks?.length ?? 0,
  },
  {
    name: "Extensions",
    route: "account.environments.detail.extensions",
    count: props.environment?.extensions?.length ?? 0,
  },
  {
    name: "Scheduled Tasks",
    route: "account.environments.detail.tasks",
    count: props.environment?.scheduledTasks?.length ?? 0,
  },
  {
    name: "Queue",
    route: "account.environments.detail.queue",
    count: props.environment?.queues?.length ?? 0,
  },
  {
    name: "Sitespeed",
    route: "account.environments.detail.sitespeed",
    count: props.environment?.sitespeeds?.length ?? 0,
  },
  {
    name: "Changelog",
    route: "account.environments.detail.changelog",
    count: props.environment?.changelogs?.length ?? 0,
  },
  {
    name: "Deployments",
    route: "account.environments.detail.deployments",
    count: props.environment?.deploymentsCount ?? 0,
  },
]);

function toggleMobileOpen() {
  isMobileOpen.value = !isMobileOpen.value;
}

watch(route, () => {
  if (isMobileOpen.value) {
    isMobileOpen.value = false;
  }
});
</script>

<style scoped>
.detail-sidebar-shell {
  display: block;
  position: relative;
}

.detail-sidebar-mobile-trigger {
  margin-bottom: 1rem;

  @media (min-width: 768px) {
    display: none;
  }
}

.detail-sidebar-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: rgba(15, 23, 42, 0.45);

  @media (min-width: 768px) {
    display: none;
  }
}

.detail-sidebar {
  position: fixed;
  inset-y: 0;
  left: 0;
  z-index: 50;
  display: flex;
  flex-direction: column;
  width: 16rem;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  background-color: var(--panel-background);
  color: var(--sidebar-nav-link-color);
  border-right: 1px solid var(--panel-border-color);
  transform: translateX(-100%);
  opacity: 0;
  pointer-events: none;
  transition:
    transform 200ms ease,
    opacity 200ms ease;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.18);

  &.mobile-open {
    transform: translateX(0);
    opacity: 1;
    pointer-events: auto;
  }

  @media (min-width: 768px) {
    position: static;
    transform: none;
    opacity: 1;
    pointer-events: auto;
    box-shadow: none;
  }
}

.detail-sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid var(--panel-border-color);
  overflow: hidden;
}

.detail-sidebar-brand {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.75rem;
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--item-background);
  box-shadow: inset 0 0 0 1px var(--panel-border-color);
}

.detail-sidebar-brand-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-sidebar-brand-fallback {
  width: 1.1rem;
  height: 1.1rem;
  color: var(--text-color-muted);
}

.detail-sidebar-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 0.15rem;
}

.detail-sidebar-organization {
  font-size: 0.75rem;
  color: var(--text-color-muted);
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-sidebar-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;

  h4 {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--text-color);
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.detail-sidebar-subtitle {
  font-size: 0.75rem;
  color: var(--text-color-muted);
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.5rem;
}

.detail-sidebar-group {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.detail-sidebar-group-label {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-color-muted);
}

.detail-sidebar-menu {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.detail-sidebar-menu-button {
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
  transition:
    background-color 150ms ease,
    color 150ms ease;

  &:hover {
    background-color: var(--button-ghost-hover-background);
    color: var(--text-color);
  }

  &:focus-visible {
    box-shadow: inset 0 0 0 1px var(--primary-color);
  }

  &.active,
  &.router-link-active {
    background-color: var(--button-ghost-hover-background);
    color: var(--text-color);
  }
}

.detail-sidebar-menu-icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  color: var(--menu-icon-color);
}

.detail-sidebar-menu-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-sidebar-menu-count {
  margin-left: auto;
  flex-shrink: 0;
  background-color: var(--item-hover-background);
  color: inherit;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.detail-sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.5rem;
  border-top: 1px solid var(--panel-border-color);
}
</style>
