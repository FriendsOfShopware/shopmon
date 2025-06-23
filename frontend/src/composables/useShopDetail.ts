import {ref, computed, onMounted, watch} from 'vue';
import {useRoute} from 'vue-router';
import {compareVersions} from 'compare-versions';
import {trpcClient} from '@/helpers/trpc';
import type {RouterOutput} from '@/helpers/trpc';
import {useAlert} from '@/composables/useAlert';

export function useShopDetail() {
    const route = useRoute();
    const {success, error} = useAlert();

    const shop = ref<RouterOutput['organization']['shop']['get'] | null>(null);
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const isCacheClearing = ref(false);
    const isSubscribed = ref(false);
    const isSubscribing = ref(false);
    const showShopRefreshModal = ref(false);
    const shopwareVersions = ref<string[] | null>(null);
    const latestShopwareVersion = ref<string | null>(null);

    // shopwareVersions and latestShopwareVersion are still needed for the update wizard in DetailShop.vue

    const shopId = computed(() => Number.parseInt(route.params.shopId as string, 10));

    async function loadShop() {
        isLoading.value = true;
        try {
            shop.value = await trpcClient.organization.shop.get.query({
                shopId: shopId.value,
            });

            // Check notification subscription status
            isSubscribed.value = await trpcClient.organization.shop.isSubscribedToNotifications.query({
                shopId: shopId.value,
            });

            const shopwareVersionsData = await trpcClient.info.getLatestShopwareVersion.query();
            shopwareVersions.value = Object.keys(shopwareVersionsData)
                .reverse()
                .filter(
                    (version) =>
                        !version.includes('-RC') &&
                        compareVersions(shop.value?.shopwareVersion ?? '', version) < 0,
                );
            latestShopwareVersion.value = shopwareVersions.value[0];

            if (shop.value?.name) {
                document.title = shop.value.name;
            }
        } catch (e) {
            error(e instanceof Error ? e.message : String(e));
        } finally {
            isLoading.value = false;
        }
    }

    async function onRefresh(sitespeed: boolean) {
        showShopRefreshModal.value = false;
        if (shop.value?.organizationId && shop.value?.id) {
            try {
                isRefreshing.value = true;
                await trpcClient.organization.shop.refreshShop.mutate({
                    shopId: shop.value.id,
                    sitespeed: sitespeed,
                });
                isRefreshing.value = false;
                await loadShop();
                success('Your Shop will refresh soon!');
            } catch (e) {
                isRefreshing.value = false;
                error(e instanceof Error ? e.message : String(e));
            }
        }
    }

    async function onCacheClear() {
        if (shop.value?.organizationId && shop.value?.id) {
            try {
                isCacheClearing.value = true;
                await trpcClient.organization.shop.clearShopCache.mutate({
                    shopId: shop.value.id,
                });
                isCacheClearing.value = false;
                await loadShop();
                success('Your Shop cache was cleared successfully');
            } catch (e) {
                isCacheClearing.value = false;
                error(e instanceof Error ? e.message : String(e));
            }
        }
    }

    async function toggleNotificationSubscription() {
        if (!shop.value) return;

        try {
            isSubscribing.value = true;

            if (isSubscribed.value) {
                await trpcClient.organization.shop.unsubscribeFromNotifications.mutate({
                    shopId: shop.value.id,
                });
                isSubscribed.value = false;
                success('You have unsubscribed from notifications for this shop');
            } else {
                await trpcClient.organization.shop.subscribeToNotifications.mutate({
                    shopId: shop.value.id,
                });
                isSubscribed.value = true;
                success('You have subscribed to notifications for this shop');
            }
        } catch (e) {
            error(e instanceof Error ? e.message : String(e));
        } finally {
            isSubscribing.value = false;
        }
    }

    onMounted(() => {
        loadShop();
    });

    // Watch for changes to the shop ID and reload the shop data when it changes
    watch(() => route.params.shopId, () => {
        loadShop();
    });

    return {
        shop,
        isLoading,
        isRefreshing,
        isCacheClearing,
        isSubscribed,
        isSubscribing,
        showShopRefreshModal,
        shopwareVersions,
        latestShopwareVersion,
        loadShop,
        onRefresh,
        onCacheClear,
        toggleNotificationSubscription,
    };
}
