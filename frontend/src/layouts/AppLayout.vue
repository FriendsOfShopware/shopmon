<template>
  <SidebarProvider>
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton @click="$router.push({ name: 'account.dashboard' })">
              <icon-fa6-solid:magnifying-glass />
              <span>Quick search...</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Overview</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem v-for="item in navigation" :key="item.route">
                <SidebarMenuButton as-child :is-active="isActive(item, $route)">
                  <RouterLink :to="{ name: item.route }">
                    <component
                      :is="$router.resolve({ name: item.route }).meta.icon"
                      v-if="$router.resolve({ name: item.route }).meta.icon"
                    />
                    <span>{{ $t($router.resolve({ name: item.route }).meta.titleKey ?? "") }}</span>
                  </RouterLink>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <SidebarGroup v-if="environments.length > 0">
          <SidebarGroupLabel>Environments</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem v-for="env in environments" :key="env.id">
                <SidebarMenuButton as-child>
                  <RouterLink
                    :to="{
                      name: 'account.environments.detail',
                      params: { organizationId: env.organizationId, environmentId: env.id },
                    }"
                  >
                    <img v-if="env.favicon" :src="env.favicon" alt="" class="size-4 rounded" />
                    <icon-fa6-solid:earth-americas v-else class="opacity-40" />
                    <span>{{ env.name }}</span>
                  </RouterLink>
                </SidebarMenuButton>
                <SidebarMenuBadge><StatusIcon :status="env.status" /></SidebarMenuBadge>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as-child>
              <RouterLink :to="{ name: 'account.settings' }">
                <icon-fa6-solid:gear />
                <span>Manage account</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>

    <SidebarInset>
      <!-- Header -->
      <header class="flex h-12 items-center justify-between border-b bg-card px-4 lg:h-13 lg:px-5">
        <div class="flex items-center gap-2">
          <SidebarTrigger class="-ml-1" />
          <Separator orientation="vertical" class="mr-2 h-4" />
          <RouterLink :to="{ name: 'home' }" class="flex items-center">
            <Logo class="h-7 w-auto" />
          </RouterLink>
          <OrganizationSwitcher />
        </div>

        <div class="flex items-center gap-1">
          <a
            href="https://github.com/FriendsOfShopware/shopmon/"
            target="_blank"
            class="inline-flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            aria-label="GitHub"
          >
            <icon-fa-brands:github class="size-[1.125rem]" />
          </a>

          <Button variant="ghost" size="icon" class="size-8" @click="toggleDarkMode" aria-label="Toggle dark mode">
            <icon-fa6-regular:moon v-if="darkMode" class="size-[1.125rem]" />
            <icon-octicon:sun-16 v-else class="size-[1.125rem]" />
          </Button>

          <Button variant="ghost" size="sm" class="h-8 px-2 text-xs font-semibold tracking-wide" @click="toggleLocale">
            {{ String(locale) === "en" ? "DE" : "EN" }}
          </Button>

          <!-- Notifications -->
          <Popover>
            <PopoverTrigger as-child>
              <Button variant="ghost" size="icon" class="relative size-8" @click="markAllRead" aria-label="Notifications">
                <icon-fa6-solid:bell class="size-[1.125rem]" />
                <span v-if="unreadNotificationCount > 0" class="absolute right-0.5 top-0.5 flex size-3.5 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-white">
                  {{ unreadNotificationCount }}
                </span>
              </Button>
            </PopoverTrigger>
            <PopoverContent align="end" class="w-80 p-0">
              <div class="flex items-center justify-between border-b px-4 py-3 text-sm font-medium">
                {{ $t("topBar.notifications", { count: notifications.length }) }}
                <Button v-if="notifications.length > 0" variant="ghost" size="icon" class="size-6" @click="deleteAllNotifications">
                  <icon-fa6-solid:trash class="size-3" />
                </Button>
              </div>

              <ScrollArea v-if="notifications.length > 0" class="max-h-80">
                <div
                  v-for="(notification, index) in notifications"
                  :key="index"
                  class="group flex gap-2 border-b px-4 py-2.5 last:border-0 hover:bg-accent"
                >
                  <div class="mt-0.5 shrink-0">
                    <icon-fa6-solid:circle-xmark v-if="notification.level === 'error'" class="size-4 text-destructive" />
                    <icon-fa6-solid:circle-info v-else class="size-4 text-info" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium">{{ notification.title }}</div>
                    <div class="text-xs text-muted-foreground">{{ formatDateTime(notification.createdAt) }}</div>
                    <div class="text-[0.8125rem] leading-snug text-muted-foreground">
                      {{ notification.message }}
                      <a v-if="notification.link" :href="notification.link.url" class="text-primary">
                        {{ notification.link.label || $t("topBar.more") }}
                      </a>
                    </div>
                  </div>
                  <button type="button" class="invisible shrink-0 rounded p-0.5 text-muted-foreground opacity-50 hover:opacity-100 group-hover:visible" @click="deleteNotification(notification.id)">
                    <icon-fa6-solid:xmark class="size-3" />
                  </button>
                </div>
              </ScrollArea>
              <div v-else class="px-4 py-6 text-center text-sm text-muted-foreground">{{ $t("notifications.notMuchGoingOn") }}</div>
            </PopoverContent>
          </Popover>

          <!-- User menu -->
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon" class="hidden size-8 overflow-hidden rounded-full bg-primary lg:flex" aria-label="User menu">
                <img class="size-full rounded-full object-cover" :src="userAvatar" alt="" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-44">
              <DropdownMenuLabel>{{ $t("topBar.greeting", { name: session!.user.name }) }}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                v-for="item in userNavigation"
                :key="item.name"
                @click="item.route === 'logout' ? logout() : $router.push({ name: item.route })"
              >
                <component :is="item.icon" class="mr-2 size-4" />
                {{ item.name }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <!-- Main content -->
      <main class="flex flex-1 flex-col overflow-y-auto bg-background">
        <ImpersonationBanner />
        <Notification />

        <div class="flex-1 p-4 lg:p-6">
          <RouterView />
        </div>

        <LayoutFooter />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import type { RouteLocationNormalizedLoaded } from "vue-router";

import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Notification from "@/components/Notification.vue";
import LayoutFooter from "@/components/layout/LayoutFooter.vue";
import OrganizationSwitcher from "@/components/layout/OrganizationSwitcher.vue";
import StatusIcon from "@/components/StatusIcon.vue";

import { useDarkMode } from "@/composables/useDarkMode";
import { useLocale } from "@/composables/useLocale";
import { useNotifications } from "@/composables/useNotifications";
import { useSession, clearSession } from "@/composables/useSession";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";
import { api, setToken } from "@/helpers/api";
import { useI18n } from "vue-i18n";
import { formatDateTime } from "@/helpers/formatter";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarSeparator,
  SidebarTrigger,
} from "@/components/ui/sidebar";

import FaGear from "~icons/fa6-solid/gear";
import FaPowerOff from "~icons/fa6-solid/power-off";

const router = useRouter();
const { session } = useSession();
const { environments } = useAccountEnvironments();

const userAvatar = ref("https://api.dicebear.com/7.x/personas/svg?seed=default?d=identicon");

if (session.value?.user.email) {
  try {
    crypto.subtle
      .digest("SHA-256", new TextEncoder().encode(session.value.user.email))
      .then((hash) => {
        const seed = Array.from(new Uint8Array(hash))
          .map((b) => b.toString(16).padStart(2, "0"))
          .join("");
        userAvatar.value = `https://api.dicebear.com/7.x/personas/svg?seed=${seed}&d=identicon`;
      });
  } catch {
    // Silent fallback
  }
}

const {
  notifications,
  unreadNotificationCount,
  markAllRead,
  deleteAllNotifications,
  deleteNotification,
} = useNotifications();
const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
const { t } = useI18n();

const navigation = [
  { route: "account.dashboard" },
  { route: "account.shop.list", active: "shop" },
  { route: "account.extension.list" },
  { route: "account.organizations.list", active: "organizations" },
];

function isActive(item: { route: string; active?: string }, $route: RouteLocationNormalizedLoaded) {
  if (item.route === $route.name) return true;
  return !!(
    $route.name &&
    typeof $route.name === "string" &&
    item.active &&
    $route.name.match(item.active)
  );
}

const userNavigation = computed(() => [
  { name: t("nav.settings"), route: "account.settings", icon: FaGear },
  { name: t("nav.logout"), route: "logout", icon: FaPowerOff },
]);

async function logout() {
  await api.POST("/auth/sign-out");
  setToken(null);
  clearSession();
  router.push({ name: "home" });
}
</script>
