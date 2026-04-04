import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { compareVersions } from "compare-versions";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useAlert } from "@/composables/useAlert";

export function useEnvironmentDetail() {
  const route = useRoute();
  const { success, error } = useAlert();

  const environment = ref<components["schemas"]["EnvironmentDetail"] | null>(null);
  const isLoading = ref(false);
  const isRefreshing = ref(false);
  const isCacheClearing = ref(false);
  const isSubscribed = ref(false);
  const isSubscribing = ref(false);
  const showEnvironmentRefreshModal = ref(false);
  const shopwareVersions = ref<string[] | null>(null);
  const latestShopwareVersion = ref<string | null>(null);

  // shopwareVersions and latestShopwareVersion are still needed for the update wizard in DetailEnvironment.vue

  const environmentId = computed(() => Number.parseInt(route.params.environmentId as string, 10));

  async function loadEnvironment() {
    isLoading.value = true;
    try {
      const { data } = await api.GET("/environments/{environmentId}", {
        params: { path: { environmentId: environmentId.value } },
      });
      environment.value = data ?? null;

      // Check notification subscription status
      const { data: subData } = await api.GET("/environments/{environmentId}/subscribe", {
        params: { path: { environmentId: environmentId.value } },
      });
      isSubscribed.value = subData?.subscribed ?? false;

      const { data: shopwareVersionsData } = await api.GET("/info/shopware-versions");
      if (shopwareVersionsData) {
        shopwareVersions.value = Object.keys(shopwareVersionsData)
          .reverse()
          .filter(
            (version) =>
              !version.includes("-RC") &&
              compareVersions(environment.value?.shopwareVersion ?? "", version) < 0,
          );
        latestShopwareVersion.value = shopwareVersions.value[0];
      }

      if (environment.value?.name) {
        const pageTitle = route.meta.title;
        document.title =
          typeof pageTitle === "string"
            ? `${pageTitle} - ${environment.value.name} | Shopmon`
            : `${environment.value.name} | Shopmon`;
      }
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isLoading.value = false;
    }
  }

  async function onRefresh(sitespeed: boolean) {
    showEnvironmentRefreshModal.value = false;
    if (environment.value?.organizationId && environment.value?.id) {
      try {
        isRefreshing.value = true;
        await api.POST("/environments/{environmentId}/refresh", {
          params: { path: { environmentId: environment.value.id } },
          body: { sitespeed },
        });
        isRefreshing.value = false;
        await loadEnvironment();
        success("Your Environment will refresh soon!");
      } catch (e) {
        isRefreshing.value = false;
        error(e instanceof Error ? e.message : String(e));
      }
    }
  }

  async function onCacheClear() {
    if (environment.value?.organizationId && environment.value?.id) {
      try {
        isCacheClearing.value = true;
        await api.POST("/environments/{environmentId}/clear-cache", {
          params: { path: { environmentId: environment.value.id } },
        });
        isCacheClearing.value = false;
        await loadEnvironment();
        success("Your Environment cache was cleared successfully");
      } catch (e) {
        isCacheClearing.value = false;
        error(e instanceof Error ? e.message : String(e));
      }
    }
  }

  async function toggleNotificationSubscription() {
    if (!environment.value) return;

    try {
      isSubscribing.value = true;

      if (isSubscribed.value) {
        await api.DELETE("/environments/{environmentId}/subscribe", {
          params: { path: { environmentId: environment.value.id } },
        });
        isSubscribed.value = false;
        success("You have unsubscribed from notifications for this environment");
      } else {
        await api.POST("/environments/{environmentId}/subscribe", {
          params: { path: { environmentId: environment.value.id } },
        });
        isSubscribed.value = true;
        success("You have subscribed to notifications for this environment");
      }
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isSubscribing.value = false;
    }
  }

  onMounted(() => {
    loadEnvironment();
  });

  // Watch for changes to the environment ID and reload the environment data when it changes
  watch(
    () => route.params.environmentId,
    () => {
      loadEnvironment();
    },
  );

  return {
    environment,
    isLoading,
    isRefreshing,
    isCacheClearing,
    isSubscribed,
    isSubscribing,
    showEnvironmentRefreshModal,
    shopwareVersions,
    latestShopwareVersion,
    loadEnvironment,
    onRefresh,
    onCacheClear,
    toggleNotificationSubscription,
  };
}
