<template>
  <div class="min-h-screen bg-background">
    <nav class="sticky top-0 z-10 border-b bg-background">
      <div class="container mx-auto flex items-center justify-between gap-4 px-4 py-2">
        <div class="flex items-center">
          <RouterLink to="/" class="flex items-center gap-2.5">
            <Logo class="size-7 shrink-0" />
            <span class="text-base font-bold tracking-tight">{{ $t("common.appName") }}</span>
          </RouterLink>

          <div class="ml-6 hidden items-center gap-1 sm:flex">
            <RouterLink
              v-for="link in navLinks"
              :key="link.to"
              :to="link.to"
              :class="[
                'inline-flex items-center rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                isLinkActive(link.match)
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
              ]"
            >
              {{ $t(link.labelKey) }}
            </RouterLink>
          </div>
        </div>

        <Button as-child variant="outline" size="sm">
          <RouterLink to="/">
            <icon-fa6-solid:house class="mr-1 size-4" />
            {{ $t("admin.backToDashboard") }}
          </RouterLink>
        </Button>
      </div>
    </nav>

    <ImpersonationBanner />

    <div class="container mx-auto px-4 py-6">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import ImpersonationBanner from "@/components/ImpersonationBanner.vue";
import Logo from "@/components/Logo.vue";
import { Button } from "@/components/ui/button";

const route = useRoute();

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
