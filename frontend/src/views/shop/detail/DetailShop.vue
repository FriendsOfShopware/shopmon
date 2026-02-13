<template>
  <div v-if="shop" class="shop-info panel">
    <h3 class="panel-title">
      <icon-fa6-solid:circle-info class="icon" />
      Shop Information
    </h3>

    <div class="shop-info-grid">
      <dl class="shop-info-list">
        <div class="shop-info-item">
          <dt>Shopware Version</dt>
          <dd>
            <span>{{ shop.shopwareVersion }}</span>
            <template v-if="latestShopwareVersion && latestShopwareVersion != shop.shopwareVersion">
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
                Compatibility Check
              </button>
            </template>
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>
            Last Checked At
            <span
              data-tooltip="Shop is updated once an hour automatically"
              class="tooltip-top-center"
            >
              <icon-fa6-solid:circle-info class="icon" />
            </span>
          </dt>

          <dd>
            {{ formatDateTime(shop.lastScrapedAt) }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Last Shop Update</dt>

          <dd>
            <template v-if="shop.lastChangelog && shop.lastChangelog.date">
              {{ formatDate(shop.lastChangelog.date) }}
            </template>

            <template v-else> never </template>
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Environment</dt>

          <dd>
            {{ shop.cacheInfo?.environment }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Cache Adapter</dt>

          <dd>
            {{ shop.cacheInfo?.cacheAdapter }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>HTTP Cache</dt>

          <dd>
            {{ shop.cacheInfo?.httpCache ? "Enabled" : "Disabled" }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Organization</dt>

          <dd>
            {{ shop.organizationName }}
          </dd>
        </div>

        <div class="shop-info-item">
          <dt>Project</dt>

          <dd>
            {{ shop.projectName }}
          </dd>
        </div>
      </dl>
    </div>
  </div>

  <div class="panel">
    <h3 class="panel-title">
      <icon-fa6-solid:circle-check class="icon" />
      Security & Health Checks
    </h3>

    <Alert v-if="sortedCriticalChecks.length === 0" type="success">
      All security checks passed
    </Alert>

    <ul v-else class="issue-list">
      <li v-for="check in sortedCriticalChecks" :key="check.id" class="issue-item">
        <status-icon :status="check.level" />

        <div class="issue-content">
          <span class="issue-message">{{ check.message }}</span>

          <a v-if="check.link" :href="check.link" target="_blank" class="issue-link">
            Learn more
            <icon-fa6-solid:arrow-up-right-from-square class="icon" />
          </a>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.shops.detail.checks',
          params: {
            slug: $route.params.slug,
            shopId: $route.params.shopId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:circle-check class="icon" />
        View all checks
      </router-link>
    </div>
  </div>

  <div class="panel">
    <h3 class="panel-title">
      <icon-fa6-solid:plug class="icon" />
      Extensions
    </h3>

    <Alert v-if="outdatedExtensions.length === 0" type="success">
      All extensions are up to date
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
            Store
            <icon-fa6-solid:arrow-up-right-from-square class="icon" />
          </a>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.shops.detail.extensions',
          params: {
            slug: $route.params.slug,
            shopId: $route.params.shopId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:plug class="icon" />
        View all extensions
      </router-link>
    </div>
  </div>

  <div class="panel">
    <h3 class="panel-title">
      <icon-fa6-solid:list-check class="icon" />
      Scheduled Tasks
    </h3>

    <Alert v-if="overdueTasks.length === 0" type="success"> No overdue scheduled tasks </Alert>

    <ul v-else class="issue-list">
      <li v-for="task in overdueTasks" :key="task.id" class="issue-item">
        <icon-fa6-solid:clock class="icon icon-warning" />
        <div class="issue-content">
          <span class="issue-message">{{ task.name }}</span>
          <span class="issue-source">
            Last run: {{ formatDateTime(task.lastExecutionTime) }} ({{
              getOverdueTime(task.nextExecutionTime)
            }}
            overdue)
          </span>
        </div>
      </li>
    </ul>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.shops.detail.tasks',
          params: {
            slug: $route.params.slug,
            shopId: $route.params.shopId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:list-check class="icon" />
        View all scheduled tasks
      </router-link>
    </div>
  </div>

  <div class="panel">
    <h3 class="panel-title">
      <icon-fa6-solid:clock-rotate-left class="icon" />
      Recent Changes
    </h3>

    <Alert v-if="recentChangelogs.length === 0" type="info"> No recent changes recorded </Alert>

    <div v-else class="changes-list">
      <div v-for="changelog in recentChangelogs" :key="changelog.id" class="change-item">
        <div class="change-date">
          {{ formatDate(changelog.date) }}
        </div>

        <div class="change-summary">
          {{ sumChanges(changelog) }}
          <span
            class="link tooltip-top-left"
            data-tooltip="Open Details"
            @click="openShopChangelog(changelog)"
            ><icon-fa6-solid:circle-info class="icon"
          /></span>
        </div>
      </div>
    </div>

    <div class="issue-more">
      <router-link
        :to="{
          name: 'account.shops.detail.changelog',
          params: {
            slug: $route.params.slug,
            shopId: $route.params.shopId,
          },
        }"
        class="btn btn-sm"
      >
        <icon-fa6-solid:file-waveform class="icon" />
        View all changes
      </router-link>
    </div>
  </div>

  <!-- Shop Changelog Modal -->
  <shop-changelog
    :show="viewShopChangelogDialog"
    :changelog="dialogShopChangelog"
    @close="closeShopChangelog"
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
import { useShopDetail } from "@/composables/useShopDetail";
import { useShopChangelogModal } from "@/composables/useShopChangelogModal";
import { useExtensionChangelogModal } from "@/composables/useExtensionChangelogModal";
import { ref, computed } from "vue";
import { trpcClient } from "@/helpers/trpc";
import { useAlert } from "@/composables/useAlert";
import { sumChanges } from "@/helpers/changelog";

const { error } = useAlert();
const { shop, shopwareVersions, latestShopwareVersion } = useShopDetail();

const {
  viewExtensionChangelogDialog,
  dialogExtension,
  openExtensionChangelog,
  closeExtensionChangelog,
} = useExtensionChangelogModal();

const { viewShopChangelogDialog, dialogShopChangelog, openShopChangelog, closeShopChangelog } =
  useShopChangelogModal();

// Computed properties for the overview panels
const allCriticalChecks = computed(() => {
  if (!shop.value?.checks) return [];
  return shop.value.checks.filter(
    (check) => check.level !== "green" && !shop.value?.ignores?.includes(check.id),
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
  if (!shop.value?.extensions) return [];
  return shop.value.extensions
    .filter((ext) => ext.installed && ext.latestVersion && ext.version !== ext.latestVersion)
    .slice(0, 5);
});

const overdueTasks = computed(() => {
  if (!shop.value?.scheduledTask) return [];
  const fifteenMinutesAgo = new Date(Date.now() - 15 * 60 * 1000);

  return shop.value.scheduledTask
    .filter((task) => task.overdue && new Date(task.nextExecutionTime) < fifteenMinutesAgo)
    .slice(0, 5);
});

const recentChangelogs = computed(() => {
  if (!shop.value?.changelog) return [];
  return shop.value.changelog.slice(0, 5);
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
const dialogUpdateWizard = ref<any[] | null>(null);

function openUpdateWizard() {
  dialogUpdateWizard.value = null;
  viewUpdateWizardDialog.value = true;
}

async function loadUpdateWizard(version: string) {
  if (!shop.value?.extensions) {
    return;
  }

  loadingUpdateWizard.value = true;

  const body = {
    currentVersion: shop.value?.shopwareVersion,
    futureVersion: version,
    extensions: shop.value.extensions.map((extension) => {
      return {
        name: extension.name,
        version: extension.version,
      };
    }),
  };

  try {
    const pluginCompatibility = await trpcClient.info.checkExtensionCompatibility.query(body);

    const extensions = JSON.parse(JSON.stringify(shop.value?.extensions)) as any[];

    for (const extension of extensions) {
      const compatibility = pluginCompatibility.find((plugin) => plugin.name === extension.name);
      extension.compatibility = null;
      if (compatibility) {
        extension.compatibility = compatibility.status;
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
.shop-info {
  &-heading {
    padding: 1.25rem 0;
    font-size: 1.125rem;
    font-weight: 500;

    .icon {
      margin-right: 0.25rem;
    }
  }

  &-grid {
    padding: 1.25rem 0;
  }

  &-list {
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

  &-item {
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
}

.auto-update-info {
  margin-top: 0.375rem;
  font-size: 0.75rem;
  color: var(--text-color-muted);
  opacity: 0.8;
  font-style: italic;
}

.issue {
  &-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  &-item {
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

  &-content {
    display: flex;
    gap: 0.5rem;
    flex-direction: column;

    @media all and (min-width: 640px) {
      align-items: center;
      flex-direction: unset;
    }
  }

  &-source {
    font-size: 0.875rem;
    color: var(--text-color-muted);
  }

  &-link {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.875rem;

    .icon {
      font-size: 0.7rem;
    }
  }

  &-more {
    margin-top: 1rem;
    text-align: center;
  }
}

.change {
  &-item {
    display: flex;
    align-items: flex-start;
    padding: 0.5rem 0;
    gap: 1rem;

    .icon {
      font-size: 0.75rem;
    }
  }

  &-date {
    padding-top: 0.175em;
    font-size: 0.875rem;
    color: var(--text-color-muted);
  }
}
</style>
