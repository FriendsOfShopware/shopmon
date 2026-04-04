<template>
  <div class="flex min-h-svh flex-col">
    <div class="relative flex flex-1 items-center justify-center p-4">
      <div class="absolute right-2 top-2 flex items-center gap-2">
        <Button variant="ghost" size="icon" class="size-8" @click="toggleDarkMode">
          <icon-fa6-regular:moon v-if="darkMode" class="size-4" />
          <icon-octicon:sun-16 v-else class="size-4" />
        </Button>
        <button class="text-xs font-bold tracking-wide text-muted-foreground hover:text-foreground" type="button" @click="toggleLocale">
          {{ String(locale) === "en" ? "DE" : "EN" }}
        </button>
      </div>

      <div class="w-full max-w-md">
        <RouterLink :to="{ name: 'home' }">
          <Logo class="mx-auto mb-8 h-24 w-auto" />
        </RouterLink>
        <Notification />
        <RouterView />
      </div>
    </div>
    <LayoutFooter />
  </div>
</template>

<script setup lang="ts">
import { useDarkMode } from "@/composables/useDarkMode";
import { useLocale } from "@/composables/useLocale";
import Notification from "@/components/Notification.vue";
import LayoutFooter from "@/components/layout/LayoutFooter.vue";
import Logo from "@/components/Logo.vue";
import { Button } from "@/components/ui/button";

const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
</script>
