<template>
  <div class="min-h-screen bg-background">
    <nav class="border-b bg-primary text-primary-foreground">
      <div class="container mx-auto flex items-center justify-between px-4 py-2">
        <div class="flex items-center">
          <Button
            as-child
            variant="ghost"
            class="text-primary-foreground hover:bg-white/10 hover:text-white"
          >
            <RouterLink to="/">
              <Logo class="h-7 w-auto brightness-0 invert" />
            </RouterLink>
          </Button>

          <div class="ml-6 hidden items-center gap-1 sm:flex">
            <RouterLink
              v-for="link in navLinks"
              :key="link.to"
              :to="link.to"
              :class="[
                'inline-flex items-center rounded px-3 py-1.5 text-sm font-medium transition-colors',
                isLinkActive(link.match)
                  ? 'bg-white/30 text-white'
                  : 'text-white/80 hover:bg-white/20 hover:text-white',
              ]"
            >
              {{ link.label }}
            </RouterLink>
          </div>
        </div>

        <Button as-child variant="secondary" size="sm">
          <RouterLink to="/">
            <icon-fa6-solid:house class="mr-1 size-4" />
            Back to Dashboard
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
  { to: "/admin/dashboard", label: "Dashboard", match: "admin.dashboard" },
  { to: "/admin/users", label: "Users", match: "admin.users" },
  { to: "/admin/organizations", label: "Organizations", match: "admin.organizations" },
  { to: "/admin/environments", label: "Environments", match: "admin.environments" },
];

function isLinkActive(match: string) {
  return route.name?.toString().startsWith(match) ?? false;
}
</script>
