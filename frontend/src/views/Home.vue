<template>
  <!-- ═══════ HERO ═══════ -->
  <section class="relative overflow-hidden bg-gradient-to-b from-[#0c4a6e] via-primary to-primary/60 pb-8 pt-16 md:pt-24">
    <!-- Radial glow -->
    <div class="pointer-events-none absolute left-1/2 top-0 -translate-x-1/2 h-[600px] w-[900px] rounded-full bg-white/5 blur-3xl" />

    <div class="relative mx-auto max-w-6xl px-6 text-center">
      <div class="mb-8 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-1.5 text-sm text-white/90 backdrop-blur">
        <icon-fa6-solid:lock-open class="size-3" />
        Open-Source &amp; Free Forever
      </div>

      <h1 class="mx-auto max-w-3xl text-4xl font-extrabold leading-[1.1] tracking-tight text-white sm:text-5xl lg:text-6xl">
        {{ $t("home.title") }}
      </h1>

      <p class="mx-auto mt-6 max-w-xl text-lg leading-relaxed text-white/70">
        {{ $t("home.description") }}
      </p>

      <div class="mt-10 flex flex-wrap items-center justify-center gap-4">
        <Button as-child size="lg" class="h-12 rounded-full bg-white px-8 text-base font-semibold text-[#0c4a6e] shadow-xl hover:bg-white/90">
          <RouterLink :to="{ name: 'account.register' }">
            {{ $t("home.startFree") }}
            <icon-fa6-solid:arrow-right class="ml-2 size-3.5" />
          </RouterLink>
        </Button>
        <Button as-child size="lg" variant="ghost" class="h-12 rounded-full border border-white/20 px-8 text-base text-white hover:bg-white/10 hover:text-white">
          <a href="https://github.com/FriendsOfShopware/shopmon/" target="_blank" rel="noopener">
            <icon-mdi:github class="mr-2 size-5" />
            GitHub
          </a>
        </Button>
      </div>

      <!-- Hero screenshot -->
      <div class="mx-auto mt-14 max-w-4xl">
        <div class="overflow-hidden rounded-xl border border-white/10 shadow-2xl ring-1 ring-white/5">
          <img :src="getThemeImage('/home/shopmon-dashboard.png')" :alt="$t('home.altDashboard')" class="w-full" />
        </div>
      </div>
    </div>
  </section>

  <!-- ═══════ BENTO FEATURES ═══════ -->
  <section class="mx-auto max-w-6xl px-6 py-24">
    <div class="mb-14 text-center">
      <h2 class="text-3xl font-bold tracking-tight md:text-4xl">{{ $t("home.subtitle") }}</h2>
      <p class="mx-auto mt-3 max-w-xl text-muted-foreground">
        {{ $t("home.description") }}
      </p>
    </div>

    <!-- Bento grid -->
    <div class="grid gap-4 md:grid-cols-6 md:grid-rows-2">
      <!-- Large card — left -->
      <Card class="group md:col-span-4 overflow-hidden border-0 bg-muted/50 p-0 transition-shadow hover:shadow-lg">
        <div class="flex h-full flex-col md:flex-row">
          <div class="flex flex-1 flex-col justify-center p-8 md:p-10">
            <div class="mb-3 flex size-10 items-center justify-center rounded-xl bg-primary/10">
              <icon-fa6-solid:gauge-high class="size-5 text-primary" />
            </div>
            <h3 class="mb-2 text-xl font-semibold">{{ $t("home.autoPerformanceChecks") }}</h3>
            <p class="text-sm leading-relaxed text-muted-foreground">{{ $t("home.autoPerformanceChecksDesc") }}</p>
          </div>
          <div class="flex flex-1 items-end overflow-hidden p-4 md:p-0 md:pr-0 md:pt-6">
            <img :src="getThemeImage('/home/shopmon-performance-checks.png')" :alt="$t('home.altPerformanceChecks')" class="w-full rounded-tl-xl border-l border-t object-cover object-left-top shadow-sm" loading="lazy" />
          </div>
        </div>
      </Card>

      <!-- Small card — top right -->
      <Card class="group md:col-span-2 flex flex-col border-0 bg-muted/50 p-8 transition-shadow hover:shadow-lg">
        <div class="mb-4 flex size-10 items-center justify-center rounded-xl bg-primary/10">
          <icon-fa6-solid:puzzle-piece class="size-5 text-primary" />
        </div>
        <h3 class="mb-2 text-lg font-semibold">{{ $t("home.extensionsTitle") }}</h3>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ $t("home.autoPerformanceChecksDesc") }}</p>
      </Card>

      <!-- Small card — bottom left -->
      <Card class="group md:col-span-2 flex flex-col border-0 bg-muted/50 p-8 transition-shadow hover:shadow-lg">
        <div class="mb-4 flex size-10 items-center justify-center rounded-xl bg-primary/10">
          <icon-fa6-solid:rocket class="size-5 text-primary" />
        </div>
        <h3 class="mb-2 text-lg font-semibold">{{ $t("home.performanceMonitoring") }}</h3>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ $t("home.performanceMonitoringDesc") }}</p>
      </Card>

      <!-- Large card — right -->
      <Card class="group md:col-span-4 overflow-hidden border-0 bg-muted/50 p-0 transition-shadow hover:shadow-lg">
        <div class="flex h-full flex-col md:flex-row-reverse">
          <div class="flex flex-1 flex-col justify-center p-8 md:p-10">
            <div class="mb-3 flex size-10 items-center justify-center rounded-xl bg-primary/10">
              <icon-fa6-solid:layer-group class="size-5 text-primary" />
            </div>
            <h3 class="mb-2 text-xl font-semibold">{{ $t("home.environmentState") }}</h3>
            <p class="text-sm leading-relaxed text-muted-foreground">{{ $t("home.environmentStateDesc") }}</p>
          </div>
          <div class="flex flex-1 items-end overflow-hidden p-4 md:p-0 md:pl-0 md:pt-6">
            <img :src="getThemeImage('/home/shopmon-extensions.png')" :alt="$t('home.altExtensions')" class="w-full rounded-tr-xl border-r border-t object-cover object-right-top shadow-sm" loading="lazy" />
          </div>
        </div>
      </Card>
    </div>
  </section>

  <!-- ═══════ DETAILED SHOWCASES ═══════ -->
  <section class="border-y bg-muted/30 py-24">
    <div class="mx-auto max-w-6xl space-y-28 px-6">
      <div v-for="(section, i) in showcaseSections" :key="i" :class="['flex items-center gap-16 max-md:flex-col', i % 2 !== 0 ? 'md:flex-row-reverse' : '']">
        <div class="flex-1 max-md:text-center">
          <span class="mb-3 inline-flex items-center gap-1.5 rounded-md bg-primary/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wider text-primary">
            <component :is="section.icon" class="size-3" />
            {{ section.badge }}
          </span>
          <h2 class="mb-4 text-2xl font-bold tracking-tight md:text-3xl">{{ section.title }}</h2>
          <p class="max-w-md leading-relaxed text-muted-foreground max-md:mx-auto">{{ section.description }}</p>
        </div>
        <div class="flex-1">
          <div class="overflow-hidden rounded-xl border shadow-lg">
            <img :src="getThemeImage(section.image)" :alt="section.alt" class="w-full" loading="lazy" />
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ═══════ FEATURE TAGS ═══════ -->
  <section class="mx-auto max-w-4xl px-6 py-20 text-center">
    <h2 class="mb-2 text-2xl font-bold tracking-tight">{{ $t("home.andMore") }}</h2>
    <p class="mb-8 text-muted-foreground">Everything you need to keep your Shopware environments healthy.</p>
    <div class="flex flex-wrap items-center justify-center gap-2">
      <Badge v-for="feat in featureList" :key="feat" variant="outline" class="rounded-full px-4 py-1.5 text-sm transition-colors hover:bg-primary/10 hover:text-primary">
        {{ feat }}
      </Badge>
    </div>
  </section>

  <!-- ═══════ SPONSORS ═══════ -->
  <section v-if="sponsors.length" class="border-t bg-muted/30 py-20">
    <div class="mx-auto max-w-5xl px-6">
      <SponsorShowcase
        :sponsors="sponsors"
        title="Sponsors"
        description="Shopmon is supported by companies that help keep the project free and moving forward for the Shopware community."
        title-tag="h2"
      />
    </div>
  </section>

  <!-- ═══════ CTA ═══════ -->
  <section class="relative overflow-hidden bg-gradient-to-br from-[#0c4a6e] via-primary to-primary/80">
    <div class="pointer-events-none absolute -right-20 -top-20 size-80 rounded-full bg-white/5 blur-3xl" />
    <div class="relative mx-auto flex max-w-5xl items-center gap-12 px-6 py-24 max-md:flex-col max-md:text-center">
      <div class="flex-1 md:text-right">
        <h2 class="mb-4 text-3xl font-bold text-white md:text-4xl">{{ $t("home.readyTitle") }}</h2>
        <p class="text-lg leading-relaxed text-white/70">{{ $t("home.readyDesc") }}</p>
      </div>
      <div class="shrink-0">
        <Button as-child size="lg" class="h-14 rounded-full bg-white px-10 text-lg font-semibold text-[#0c4a6e] shadow-xl hover:bg-white/90">
          <RouterLink :to="{ name: 'account.register' }">
            {{ $t("home.startFree") }}
            <icon-fa6-solid:arrow-right class="ml-2 size-4" />
          </RouterLink>
        </Button>
      </div>
    </div>
  </section>

  <!-- ═══════ VALUE PROPS ═══════ -->
  <section class="mx-auto max-w-5xl px-6 py-24">
    <div class="grid gap-12 md:grid-cols-3">
      <div v-for="prop in valueProps" :key="prop.title" class="text-center">
        <div class="mx-auto mb-4 flex size-14 items-center justify-center rounded-2xl bg-primary/10">
          <component :is="prop.icon" class="size-7 text-primary" aria-hidden="true" />
        </div>
        <h3 class="mb-2 text-lg font-semibold">{{ prop.title }}</h3>
        <p class="mx-auto max-w-xs text-sm leading-relaxed text-muted-foreground">{{ prop.description }}</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RouterLink } from "vue-router";
import SponsorShowcase from "@/components/SponsorShowcase.vue";
import { useDarkMode } from "@/composables/useDarkMode";
import { useI18n } from "vue-i18n";
import { sponsors } from "@/data/sponsors";

import FaMoneyOff from "~icons/fluent/money-off-24-regular";
import FaPeopleCommunity from "~icons/fluent/people-community-12-filled";
import FaGithub from "~icons/mdi/github";

import FaChartLine from "~icons/fa6-solid/chart-line";
import FaShieldHalved from "~icons/fa6-solid/shield-halved";
import FaPuzzlePiece from "~icons/fa6-solid/puzzle-piece";
import FaBolt from "~icons/fa6-solid/bolt";

const { getThemeImage } = useDarkMode();
const { t } = useI18n();

const showcaseSections = computed(() => [
  {
    badge: "Health Checks",
    icon: FaShieldHalved,
    title: t("home.froshToolsTitle"),
    description: t("home.froshToolsDesc"),
    image: "/home/shopmon-performance-checks.png",
    alt: t("home.altPerformanceChecks"),
  },
  {
    badge: "Performance",
    icon: FaChartLine,
    title: t("home.sitespeedTitle"),
    description: t("home.sitespeedDesc"),
    image: "/home/shopmon-sitespeed.png",
    alt: t("home.altSitespeed"),
  },
]);

const featureList = computed(() =>
  t("home.featureList").split("|").map((s: string) => s.trim()),
);

const valueProps = computed(() => [
  { icon: FaMoneyOff, title: t("home.foreverFree"), description: t("home.foreverFreeDesc") },
  { icon: FaPeopleCommunity, title: t("home.communityDriven"), description: t("home.communityDrivenDesc") },
  { icon: FaGithub, title: t("home.openSource"), description: t("home.openSourceDesc") },
]);
</script>
