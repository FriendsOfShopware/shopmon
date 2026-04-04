<template>
  <SidebarProvider>
    <Sidebar collapsible="icon">
      <!-- Logo + brand -->
      <SidebarHeader class="border-b">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as-child class="h-10 hover:bg-transparent active:bg-transparent">
              <RouterLink :to="{ name: 'home' }" class="flex items-center gap-2.5">
                <Logo class="size-7 shrink-0" />
                <span class="text-base font-bold tracking-tight">Shopmon</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <!-- Org switcher -->
        <SidebarGroup class="py-2">
          <SidebarGroupContent>
            <OrganizationSwitcher />
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <!-- Main nav -->
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
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

        <!-- Environments -->
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

      <!-- User menu in sidebar footer -->
      <SidebarFooter class="border-t">
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <SidebarMenuButton class="h-10">
                  <img
                    class="size-6 shrink-0 rounded-full bg-primary object-cover"
                    :src="userAvatar"
                    alt=""
                  />
                  <div class="flex min-w-0 flex-1 flex-col leading-tight">
                    <span class="truncate text-sm font-medium">{{ session?.user.name }}</span>
                    <span class="truncate text-xs text-muted-foreground">{{
                      session?.user.email
                    }}</span>
                  </div>
                  <icon-fa6-solid:ellipsis-vertical class="ml-auto size-3 text-muted-foreground" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" class="w-56">
                <DropdownMenuLabel class="font-normal">
                  <div class="flex flex-col gap-0.5">
                    <span class="text-sm font-medium">{{ session?.user.name }}</span>
                    <span class="text-xs text-muted-foreground">{{ session?.user.email }}</span>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem @click="$router.push({ name: 'account.settings' })">
                  <icon-fa6-solid:gear class="mr-2 size-3.5" />
                  {{ $t("nav.settings") }}
                </DropdownMenuItem>
                <DropdownMenuItem @click="toggleDarkMode">
                  <icon-fa6-regular:moon v-if="darkMode" class="mr-2 size-3.5" />
                  <icon-octicon:sun-16 v-else class="mr-2 size-3.5" />
                  {{ darkMode ? "Light mode" : "Dark mode" }}
                </DropdownMenuItem>
                <DropdownMenuItem @click="toggleLocale">
                  <icon-fa6-solid:globe class="mr-2 size-3.5" />
                  {{ String(locale) === "en" ? "Deutsch" : "English" }}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  as="a"
                  href="https://github.com/FriendsOfShopware/shopmon/"
                  target="_blank"
                >
                  <icon-fa-brands:github class="mr-2 size-3.5" />
                  GitHub
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem class="text-destructive focus:text-destructive" @click="logout">
                  <icon-fa6-solid:power-off class="mr-2 size-3.5" />
                  {{ $t("nav.logout") }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>

    <SidebarInset>
      <!-- Top bar — minimal: trigger + breadcrumb area + notifications -->
      <header
        class="sticky top-0 z-10 flex h-12 items-center justify-between border-b bg-background px-4"
      >
        <div class="flex items-center gap-2">
          <SidebarTrigger class="-ml-1" />
        </div>

        <div class="flex items-center gap-1">
          <!-- Notifications -->
          <Popover>
            <PopoverTrigger as-child>
              <Button
                variant="ghost"
                size="icon"
                class="relative size-8"
                @click="markAllRead"
                aria-label="Notifications"
              >
                <icon-fa6-solid:bell class="size-4" />
                <span
                  v-if="unreadNotificationCount > 0"
                  class="absolute right-0.5 top-0.5 flex size-3.5 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-white"
                >
                  {{ unreadNotificationCount }}
                </span>
              </Button>
            </PopoverTrigger>
            <PopoverContent align="end" class="w-80 max-w-[calc(100vw-2rem)] p-0">
              <div class="flex items-center justify-between border-b px-4 py-3 text-sm font-medium">
                {{ $t("topBar.notifications", { count: notifications.length }) }}
                <Button
                  v-if="notifications.length > 0"
                  variant="ghost"
                  size="icon"
                  class="size-6"
                  @click="deleteAllNotifications"
                >
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
                    <icon-fa6-solid:circle-xmark
                      v-if="notification.level === 'error'"
                      class="size-4 text-destructive"
                    />
                    <icon-fa6-solid:circle-info v-else class="size-4 text-info" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium">{{ notification.title }}</div>
                    <div class="text-xs text-muted-foreground">
                      {{ formatDateTime(notification.createdAt) }}
                    </div>
                    <div class="text-[0.8125rem] leading-snug text-muted-foreground">
                      {{ notification.message }}
                      <a
                        v-if="notification.link"
                        :href="notification.link.url"
                        class="text-primary"
                      >
                        {{ notification.link.label || $t("topBar.more") }}
                      </a>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="invisible shrink-0 rounded p-0.5 text-muted-foreground opacity-50 hover:opacity-100 group-hover:visible"
                    @click="deleteNotification(notification.id)"
                  >
                    <icon-fa6-solid:xmark class="size-3" />
                  </button>
                </div>
              </ScrollArea>
              <div v-else class="px-4 py-6 text-center text-sm text-muted-foreground">
                {{ $t("notifications.notMuchGoingOn") }}
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </header>

      <!-- Main content -->
      <main class="flex flex-1 flex-col bg-background">
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
import { ref } from "vue";
import { useRouter } from "vue-router";
import type { RouteLocationNormalizedLoaded } from "vue-router";

import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Notification from "@/components/Notification.vue";
import LayoutFooter from "@/components/layout/LayoutFooter.vue";
import OrganizationSwitcher from "@/components/layout/OrganizationSwitcher.vue";
import StatusIcon from "@/components/StatusIcon.vue";
import Logo from "@/components/Logo.vue";

import { useDarkMode } from "@/composables/useDarkMode";
import { useLocale } from "@/composables/useLocale";
import { useNotifications } from "@/composables/useNotifications";
import { useSession, clearSession } from "@/composables/useSession";
import { useAccountEnvironments } from "@/composables/useAccountEnvironments";
import { api, setToken } from "@/helpers/api";
import { formatDateTime } from "@/helpers/formatter";

import { Button } from "@/components/ui/button";
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

async function logout() {
  await api.POST("/auth/sign-out");
  setToken(null);
  clearSession();
  router.push({ name: "home" });
}
</script>
