<template>
  <div class="flex min-h-screen flex-col">
    <!-- Transparent nav — floats over hero -->
    <nav class="absolute inset-x-0 top-0 z-20">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <RouterLink :to="{ name: 'home' }" class="flex items-center gap-2">
          <Logo class="h-8 w-auto" />
          <span class="text-lg font-bold text-white">Shopmon</span>
        </RouterLink>

        <div class="flex items-center gap-2">
          <Button
            v-if="session"
            as-child
            size="sm"
            class="rounded-full bg-white/10 text-white backdrop-blur hover:bg-white/20 hover:text-white border-white/20 border"
          >
            <RouterLink :to="{ name: 'account.dashboard' }">
              <icon-ri:dashboard-fill class="mr-1.5 size-3.5" />
              {{ $t("nav.dashboard") }}
            </RouterLink>
          </Button>
          <template v-else>
            <Button
              as-child
              size="sm"
              variant="ghost"
              class="text-white/80 hover:bg-white/10 hover:text-white"
            >
              <RouterLink :to="{ name: 'account.login' }">
                {{ $t("nav.login") }}
              </RouterLink>
            </Button>
            <Button
              as-child
              size="sm"
              class="rounded-full bg-white px-5 text-[#0c4a6e] hover:bg-white/90"
            >
              <RouterLink :to="{ name: 'account.register' }">
                {{ $t("nav.register") }}
              </RouterLink>
            </Button>
          </template>

          <Button
            variant="ghost"
            size="icon"
            class="size-8 text-white/60 hover:bg-white/10 hover:text-white"
            @click="toggleDarkMode"
          >
            <icon-fa6-regular:moon v-if="darkMode" class="size-4" />
            <icon-octicon:sun-16 v-else class="size-4" />
          </Button>

          <button
            class="text-xs font-bold tracking-wide text-white/50 hover:text-white"
            type="button"
            @click="toggleLocale"
          >
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
