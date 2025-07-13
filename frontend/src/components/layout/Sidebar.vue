<template>
    <div :class="['sidebar-wrapper', currentViewClass]">
        <div class="main-sidebar sidebar">
            <nav class="nav-main">
                <router-link
                    v-for="item in navigation"
                    :key="item.route"
                    :to="{ name: item.route }"
                    :class="{
                        'nav-link': true,
                        'active': isActive(item, $route),
                    }"
                    :aria-current="isActive(item, $route) ? 'page' : undefined"
                >
                    <component :is="$router.resolve({name: item.route}).meta.icon" v-if="$router.resolve({name: item.route}).meta.icon" class="nav-link-icon"/>
                    <span class="nav-link-name">{{ $router.resolve({name: item.route}).meta.title }}</span>
                </router-link>
            </nav>

            <!-- Store List -->
            <ul class="nav-stores">
                <li v-for="shop in shops" :key="shop.id" class="nav-link-item">
                    <router-link
                        :to="{
                            name: 'account.shops.detail',
                            params: {
                                slug: shop.organizationSlug,
                                shopId: shop.id
                            }
                        }"
                        active-class=""
                        class="nav-link"
                    >
                        <div class="nav-link-icon">
                            <img v-if="shop.favicon" :src="shop.favicon" alt="Shop Logo" class="item-logo-img">
                            <icon-fa6-solid:earth-americas
                                v-else
                                class="placeholder-image"
                            />
                        </div>

                        <span class="nav-link-name">{{ shop.name }}</span>

                        <div class="nav-link-state">
                            <status-icon :status="shop.status" />
                        </div>
                    </router-link>
                </li>
            </ul>
        </div>
    </div>
</template>

<script setup lang="ts">
import type {RouteLocationNormalizedLoaded} from "vue-router";

const shops = ref<RouterOutput['account']['currentUserShops']>([]);

trpcClient.account.currentUserShops.query().then((data) => {
    shops.value = data;
});

import {ref, computed} from "vue";
import {RouterOutput, trpcClient} from "@/helpers/trpc";
import {useRoute} from "vue-router";

const route = useRoute();

const currentViewClass = computed(() => {
    const routeName = route.name as string;

    if (!routeName) {
        return 'is-default';
    }

    return 'is-' + routeName.replace(/\./g, '-');
});

const navigation = [
    { route: 'home' },
    { route: 'account.project.list', active: 'shop' },
    { route: 'account.extension.list' },
    { route: 'account.organizations.list', active: 'organizations', },
];

function isActive(
    item: { route: string; active?: string },
    $route: RouteLocationNormalizedLoaded,
) {
    if (item.route === $route.name) {
        return true;
    }
    return !!($route.name &&
        typeof $route.name === 'string' &&
        item.active &&
        $route.name.match(item.active));
}
</script>

<style>
.sidebar-wrapper {
    display: none;
    grid-area: sidebar;

    @media all and (min-width: 1024px) {
        display: block;
        width: 250px;

        &[class*="is-account-shops-detail"] {
            width: 60px;

            .nav-link-name,
            .nav-link-state {
                display: none;
            }

            .nav-link {
                padding: 0.5rem;
            }
        }
    }
}

.sidebar {
    background-color: var(--panel-background);
    padding: 0.75rem;
    border-radius: 0.375rem;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
    width: 100%;
    flex-shrink: 0;

    .dark & {
        box-shadow: none;
    }
}

/* Main */
.nav-main {
    padding-bottom: 0.3rem;
    margin-bottom: 0.5rem;
    border-bottom: 1px solid var(--panel-border-color);
}

.nav-link {
    display: flex;
    color: var(--sidebar-nav-link-color);
    margin-bottom: 0.2rem;
    padding: 0.5rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 1rem;
    font-weight: 300;
    gap: 0.875rem;
    align-items: center;
    transition: all 0.2s ease-in-out;

    &:hover,
    &.active {
        background-color: var(--primary-color);
        color: #ffffff;
    }

    .nav-link-icon {
        width: 1.25rem;
    }

    .nav-link-state {
        margin-left: auto;
    }
}

.nav-stores {
    .nav-link-name {
        overflow: hidden;
        max-width: 8.35rem;
        white-space: nowrap;
        text-overflow: ellipsis;
    }
}
</style>
