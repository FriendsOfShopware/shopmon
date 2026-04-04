<template>
  <div v-if="props.environment" class="sidebar-wrapper">
    <div
      ref="toggleRef"
      class="sidebar-detail-toggle btn btn-primary btn-block"
      @click="toggleMobileOpen"
    >
      {{ environment?.organizationName + " / " + environment?.name }}
      {{ currentRouteTitle && route.name !== "account.environments.detail" ? " / " : "" }}
      {{ currentRouteTitle }}
    </div>

    <div ref="sidebarRef" :class="['sidebar', 'sidebar-detail', { 'mobile-open': isMobileOpen }]">
      <div class="environment-image-container">
        <img
          v-if="environment?.environmentImage"
          :src="`${environment.environmentImage}`"
          class="environment-image"
        />
        <icon-fa6-solid:image v-else class="placeholder-image" />
      </div>

      <div class="environment-name">
        <div class="sidebar-section-label">Environment:</div>
        <h4>{{ environment?.name }}</h4>
        <status-icon :status="environment?.status ?? ''" />
        <div>
          <a :href="environment?.url" data-tooltip="Go to storefront" target="_blank">
            <icon-fa6-solid:store />
            Storefront
          </a>
          &nbsp;/&nbsp;
          <a
            :href="(environment?.url ?? '') + '/admin'"
            data-tooltip="Go to shopware admin"
            target="_blank"
          >
            <icon-fa6-solid:user-gear />
            Admin
          </a>
        </div>
      </div>

      <nav class="nav-main">
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
            'nav-link': true,
            active: isRouteActive(item.route),
          }"
          active-class=""
          exact-active-class=""
        >
          <component
            :is="$router.resolve({ name: item.route }).meta.icon"
            v-if="$router.resolve({ name: item.route }).meta.icon"
            class="nav-link-icon"
          />
          <span class="nav-link-name">{{
            $t($router.resolve({ name: item.route }).meta.titleKey ?? "")
          }}</span>
          <span v-if="item.count !== undefined" class="nav-link-count">{{ item.count }}</span>
        </router-link>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { components } from "@/types/api";
import { computed, ref, onMounted, onUnmounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import StatusIcon from "@/components/StatusIcon.vue";

const route = useRoute();
const { t } = useI18n();

const isMobileOpen = ref(false);
const sidebarRef = ref<HTMLElement | null>(null);
const toggleRef = ref<HTMLElement | null>(null);

const props = defineProps<{
  environment: components["schemas"]["EnvironmentDetail"] | null;
}>();

const currentRouteTitle = computed(() => {
  const titleKey = route.meta.titleKey;
  return typeof titleKey === "string" ? t(titleKey) : "";
});

const isRouteActive = (routeName: string) => {
  // Exact match
  if (route.name === routeName) return true;

  // Check if current route is a child of deployments
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

function handleClickOutside(event: MouseEvent) {
  if (!isMobileOpen.value) return;

  // Check if click was outside both the sidebar and the toggle button
  const clickedElement = event.target as Node;
  const isClickInsideSidebar = sidebarRef.value?.contains(clickedElement) ?? false;
  const isClickOnToggle = toggleRef.value?.contains(clickedElement) ?? false;

  if (!isClickInsideSidebar && !isClickOnToggle) {
    isMobileOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});

// Close sidebar when route changes
watch(route, () => {
  if (isMobileOpen.value) {
    isMobileOpen.value = false;
  }
});
</script>

<style scoped>
.sidebar-wrapper {
  display: block;
  position: relative;
}

.sidebar-detail-toggle {
  border-color: transparent;
  margin-bottom: 1rem;
  background-color: #0284c7;

  &:after {
    content: "";
    position: absolute;
    right: 1rem;
    top: 50%;
    margin-top: -2px;
    border: 5px solid transparent;
    border-top-color: #fff;
  }

  @media all and (min-width: 1024px) {
    display: none;
  }
}

.sidebar {
  position: absolute;
  z-index: 5;
  top: calc(100% - 0.9rem);
  left: 0;
  right: 0;
  display: none;
  border: 1px solid var(--panel-border-color);

  &.mobile-open {
    display: block;
  }

  @media all and (min-width: 1024px) {
    position: static;
    display: block;
    border: unset;
  }
}

.environment-image-container {
  margin-top: 1.5rem;
  display: flex;
  justify-content: center;
  padding-bottom: 1rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--panel-border-color);

  @media (min-width: 640px) {
    grid-column: 2 / span 1;
    grid-row: 1 / span 1;
    margin-top: 0;
  }

  @media (min-width: 960px) {
    grid-column: 3 / span 1;
    grid-row: 1 / span 1;
  }
}

.placeholder-image {
  color: #e5e7eb;
  font-size: 9rem;
}

.environment-name {
  padding: 0 0 1rem 0.75rem;
  margin-bottom: 0.5rem;
  border-bottom: 1px solid var(--panel-border-color);
  display: flex;
  flex-flow: row wrap;
  justify-content: space-between;
  gap: 0.5rem;

  .sidebar-section-label {
    flex: 1;
    width: 100%;
  }

  h4 {
    width: calc(100% - 1.5rem);
  }

  .icon-status {
    margin-top: 0.2rem;
    flex-shrink: 0;
  }
}

.nav-link-count {
  margin-left: auto;
  background-color: rgba(0, 0, 0, 0.1);
  color: inherit;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
}
</style>
