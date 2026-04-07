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
const { config: instanceConfig } = useInstanceConfig();

const allSections = [
  {
    id: "getting-started",
    title: "Getting Started",
    icon: FaRocket,
    content: `<p>Shopmon is an open-source monitoring dashboard for Shopware 6 shops. It lets you monitor all your Shopware environments from one central place — shop status, performance metrics, extension updates, and security alerts in real-time.</p>
<p>To start monitoring, you need to:</p>
<ol><li>Create an <a href="#organizations">Organization</a> to group your team and shops</li><li>Create a <a href="#projects">Project</a> within that organization</li><li><a href="#connecting-shop">Connect a Shop</a> by providing API credentials</li></ol>`,
  },
  {
    id: "connecting-shop",
    title: "Connecting a Shop",
    icon: FaPlug,
    content: `<p>There are two ways to connect your Shopware shop to Shopmon:</p>
<h4>Option 1: Using the Shopmon Plugin (Recommended)</h4>
<p>Install the <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank">FroshShopmon Plugin</a> in your Shopware shop. The plugin automatically creates the required integration and provides a base64-encoded connection string.</p>
<ol><li>Install and activate the FroshShopmon plugin in your Shopware Administration</li><li>Copy the base64 connection string from the plugin configuration</li><li>In Shopmon, click "Connect using Shopmon Plugin" when adding a shop</li><li>Paste the connection string — credentials are filled automatically</li></ol>
<h4>Option 2: Manual Integration</h4>
<p>Create an integration in your Shopware Administration under <strong>Settings → System → Integrations</strong>. The integration needs read access to system info, plugins, scheduled tasks, and message queue stats; and write access for cache operations and task rescheduling.</p>
<p>See the <a href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions" target="_blank">plugin documentation</a> for the exact permission list.</p>
<div class="callout">Your Client-Secret is encrypted before being stored. Shopmon uses it exclusively to authenticate with your Shopware API.</div>`,
  },
  {
    id: "organizations",
    title: "Organizations",
    icon: FaBuilding,
    content: `<p>Organizations are the top-level grouping entity. Every shop belongs to an organization, and organizations can have multiple members.</p>
<h4>Managing Members</h4>
<p>Organization owners can invite new members by email. Invitees receive an email with accept/reject links. Expired invitations are automatically cleaned up.</p>
<h4>Creating an Organization</h4>
<p>Go to My Organizations and click "New Organization". Give it a name to get started.</p>`,
  },
  {
    id: "projects",
    title: "Projects",
    icon: FaFolder,
    content: `<p>Projects sit within an organization and group related shops together — for example, a client project with staging and production environments.</p>
<ul><li>Each project has a name, description, and optional Git repository URL</li><li>The Git URL is used to link commits in the <a href="#deployments">Deployments</a> view</li><li>Projects can have <strong>API Keys</strong> with specific scopes for CI/CD integration</li><li>Projects can manage <a href="#packages-mirror">Packages Mirror</a> tokens</li></ul>`,
  },
  {
    id: "dashboard-overview",
    title: "Dashboard Overview",
    icon: FaDashboard,
    content: `<p>The Dashboard gives you a quick overview of all your shops and organizations. Each shop card shows:</p>
<ul><li><strong>Shop name</strong> and favicon</li><li><strong>Project</strong> it belongs to</li><li><strong>Shopware version</strong></li><li><strong>Status indicator</strong> (green, yellow, or red) based on health checks</li></ul>
<p>The dashboard also shows recent changes across all your shops.</p>`,
  },
  {
    id: "health-checks",
    title: "Health Checks",
    icon: FaCircleCheck,
    content: `<p>After each scrape, Shopmon runs automated health checks and assigns a status (green, yellow, or red) to your shop.</p>
<h4>Built-in Checks</h4>
<ul><li><strong>Security</strong> — Cross-references your Shopware version against known CVE advisories</li><li><strong>Environment</strong> — Verifies production/staging mode (not dev)</li><li><strong>Scheduled Tasks</strong> — Flags overdue tasks</li><li><strong>Admin Worker</strong> — Warns if still enabled in production</li><li><strong>Frosh Tools</strong> — Additional checks for PHP, MySQL, opcache, memory limits, queue storage, and mail config</li></ul>
<div class="callout">You can suppress individual checks per shop using the ignore feature.</div>`,
  },
  {
    id: "extensions",
    title: "Extensions",
    icon: FaPlugExt,
    content: `<p>Shopmon tracks all installed plugins and apps:</p>
<ul><li>View installed versions and available updates</li><li>See Shopware Store ratings and links</li><li>Get notified when updates are available</li><li>View extension changelogs for available updates</li></ul>
<h4>Cross-Shop Extension View</h4>
<p>The My Extensions page aggregates all extensions across all your shops to identify outdated or inconsistent versions.</p>
<h4>Compatibility Check</h4>
<p>When a new Shopware version is available, use the built-in Compatibility Check to verify extension support before upgrading.</p>`,
  },
  {
    id: "scheduled-tasks",
    title: "Scheduled Tasks",
    icon: FaListCheck,
    content: `<p>Monitor all registered scheduled tasks. The overview shows each task's name, status, interval, and execution times.</p>
<ul><li>Overdue tasks are highlighted with a warning</li><li>You can <strong>reschedule stuck tasks</strong> directly from Shopmon</li></ul>`,
  },
  {
    id: "queue",
    title: "Queue Monitoring",
    icon: FaBars,
    content: `<p>View the message queue sizes of your Shopware shop. This helps you identify if messages are piling up and workers aren't keeping up.</p>`,
  },
  {
    id: "sitespeed",
    title: "Performance Monitoring",
    icon: FaChartLine,
    feature: "sitespeedEnabled" as const,
    content: `<p>Shopmon can perform daily automated page speed checks using <a href="https://www.sitespeed.io" target="_blank">sitespeed.io</a>.</p>
<h4>What's Measured</h4>
<ul><li><strong>TTFB</strong> — Time to First Byte</li><li><strong>Fully Loaded</strong> — Total page load time</li><li><strong>LCP</strong> — Largest Contentful Paint</li><li><strong>FCP</strong> — First Contentful Paint</li><li><strong>CLS</strong> — Cumulative Layout Shift</li><li><strong>Transfer Size</strong> — Total bytes transferred</li></ul>
<p>Configure up to 5 URLs per shop. When linked with <a href="#deployments">Deployments</a>, performance changes can be correlated with specific releases.</p>`,
  },
  {
    id: "deployments",
    title: "Deployments",
    icon: FaCodeBranch,
    content: `<p>Track deployment history using <a href="https://github.com/FriendsOfShopware/shopmon-cli" target="_blank">shopmon-cli</a> from your CI/CD pipeline.</p>
<ol><li>Create an API key with the <code>deployments</code> scope in Project Settings</li><li>Install <code>shopmon-cli</code> in your pipeline</li><li>The CLI reports deployment details, git reference, composer packages, and log output</li></ol>
<div class="callout">Sitespeed measurements are automatically tagged with deployments for performance correlation.</div>`,
  },
  {
    id: "changelog",
    title: "Changelog",
    icon: FaFileWaveform,
    content: `<p>Shopmon automatically detects and records changes on each scrape:</p>
<ul><li>Shopware version upgrades or downgrades</li><li>Extension installations, updates, removals</li><li>Extension activations and deactivations</li></ul>`,
  },
  {
    id: "notifications",
    title: "Notifications",
    icon: FaBell,
    content: `<h4>In-App Notifications</h4>
<p>Notifications appear in the sidebar when a shop's status worsens or authentication errors occur.</p>
<h4>Email Alerts</h4>
<p>Subscribe per shop to receive emails on status changes. Duplicate alerts are prevented.</p>`,
  },
  {
    id: "shop-token",
    title: "Bypass Authentication",
    icon: FaKey,
    content: `<p>If your shop is behind HTTP authentication, configure your web server to allow requests with the header <code>shopmon-shop-token</code> matching your shop's token value. The token is shown on the shop information page.</p>`,
  },
  {
    id: "packages-mirror",
    title: "Packages Mirror",
    icon: FaCube,
    feature: "packageMirrorEnabled" as const,
    content: `<p>Serve Shopware store packages through a Global CDN — ~80ms instead of ~6s from <code>packages.shopware.com</code>.</p>
<ol><li>Add your Shopware store token in the project settings</li><li>Replace <code>packages.shopware.com</code> with the mirror URL in <code>composer.json</code></li><li>Authenticate with <code>composer config --auth</code></li></ol>
<div class="callout">Tokens are validated against the Shopware store and synced automatically every hour.</div>`,
  },
  {
    id: "sso",
    title: "Single Sign-On (SSO)",
    icon: FaSignIn,
    content: `<p>Organizations can configure OIDC-based SSO. Auto-discovery via <code>/.well-known/openid-configuration</code> is supported — just provide the issuer URL. Users with matching email domains are automatically redirected.</p>`,
  },
  {
    id: "scraping",
    title: "How Scraping Works",
    icon: FaRotate,
    content: `<p>Shopmon scrapes your shop once per hour, collecting Shopware version, plugins, scheduled tasks, message queues, cache config, and favicon. After scraping, health checks run and notifications trigger on status changes.</p>
<h4>Remote Actions</h4>
<ul><li><strong>Clear Cache</strong> — Remotely clears HTTP and application cache</li><li><strong>Reschedule Tasks</strong> — Resets stuck tasks back to "scheduled"</li></ul>`,
  },
];

const sections = computed(() =>
  allSections.filter(
    (s) =>
      !("feature" in s) ||
      instanceConfig.value?.[s.feature as keyof typeof instanceConfig.value] !== false,
  ),
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
