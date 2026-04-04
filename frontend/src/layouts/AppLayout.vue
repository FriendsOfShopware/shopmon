<template>
  <div class="app-shell">
    <!-- Top header bar -->
    <header class="app-header">
      <div class="app-header-start">
        <router-link :to="{ name: 'home' }" class="header-logo-link">
          <logo class="header-logo-img" />
        </router-link>
        <OrganizationSwitcher />
      </div>

      <div class="app-header-end">
        <a
          href="https://github.com/FriendsOfShopware/shopmon/"
          target="_blank"
          class="header-action"
          aria-label="GitHub"
        >
          <icon-fa-brands:github class="header-action-icon" />
        </a>

        <UiButton class="header-action" type="button" @click="toggleDarkMode" aria-label="Toggle dark mode">
          <icon-fa6-regular:moon v-if="darkMode" class="header-action-icon" />
          <icon-octicon:sun-16 v-else class="header-action-icon" />
        </UiButton>

        <UiButton class="header-action header-action-locale" type="button" @click="toggleLocale">
          {{ String(locale) === "en" ? "DE" : "EN" }}
        </UiButton>

        <!-- Notifications -->
        <popover class="header-notifications">
          <popover-button class="header-action" @click="markAllRead" aria-label="Notifications">
            <icon-fa6-solid:bell class="header-action-icon" />
            <span v-if="unreadNotificationCount > 0" class="header-notification-badge">
              {{ unreadNotificationCount }}
            </span>
          </popover-button>

          <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="transform opacity-0 scale-95"
            enter-to-class="transform opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="transform opacity-100 scale-100"
            leave-to-class="transform opacity-0 scale-95"
          >
            <popover-panel class="notifications-panel">
              <div class="notifications-panel-header">
                {{ $t("topBar.notifications", { count: notifications.length }) }}
                <UiButton
                  v-if="notifications.length > 0"
                  class="notifications-delete-btn"
                  type="button"
                  @click="deleteAllNotifications"
                >
                  <icon-fa6-solid:trash class="icon-xs" />
                </UiButton>
              </div>

              <ul v-if="notifications.length > 0" class="notifications-list">
                <li
                  v-for="(notification, index) in notifications"
                  :key="index"
                  class="notification-item"
                >
                  <div class="notification-icon">
                    <icon-fa6-solid:circle-xmark
                      v-if="notification.level === 'error'"
                      class="icon-sm icon-error"
                    />
                    <icon-fa6-solid:circle-info v-else class="icon-sm icon-warning" />
                  </div>
                  <div class="notification-item-content">
                    <div class="notification-item-title">{{ notification.title }}</div>
                    <div class="notification-item-date">{{ formatDateTime(notification.createdAt) }}</div>
                    <div class="notification-item-message">
                      {{ notification.message }}
                      <a v-if="notification.link" :href="notification.link.url">
                        {{ notification.link.label || $t("topBar.more") }}
                      </a>
                    </div>
                  </div>
                  <button type="button" class="notification-item-dismiss" @click="deleteNotification(notification.id)">
                    <icon-fa6-solid:xmark class="icon-xs" />
                  </button>
                </li>
              </ul>
              <div v-else class="notifications-empty">{{ $t("notifications.notMuchGoingOn") }}</div>
            </popover-panel>
          </transition>
        </popover>

        <!-- User menu -->
        <menu-container as="div" class="header-user-menu">
          <menu-button class="header-action header-action-user" aria-label="User menu">
            <img class="header-avatar" :src="userAvatar" alt="" />
          </menu-button>

          <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="transform opacity-0 scale-95"
            enter-to-class="transform opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="transform opacity-100 scale-100"
            leave-to-class="transform opacity-0 scale-95"
          >
            <menu-items class="user-menu-panel">
              <div class="user-menu-header">
                {{ $t("topBar.greeting", { name: session.user.name }) }}
              </div>
              <menu-item v-for="item in userNavigation" :key="item.name">
                <button
                  class="user-menu-item"
                  type="button"
                  @click="item.route === 'logout' ? logout() : $router.push({ name: item.route })"
                >
                  <component :is="item.icon" class="icon-sm" />
                  {{ item.name }}
                </button>
              </menu-item>
            </menu-items>
          </transition>
        </menu-container>

        <!-- Mobile sidebar toggle -->
        <button
          class="header-action header-mobile-toggle"
          type="button"
          @click="mobileSidebarOpen = !mobileSidebarOpen"
          aria-label="Toggle navigation"
        >
          <icon-fa6-solid:bars-staggered v-if="!mobileSidebarOpen" class="header-action-icon" />
          <icon-fa6-solid:xmark v-else class="header-action-icon" />
        </button>
      </div>
    </header>

    <!-- Body: Sidebar + Content -->
    <div class="app-body">
      <!-- Mobile overlay -->
      <div
        v-if="mobileSidebarOpen"
        class="app-sidebar-overlay"
        @click="mobileSidebarOpen = false"
      />

      <!-- Sidebar -->
      <sidebar :class="{ 'sidebar-mobile-open': mobileSidebarOpen }" />

      <!-- Main content -->
      <main class="app-main">
        <ImpersonationBanner />
        <Notification />

        <div class="app-main-content">
          <router-view />
        </div>

        <layout-footer />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";

import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Notification from "@/components/Notification.vue";
import LayoutFooter from "@/components/layout/LayoutFooter.vue";
import OrganizationSwitcher from "@/components/layout/OrganizationSwitcher.vue";

import {
  useDarkMode,
} from "@/composables/useDarkMode";
import { useLocale } from "@/composables/useLocale";
import { useNotifications } from "@/composables/useNotifications";
import { useSession, clearSession } from "@/composables/useSession";
import { api, setToken } from "@/helpers/api";
import { useI18n } from "vue-i18n";

import {
  MenuButton,
  Menu as MenuContainer,
  MenuItem,
  MenuItems,
  Popover,
  PopoverButton,
  PopoverPanel,
} from "@headlessui/vue";

import FaGear from "~icons/fa6-solid/gear";
import FaPowerOff from "~icons/fa6-solid/power-off";
import { formatDateTime } from "@/helpers/formatter";

const router = useRouter();
const { session } = useSession();
const mobileSidebarOpen = ref(false);

const userAvatar = ref("https://api.dicebear.com/7.x/personas/svg?seed=default?d=identicon");

if (session.value?.user.email) {
  try {
    crypto.subtle
      .digest("SHA-256", new TextEncoder().encode(session.value.user.email))
      .then((hash) => {
        const seed = Array.from(new Uint8Array(hash))
          .map((b) => b.toString(16).padStart(2, "0"))
          .join("");
        userAvatar.value = `https://api.dicebear.com/7.x/personas/svg?seed=${seed}&d=identicon`;
      });
  } catch {
    // Silent fallback
  }
}

const {
  notifications,
  unreadNotificationCount,
  markAllRead,
  deleteAllNotifications,
  deleteNotification,
} = useNotifications();
const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
const { t } = useI18n();

const userNavigation = computed(() => [
  { name: t("nav.settings"), route: "account.settings", icon: FaGear },
  { name: t("nav.logout"), route: "logout", icon: FaPowerOff },
]);

async function logout() {
  await api.POST("/auth/sign-out");
  setToken(null);
  clearSession();
  router.push({ name: "home" });
}
</script>

<style scoped>
/* ── App Shell ──
   Matches Kumo's pattern: full viewport height, flex column,
   no scroll on body — all scrolling happens inside .app-main */
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background-color: var(--background-color);
}

/* ── Header ──
   Clean top bar: flex row, items center, space-between,
   subtle bottom border, neutral background */
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 3rem; /* 48px — slightly shorter than before */
  padding: 0 1rem;
  flex-shrink: 0;
  background-color: var(--panel-background);
  border-bottom: 1px solid var(--panel-border-color);
  z-index: 30;

  @media (min-width: 1024px) {
    padding: 0 1.25rem;
    height: 3.25rem; /* 52px on desktop */
  }
}

.app-header-start {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.header-logo-link {
  display: flex;
  align-items: center;
  line-height: 0;
}

.header-logo-img {
  height: 1.75rem;
  width: auto;
}

/* Org switcher in header — base styles from OrganizationSwitcher.vue are sufficient */
.app-header-start .org-switcher {
  margin-left: 0;
}

/* Header end: action buttons */
.app-header-end {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: auto;
}

/* Action button — matches Kumo's subtle header action style */
.header-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  color: var(--text-color-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 150ms ease, background-color 150ms ease;
  position: relative;

  &:hover {
    color: var(--text-color);
    background-color: var(--item-hover-background);
  }

  &:focus-visible {
    outline: none;
    box-shadow: inset 0 0 0 1px var(--primary-color);
  }
}

.header-action-icon {
  width: 1.125rem;
  height: 1.125rem;
}

.header-action-locale {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.025em;
  width: auto;
  padding: 0 0.5rem;
}

/* User avatar */
.header-action-user {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  overflow: hidden;
  padding: 0;
  background-color: var(--primary-color);

  @media (min-width: 1024px) {
    display: flex;
  }

  /* Hide on small screens */
  @media (max-width: 1023px) {
    display: none;
  }

  .header-avatar {
    width: 100%;
    height: 100%;
    border-radius: 9999px;
    object-fit: cover;
  }
}

/* Notification badge */
.header-notification-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 14px;
  height: 14px;
  padding: 0 3px;
  font-size: 10px;
  font-weight: 700;
  line-height: 14px;
  color: #ffffff;
  background-color: var(--error-color);
  border-radius: 9999px;
  text-align: center;
}

/* Mobile toggle */
.header-mobile-toggle {
  @media (min-width: 768px) {
    display: none;
  }
}

/* ── Notifications Panel ── */
.header-notifications {
  position: relative;

  @media (min-width: 768px) {
    position: relative;
  }
}

.notifications-panel {
  position: absolute;
  top: calc(100% + 0.375rem);
  right: 0;
  z-index: 50;
  width: 20rem;
  max-width: calc(100vw - 2rem);
  border-radius: 0.5rem;
  background-color: var(--panel-background);
  border: 1px solid var(--panel-border-color);
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.1),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.notifications-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-bottom: 1px solid var(--panel-border-color);
}

.notifications-delete-btn {
  padding: 0.25rem;
  border-radius: 0.25rem;
  color: var(--text-color-muted);
  background: transparent;
  border: none;
  cursor: pointer;

  &:hover {
    color: var(--text-color);
  }
}

.notifications-list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 20rem;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  background-color: var(--panel-background);

  &:not(:last-child) {
    border-bottom: 1px solid var(--panel-border-color);
  }

  &:hover {
    background-color: var(--item-hover-background);
  }

  .notification-item-content {
    flex: 1;
    min-width: 0;
  }

  .notification-item-title {
    font-weight: 500;
    font-size: 0.875rem;
  }

  .notification-item-date {
    font-size: 0.75rem;
    color: var(--text-color-muted);
    margin-bottom: 0.125rem;
  }

  .notification-item-message {
    font-size: 0.8125rem;
    color: var(--text-color-muted);
    line-height: 1.4;
  }

  .notification-item-dismiss {
    visibility: hidden;
    padding: 0.125rem;
    border-radius: 0.25rem;
    color: var(--text-color-muted);
    background: transparent;
    border: none;
    cursor: pointer;
    flex-shrink: 0;
    opacity: 0.5;

    &:hover {
      opacity: 1;
    }
  }

  &:hover .notification-item-dismiss {
    visibility: visible;
  }
}

.notifications-empty {
  padding: 1.5rem 1rem;
  text-align: center;
  color: var(--text-color-muted);
  font-size: 0.875rem;
}

.notification-icon {
  flex-shrink: 0;
  padding-top: 2px;
}

.icon-sm {
  width: 1rem;
  height: 1rem;
}

.icon-error {
  color: var(--error-color);
}

.icon-warning {
  color: var(--info-color);
}

/* ── User Menu ── */
.header-user-menu {
  position: relative;
}

.user-menu-panel {
  position: absolute;
  top: calc(100% + 0.375rem);
  right: 0;
  z-index: 50;
  min-width: 11rem;
  border-radius: 0.5rem;
  background-color: var(--panel-background);
  border: 1px solid var(--panel-border-color);
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.1),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  padding: 0.25rem;
}

.user-menu-header {
  padding: 0.625rem 0.875rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-bottom: 1px solid var(--panel-border-color);
  margin-bottom: 0.25rem;
}

.user-menu-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  color: var(--text-color);
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: background-color 150ms ease;

  &:hover {
    background-color: var(--item-hover-background);
  }
}

/* ── App Body: Sidebar + Main ── */
.app-body {
  display: flex;
  flex: 1;
  min-height: 0; /* Critical: allows flex child to shrink below content size */
  overflow: hidden;
}

/* Mobile sidebar overlay */
.app-sidebar-overlay {
  display: block;
  position: fixed;
  inset: 0;
  z-index: 40;
  background-color: rgba(0, 0, 0, 0.4);

  @media (min-width: 768px) {
    display: none;
  }
}

/* ── Main Content Area ──
   Scrollable region containing page content + footer */
.app-main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  background-color: var(--background-color);
}

.app-main-content {
  flex: 1;
  padding: 1.25rem 1rem;

  @media (min-width: 1024px) {
    padding: 1.5rem 2rem;
  }
}
</style>
