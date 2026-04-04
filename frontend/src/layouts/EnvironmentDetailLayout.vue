<template>
  <div v-if="environment" class="space-y-4">
    <!-- Breadcrumb -->
    <Breadcrumbs :items="breadcrumbItems" />

    <!-- Environment header strip -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center gap-3">
        <div
          class="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border bg-muted"
        >
          <img
            v-if="environment.environmentImage"
            :src="environment.environmentImage"
            class="size-full object-cover"
            alt=""
          />
          <icon-fa6-solid:image v-else class="size-4 text-muted-foreground" />
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <!-- Environment switcher -->
            <DropdownMenu v-if="siblingEnvironments.length > 1">
              <DropdownMenuTrigger
                class="group flex items-center gap-1.5 rounded-md px-1 -ml-1 hover:bg-accent transition-colors"
              >
                <h1 class="truncate text-xl font-bold">{{ environment.name }}</h1>
                <StatusIcon :status="environment.status" />
                <icon-fa6-solid:chevron-down
                  class="size-2.5 text-muted-foreground transition-transform group-data-[state=open]:rotate-180"
                />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" class="w-64">
                <DropdownMenuLabel>Switch environment</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  v-for="env in siblingEnvironments"
                  :key="env.id"
                  :class="{ 'bg-accent': env.id === environment.id }"
                  @click="
                    $router.push({
                      name: currentTabRoute,
                      params: { organizationId: env.organizationId, environmentId: env.id },
                    })
                  "
                >
                  <div class="flex items-center gap-2.5 w-full">
                    <img
                      v-if="env.favicon"
                      :src="env.favicon"
                      alt=""
                      class="size-4 shrink-0 rounded"
                    />
                    <icon-fa6-solid:earth-americas
                      v-else
                      class="size-3.5 shrink-0 text-muted-foreground/50"
                    />
                    <span class="flex-1 truncate">{{ env.name }}</span>
                    <StatusIcon :status="env.status" />
                  </div>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <!-- No switcher if only 1 environment -->
            <template v-else>
              <h1 class="truncate text-xl font-bold">{{ environment.name }}</h1>
              <StatusIcon :status="environment.status" />
            </template>
          </div>
          <div class="flex items-center gap-2 text-sm text-muted-foreground">
            <span class="truncate">{{ environment.organizationName }}</span>
            <template v-if="environmentHost">
              <span class="text-border">/</span>
              <a
                :href="environment.url"
                target="_blank"
                rel="noopener noreferrer"
                class="truncate hover:text-primary"
                >{{ environmentHost }}</a
              >
            </template>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" class="max-sm:hidden" as-child>
          <a :href="environment.url" target="_blank" rel="noopener noreferrer">
            <icon-fa6-solid:store class="mr-1.5 size-3" />
            Storefront
          </a>
        </Button>
        <Button variant="outline" size="sm" class="max-sm:hidden" as-child>
          <a :href="(environment.url ?? '') + '/admin'" target="_blank" rel="noopener noreferrer">
            <icon-fa6-solid:user-gear class="mr-1.5 size-3" />
            Admin
          </a>
        </Button>

        <Separator orientation="vertical" class="mx-1 h-6 max-sm:hidden" />

        <Button
          variant="ghost"
          size="icon"
          class="size-8"
          :title="isSubscribed ? 'Unwatch environment' : 'Watch environment'"
          :disabled="isSubscribing"
          @click="toggleNotificationSubscription"
        >
          <icon-fa6-solid:bell
            v-if="isSubscribed"
            :class="['size-3.5', { 'animate-pulse': isSubscribing }]"
          />
          <icon-fa6-regular:bell v-else :class="['size-3.5', { 'animate-pulse': isSubscribing }]" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          class="size-8"
          title="Clear environment cache"
          :disabled="isCacheClearing"
          @click="onCacheClear"
        >
          <icon-ic:baseline-cleaning-services
            :class="['size-3.5', { 'animate-pulse': isCacheClearing }]"
          />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          class="size-8"
          title="Refresh environment data"
          :disabled="isRefreshing"
          @click="showEnvironmentRefreshModal = true"
        >
          <icon-fa6-solid:rotate :class="['size-3.5', { 'animate-spin': isRefreshing }]" />
        </Button>

        <Button as-child size="sm">
          <RouterLink
            :to="{
              name: 'account.environments.edit',
              params: {
                organizationId: route.params.organizationId,
                environmentId: environment.id,
              },
            }"
          >
            <icon-fa6-solid:pencil class="mr-1.5 size-3" />
            Edit
          </RouterLink>
        </Button>
      </div>
    </div>

    <!-- Error banner -->
    <Alert
      v-if="environment.lastScrapedError"
      variant="destructive"
      class="border-destructive/30 bg-destructive/10"
    >
      <CircleX class="size-4" />
      <AlertDescription>
        This environment will be not automatically updated anymore. Please update the API
        credentials or URL to fix this issue.
      </AlertDescription>
    </Alert>

    <!-- Tab navigation -->
    <nav class="flex gap-1 overflow-x-auto border-b" v-if="environment.lastScrapedAt">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.route"
        :to="{
          name: tab.route,
          params: {
            organizationId: route.params.organizationId,
            environmentId: route.params.environmentId,
          },
        }"
        :class="[
          'inline-flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-2 text-sm font-medium transition-colors whitespace-nowrap',
          isTabActive(tab.route)
            ? 'border-primary text-primary'
            : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground',
        ]"
        active-class=""
        exact-active-class=""
      >
        <component :is="tab.icon" class="size-3.5" />
        {{ tab.label }}
        <Badge v-if="tab.count" variant="secondary" class="ml-0.5 h-5 min-w-5 px-1 text-[10px]">{{
          tab.count
        }}</Badge>
      </RouterLink>
    </nav>

    <!-- Page content -->
    <div v-if="environment.lastScrapedAt">
      <RouterView />
    </div>

    <!-- Refresh modal -->
    <Dialog
      :open="showEnvironmentRefreshModal"
      @update:open="(v: boolean) => !v && (showEnvironmentRefreshModal = false)"
    >
      <DialogContent class="max-w-md">
        <DialogHeader>
          <div class="flex items-start gap-3">
            <FaRotate class="mt-0.5 size-5 shrink-0 text-info" aria-hidden="true" />
            <DialogTitle>Refresh {{ environment.name }}</DialogTitle>
          </div>
        </DialogHeader>
        <p>Do you also want to have a new pagespeed test?</p>
        <DialogFooter>
          <Button variant="destructive" @click="onRefresh(false)">No</Button>
          <Button @click="onRefresh(true)">Yes</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";
import Breadcrumbs from "@/components/layout/Breadcrumbs.vue";
import StatusIcon from "@/components/StatusIcon.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CircleX } from "lucide-vue-next";
import type { BreadcrumbItem } from "@/components/layout/breadcrumbs";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import FaRotate from "~icons/fa6-solid/rotate";

import FaShop from "~icons/fa6-solid/shop";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaPlug from "~icons/fa6-solid/plug";
import FaListCheck from "~icons/fa6-solid/list-check";
import FaRocket from "~icons/fa6-solid/rocket";
import FaFileWaveform from "~icons/fa6-solid/file-waveform";
import FaCodeBranch from "~icons/fa6-solid/code-branch";

const route = useRoute();
const { t } = useI18n();

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

const { environments: allEnvironments } = useAccountEnvironments();

// Environments in the same project (same shopId)
const siblingEnvironments = computed(() => {
  if (!environment.value?.shopId) return [];
  return allEnvironments.value.filter((e) => e.shopId === environment.value!.shopId);
});

// Keep the user on the same tab when switching environments
const currentTabRoute = computed(() => {
  const name = route.name as string;
  // If on a sub-detail page (e.g. deployment/:id), go to the parent list
  if (name === "account.environments.detail.deployment")
    return "account.environments.detail.deployments";
  // If it's a known tab route, keep it
  if (name.startsWith("account.environments.detail")) return name;
  return "account.environments.detail";
});

const environmentHost = computed(() => {
  if (!environment.value?.url) return "";
  try {
    return new URL(environment.value.url).host;
  } catch {
    return environment.value.url;
  }
});

const tabs = computed(() => [
  {
    route: "account.environments.detail",
    label: t("nav.environmentInformation"),
    icon: FaShop,
    count: 0,
  },
  {
    route: "account.environments.detail.checks",
    label: t("nav.checks"),
    icon: FaCircleCheck,
    count: environment.value?.checks?.length ?? 0,
  },
  {
    route: "account.environments.detail.extensions",
    label: t("nav.extensions"),
    icon: FaPlug,
    count: environment.value?.extensions?.length ?? 0,
  },
  {
    route: "account.environments.detail.tasks",
    label: t("nav.scheduledTasks"),
    icon: FaListCheck,
    count: environment.value?.scheduledTasks?.length ?? 0,
  },
  {
    route: "account.environments.detail.queue",
    label: t("nav.queue"),
    icon: FaCircleCheck,
    count: environment.value?.queues?.length ?? 0,
  },
  {
    route: "account.environments.detail.sitespeed",
    label: t("nav.sitespeed"),
    icon: FaRocket,
    count: environment.value?.sitespeeds?.length ?? 0,
  },
  {
    route: "account.environments.detail.changelog",
    label: t("nav.changelog"),
    icon: FaFileWaveform,
    count: environment.value?.changelogs?.length ?? 0,
  },
  {
    route: "account.environments.detail.deployments",
    label: t("nav.deployments"),
    icon: FaCodeBranch,
    count: environment.value?.deploymentsCount ?? 0,
  },
]);

function isTabActive(tabRoute: string): boolean {
  if (route.name === tabRoute) return true;
  // Deployment detail page highlights the deployments tab
  if (
    tabRoute === "account.environments.detail.deployments" &&
    route.name === "account.environments.detail.deployment"
  )
    return true;
  return false;
}

const breadcrumbItems = computed<BreadcrumbItem[]>(() => {
  if (!environment.value) return [];

  const currentRouteTitle = (() => {
    const titleKey = route.meta.titleKey;
    return typeof titleKey === "string" ? t(titleKey) : "";
  })();

  const items: BreadcrumbItem[] = [
    {
      label: environment.value.organizationName,
      to: {
        name: "account.organizations.detail",
        params: { organizationId: environment.value.organizationId },
      },
    },
  ];

  const isOverview = route.name === "account.environments.detail";

  items.push({
    label: environment.value.name,
    to: isOverview
      ? undefined
      : {
          name: "account.environments.detail",
          params: {
            organizationId: route.params.organizationId,
            environmentId: route.params.environmentId,
          },
        },
  });

  if (!isOverview && currentRouteTitle) {
    items.push({ label: currentRouteTitle });
  }

  return items;
});
</script>
