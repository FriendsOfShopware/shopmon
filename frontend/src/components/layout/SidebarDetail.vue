<template>
    <div v-if="props.shop" class="sidebar-wrapper">
        <div ref="toggleRef" class="sidebar-detail-toggle btn btn-primary btn-block" @click="toggleMobileOpen">
            {{ shop.nameCombined }} {{ route.meta.title && route.name !== 'account.shops.detail' ? ' / ' : '' }} {{ route.meta.title }}
        </div>

        <div ref="sidebarRef" :class="[
            'sidebar',
            'sidebar-detail',
            { 'mobile-open': isMobileOpen }
            ]"
        >
            <div class="shop-image-container">
                <img
                    v-if="shop.shopImage"
                    :src="`${shop.shopImage}`"
                    class="shop-image"
                >
                <icon-fa6-solid:image
                    v-else
                    class="placeholder-image"
                />
            </div>

            <div class="shop-name">
                <div class="sidebar-section-label">Environment:</div>
                <h4>{{ shop.name }}</h4>
            </div>

            <nav class="nav-main">
                <router-link
                    v-for="item in detailNavigation"
                    :key="item.name"
                    :to="{
                    name: item.route,
                    params: {
                        slug: route.params.slug,
                        shopId: route.params.shopId
                    }
                }"
                    :class="{
                    'nav-link': true
                }"
                    active-class=""
                    exact-active-class="active"
                >
                    <component :is="$router.resolve({name: item.route}).meta.icon" v-if="$router.resolve({name: item.route}).meta.icon" class="nav-link-icon"/>
                    <span class="nav-link-name">{{ $router.resolve({name: item.route}).meta.title }}</span>
                    <span v-if="item.count !== undefined" class="nav-link-count">{{ item.count }}</span>
                </router-link>
            </nav>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { RouterOutput } from '@/helpers/trpc';
import { computed, ref, onMounted, onUnmounted, watch } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

const isMobileOpen = ref(false);
const sidebarRef = ref<HTMLElement | null>(null);
const toggleRef = ref<HTMLElement | null>(null);

const props = defineProps<{
    shop: RouterOutput['organization']['shop']['get'] | null
}>();

const detailNavigation = computed(() => [
    { 
        name: 'Shop Information', 
        route: 'account.shops.detail'
    },
    { 
        name: 'Checks', 
        route: 'account.shops.detail.checks',
        count: props.shop?.checks?.length ?? 0
    },
    { 
        name: 'Extensions', 
        route: 'account.shops.detail.extensions',
        count: props.shop?.extensions?.length ?? 0
    },
    { 
        name: 'Scheduled Tasks', 
        route: 'account.shops.detail.tasks',
        count: props.shop?.scheduledTask?.length ?? 0
    },
    { 
        name: 'Queue', 
        route: 'account.shops.detail.queue',
        count: props.shop?.queueInfo?.length ?? 0
    },
    { 
        name: 'Sitespeed', 
        route: 'account.shops.detail.sitespeed',
        count: props.shop?.sitespeed?.length ?? 0
    },
    { 
        name: 'Changelog', 
        route: 'account.shops.detail.changelog',
        count: props.shop?.changelog?.length ?? 0
    }
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
    document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside);
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
        content: '';
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

.shop-image-container {
    margin-top: 1.5rem;
    display: flex;
    justify-content: center;
    padding-bottom: 0.5rem;
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

.shop-image {
    height: 6.25rem;

    @media (min-width: 640px) {
        height: 25rem;
    }

    @media (min-width: 960px) {
        height: 12.5rem;
    }
}

.shop-name {
    padding: 0 0 1rem 0.75rem;
    margin-bottom: 0.5rem;
    border-bottom: 1px solid var(--panel-border-color);
}

.placeholder-image {
    color: #e5e7eb;
    font-size: 9rem;
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
