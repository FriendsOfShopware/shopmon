<template>
  <div class="detail-wrapper">
    <sidebar-detail :environment="environment" />

    <header-container
      v-if="environment"
      :breadcrumb="breadcrumbItems"
      :title="environment.name"
      :titleMobileHide="true"
    >
      <slot name="header-actions">
        <UiButton
          icon-only
          :data-tooltip="isSubscribed ? 'Unwatch environment' : 'Watch environment'"
          :disabled="isSubscribing"
          type="button"
          @click="toggleNotificationSubscription"
        >
          <icon-fa6-solid:bell
            v-if="isSubscribed"
            :class="{ 'animate-pulse': isSubscribing }"
            class="icon"
          />
          <icon-fa6-regular:bell v-else :class="{ 'animate-pulse': isSubscribing }" class="icon" />
        </UiButton>

        <UiButton
          icon-only
          data-tooltip="Clear environment cache"
          :disabled="isCacheClearing"
          type="button"
          @click="onCacheClear"
        >
          <icon-ic:baseline-cleaning-services
            :class="{ 'animate-pulse': isCacheClearing }"
            class="icon"
          />
        </UiButton>

        <UiButton
          icon-only
          data-tooltip="Refresh environment data"
          :disabled="isRefreshing"
          type="button"
          @click="showEnvironmentRefreshModal = true"
        >
          <icon-fa6-solid:rotate :class="{ icon: true, 'animate-spin': isRefreshing }" />
        </UiButton>

        <UiButton
          :to="{
            name: 'account.environments.edit',
            params: {
              organizationId: route.params.organizationId,
              environmentId: environment.id,
            },
          }"
          type="button"
          variant="primary"
        >
          <icon-fa6-solid:pencil class="icon" aria-hidden="true" />
          Edit Environment
        </UiButton>
      </slot>
    </header-container>

    <main-container v-if="environment && environment.lastScrapedAt">
      <Banner v-if="environment.lastScrapedError" class="environment-scrape-error" variant="error">
        This environment will be not automatically updated anymore. Please update the API
        credentials or URL to fix this issue.
      </Banner>

      <router-view />

      <modal
        :show="showEnvironmentRefreshModal"
        close-x-mark
        @close="showEnvironmentRefreshModal = false"
      >
        <template #icon>
          <FaRotate class="icon icon-info" aria-hidden="true" />
        </template>

        <template #title> Refresh {{ environment.name }} </template>

        <template #content> Do you also want to have a new pagespeed test? </template>

        <template #footer>
          <UiButton type="button" variant="primary" @click="onRefresh(true)">Yes</UiButton>

          <UiButton
            ref="cancelButtonRef"
            type="button"
            variant="destructive"
            @click="onRefresh(false)"
          >
            No
          </UiButton>
        </template>
      </modal>
    </main-container>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import SidebarDetail from "@/components/layout/SidebarDetail.vue";
import HeaderContainer from "@/components/layout/HeaderContainer.vue";
import MainContainer from "@/components/layout/MainContainer.vue";
import Banner from "@/components/layout/Banner.vue";
import Modal from "@/components/layout/Modal.vue";
import type { BreadcrumbItem } from "@/components/layout/breadcrumbs";
import FaRotate from "~icons/fa6-solid/rotate";

const route = useRoute();
const { t } = useI18n();

const currentRouteTitle = computed(() => {
  const titleKey = route.meta.titleKey;
  return typeof titleKey === "string" ? t(titleKey) : "";
});

const {
  environment,
  isRefreshing,
  isCacheClearing,
  isSubscribed,
  isSubscribing,
  showEnvironmentRefreshModal,
  onRefresh,
  onCacheClear,
  toggleNotificationSubscription,
} = useEnvironmentDetail();

const breadcrumbItems = computed<BreadcrumbItem[]>(() => {
  if (!environment.value) {
    return [];
  }

  const items: BreadcrumbItem[] = [
    {
      label: environment.value.organizationName,
      to: {
        name: "account.organizations.detail",
        params: { organizationId: environment.value.organizationId },
      },
    },
  ];

  const isEnvironmentOverview = route.name === "account.environments.detail";

  items.push({
    label: environment.value.name,
    to: isEnvironmentOverview
      ? undefined
      : {
          name: "account.environments.detail",
          params: {
            organizationId: route.params.organizationId,
            environmentId: route.params.environmentId,
          },
        },
  });

  if (!isEnvironmentOverview && currentRouteTitle.value) {
    items.push({
      label: currentRouteTitle.value,
    });
  }

  return items;
});
</script>

<style scoped>
.detail-wrapper {
  @media all and (min-width: 768px) {
    display: grid;
    column-gap: 1rem;
    grid-template-areas:
      "sidebar header"
      "sidebar content";
    grid-template-columns: 16rem 1fr;
  }
}

.environment-scrape-error {
  margin-bottom: 1rem;
}

.main-container {
  min-height: 80vh;
}
</style>
