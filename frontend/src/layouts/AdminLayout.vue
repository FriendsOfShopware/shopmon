<template>
  <div class="min-h-screen bg-background">
    <nav class="sticky top-0 z-20 border-b bg-background/80 backdrop-blur-md">
      <div class="container mx-auto flex items-center justify-between gap-4 px-4 py-2.5">
        <div class="flex items-center">
          <Sheet v-model:open="menuOpen">
            <SheetTrigger as-child>
              <Button variant="outline" size="sm" class="mr-2 sm:hidden">
                <icon-fa6-solid:bars class="size-4" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" class="w-80 p-0">
              <div class="flex h-full flex-col">
                <SheetHeader class="border-b px-4 py-3 text-left">
                  <div class="flex items-center gap-2">
                    <SheetTitle class="text-base">{{ $t("common.appName") }}</SheetTitle>
                    <Badge
                      variant="outline"
                      class="border-primary/30 bg-primary/10 text-[10px] font-bold uppercase tracking-wider text-primary"
                    >
                      Admin
                    </Badge>
                  </div>
                </SheetHeader>
                <div class="flex-1 flex flex-col gap-1 px-4 py-4">
                  <RouterLink
                    v-for="link in navLinks"
                    :key="link.to"
                    :to="link.to"
                    :class="[
                      'inline-flex items-center rounded-lg px-3 py-2 text-sm font-medium transition-all duration-150',
                      isLinkActive(link.match)
                        ? 'bg-primary/10 font-semibold text-primary shadow-xs'
                        : 'text-muted-foreground hover:bg-accent hover:text-foreground',
                    ]"
                    @click="menuOpen = false"
                  >
                    {{ $t(link.labelKey) }}
                  </RouterLink>
                </div>
                <Separator />
                <div class="p-4 space-y-2">
                  <Button
                    variant="outline"
                    size="sm"
                    class="w-full justify-start text-muted-foreground"
                    @click="toggleDarkMode"
                  >
                    <icon-fa6-solid:moon v-if="!darkMode" class="mr-2 size-4 text-amber-500" />
                    <icon-fa6-solid:sun v-else class="mr-2 size-4 text-amber-400" />
                    {{ darkMode ? "Light Mode" : "Dark Mode" }}
                  </Button>

                  <RouterLink
                    to="/"
                    class="inline-flex w-full items-center rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
                    @click="menuOpen = false"
                  >
                    <icon-fa6-solid:house class="mr-2 size-4" />
                    {{ $t("admin.backToDashboard") }}
                  </RouterLink>
                </div>
              </div>
            </SheetContent>
          </Sheet>

          <RouterLink to="/" class="flex items-center gap-2.5">
            <Logo class="size-7 shrink-0 transition-transform hover:scale-105" />
            <span class="text-base font-bold tracking-tight">{{ $t("common.appName") }}</span>
            <Badge
              variant="outline"
              class="hidden sm:inline-flex border-primary/30 bg-primary/10 text-[10px] font-bold uppercase tracking-wider text-primary px-1.5 py-0.5"
            >
              Admin
            </Badge>
          </RouterLink>

          <div class="ml-6 hidden items-center gap-1 sm:flex">
            <RouterLink
              v-for="link in navLinks"
              :key="link.to"
              :to="link.to"
              :class="[
                'inline-flex items-center rounded-lg px-3 py-1.5 text-sm font-medium transition-all duration-150',
                isLinkActive(link.match)
                  ? 'bg-primary/10 font-semibold text-primary shadow-xs'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground',
              ]"
            >
              {{ $t(link.labelKey) }}
            </RouterLink>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            class="size-8 rounded-lg text-muted-foreground hover:text-foreground"
            :title="darkMode ? 'Switch to light mode' : 'Switch to dark mode'"
            @click="toggleDarkMode"
          >
            <icon-fa6-solid:moon v-if="!darkMode" class="size-4 text-slate-700" />
            <icon-fa6-solid:sun v-else class="size-4 text-amber-400" />
            <span class="sr-only">Toggle theme</span>
          </Button>

          <Button as-child variant="outline" size="sm" class="h-8 gap-1.5 text-xs font-medium">
            <RouterLink to="/">
              <icon-fa6-solid:house class="size-3.5" />
              {{ $t("admin.backToDashboard") }}
            </RouterLink>
          </Button>
        </div>
      </div>
    </nav>

    <ImpersonationBanner />

    <div class="container mx-auto px-4 py-6">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRoute } from "vue-router";
import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Logo from "@/components/Logo.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { useDarkMode } from "@/composables/useDarkMode";

const route = useRoute();
const menuOpen = ref(false);
const { darkMode, toggleDarkMode } = useDarkMode();

const navLinks = [
  { to: "/admin/dashboard", labelKey: "admin.dashboard", match: "admin.dashboard" },
  { to: "/admin/users", labelKey: "common.users", match: "admin.users" },
  { to: "/admin/organizations", labelKey: "common.organizations", match: "admin.organizations" },
  { to: "/admin/environments", labelKey: "common.environments", match: "admin.environments" },
  { to: "/admin/audit-log", labelKey: "admin.auditLog", match: "admin.auditLog" },
];

function isLinkActive(match: string) {
  return route.name?.toString().startsWith(match) ?? false;
}
</script>
