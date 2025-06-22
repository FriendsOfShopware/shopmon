<template>
    <div v-if="shop" class="shop-info panel">
        <h3 class="panel-title">
            <status-icon :status="shop.status" />
            Shop Information
        </h3>

        <div class="shop-info-grid">
            <dl class="shop-info-list">
                <div class="shop-info-item">
                    <dt>
                        Shopware Version
                    </dt>
                    <dd>
                        <span>{{ shop.shopwareVersion }}</span>
                        <template v-if="latestShopwareVersion && latestShopwareVersion != shop.shopwareVersion">
                            <a
                                class="badge badge-warning"
                                :href="'https://github.com/shopware/platform/releases/tag/v' + latestShopwareVersion"
                                target="_blank"
                            >
                                {{ latestShopwareVersion }}
                            </a>

                            <button
                                class="badge badge-info"
                                type="button"
                                @click="openUpdateWizard"
                            >
                                <FaRotate />
                                Compatibility Check
                            </button>
                        </template>
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        Last Shop Update
                    </dt>

                    <dd>
                        <template v-if="shop.lastChangelog && shop.lastChangelog.date">
                            {{ formatDate(shop.lastChangelog.date) }}
                        </template>

                        <template v-else>
                            never
                        </template>
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        Organization
                    </dt>

                    <dd>
                        {{ shop.organizationName }}
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        Last Checked At
                    </dt>

                    <dd>
                        {{ formatDateTime(shop.lastScrapedAt) }}
                        <div class="auto-update-info">
                            Shop is updated once an hour automatically
                        </div>
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        Environment
                    </dt>

                    <dd>
                        {{ shop.cacheInfo?.environment }}
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        HTTP Cache
                    </dt>

                    <dd>
                        {{ shop.cacheInfo?.httpCache ? 'Enabled' : 'Disabled' }}
                    </dd>
                </div>

                <div class="shop-info-item">
                    <dt>
                        URL
                    </dt>

                    <dd>
                        <a
                            :href="shop.url"
                            data-tooltip="Go to storefront"
                            target="_blank"
                        >
                            <FaStore /> Storefront
                        </a>
                        &nbsp;/&nbsp;
                        <a
                            :href="shop.url + '/admin'"
                            data-tooltip="Go to shopware admin"
                            target="_blank"
                        >
                            <FaShieldHalved /> Admin
                        </a>
                    </dd>
                </div>
            </dl>
        </div>
    </div>

    <modal
        class="update-wizard"
        :show="viewUpdateWizardDialog"
        close-x-mark
        @close="viewUpdateWizardDialog = false"
    >
        <template #title>
            <FaRotate /> Shopware Extension Compatibility Check
        </template>

        <template #content>
            <select
                class="field"
                @change="(event: Event) => loadUpdateWizard((event.target as HTMLSelectElement).value)"
            >
                <option disabled selected>
                    Select update Version
                </option>

                <option
                    v-for="version in shopwareVersions"
                    :key="version"
                >
                    {{ version }}
                </option>
            </select>

            <template v-if="loadingUpdateWizard">
                <div class="update-wizard-loader">
                    Loading <FaRotate class="animate-spin" />
                </div>
            </template>

            <div
                v-if="dialogUpdateWizard"
                :class="{ 'update-wizard-refresh': loadingUpdateWizard }"
            >
                <h2 class="update-wizard-plugins-heading">
                    Extension Compatibility
                </h2>

                <ul>
                    <li
                        v-for="extension in dialogUpdateWizard"
                        :key="extension.name"
                        class="update-wizard-plugin"
                    >
                        <div class="update-wizard-plugin-icon">
                            <FaRegularCircle
                                v-if="!extension.active"
                                class="icon icon-muted"
                            />
                            <FaCircleInfo
                                v-else-if="!extension.compatibility"
                                class="icon icon-warning"
                            />
                            <FaCircleXmark
                                v-else-if="extension.compatibility.type == 'red'"
                                class="icon icon-error"
                            />
                            <FaRotate
                                v-else-if="extension.compatibility.label === 'Available now'"
                                class="icon icon-info"
                            />
                            <FaCircleCheck
                                v-else
                                class="icon icon-success"
                            />
                        </div>

                        <div>
                            <component
                                :is="extension.storeLink ? 'a' : 'span'"
                                v-bind="extension.storeLink ? {href: extension.storeLink, target: '_blank'} : {}"
                            >
                                <strong>{{ extension.label }}</strong>
                            </component>
                            <span class="update-wizard-plugin-technical-name"> ({{ extension.name }})</span>

                            <div v-if="!extension.compatibility || !extension.storeLink">
                                This plugin is not available in the Store. Please contact the
                                plugin manufacturer.
                            </div>
                            <div v-else>
                                {{ extension.compatibility.label }}
                            </div>
                        </div>
                    </li>
                </ul>
            </div>
        </template>
    </modal>
</template>

<script setup lang="ts">
import { formatDate, formatDateTime } from '@/helpers/formatter';
import { useShopDetail } from '@/composables/useShopDetail';
import { ref } from 'vue';
import { trpcClient } from '@/helpers/trpc';
import { useAlert } from '@/composables/useAlert';
import Modal from '@/components/layout/Modal.vue';
import StatusIcon from '@/components/StatusIcon.vue';

// Icon imports
import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaCircleInfo from '~icons/fa6-solid/circle-info';
import FaCircleXmark from '~icons/fa6-solid/circle-xmark';
import FaRegularCircle from '~icons/fa6-regular/circle';
import FaRotate from '~icons/fa6-solid/rotate';
import FaStore from '~icons/fa6-solid/store';
import FaShieldHalved from '~icons/fa6-solid/shield-halved';

const { error } = useAlert();
const {
    shop,
    shopwareVersions,
    latestShopwareVersion,
} = useShopDetail();

// For update wizard
const viewUpdateWizardDialog = ref(false);
const loadingUpdateWizard = ref(false);
const dialogUpdateWizard = ref<any[] | null>(null);

function openUpdateWizard() {
  dialogUpdateWizard.value = null;
  viewUpdateWizardDialog.value = true;
}

async function loadUpdateWizard(version: string) {
  if (!shop.value || !shop.value.extensions) {
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
      const compatibility = pluginCompatibility.find(
        (plugin) => plugin.name === extension.name,
      );
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
    margin-bottom: 3rem;

    &-heading {
        padding: 1.25rem 0;
        font-size: 1.125rem;
        font-weight: 500;

        .icon {
            margin-right: 0.25rem;
        }
    }
}

.shop-info-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    padding: 1.25rem 0;

    @media (min-width: 640px) {
        grid-template-columns: repeat(2, 1fr);
    }
}

.shop-info-list {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;
    gap: 0.5rem 1.5rem;

    @media (min-width: 960px) {
        grid-column: 1 / span 2;
        grid-template-columns: repeat(2, 1fr);
    }
}

.shop-info-item {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    grid-auto-rows: min-content;

    @media (min-width: 960px) {
        grid-column: span 1;
    }

    dt {
        font-size: 0.875rem;
        font-weight: 500;
    }

    dd {
        margin-top: 0.25rem;
        font-size: 0.875rem;
        color: var(--text-color-muted);
    }
}

.update-wizard {
    .field {
        margin-bottom: .75rem;
    }

    &-loader {
        text-align:center;
    }

    &-refresh {
        opacity: .2;
    }

    &-plugins {
        &-heading {
            font-size: 1.125rem;
            font-weight: 500;
            margin-bottom: .5rem;
        }
    }

    &-plugin {
        background-color: var(--item-background);
        padding: .5rem;
        display: flex;

        &:nth-child(odd) {
            background-color: var(--item-odd-background);
        }

        &:hover {
            background-color: var(--item-hover-background);
        }

        &-icon {
            margin-right: .5rem;
        }

        &-technical-name {
            opacity: .6;
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
</style>
