<template>
  <div class="mx-auto max-w-5xl px-6 py-8">
    <h1 class="mb-8 text-2xl font-bold tracking-tight lg:hidden">{{ $t("nav.documentation") }}</h1>

    <div class="flex gap-10">
      <!-- Sticky sidebar TOC -->
      <aside class="hidden w-52 shrink-0 self-start lg:sticky lg:top-16 lg:block">
        <div class="mb-3 text-sm font-bold tracking-tight">{{ $t("nav.documentation") }}</div>
        <a
          v-for="item in tocItems"
          :key="item.href"
          :href="item.href"
          :class="[
            'block rounded-md px-3 py-1.5 text-[13px] transition-colors',
            activeSection === item.href.slice(1)
              ? 'bg-primary/10 font-medium text-primary'
              : 'text-muted-foreground hover:text-foreground',
          ]"
        >
          {{ item.label }}
        </a>
      </aside>

      <!-- Content -->
      <div class="min-w-0 flex-1 space-y-12">
        <!-- Mobile TOC -->
        <Card class="lg:hidden">
          <CardContent class="p-4">
            <button
              class="flex w-full items-center justify-between text-sm font-medium"
              @click="mobileTocOpen = !mobileTocOpen"
            >
              {{ $t("nav.onThisPage") }}
              <icon-fa6-solid:chevron-down
                :class="[
                  'size-3 text-muted-foreground transition-transform',
                  mobileTocOpen ? 'rotate-180' : '',
                ]"
              />
            </button>
            <nav v-if="mobileTocOpen" class="mt-3 flex flex-col gap-0.5 border-t pt-3">
              <a
                v-for="item in tocItems"
                :key="item.href"
                :href="item.href"
                class="rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground"
                @click="mobileTocOpen = false"
              >
                {{ item.label }}
              </a>
            </nav>
          </CardContent>
        </Card>

        <section
          v-for="section in sections"
          :key="section.id"
          :id="section.id"
          class="scroll-mt-16"
        >
          <div class="mb-4 flex items-center gap-2.5">
            <div class="flex size-8 items-center justify-center rounded-lg bg-primary/10">
              <component :is="section.icon" class="size-4 text-primary" />
            </div>
            <h2 class="text-xl font-bold">{{ section.title }}</h2>
          </div>

          <!-- eslint-disable vue/no-v-html -->
          <div
            class="docs-prose text-[0.9375rem] leading-7 text-muted-foreground"
            v-html="section.content"
          />
          <!-- eslint-enable vue/no-v-html -->
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Card, CardContent } from "@/components/ui/card";
import { useI18n } from "vue-i18n";
import { useInstanceConfig } from "@/composables/useInstanceConfig";

import FaRocket from "~icons/fa6-solid/rocket";
import FaPlug from "~icons/fa6-solid/plug";
import FaBuilding from "~icons/fa6-solid/building";
import FaFolder from "~icons/fa6-solid/folder";
import FaDashboard from "~icons/ri/dashboard-fill";
import FaCircleCheck from "~icons/fa6-solid/circle-check";
import FaPlugExt from "~icons/fa6-solid/puzzle-piece";
import FaListCheck from "~icons/fa6-solid/list-check";
import FaBars from "~icons/fa6-solid/bars-staggered";
import FaChartLine from "~icons/fa6-solid/chart-line";
import FaCodeBranch from "~icons/fa6-solid/code-branch";
import FaFileWaveform from "~icons/fa6-solid/file-waveform";
import FaBell from "~icons/fa6-solid/bell";
import FaKey from "~icons/fa6-solid/key";
import FaCube from "~icons/fa6-solid/cube";
import FaSignIn from "~icons/fa6-solid/right-to-bracket";
import FaRotate from "~icons/fa6-solid/rotate";

const mobileTocOpen = ref(false);
const activeSection = ref("getting-started");
const { t } = useI18n();
const { config: instanceConfig } = useInstanceConfig();

const sectionDefinitions = [
  {
    id: "getting-started",
    titleKey: "docs.sections.gettingStarted.title",
    contentKey: "docs.sections.gettingStarted.content",
    icon: FaRocket,
  },
  {
    id: "connecting-shop",
    titleKey: "docs.sections.connectingShop.title",
    contentKey: "docs.sections.connectingShop.content",
    icon: FaPlug,
  },
  {
    id: "organizations",
    titleKey: "docs.sections.organizations.title",
    contentKey: "docs.sections.organizations.content",
    icon: FaBuilding,
  },
  {
    id: "projects",
    titleKey: "docs.sections.projects.title",
    contentKey: "docs.sections.projects.content",
    icon: FaFolder,
  },
  {
    id: "dashboard-overview",
    titleKey: "docs.sections.dashboardOverview.title",
    contentKey: "docs.sections.dashboardOverview.content",
    icon: FaDashboard,
  },
  {
    id: "health-checks",
    titleKey: "docs.sections.healthChecks.title",
    contentKey: "docs.sections.healthChecks.content",
    icon: FaCircleCheck,
  },
  {
    id: "extensions",
    titleKey: "docs.sections.extensions.title",
    contentKey: "docs.sections.extensions.content",
    icon: FaPlugExt,
  },
  {
    id: "scheduled-tasks",
    titleKey: "docs.sections.scheduledTasks.title",
    contentKey: "docs.sections.scheduledTasks.content",
    icon: FaListCheck,
  },
  {
    id: "queue",
    titleKey: "docs.sections.queue.title",
    contentKey: "docs.sections.queue.content",
    icon: FaBars,
  },
  {
    id: "sitespeed",
    titleKey: "docs.sections.sitespeed.title",
    contentKey: "docs.sections.sitespeed.content",
    icon: FaChartLine,
    feature: "sitespeedEnabled" as const,
  },
  {
    id: "deployments",
    titleKey: "docs.sections.deployments.title",
    contentKey: "docs.sections.deployments.content",
    icon: FaCodeBranch,
  },
  {
    id: "changelog",
    titleKey: "docs.sections.changelog.title",
    contentKey: "docs.sections.changelog.content",
    icon: FaFileWaveform,
  },
  {
    id: "notifications",
    titleKey: "docs.sections.notifications.title",
    contentKey: "docs.sections.notifications.content",
    icon: FaBell,
  },
  {
    id: "shop-token",
    titleKey: "docs.sections.shopToken.title",
    contentKey: "docs.sections.shopToken.content",
    icon: FaKey,
  },
  {
    id: "packages-mirror",
    titleKey: "docs.sections.packagesMirror.title",
    contentKey: "docs.sections.packagesMirror.content",
    icon: FaCube,
    feature: "packageMirrorEnabled" as const,
  },
  {
    id: "sso",
    titleKey: "docs.sections.sso.title",
    contentKey: "docs.sections.sso.content",
    icon: FaSignIn,
  },
  {
    id: "scraping",
    titleKey: "docs.sections.scraping.title",
    contentKey: "docs.sections.scraping.content",
    icon: FaRotate,
  },
];

const sections = computed(() =>
  sectionDefinitions
    .filter(
      (s) =>
        !("feature" in s) ||
        instanceConfig.value?.[s.feature as keyof typeof instanceConfig.value] !== false,
    )
    .map((s) => ({
      ...s,
      title: t(s.titleKey),
      content: t(s.contentKey),
    })),
);

const tocItems = computed(() => sections.value.map((s) => ({ href: `#${s.id}`, label: s.title })));

// Intersection observer for active section tracking
function updateActiveSection() {
  const headerOffset = 80;
  let current = sections.value[0]?.id ?? "getting-started";

  for (const section of sections.value) {
    const el = document.getElementById(section.id);
    if (el) {
      const top = el.getBoundingClientRect().top;
      if (top <= headerOffset) {
        current = section.id;
      }
    }
  }

  activeSection.value = current;
}

onMounted(() => {
  window.addEventListener("scroll", updateActiveSection, { passive: true });
  updateActiveSection();
});

onUnmounted(() => {
  window.removeEventListener("scroll", updateActiveSection);
});
</script>

<style>
@reference "../app.css";

.docs-prose h4 {
  @apply mt-6 mb-2 text-base font-semibold text-foreground;
}

.docs-prose h4:first-child {
  @apply mt-0;
}

.docs-prose p {
  @apply mb-3;
}

.docs-prose p:last-child {
  @apply mb-0;
}

.docs-prose ul,
.docs-prose ol {
  @apply mb-3 space-y-1 pl-5;
}

.docs-prose ul {
  @apply list-disc;
}

.docs-prose ol {
  @apply list-decimal;
}

.docs-prose code {
  @apply rounded bg-muted px-1.5 py-0.5 text-sm text-foreground;
}

.docs-prose a {
  @apply text-primary underline underline-offset-2;
}

.docs-prose strong {
  @apply text-foreground;
}

.docs-prose .callout {
  @apply mt-3 rounded-lg border border-primary/20 bg-primary/5 px-4 py-3 text-sm;
}
</style>
