<template>
  <div class="flex flex-1 flex-col bg-gradient-to-b from-[#0a3d5c] via-primary to-primary/30">
    <div
      class="mx-auto flex w-full max-w-2xl flex-1 flex-col items-center justify-center px-6 py-32 text-center"
    >
      <p class="text-7xl font-bold tracking-tight text-white sm:text-8xl">404</p>
      <h1 class="mt-4 text-2xl font-semibold text-white">{{ $t("notFound.title") }}</h1>
      <p class="mt-2 text-sm text-white/70">{{ $t("notFound.description") }}</p>
      <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
        <Button
          v-if="session"
          as-child
          class="rounded-full bg-white px-5 text-[#0c4a6e] hover:bg-white/90"
        >
          <RouterLink :to="{ name: 'account.dashboard' }">
            {{ $t("nav.dashboard") }}
          </RouterLink>
        </Button>
        <Button
          as-child
          variant="ghost"
          class="rounded-full border border-white/20 bg-white/10 px-5 text-white backdrop-blur hover:bg-white/20 hover:text-white"
        >
          <RouterLink :to="{ name: 'home' }">
            {{ $t("notFound.backHome") }}
          </RouterLink>
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSession } from "@/composables/useSession";
import { Button } from "@/components/ui/button";

const { session } = useSession();
</script>
