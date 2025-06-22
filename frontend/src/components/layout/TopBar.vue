<template>
    <disclosure
        v-if="session.data"
        v-slot="{ open }"
        as="div"
        class="top-bar"
    >
        <div class="top-bar-container container">
            <div class="top-bar-logo">
                <router-link :to="{ name: 'home' }">
                    <logo class="nav-logo-img"/>
                </router-link>
            </div>

            <div class="top-bar-actions">
                <a
                    href="https://github.com/FriendsOfShopware/shopmon/"
                    target="_blank"
                    class="action action-github"
                >
                    <icon-fa-brands:github class="icon"/>
                </a>

                <button
                    class="action action-dark-mode"
                    type="button"
                    @click="toggleDarkMode"
                >
                    <icon-fa6-regular:moon
                        v-if="darkMode"
                        class="icon"
                    />

                    <icon-octicon:sun-16
                        v-else
                        class="icon"
                    />
                </button>

                <popover class="notifications">
                    <popover-button
                        class="action notifications"
                        @click="markAllRead"
                    >
                        <span class="sr-only">View notifications</span>

                        <icon-fa6-solid:bell
                            class="icon"
                            aria-hidden="true"
                        />

                        <div
                            v-if="unreadNotificationCount > 0"
                            class="notifications-count"
                        >
                            {{ unreadNotificationCount }}
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
                                Notifications ({{ notifications.length }})

                                <button
                                    v-if="notifications.length > 0"
                                    class="notification-delete"
                                    type="button"
                                    @click="deleteAllNotifications"
                                >
                                    <icon-fa6-solid:trash class="icon"/>
                                </button>
                            </div>

                            <ul
                                v-if="notifications.length > 0"
                                class="notifications-list"
                            >
                                <li
                                    v-for="(notification, index) in notifications"
                                    :key="index"
                                    class="notification-item"
                                >
                                    <div class="notification-icon">
                                        <icon-fa6-solid:circle-xmark
                                            v-if="notification.level === 'error'"
                                            class="icon icon-error"
                                        />

                                        <icon-fa6-solid:circle-info
                                            v-else
                                            class="icon icon-warning"
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
                                            {{ notification.message }}
                                            <router-link
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
                                            @click="deleteNotification(notification.id)"
                                        >
                                            <icon-fa6-solid:xmark class="icon"/>
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
                    class="menu"
                >
                    <menu-button class="action action-user">
                        <span class="sr-only">Open user menu</span>
                        <img
                            class="user-avatar"
                            :src="userAvatar"
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
                        <menu-items class="menu-panel">
                            <div class="menu-header">
                                <div class="menu-name">
                                    Hi {{ session.data.user.name }}
                                </div>
                            </div>

                            <menu-item
                                v-for="item in userNavigation"
                                :key="item.name"
                            >
                                <button
                                    class="menu-item"
                                    type="button"
                                    @click="item.route === 'logout' ? logout() : $router.push({ name: item.route })"
                                >
                                    <component :is="item.icon" class="icon" />
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

                <!-- Mobile menu -->
                <disclosure-panel class="top-bar-mobile-menu">
                    <div class="top-bar-mobile-links">
                        <disclosure-button
                            v-for="item in navigation"
                            :key="item.name"
                            as="a"
                            class="top-bar-mobile-link"
                            @click="$router.push({ name: item.route })"
                        >
                            <component :is="$router.resolve({name: item.route}).meta.icon" v-if="$router.resolve({name: item.route}).meta.icon" class="icon"/>
                            {{ $router.resolve({name: item.route}).meta.title }}
                        </disclosure-button>
                    </div>

                    <div class="top-bar-mobile-user">
                        <div class="top-bar-mobile-user-info">
                            <div class="top-bar-mobile-user-avatar">
                                <img
                                    :src="userAvatar"
                                    alt="avatar"
                                    class="user-avatar"
                                >
                            </div>

                            <div class="top-bar-mobile-user-details">
                                <div class="top-bar-mobile-user-name">
                                    {{ session.data.user.name }}
                                </div>

                                <div class="top-bar-mobile-user-email">
                                    {{ session.data.user.email }}
                                </div>
                            </div>
                        </div>

                        <div class="top-bar-mobile-links">
                            <disclosure-button
                                v-for="item in userNavigation"
                                :key="item.name"
                                as="a"
                                class="top-bar-mobile-link"
                                @click="
                            item.route === 'logout'
                                ? logout()
                                : $router.push({ name: item.route })
                        "
                            >
                                <component :is="item.icon" class="icon" />
                                {{ item.name }}
                            </disclosure-button>
                        </div>
                    </div>
                </disclosure-panel>
            </div>
        </div>
    </disclosure>
</template>

<script setup lang="ts">
import { useDarkMode } from '@/composables/useDarkMode';
import { useNotifications } from '@/composables/useNotifications';

import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    MenuButton,
    Menu as MenuContainer,
    MenuItem,
    MenuItems,
    Popover,
    PopoverButton,
    PopoverPanel,
} from '@headlessui/vue';

import { ref } from 'vue';
import FaGear from '~icons/fa6-solid/gear';
import FaPowerOff from '~icons/fa6-solid/power-off';

import { authClient } from '@/helpers/auth-client';
import { formatDateTime } from '@/helpers/formatter';

const session = authClient.useSession();

const userAvatar = ref(
    'https://api.dicebear.com/7.x/personas/svg?seed=default?d=identicon',
);

const navigation = [
    { route: 'home' },
    { route: 'account.project.list', active: 'shop' },
    { route: 'account.extension.list' },
    { route: 'account.organizations.list', active: 'organizations', },
];

if (session.value.data?.user.email) {
    try {
        crypto.subtle
            .digest(
                'SHA-256',
                new TextEncoder().encode(session.value.data.user.email),
            )
            .then((hash) => {
                const seed = Array.from(new Uint8Array(hash))
                    .map((b) => b.toString(16).padStart(2, '0'))
                    .join('');
                userAvatar.value = `https://api.dicebear.com/7.x/personas/svg?seed=${seed}&d=identicon`;
            });
        // eslint-disable-next-line no-unused-vars
    } catch (error) {
        // Web crypto API is not supported, fallback to default avatar
        // Silent fallback, no need for console error
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

const userNavigation = [
    {name: 'Settings', route: 'account.settings', icon: FaGear},
    {name: 'Logout', route: 'logout', icon: FaPowerOff},
];

async function logout() {
    await authClient.signOut();
    window.location.reload();
}

</script>

<style>
.top-bar {
    position:relative;
}

.top-bar-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 4rem;
}

/* Logo */
.top-bar-logo {
    flex-shrink: 0;

    img {
        height: 3.5rem;
        width: auto;
    }
}

/* Actions */
.top-bar-actions {
    display: flex;
    margin: 0 0 0 1rem;
    gap: 0.75rem;

    .action {
        height: 2rem;
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
            height: 2rem;
            width: 2rem;
            background-color: #38bdf8;
            border-radius: 9999px;
            display: none;
            align-items: center;
            margin-left: 1rem;

            @media (min-width: 1024px) {
                display: flex;
            }
            .user-avatar {
                border-radius: 9999px;
            }
        }

        &.action-mobile-nav-toggle {
            margin-left: 1rem;

            @media (min-width: 1024px) {
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

        @media (min-width: 1024px) {
            display: block;
        }
    }
}

/* Benachrichtigungen */
.notifications {
    @media all and (min-width: 768px) {
        position: relative;
    }

    &-count {
        position: absolute;
        right: -0.4rem;
        top: 0;
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
    top: 110%;
    left: 0.5rem;
    right: 0.5rem;
    z-index: 20;
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
        min-width: 12rem;
        left: auto;
        right: 0;
        padding-left: 0;
        padding-right: 0;
    }
}

.notifications-list {
    overflow-y: auto;
    max-height: 20rem;
}

.notification-item {
    padding: 0.5rem 1rem;
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

/* Mobile Menü */
.top-bar-mobile-menu {
    width: 100%;
    position: absolute;
    left: 0;
    top: 110%;
    background-color: var(--primary-color);
    z-index: 10;
    filter: drop-shadow(0 10px 8px rgba(0, 0, 0, 0.04)) drop-shadow(0 4px 3px rgba(0, 0, 0, 0.1));

    @media (min-width: 1024px) {
        display: none;
    }
}

.top-bar-mobile-links {
    padding: 0.5rem 0.5rem 0.75rem;
    gap: 0.25rem;
    display: flex;
    flex-direction: column;

    @media (min-width: 640px) {
        padding-left: 0.75rem;
        padding-right: 0.75rem;
    }
}

.top-bar-mobile-link {
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

    .icon {
        width: 1rem;
        height: 1rem;
        display: inline-block;
        margin-right: 0.25rem;
    }
}

.top-bar-mobile-user {
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
            background: #38bdf8;
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
}
</style>
