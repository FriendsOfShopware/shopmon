<template>
  <!-- ═══════ HERO ═══════ -->
  <section
    class="relative overflow-hidden bg-gradient-to-b from-[#0a3d5c] via-[#0c6ea6] to-primary/50 pb-12 pt-28 sm:pt-36 md:pb-20"
  >
    <!-- Subtle pattern -->
    <div
      class="pointer-events-none absolute inset-0 opacity-[0.04]"
      style="
        background-image: radial-gradient(circle at 1px 1px, white 1px, transparent 0);
        background-size: 32px 32px;
      "
    />
    <div
      class="pointer-events-none absolute left-1/2 top-1/4 -translate-x-1/2 -translate-y-1/2 h-[500px] w-[800px] rounded-full bg-white/[0.03] blur-3xl"
    />

    <div class="relative mx-auto max-w-5xl px-6 text-center">
      <div
        class="mb-6 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/[0.08] px-4 py-1.5 text-sm text-white/80 backdrop-blur-sm"
      >
        <icon-fa6-solid:lock-open class="size-3" />
        Open-Source
      </div>

      <h1
        class="mx-auto max-w-4xl text-3xl font-extrabold leading-[1.15] tracking-tight text-white sm:text-5xl lg:text-[3.5rem]"
      >
        The monitoring dashboard<br class="hidden sm:block" />
        <span class="text-white/60">your Shopware shops deserve</span>
      </h1>

      <p class="mx-auto mt-5 max-w-xl text-base leading-relaxed text-white/60 sm:text-lg">
        {{ $t("home.description") }}
      </p>

      <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
        <Button
          as-child
          size="lg"
          class="h-11 rounded-full bg-white px-7 text-sm font-semibold text-[#0a3d5c] shadow-lg hover:bg-white/90 sm:h-12 sm:px-8 sm:text-base"
        >
          <RouterLink :to="{ name: 'account.register' }">
            {{ $t("home.startFree") }}
            <icon-fa6-solid:arrow-right class="ml-2 size-3.5" />
          </RouterLink>
        </Button>
        <Button
          as-child
          size="lg"
          variant="ghost"
          class="h-11 rounded-full border border-white/15 px-7 text-sm text-white/80 hover:bg-white/10 hover:text-white sm:h-12 sm:px-8 sm:text-base"
        >
          <a href="https://github.com/FriendsOfShopware/shopmon/" target="_blank" rel="noopener">
            <icon-mdi:github class="mr-2 size-5" />
            GitHub
          </a>
        </Button>
      </div>

      <!-- Hero screenshot -->
      <div class="mx-auto mt-12 max-w-4xl sm:mt-16">
        <div
          class="rounded-xl bg-gradient-to-b from-white/10 to-white/5 p-1.5 shadow-2xl ring-1 ring-white/10"
        >
          <div class="overflow-hidden rounded-lg">
            <img
              :src="getThemeImage('/home/shopmon-dashboard.png')"
              :alt="$t('home.altDashboard')"
              class="w-full"
              fetchpriority="high"
            />
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ═══════ FEATURES — 3-column ═══════ -->
  <section class="mx-auto max-w-6xl px-6 py-20 md:py-28">
    <div class="mb-12 text-center md:mb-16">
      <h2 class="text-2xl font-bold tracking-tight sm:text-3xl">{{ $t("home.subtitle") }}</h2>
      <p class="mx-auto mt-3 max-w-lg text-muted-foreground">
        One dashboard to monitor status, performance, extensions, and deployments across all your
        environments.
      </p>
    </div>

    <div class="grid gap-6 md:grid-cols-3">
      <div
        v-for="feature in features"
        :key="feature.title"
        class="group rounded-2xl border bg-card p-7 transition-all duration-200 hover:shadow-lg"
      >
        <div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-primary/10">
          <img :src="feature.icon" :alt="feature.alt" class="size-6 object-contain" />
        </div>
        <h3 class="mb-2 text-lg font-semibold group-hover:text-primary transition-colors">
          {{ feature.title }}
        </h3>
        <p class="text-sm leading-relaxed text-muted-foreground">{{ feature.description }}</p>
      </div>
    </div>
  </section>

  <!-- ═══════ SHOWCASE SECTIONS ═══════ -->
  <section class="border-y bg-muted/30">
    <div class="mx-auto max-w-6xl divide-y px-6">
      <div
        v-for="(section, i) in showcaseSections"
        :key="i"
        :class="[
          'flex items-center gap-10 py-16 md:gap-16 md:py-24 max-md:flex-col',
          i % 2 !== 0 ? 'md:flex-row-reverse' : '',
        ]"
      >
        <div class="flex-1 max-md:text-center">
          <span
            class="mb-3 inline-flex items-center gap-1.5 rounded-md bg-primary/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wider text-primary"
          >
            <component :is="section.icon" class="size-3" />
            {{ section.badge }}
          </span>
          <h2 class="mb-3 text-2xl font-bold tracking-tight md:text-3xl">{{ section.title }}</h2>
          <p class="max-w-md text-[0.9375rem] leading-relaxed text-muted-foreground max-md:mx-auto">
            {{ section.description }}
          </p>
        </div>
        <div class="flex-1">
          <div
            class="overflow-hidden rounded-xl border shadow-md transition-shadow hover:shadow-lg"
          >
            <img
              :src="getThemeImage(section.image)"
              :alt="section.alt"
              class="w-full"
              loading="lazy"
            />
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ═══════ MORE FEATURES ═══════ -->
  <section class="mx-auto max-w-4xl px-6 py-16 text-center md:py-20">
    <h2 class="mb-2 text-2xl font-bold tracking-tight">{{ $t("home.andMore") }}</h2>
    <p class="mb-6 text-sm text-muted-foreground">
      Everything you need to keep your Shopware environments healthy.
    </p>
    <div class="flex flex-wrap items-center justify-center gap-2">
      <Badge
        v-for="feat in featureList"
        :key="feat"
        variant="outline"
        class="rounded-full px-4 py-1.5 text-sm transition-colors hover:bg-primary/10 hover:text-primary"
      >
        {{ feat }}
      </Badge>
    </div>
  </section>

  <!-- ═══════ SPONSORS ═══════ -->
  <section v-if="sponsors.length" class="border-t bg-muted/30 py-16 md:py-20">
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
  <section
    class="relative overflow-hidden bg-gradient-to-br from-[#0a3d5c] via-[#0c6ea6] to-primary/70"
  >
    <div
      class="pointer-events-none absolute -right-20 -top-20 size-72 rounded-full bg-white/[0.04] blur-3xl"
    />
    <div
      class="relative mx-auto flex max-w-5xl items-center gap-8 px-6 py-16 max-md:flex-col max-md:text-center md:gap-12 md:py-20"
    >
      <div class="flex-1 md:text-right">
        <h2 class="mb-3 text-2xl font-bold text-white sm:text-3xl">{{ $t("home.readyTitle") }}</h2>
        <p class="text-base leading-relaxed text-white/60">{{ $t("home.readyDesc") }}</p>
      </div>
      <div class="shrink-0">
        <Button
          as-child
          size="lg"
          class="h-12 rounded-full bg-white px-8 text-base font-semibold text-[#0a3d5c] shadow-lg hover:bg-white/90"
        >
          <RouterLink :to="{ name: 'account.register' }">
            {{ $t("home.startFree") }}
            <icon-fa6-solid:arrow-right class="ml-2 size-3.5" />
          </RouterLink>
        </Button>
      </div>
    </div>
  </section>

  <!-- ═══════ VALUE PROPS ═══════ -->
  <section class="mx-auto max-w-5xl px-6 py-16 md:py-20">
    <div class="grid gap-10 md:grid-cols-3">
      <div v-for="prop in valueProps" :key="prop.title" class="text-center">
        <div class="mx-auto mb-4 flex size-12 items-center justify-center rounded-xl bg-primary/10">
          <component :is="prop.icon" class="size-6 text-primary" aria-hidden="true" />
        </div>
        <h3 class="mb-1.5 text-base font-semibold">{{ prop.title }}</h3>
        <p class="mx-auto max-w-xs text-sm leading-relaxed text-muted-foreground">
          {{ prop.description }}
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
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

const { getThemeImage } = useDarkMode();
const { t } = useI18n();

const features = computed(() => [
  {
    icon: "/home/dashboard.svg",
    alt: t("home.altDashboard"),
    title: t("home.dashboardOverview"),
    description: t("home.dashboardOverviewDesc"),
  },
  {
    icon: "/home/automatic-performancechecks.svg",
    alt: t("home.altAutoChecks"),
    title: t("home.autoPerformanceChecks"),
    description: t("home.autoPerformanceChecksDesc"),
  },
  {
    icon: "/home/speed.png",
    alt: t("home.altSitespeedLogo"),
    title: t("home.performanceMonitoring"),
    description: t("home.performanceMonitoringDesc"),
  },
]);

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
  t("home.featureList")
    .split("|")
    .map((s: string) => s.trim()),
);

const valueProps = computed(() => [
  { icon: FaMoneyOff, title: t("home.free"), description: t("home.freeDesc") },
  {
    icon: FaPeopleCommunity,
    title: t("home.communityDriven"),
    description: t("home.communityDrivenDesc"),
  },
  { icon: FaGithub, title: t("home.openSource"), description: t("home.openSourceDesc") },
]);
</script>
