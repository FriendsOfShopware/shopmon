<template xmlns="http://www.w3.org/1999/html">
    <div class="sidebar-wrapper">

        <div class="detail-sidebar sidebar" v-if="props.shop">
            <div class="shop-image-container">
                <img
                    v-if="shop.shopImage"
                    :src="`/${shop.shopImage}`"
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
                    <span class="nav-link-name">{{ $router.resolve({name: item.route}).meta.title || item.name }}</span>
                    <span v-if="item.count !== undefined" class="nav-link-count">{{ item.count }}</span>
                </router-link>
            </nav>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { RouterOutput } from '@/helpers/trpc';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

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
        name: 'Pagespeed', 
        route: 'account.shops.detail.pagespeed',
        count: props.shop?.pageSpeed?.length ?? 0
    },
    { 
        name: 'Changelog', 
        route: 'account.shops.detail.changelog',
        count: props.shop?.changelog?.length ?? 0
    }
]);
</script>

<style scoped>
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
