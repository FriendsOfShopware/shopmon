<template>
    <div class="detail-wrapper">
        <sidebar-detail :shop="shop" />

        <header-container
            v-if="shop"
            :breadcrumb="shop.nameCombined + (route.meta.title && route.name !== 'account.shops.detail' ? ' / ' + route.meta.title : '')"
            :title="shop.name"
            :titleMobileHide="true"
        >
            <slot name="header-actions">
                <button
                    class="btn icon-only"
                    :data-tooltip="isSubscribed ? 'Unwatch shop' : 'Watch shop'"
                    :disabled="isSubscribing"
                    type="button"
                    @click="toggleNotificationSubscription"
                >
                    <icon-fa6-solid:bell
                        v-if="isSubscribed"
                        :class="{ 'animate-pulse': isSubscribing }"
                        class="icon"
                    />
                    <icon-fa6-regular:bell
                        v-else
                        :class="{ 'animate-pulse': isSubscribing }"
                        class="icon"
                    />
                </button>

                <button
                    class="btn icon-only"
                    data-tooltip="Clear shop cache"
                    :disabled="isCacheClearing"
                    type="button"
                    @click="onCacheClear"
                >
                    <icon-ic:baseline-cleaning-services
                        :class="{ 'animate-pulse': isCacheClearing }"
                        class="icon"
                    />
                </button>

                <button
                    class="btn icon-only"
                    data-tooltip="Refresh shop data"
                    :disabled="isRefreshing"
                    type="button"
                    @click="showShopRefreshModal = true"
                >
                    <icon-fa6-solid:rotate :class="{ 'icon': true, 'animate-spin': isRefreshing }" />
                </button>

                <router-link
                    :to="{
                        name: 'account.shops.edit',
                        params: {
                            slug: route.params.slug,
                            shopId: shop.id
                        }
                    }"
                    type="button"
                    class="btn btn-primary"
                >
                    <icon-fa6-solid:pencil class="icon" aria-hidden="true" />
                    Edit Shop
                </router-link>
            </slot>
        </header-container>

        <main-container v-if="shop && shop.lastScrapedAt">
            <alert
                v-if="shop.connectionIssueCount >= 3"
                class="shop-scrape-error"
                type="error"
            >
                This shop will be not automatically updated anymore. Please update the API credentials or Shop URL to fix this issue.
            </alert>

            <router-view />

            <modal
                :show="showShopRefreshModal"
                close-x-mark
                @close="showShopRefreshModal = false"
            >
                <template #icon>
                    <FaRotate
                        class="icon icon-info"
                        aria-hidden="true"
                    />
                </template>

                <template #title>
                    Refresh {{ shop.name }}
                </template>

                <template #content>
                    Do you also want to have a new pagespeed test?
                </template>

                <template #footer>
                    <button
                        type="button"
                        class="btn btn-primary"
                        @click="onRefresh(true)"
                    >
                        Yes
                    </button>

                    <button
                        ref="cancelButtonRef"
                        type="button"
                        class="btn btn-danger"
                        @click="onRefresh(false)"
                    >
                        No
                    </button>
                </template>
            </modal>
        </main-container>
    </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { useShopDetail } from '@/composables/useShopDetail';
import SidebarDetail from '@/components/layout/SidebarDetail.vue';
import HeaderContainer from '@/components/layout/HeaderContainer.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import Alert from '@/components/layout/Alert.vue';
import Modal from '@/components/layout/Modal.vue';
import FaRotate from '~icons/fa6-solid/rotate';

const route = useRoute();

const {
    shop,
    isRefreshing,
    isCacheClearing,
    isSubscribed,
    isSubscribing,
    showShopRefreshModal,
    onRefresh,
    onCacheClear,
    toggleNotificationSubscription,
} = useShopDetail();
</script>

<style scoped>
.detail-wrapper {
    @media all and (min-width: 1024px) {
        display: grid;
        column-gap: 1rem;
        grid-template-areas:
        "sidebar header"
        "sidebar content";
        grid-template-columns: 250px 1fr;
    }
}

.shop-scrape-error {
    margin-bottom: 1rem;
}

.main-container {
    min-height: 80vh;
}
</style>
