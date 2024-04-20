<template>
    <disclosure
        v-if="authStore.user"
        v-slot="{ open }"
        as="nav"
        class="nav"
    >
        <div class="nav-container container">
            <div class="nav-main">
                <div class="nav-logo">
                    <router-link to="/">
                        <logo class="nav-logo-img" />
                    </router-link>
                </div>

                <div class="nav-main-menue">
                    <router-link
                        v-for="item in navigation"
                        :key="item.name"
                        :to="item.route"
                        :class="[
                            'nav-link',
                            item.route == $route.path
                                ? 'active'
                                : '',
                        ]"
                        :aria-current="item.route == $route.path ? 'page' : undefined"
                    >
                        {{ item.name }}
                    </router-link>
                </div>
            </div>

            <div class="nav-actions">
                <a
                    href="https://github.com/FriendsOfShopware/shopmon/"
                    target="_blank"
                    class="action action-github"
                >
                    <icon-fa-brands:github class="icon" />
                </a>

                <button
                    class="action action-dark-mode"
                    type="button"
                    @click="darkModeStore.toggleDarkMode"
                >
                    <icon-fa6-regular:moon
                        v-if="darkModeStore.darkMode"
                        class="icon"
                    />
                    <icon-fa6-regular:sun
                        v-else
                        class="icon"
                    />
                </button>

                <popover class="notifications">
                    <popover-button
                        class="action notifications"
                        @click="notificationStore.markAllRead"
                    >
                        <span class="sr-only">View notifications</span>
                        <icon-fa6-solid:bell
                            class="icon"
                            aria-hidden="true"
                        />
                        <div
                            v-if="notificationStore.unreadNotificationCount > 0"
                            class="notifications-count"
                        >
                            {{ notificationStore.unreadNotificationCount }}
                        </div>
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
                            <div class="notifications-header">
                                Notifications ({{ notificationStore.notifications.length }})
                                <button
                                    v-if="notificationStore.notifications.length > 0"
                                    class="notification-delete"
                                    type="button"
                                    @click="notificationStore.deleteAllNotifications"
                                >
                                    <icon-fa6-solid:trash class="icon" />
                                </button>
                            </div>
                            <ul
                                v-if="notificationStore.notifications.length > 0"
                                class="notifications-list"
                            >
                                <li
                                    v-for="(notification, index) in notificationStore.notifications"
                                    :key="index"
                                    class="notification-item"
                                >
                                    <div class="notification-icon">
                                        <icon-fa6-solid:circle-xmark
                                            v-if="notification.level === 'error'"
                                            class="icon-error"
                                        />
                                        <icon-fa6-solid:circle-info
                                            v-else
                                            class="icon-warning"
                                        />
                                    </div>

                                    <div class="notification-item-content">
                                        <div class="notification-item-title">
                                            {{ notification.title }}
                                        </div>
                                        <div class="notification-item-date">
                                            {{ formatDateTime(notification.createdAt) }}
                                        </div>
                                        <div class="notification-item-message">
                                            {{ notification.message }} <router-link
                                                v-if="notification.link"
                                                :to="notification.link"
                                                type="a"
                                            >
                                                more ...
                                            </router-link>
                                        </div>
                                    </div>

                                    <div class="notification-delete">
                                        <button
                                            type="button"
                                            @click="notificationStore.deleteNotification(notification.id)"
                                        >
                                            <icon-fa6-solid:xmark class="icon" />
                                        </button>
                                    </div>
                                </li>
                            </ul>

                            <div
                                v-else
                                class="notifications-empty"
                            >
                                Not much going on here
                            </div>
                        </popover-panel>
                    </transition>
                </popover>


                <!-- Profile dropdown -->
                <menu-container
                    as="div"
                    class="user-menu"
                >
                    <menu-button class="action action-user">
                        <span class="sr-only">Open user menu</span>
                        <img
                            class="user-avatar"
                            :src="authStore.user.avatar"
                            alt=""
                        >
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
                                <div class="user-menu-name">
                                    Hi {{ authStore.user.displayName }}
                                </div>
                            </div>
                            <menu-item
                                v-for="item in userNavigation"
                                :key="item.name"
                                v-slot="{ active }"
                            >
                                <button
                                    class="user-menu-item"
                                    :class="[
                                        active && 'active',
                                    ]"
                                    type="button"
                                    @click="item.route === '/logout' ? authStore.logout() : $router.push(item.route)"
                                >
                                    <component
                                        :is="item.icon"
                                        class="icon"
                                    />
                                    {{ item.name }}
                                </button>
                            </menu-item>
                        </menu-items>
                    </transition>
                </menu-container>

                <!-- Mobile menu button -->
                <disclosure-button class="action action-mobile-nav-toggle">
                    <span class="sr-only">Open main menu</span>
                    <icon-fa6-solid:bars-staggered
                        v-if="!open"
                        class="icon"
                        aria-hidden="true"
                    />
                    <icon-fa6-solid:xmark
                        v-else
                        class="icon"
                        aria-hidden="true"
                    />
                </disclosure-button>
            </div>
        </div>

        <!-- Mobile menu -->
        <disclosure-panel class="nav-mobile-menu">
            <div class="nav-mobile-links">
                <disclosure-button
                    v-for="item in navigation"
                    :key="item.name"
                    as="a"
                    class="nav-mobile-link"
                    :class="[
                        item.route == $route.path && 'active',
                    ]"
                    :aria-current="item.route == $route.path ? 'page' : undefined"
                    @click="$router.push(item.route)"
                >
                    {{ item.name }}
                </disclosure-button>
            </div>
            <div class="nav-mobile-user">
                <div class="nav-mobile-user-info">
                    <div class="nav-mobile-user-avatar">
                        <img
                            :src="authStore.user.avatar"
                            alt="avatar"
                            class="user-avatar"
                        >
                    </div>
                    <div class="nav-mobile-user-details">
                        <div class="nav-mobile-user-name">
                            {{ authStore.user.displayName }}
                        </div>
                        <div class="nav-mobile-user-email">
                            {{ authStore.user.email }}
                        </div>
                    </div>
                </div>

                <div class="nav-mobile-user-links">
                    <disclosure-button
                        v-for="item in userNavigation"
                        :key="item.name"
                        as="a"
                        class="nav-mobile-user-link"
                        @click="
                            item.route === '/logout'
                                ? authStore.logout()
                                : $router.push(item.route)
                        "
                    >
                        <component
                            :is="item.icon"
                            class="icon"
                        />
                        {{ item.name }}
                    </disclosure-button>
                </div>
            </div>
        </disclosure-panel>
    </disclosure>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth.store';
import { useNotificationStore } from '@/stores/notification.store';
import { useDarkModeStore } from '@/stores/darkMode.store';

import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Menu as MenuContainer,
    MenuButton,
    MenuItem,
    MenuItems,
    Popover,
    PopoverButton,
    PopoverPanel,
} from '@headlessui/vue';

import FaGear from '~icons/fa6-solid/gear';
import FaPowerOff from '~icons/fa6-solid/power-off';

import { formatDateTime } from '@/helpers/formatter';

const authStore = useAuthStore();
const notificationStore = useNotificationStore();
const darkModeStore = useDarkModeStore();

const navigation = [
    { name: 'Dashboard', route: '/' },
    { name: 'My Shops', route: '/account/shops' },
    { name: 'My Apps', route: '/account/extensions' },
    { name: 'My Organizations', route: '/account/organizations' },
];

const userNavigation = [
    { name: 'Settings', route: '/account/settings', icon: FaGear },
    { name: 'Logout', route: '/logout', icon: FaPowerOff },
];
</script>

<style>
.nav {
    background-color: var(--primary-color);
}

.nav-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 4rem;
}

/* Logo */
.nav-logo {
    flex-shrink: 0;

    img {
        height: 2.5rem;
        width: auto;
    }
}

/* Main */
.nav-main {
    display: flex;

    &-menue {
        display: none;

        @media (min-width: 768px) {
            display: flex;
            margin-left: 1.5rem;
            align-items: baseline;
            gap: 1rem;
        }
    }
}

.nav-link {
    color: #fff;
    padding: 0.5rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;

    &:hover,
    &.active {
        background-color: #0c84c2;
        color: #ffffff;
    }
}

/* Actions */
.nav-actions {
    display: flex;
    margin: 0 0 0 1rem;
    align-items: center;
    gap: 0.75rem;

    .action {
        height: 1.25rem;
        width: 1.25rem;
        color: #bae6fd;
        display: flex;
        justify-content: center;
        align-items: center;

        &:hover {
            color: #ffffff;
        }

        .icon {
            height: 1.25rem;
            width: 1.25rem;
        }

        &.action-user {
            display: block;
            height: 2rem;
            width: 2rem;
            background-color: #38bdf8;
            border-radius: 9999px;
            display: flex;
            align-items: center;
            margin-left: 1rem;

            .user-avatar {
                border-radius: 9999px;
            }
        }

        &.action-mobile-nav-toggle {
            margin-left: 1rem;

            @media (min-width: 768px) {
                display: none;
            }

            .icon {
                height: 1.5rem;
                width: 1.5rem;
            }
        }
    }

    .user-menu {
        display: none;

        @media (min-width: 768px) {
            display: block;
        }
    }
}

/* Benachrichtigungen */
.notifications {
    position: relative;

    &-count {
        position: absolute;
        right: -0.5rem;
        top: -0.5rem;
        background-color: #ef4444;
        border-radius: 9999px;
        padding: 2px;
        min-width: 16px;
        font-size: 10px;
        line-height: 12px;
        color: #ffffff;
        font-weight: bold;
    }

    &-header {
        padding: 1rem;
        font-weight: 500;
        border-bottom: 1px solid var(--notification-border-color);
        display: flex;
        justify-content: space-between;
    }

    &-empty {
        padding: 1rem;
    }

    &-delete {
        opacity: 0.3;

        &:hover {
            opacity: 1;
        }
    }
}

.notifications-panel {
    position: absolute;
    top: 140%;
    left: 0;
    right: 0;
    z-index: 20;
    width: 100vw;
    border-radius: 0.375rem;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    background-color: var(--panel-background);
    overflow: hidden;

    @media (min-width: 640px) {
        padding-left: 1.5rem;
        padding-right: 1.5rem;
    }

    @media (min-width: 768px) {
        max-width: 25rem;
        left: auto;
        padding-left: 0;
        padding-right: 0;
    }
}

.notifications-list {
    overflow-y: auto;
    max-height: 20rem;
}

.notification-item {
    padding-left: 1rem;
    padding-right: 1rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
    display: flex;
    gap: 0.5rem;
    background-color: var(--item-background);

    &:not(:last-child) {
        border-bottom: 1px solid var(--notification-border-color);
    }

    &:nth-child(odd) {
        background-color: var(--item-odd-background);
    }

    &:hover {
        background-color: var(--item-hover-background);
    }

    .notification-item-content {
        flex: 1;
    }

    .notification-item-title {
        font-weight: 500;
    }

    .notification-item-message {
        color: var(--text-color-muted);
    }

    .notification-item-date {
        font-size: 0.75rem;
        margin-bottom: 0.375rem;
        color: var(--text-color-muted);
    }

    .notification-delete {
        visibility: hidden;
    }

    &:hover {
        .notification-delete {
            visibility: visible;
        }
    }
}

.notification-delete {
    opacity: .5;

    &:hover {
        opacity: 1;
    }
}

/* Nutzermenu */
.user-menu {
    position: relative;

    &-panel {
        position: absolute;
        top: 110%;
        right: 0;
        width: 12rem;
        border-radius: 0.375rem;
        box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
        padding-top: 0.25rem;
        padding-bottom: 0.25rem;
        background-color: var(--panel-background);
        z-index: 20;
    }

    &-header {
        padding: 0.5rem 1rem;
        border-bottom: 1px solid var(--panel-border-color);
    }

    &-name {
        font-size: 1rem;
        font-weight: 500;
    }

    &-item {
        width: 100%;
        text-align: left;
        padding-left: 1rem;
        padding-right: 1rem;
        padding-top: 0.5rem;
        padding-bottom: 0.5rem;
        font-size: 0.875rem;

        &.active {
            background-color: var(--user-menu-item-active-background-color);
        }
    }

    .icon {
        width: 1rem;
        height: 1rem;
        display: inline-block;
        color: var(--user-menu-icon-color);
        margin-right: 0.25rem;
    }
}

/* Mobile Menü */
.nav-mobile-menu {
    width: 100%;
    position: absolute;
    background-color: var(--primary-color);
    z-index: 10;
    filter: drop-shadow(0 10px 8px rgba(0, 0, 0, 0.04)) drop-shadow(0 4px 3px rgba(0, 0, 0, 0.1));

    @media (min-width: 768px) {
        display: none;
    }
}

.nav-mobile-links {
    padding: 0.5rem 0.5rem 0.75rem;
    gap: 0.25rem;
    display: flex;
    flex-direction: column;

    @media (min-width: 640px) {
        padding-left: 0.75rem;
        padding-right: 0.75rem;
    }
}

.nav-mobile-link {
    display: block;
    padding: 0.5rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 1rem;
    font-weight: 500;
    color: #ffffff;

    &:hover,
    &.active {
        color: #ffffff;
        background-color: #0c84c2;
    }
}

.nav-mobile-user {
    padding: 1rem 0 0.75rem;
    border-top: 1px solid #38bdf8;

    &-info {
        padding-left: 1.25rem;
        padding-right: 1.25rem;
        display: flex;
        align-items: center;
    }

    &-avatar {
        flex-shrink: 0;

        .user-avatar {
            height: 2.5rem;
            width: 2.5rem;
            border-radius: 9999px;
        }
    }

    &-details {
        margin-left: 0.75rem;
    }

    &-name {
        font-size: 1rem;
        font-weight: 500;
        color: #ffffff;
    }

    &-email {
        font-size: 0.875rem;
        font-weight: 500;
        color: #e7e5e4;
    }

    &-links {
        margin-top: 0.75rem;
        padding: 0 0.5rem;
        gap: 0.25rem;
        display: flex;
        flex-direction: column;
    }

    &-link {
        display: block;
        padding: 0.25rem 0.75rem;
        border-radius: 0.25rem;
        font-size: 1rem;
        font-weight: 500;
        color: #ffffff;

        &:hover {
            background-color: #0c84c2;
        }

        .icon {
            width: 1rem;
            height: 1rem;
            display: inline-block;
            margin-right: 0.25rem;
        }
    }
}
</style>
