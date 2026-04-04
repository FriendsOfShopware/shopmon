<template>
  <div class="flex min-h-screen flex-col bg-background">
    <nav class="border-b bg-primary text-primary-foreground">
      <div class="container mx-auto flex items-center justify-between px-4 py-2">
        <RouterLink :to="{ name: 'home' }" class="flex items-center">
          <Logo class="h-8 w-auto brightness-0 invert" />
        </RouterLink>

        <div class="flex items-center gap-1">
          <Button
            v-if="session"
            as-child
            variant="secondary"
            size="sm"
          >
            <RouterLink :to="{ name: 'account.dashboard' }">
              <icon-ri:dashboard-fill class="mr-1 size-4" />
              {{ $t("nav.dashboard") }}
            </RouterLink>
          </Button>
          <template v-else>
            <Button as-child variant="secondary" size="sm">
              <RouterLink :to="{ name: 'account.login' }">
                <icon-fa6-solid:right-to-bracket class="mr-1 size-4" />
                {{ $t("nav.login") }}
              </RouterLink>
            </Button>
            <Button as-child variant="secondary" size="sm">
              <RouterLink :to="{ name: 'account.register' }">
                <icon-fa6-solid:user-plus class="mr-1 size-4" />
                {{ $t("nav.register") }}
              </RouterLink>
            </Button>
          </template>

          <Button variant="ghost" size="icon" class="size-8 text-primary-foreground hover:bg-white/10 hover:text-white" @click="toggleDarkMode">
            <icon-fa6-regular:moon v-if="darkMode" class="size-4" />
            <icon-octicon:sun-16 v-else class="size-4" />
          </Button>

          <button class="ml-1 text-xs font-bold tracking-wide text-primary-foreground/70 hover:text-white" type="button" @click="toggleLocale">
            {{ String(locale) === "en" ? "DE" : "EN" }}
          </button>
        </div>
      </div>
    </nav>

    <RouterView />

    <LayoutFooter />
  </div>
</template>

<script setup lang="ts">
import { useDarkMode } from "@/composables/useDarkMode";
import { useSession } from "@/composables/useSession";
import { useLocale } from "@/composables/useLocale";
import { Button } from "@/components/ui/button";
import LayoutFooter from "@/components/layout/LayoutFooter.vue";
import Logo from "@/components/Logo.vue";

const { session } = useSession();
const { darkMode, toggleDarkMode } = useDarkMode();
const { locale, toggleLocale } = useLocale();
</script>
