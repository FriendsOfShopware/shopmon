import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { compareVersions } from "compare-versions";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useAlert } from "@/composables/useAlert";

export function useShopDetail() {
  const route = useRoute();
  const { success, error } = useAlert();

  const shop = ref<components["schemas"]["ShopDetail"] | null>(null);
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
      const { data } = await api.GET("/shops/{shopId}", {
        params: { path: { shopId: shopId.value } },
      });
      shop.value = data ?? null;

      // Check notification subscription status
      const { data: subData } = await api.GET("/shops/{shopId}/subscribe", {
        params: { path: { shopId: shopId.value } },
      });
      isSubscribed.value = subData?.subscribed ?? false;

      const { data: shopwareVersionsData } = await api.GET("/info/shopware-versions");
      if (shopwareVersionsData) {
        shopwareVersions.value = Object.keys(shopwareVersionsData)
          .reverse()
          .filter(
            (version) =>
              !version.includes("-RC") &&
              compareVersions(shop.value?.shopwareVersion ?? "", version) < 0,
          );
        latestShopwareVersion.value = shopwareVersions.value[0];
      }

      if (shop.value?.name) {
        const pageTitle = route.meta.title;
        document.title =
          typeof pageTitle === "string"
            ? `${pageTitle} - ${shop.value.name} | Shopmon`
            : `${shop.value.name} | Shopmon`;
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
        await api.POST("/shops/{shopId}/refresh", {
          params: { path: { shopId: shop.value.id } },
          body: { sitespeed },
        });
        isRefreshing.value = false;
        await loadShop();
        success("Your Shop will refresh soon!");
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
        await api.POST("/shops/{shopId}/clear-cache", {
          params: { path: { shopId: shop.value.id } },
        });
        isCacheClearing.value = false;
        await loadShop();
        success("Your Shop cache was cleared successfully");
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
        await api.DELETE("/shops/{shopId}/subscribe", {
          params: { path: { shopId: shop.value.id } },
        });
        isSubscribed.value = false;
        success("You have unsubscribed from notifications for this shop");
      } else {
        await api.POST("/shops/{shopId}/subscribe", {
          params: { path: { shopId: shop.value.id } },
        });
        isSubscribed.value = true;
        success("You have subscribed to notifications for this shop");
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
  watch(
    () => route.params.shopId,
    () => {
      loadShop();
    },
  );

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
