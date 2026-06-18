import { ref, computed, watch } from "vue";
import { useHead } from "@unhead/vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { compareVersions } from "compare-versions";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useAlert } from "@/composables/useAlert";

// Shared state across all callers. The composable is used by both the
// environment detail layout and the route child views; without a single shared
// instance each caller would fetch the same endpoints independently, firing
// every request (e.g. /info/shopware-versions) once per consumer.
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

// The environment id the shared state currently reflects, plus an in-flight
// promise so concurrent callers reuse the same load instead of duplicating it.
let loadedEnvironmentId: number | null = null;
let loadPromise: Promise<void> | null = null;

export function useEnvironmentDetail() {
  const route = useRoute();
  const { t } = useI18n();
  const { success, error } = useAlert();

  const environmentId = computed(() => Number.parseInt(route.params.environmentId as string, 10));
  const pageTitle = computed(() => {
    const titleKey = route.meta.titleKey;
    return typeof titleKey === "string" ? t(titleKey) : "";
  });

  useHead({
    title: computed(() => {
      if (!environment.value?.name) {
        return pageTitle.value || undefined;
      }

      return pageTitle.value
        ? `${pageTitle.value} - ${environment.value.name}`
        : environment.value.name;
    }),
  });

  async function fetchEnvironment(id: number) {
    isLoading.value = true;
    try {
      const { data } = await api.GET("/environments/{environmentId}", {
        params: { path: { environmentId: id } },
      });
      environment.value = data ?? null;

      // Check notification subscription status
      const { data: subData } = await api.GET("/environments/{environmentId}/subscribe", {
        params: { path: { environmentId: id } },
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
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isLoading.value = false;
    }
  }

  // Load the current environment, deduplicating concurrent callers. Passing
  // force re-fetches the same id (used after refresh/cache-clear actions).
  function loadEnvironment(force = false) {
    const id = environmentId.value;
    if (Number.isNaN(id)) return Promise.resolve();

    if (!force && loadedEnvironmentId === id && loadPromise) {
      return loadPromise;
    }

    loadedEnvironmentId = id;
    loadPromise = fetchEnvironment(id).finally(() => {
      loadPromise = null;
    });
    return loadPromise;
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
        await loadEnvironment(true);
        success(t("environment.refreshQueued"));
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
        await loadEnvironment(true);
        success(t("environment.cacheCleared"));
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
        success(t("environment.unsubscribed"));
      } else {
        await api.POST("/environments/{environmentId}/subscribe", {
          params: { path: { environmentId: environment.value.id } },
        });
        isSubscribed.value = true;
        success(t("environment.subscribed"));
      }
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isSubscribing.value = false;
    }
  }

  // Load on mount and reload whenever the environment ID changes. A single
  // immediate watch covers both cases; loadEnvironment() dedupes so multiple
  // consumers mounting together still trigger only one fetch per id.
  watch(
    () => route.params.environmentId,
    () => {
      void loadEnvironment();
    },
    { immediate: true },
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
