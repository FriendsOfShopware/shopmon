<template>
  <Panel v-if="environment" class="shop-info">
    <template #title>
      <icon-fa6-solid:circle-info class="icon" />
      {{ $t("shopDetail.shopInfo") }}
    </template>

    <div class="shop-info-grid">
      <dl class="shop-info-list">
        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.shopwareVersion") }}</dt>
          <dd>
            <span>{{ environment.shopwareVersion }}</span>
            <template
              v-if="latestShopwareVersion && latestShopwareVersion != environment.shopwareVersion"
            >
              <a
                class="badge badge-warning"
                :href="
                  'https://github.com/shopware/platform/releases/tag/v' + latestShopwareVersion
                "
                target="_blank"
              >
                {{ latestShopwareVersion }}
              </a>

              <button class="badge badge-info" type="button" @click="openUpdateWizard">
                <icon-fa6-solid:rotate />
                {{ $t("shopDetail.compatibilityCheck") }}
              </button>
            </template>
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>
            {{ $t("shopDetail.lastCheckedAt") }}
            <span :data-tooltip="$t('shopDetail.shopUpdateTooltip')" class="tooltip-top-center">
              <icon-fa6-solid:circle-info class="icon" />
            </span>
          </dt>

          <dd>
            {{ formatDateTime(environment.lastScrapedAt) }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.lastShopUpdate") }}</dt>

          <dd>
            <template v-if="environment.lastChangelog && environment.lastChangelog.date">
              {{ formatDate(environment.lastChangelog.date) }}
            </template>

            <template v-else> {{ $t("common.never") }} </template>
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.lastDeployment") }}</dt>

          <dd v-if="environment.deploymentsCount > 0">
            <router-link
              :to="{
                name: 'account.environments.detail.deployments',
                params: {
                  organizationId: $route.params.organizationId,
                  shopId: $route.params.environmentId,
                },
              }"
            >
              {{ environment.deploymentsCount }} deployment{{
                environment.deploymentsCount !== 1 ? "s" : ""
              }}
            </router-link>
          </dd>

          <dd v-else>{{ $t("common.never") }}</dd>
        </div>

        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.environment") }}</dt>

          <dd>
            {{ environment.cache?.environment }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.cacheAdapter") }}</dt>

          <dd>
            {{ environment.cache?.cacheAdapter }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>{{ $t("shopDetail.httpCache") }}</dt>

          <dd>
            {{ environment.cache?.httpCache ? $t("common.enabled") : $t("common.disabled") }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Organization</dt>

          <dd>
            {{ environment.organizationName }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Project</dt>

          <dd>
            {{ environment.shopName }}
          </dd>
        </div>

        <div class="shop-info-item shop-token-item">
          <dt>
            Bypass Authentication Header
            <span
              data-tooltip="If your website is protected by authentication, configure the header 'shopmon-shop-token' with this value to be excluded"
              class="tooltip-top-center"
            >
              <icon-fa6-solid:circle-info class="icon" />
            </span>
          </dt>

          <dd class="shop-token-value">
            <code>{{ environment.environmentToken }}</code>

            <button
              type="button"
              class="btn btn-sm btn-icon tooltip-top-center"
              data-tooltip="Copy token"
              @click="copyToken"
            >
              <icon-fa6-solid:copy />
            </button>
          </dd>
        </div>
      </dl>
    </div>
  </Panel>

  <Panel>
    <template #title>
      <icon-fa6-solid:circle-check class="icon" />
      {{ $t("shopDetail.securityChecks") }}
    </template>

    <Alert v-if="sortedCriticalChecks.length === 0" type="success">
      {{ $t("shopDetail.allChecksPassed") }}
    </Alert>

    <ul v-else class="issue-list">
      <li v-for="check in sortedCriticalChecks" :key="check.id" class="issue-item">
        <status-icon :status="check.level" />

        <div class="issue-content">
          <span class="issue-message">{{ check.message }}</span>

          <a v-if="check.link" :href="check.link" target="_blank" class="issue-link">
            {{ $t("shopDetail.learnMore") }}
            <icon-fa6-solid:arrow-up-right-from-square class="icon" />
          </a>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.environments.detail.checks',
          params: {
            organizationId: $route.params.organizationId,
            shopId: $route.params.environmentId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:circle-check class="icon" />
        {{ $t("shopDetail.viewAllChecks") }}
      </router-link>
    </div>
  </Panel>

  <Panel>
    <template #title>
      <icon-fa6-solid:plug class="icon" />
      {{ $t("shopDetail.extensions") }}
    </template>

    <Alert v-if="outdatedExtensions.length === 0" type="success">
      {{ $t("shopDetail.allExtensionsUpToDate") }}
    </Alert>

    <ul v-else class="issue-list">
      <li v-for="extension in outdatedExtensions" :key="extension.name" class="issue-item">
        <icon-fa6-solid:arrow-up class="icon icon-info" />

        <div class="issue-content">
          <span class="issue-message">
            {{ extension.label }}
            <span class="link" @click="openExtensionChangelog(extension)"
              >({{ extension.version }} → {{ extension.latestVersion }})</span
            >
          </span>

          <span class="issue-source">{{ extension.name }}</span>

          <a
            v-if="extension.storeLink"
            :href="extension.storeLink"
            target="_blank"
            class="issue-link"
          >
            {{ $t("shopDetail.store") }}
            <icon-fa6-solid:arrow-up-right-from-square class="icon" />
          </a>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.environments.detail.extensions',
          params: {
            organizationId: $route.params.organizationId,
            shopId: $route.params.environmentId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:plug class="icon" />
        {{ $t("shopDetail.viewAllExtensions") }}
      </router-link>
    </div>
  </Panel>

  <Panel>
    <template #title>
      <icon-fa6-solid:list-check class="icon" />
      {{ $t("shopDetail.scheduledTasks") }}
    </template>

    <Alert v-if="overdueTasks.length === 0" type="success">
      {{ $t("shopDetail.noOverdueTasks") }}
    </Alert>

    <ul v-else class="issue-list">
      <li v-for="task in overdueTasks" :key="task.id" class="issue-item">
        <icon-fa6-solid:clock class="icon icon-warning" />
        <div class="issue-content">
          <span class="issue-message">{{ task.name }}</span>
          <span class="issue-source">
            Last run:
            {{ task.lastExecutionTime ? formatDateTime(task.lastExecutionTime) : "-" }} ({{
              getOverdueTime(task.nextExecutionTime ?? "")
            }}
            overdue)
          </span>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.environments.detail.tasks',
          params: {
            organizationId: $route.params.organizationId,
            shopId: $route.params.environmentId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:list-check class="icon" />
        {{ $t("shopDetail.viewAllScheduledTasks") }}
      </router-link>
    </div>
  </Panel>

  <Panel>
    <template #title>
      <icon-fa6-solid:clock-rotate-left class="icon" />
      {{ $t("shopDetail.recentChanges") }}
    </template>

    <Alert v-if="recentChangelogs.length === 0" type="info">
      {{ $t("shopDetail.noRecentChanges") }}
    </Alert>

    <div v-else class="changes-list">
      <div v-for="changelog in recentChangelogs" :key="changelog.id" class="change-item">
        <div class="change-date">
          {{ formatDate(changelog.date) }}
        </div>

        <div class="change-summary">
          {{ sumChanges(changelog) }}
          <span
            class="link tooltip-top-left"
            :data-tooltip="$t('shopDetail.openDetails')"
            @click="openEnvironmentChangelog(changelog)"
            ><icon-fa6-solid:circle-info class="icon"
          /></span>
        </div>
      </div>
    </div>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.environments.detail.changelog',
          params: {
            organizationId: $route.params.organizationId,
            shopId: $route.params.environmentId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:file-waveform class="icon" />
        {{ $t("shopDetail.viewAllChanges") }}
      </router-link>
    </div>
  </Panel>

  <!-- Shop Changelog Modal -->
  <environment-changelog
    :show="viewEnvironmentChangelogDialog"
    :changelog="dialogEnvironmentChangelog"
    @close="closeEnvironmentChangelog"
  />

  <!-- Extension Changelog Modal -->
  <extension-changelog
    :show="viewExtensionChangelogDialog"
    :extension="dialogExtension"
    @close="closeExtensionChangelog"
  />

  <!-- Update Wizard Modal -->
  <shopware-update-wizard
    :show="viewUpdateWizardDialog"
    :shopware-versions="shopwareVersions"
    :loading="loadingUpdateWizard"
    :extensions="dialogUpdateWizard"
    @close="viewUpdateWizardDialog = false"
    @version-selected="loadUpdateWizard"
  />
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from "@/helpers/formatter";
import { useEnvironmentDetail } from "@/composables/useEnvironmentDetail";
import { useEnvironmentChangelogModal } from "@/composables/useEnvironmentChangelogModal";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import { ref, computed } from "vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { useAlert } from "@/composables/useAlert";
import { sumChanges } from "@/helpers/changelog";

type Extension = components["schemas"]["EnvironmentExtension"];

type ExtensionWithCompatibility = Extension & {
  compatibility?: {
    name: string;
    label: string;
    type: string;
  };
};

const { error, success } = useAlert();
const { environment, shopwareVersions, latestShopwareVersion } = useEnvironmentDetail();

async function copyToken() {
  if (environment.value?.environmentToken) {
    await navigator.clipboard.writeText(environment.value.environmentToken);
    success("Token copied to clipboard");
  }
}

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

const {
  viewEnvironmentChangelogDialog,
  dialogEnvironmentChangelog,
  openEnvironmentChangelog,
  closeEnvironmentChangelog,
} = useEnvironmentChangelogModal();

// Computed properties for the overview panels
const allCriticalChecks = computed(() => {
  if (!environment.value?.checks) return [];
  return environment.value.checks.filter(
    (check) => check.level !== "green" && !environment.value?.ignores?.includes(check.id),
  );
});

const sortedCriticalChecks = computed(() => {
  return [...allCriticalChecks.value]
    .sort((a, b) => {
      if (a.level === "red" && b.level !== "red") return -1;
      if (a.level !== "red" && b.level === "red") return 1;
      if (a.level === "yellow" && b.level === "green") return -1;
      if (a.level === "green" && b.level === "yellow") return 1;
      return 0;
    })
    .slice(0, 5); // Limit to first 10
});

const outdatedExtensions = computed(() => {
  if (!environment.value?.extensions) return [];
  return environment.value.extensions
    .filter((ext) => ext.installed && ext.latestVersion && ext.version !== ext.latestVersion)
    .slice(0, 5);
});

const overdueTasks = computed(() => {
  if (!environment.value?.scheduledTasks) return [];
  const fifteenMinutesAgo = new Date(Date.now() - 15 * 60 * 1000);

  return environment.value.scheduledTasks
    .filter(
      (task) =>
        task.overdue &&
        task.status !== "inactive" &&
        task.nextExecutionTime &&
        new Date(task.nextExecutionTime) < fifteenMinutesAgo,
    )
    .slice(0, 5);
});

const recentChangelogs = computed(() => {
  if (!environment.value?.changelogs) return [];
  return environment.value.changelogs.slice(0, 5);
});

function getOverdueTime(nextExecutionTime: string): string {
  const now = new Date();
  const scheduledTime = new Date(nextExecutionTime);
  const diffMs = now.getTime() - scheduledTime.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));

  if (diffMinutes < 60) {
    return `${diffMinutes} minutes`;
  }

  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) {
    return `${diffHours} hours`;
  }

  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays} days`;
}

// For update wizard
const viewUpdateWizardDialog = ref(false);
const loadingUpdateWizard = ref(false);
const dialogUpdateWizard = ref<ExtensionWithCompatibility[] | null>(null);

function openUpdateWizard() {
  dialogUpdateWizard.value = null;
  viewUpdateWizardDialog.value = true;
}

async function loadUpdateWizard(version: string) {
  if (!environment.value?.extensions) {
    return;
  }

  loadingUpdateWizard.value = true;

  const body = {
    currentVersion: environment.value?.shopwareVersion,
    futureVersion: version,
    extensions: environment.value.extensions.map((extension) => {
      return {
        name: extension.name,
        version: extension.version,
      };
    }),
  };

  try {
    const { data: pluginCompatibility } = await api.POST("/info/extension-compatibility", {
      body,
    });

    if (!pluginCompatibility) return;

    const extensions = JSON.parse(
      JSON.stringify(environment.value?.extensions),
    ) as ExtensionWithCompatibility[];

    for (const extension of extensions) {
      const compatibility = pluginCompatibility.find((plugin) => plugin.name === extension.name);
      extension.compatibility = undefined;
      if (compatibility) {
        extension.compatibility =
          compatibility.status as ExtensionWithCompatibility["compatibility"];
      }
    }

    dialogUpdateWizard.value = extensions.sort((a, b) => {
      // Sort by active status (active first)
      if (a.active !== b.active) {
        return a.active ? -1 : 1;
      }
      // Then sort by label
      return a.label.localeCompare(b.label);
    });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  } finally {
    loadingUpdateWizard.value = false;
  }
}
</script>

<style scoped>
.shop-info-heading {
  padding: 1.25rem 0;
  font-size: 1.125rem;
  font-weight: 500;

  .icon {
    margin-right: 0.25rem;
  }
}

.shop-info-grid {
  padding: 1.25rem 0;
}

.shop-info-list {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  grid-auto-rows: min-content;
  gap: 1rem 1.5rem;

  @media (min-width: 640px) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (min-width: 960px) {
    grid-column: 1 / span 2;
    grid-template-columns: repeat(3, 1fr);
  }
}

.shop-info-item {
  dt {
    font-weight: 500;

    .icon {
      color: var(--info-color);
      font-size: 0.875rem;
      margin-left: 0.25rem;
    }
  }

  dd {
    color: var(--text-color-muted);
  }
}

.shop-token-item {
  grid-column: 1 / -1;
}

.shop-token-value {
  display: flex;
  align-items: center;
  gap: 0.5rem;

  code {
    font-size: 0.8rem;
    word-break: break-all;
  }
}

.deployment-info {
  display: flex;
  flex-direction: column;
}

.auto-update-info {
  margin-top: 0.375rem;
  font-size: 0.75rem;
  color: var(--text-color-muted);
  opacity: 0.8;
  font-style: italic;
}

.issue-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.issue-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.5rem 0;

  @media all and (min-width: 640px) {
    align-items: center;
  }

  > .icon {
    margin-top: 0.25rem;
    flex-shrink: 0;

    @media all and (min-width: 640px) {
      margin-top: unset;
    }
  }
}

.issue-content {
  display: flex;
  gap: 0.5rem;
  flex-direction: column;

  @media all and (min-width: 640px) {
    align-items: center;
    flex-direction: unset;
  }
}

.issue-source {
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.issue-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;

  .icon {
    font-size: 0.7rem;
  }
}

.issue-more {
  margin-top: 1rem;
  text-align: center;
}

.change-item {
  display: flex;
  align-items: flex-start;
  padding: 0.5rem 0;
  gap: 1rem;

  .icon {
    font-size: 0.75rem;
  }
}

.change-date {
  padding-top: 0.175em;
  font-size: 0.875rem;
  color: var(--text-color-muted);
}
</style>
