<template>
  <header-container
    v-if="environment"
    :title="$t('environment.editEnvironment', { name: environment.name })"
  >
    <UiButton
      :to="{
        name: 'account.environments.detail',
        params: {
          organizationId: route.params.organizationId as string,
          environmentId: environment.id,
        },
      }"
      type="button"
    >
      {{ $t("common.cancel") }}
    </UiButton>
  </header-container>

  <main-container v-if="environment">
    <Panel>
      <vee-form
        ref="formRef"
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="environment"
        @submit="onSubmit"
      >
        <form-group :title="$t('environment.environmentInfo')">
          <InputField
            name="name"
            :label="$t('common.name')"
            autocomplete="name"
            :error="errors.name"
          />

          <BaseSelect
            v-model="selectedShopId"
            name="shopId"
            :label="$t('environment.shop')"
            :error="errors.shopId"
          >
            <option v-for="shop in shops" :key="shop.id" :value="shop.id">
              {{ shop.organizationName + " / " + shop.name }}
            </option>
          </BaseSelect>

          <InputField
            name="shopUrl"
            :label="$t('common.url')"
            autocomplete="url"
            :error="errors.shopUrl"
          />
        </form-group>

        <form-group :title="$t('environment.integration')">
          <template #info>
            <i18n-t keypath="environment.integrationDesc" tag="span">
              <template #pluginLink>
                <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank">{{
                  $t("environment.pluginName")
                }}</a>
              </template>
            </i18n-t>
            <a
              href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions"
            >
              {{ $t("environment.permissions") }}
            </a>
          </template>

          <UiButton type="button" @click="openPluginModal">
            {{ $t("environment.connectPlugin") }}
          </UiButton>

          <InputField
            name="clientId"
            :label="$t('environment.clientId')"
            :error="errors.clientId"
          />
          <InputField
            name="clientSecret"
            :label="$t('environment.clientSecret')"
            :error="errors.clientSecret"
          />
        </form-group>

        <div class="form-submit">
          <UiButton :disabled="isSubmitting" type="submit" variant="primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </UiButton>
        </div>
      </vee-form>
    </Panel>

    <Panel v-if="environment" :title="$t('environment.sitespeed')">
      <p>
        {{ $t("environment.sitespeedDesc") }}
      </p>
      <p>
        {{ $t("environment.sitespeedActivateDesc") }}
      </p>
      <p>
        {{ $t("environment.sitespeedScheduleDesc") }}
      </p>

      <form @submit.prevent="onSitespeedSubmit">
        <div class="mb-1">
          <label for="sitespeedEnabled">{{ $t("environment.sitespeedEnabled") }}</label>
          <input id="sitespeedEnabled" v-model="sitespeedEnabled" type="checkbox" class="field" />
        </div>

        <div v-if="sitespeedEnabled">
          <label for="sitespeedUrls">{{ $t("environment.sitespeedUrls") }}</label>

          <div class="sitespeed-urls-container">
            <div v-for="(url, index) in sitespeedUrls" :key="index" class="sitespeed-url-row">
              <BaseInput
                :model-value="sitespeedUrls[index]"
                type="url"
                placeholder="https://example.com"
                @update:model-value="sitespeedUrls[index] = $event"
              >
                <template #append>
                  <UiButton type="button" icon-only @click="removeSitespeedUrl(index)">
                    <icon-fa6-solid:xmark />
                  </UiButton>
                </template>
              </BaseInput>
            </div>

            <UiButton
              v-if="sitespeedUrls.length < 5"
              type="button"
              @click="addSitespeedUrl"
            >
              <icon-fa6-solid:plus class="icon" />
              {{ $t("environment.newUrl") }}
            </UiButton>
          </div>
        </div>

        <div class="form-submit">
          <UiButton
            :disabled="isSitespeedSubmitting || !isSitespeedFormValid"
            type="submit"
            variant="primary"
          >
            <icon-fa6-solid:floppy-disk
              v-if="!isSitespeedSubmitting"
              class="icon"
              aria-hidden="true"
            />

            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("environment.saveSitespeedSettings") }}
          </UiButton>
        </div>
      </form>
    </Panel>

    <Panel :title="$t('environment.deleteEnvironmentTitle', { name: environment.name })">
      <p>{{ $t("environment.deleteEnvironmentWarning") }}</p>

      <UiButton
        type="button"
        variant="destructive"
        @click="showEnvironmentDeletionModal = true"
      >
        <icon-fa6-solid:trash class="icon icon-delete" />
        {{ $t("environment.deleteEnvironment") }}
      </UiButton>
    </Panel>

    <delete-confirmation-modal
      :show="showEnvironmentDeletionModal"
      :title="$t('environment.deleteEnvironment')"
      :entity-name="environment?.name || 'this environment'"
      @close="showEnvironmentDeletionModal = false"
      @confirm="deleteEnvironment"
    />

    <!-- Plugin Connection Modal -->
    <plugin-connection-modal
      v-model:base64="pluginBase64"
      :show="showPluginModal"
      :error="pluginError"
      @close="closePluginModal"
      @import="processPluginData"
    />
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";

import { Form as VeeForm } from "vee-validate";
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { error, success } = useAlert();
const router = useRouter();
const route = useRoute();
const environment = ref<components["schemas"]["EnvironmentDetail"] | null>(null);
const isLoading = ref(false);
const shops = ref<components["schemas"]["AccountShop"][]>([]);
const selectedShopId = ref<number>(0);

const environmentId = Number.parseInt(route.params.environmentId as string, 10);

api.GET("/account/shops").then(({ data }) => {
  if (data) shops.value = data;
});

async function loadEnvironment() {
  isLoading.value = true;
  const { data } = await api.GET("/environments/{environmentId}", {
    params: { path: { environmentId } },
  });
  environment.value = data ?? null;

  // Load shops for the organization
  if (environment.value?.organizationId) {
    const { data: shopsData } = await api.GET("/account/shops");
    if (shopsData) shops.value = shopsData;
    selectedShopId.value = environment.value.shopId ?? 0;
  }

  // Initialize sitespeed settings
  if (environment.value) {
    sitespeedEnabled.value = environment.value.sitespeedEnabled ?? false;
    sitespeedUrls.value = environment.value.sitespeedUrls
      ? [...environment.value.sitespeedUrls]
      : [];
  }

  isLoading.value = false;
}

loadEnvironment();

const showEnvironmentDeletionModal = ref(false);
const sitespeedUrls = ref<string[]>([]);
const sitespeedEnabled = ref(false);
const isSitespeedSubmitting = ref(false);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

const formRef = ref();

const isSitespeedFormValid = computed(() => {
  if (!sitespeedEnabled.value) {
    return true; // Always valid when disabled
  }
  // Check if at least one non-empty URL exists
  const hasNonEmptyUrl = sitespeedUrls.value.some((url) => url.trim() !== "");
  return hasNonEmptyUrl;
});

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.environmentNameRequired")),
  url: Yup.string().required(t("validation.environmentUrlRequired")).url(),
  shopId: Yup.number().required(t("validation.shopRequired")),
  clientId: Yup.string().when("url", {
    is: (url: string) => url !== environment.value?.url,
    // eslint-disable-next-line unicorn/no-thenable -- Valid Yup schema method
    then: () => Yup.string().required(t("validation.urlChangeClientId")),
  }),
  clientSecret: Yup.string().when("url", {
    is: (url: string) => url !== environment.value?.url,
    // eslint-disable-next-line unicorn/no-thenable -- Valid Yup schema method
    then: () => Yup.string().required(t("validation.urlChangeClientSecret")),
  }),
});

async function onSubmit(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  if (environment.value) {
    try {
      if (typedValues.url) {
        typedValues.url = typedValues.url.replace(/\/+$/, "");
      }
      await api.PATCH("/environments/{environmentId}", {
        params: { path: { environmentId: environment.value.id } },
        body: {
          name: typedValues.name,
          shopUrl: typedValues.url,
          clientId: typedValues.clientId,
          clientSecret: typedValues.clientSecret,
          shopId: selectedShopId.value,
        },
      });

      router.push({
        name: "account.environments.detail",
        params: {
          organizationId: route.params.organizationId as string,
          environmentId: environment.value.id,
        },
      });
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    }
  }
}

async function deleteEnvironment() {
  if (environment.value) {
    try {
      await api.DELETE("/environments/{environmentId}", {
        params: { path: { environmentId: environment.value.id } },
      });

      router.push({ name: "account.shop.list" });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
}

function addSitespeedUrl() {
  sitespeedUrls.value.push("");
}

function removeSitespeedUrl(index: number) {
  sitespeedUrls.value.splice(index, 1);
}

async function onSitespeedSubmit() {
  if (environment.value) {
    try {
      isSitespeedSubmitting.value = true;

      // Validate that if enabled, at least one URL is provided
      if (sitespeedEnabled.value && sitespeedUrls.value.length === 0) {
        error(t("environment.sitespeedUrlRequired"));
        return;
      }

      await api.PUT("/environments/{environmentId}/sitespeed-settings", {
        params: { path: { environmentId: environment.value.id } },
        body: {
          enabled: sitespeedEnabled.value,
          urls: sitespeedUrls.value,
        },
      });

      // Reload environment data to refresh the UI
      await loadEnvironment();
      success(t("environment.sitespeedSaved"));
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    } finally {
      isSitespeedSubmitting.value = false;
    }
  }
}

function openPluginModal() {
  showPluginModal.value = true;
  pluginBase64.value = "";
  pluginError.value = "";
}

const closePluginModal = () => {
  showPluginModal.value = false;
  pluginBase64.value = "";
  pluginError.value = "";
};

function processPluginData() {
  try {
    pluginError.value = "";

    if (!pluginBase64.value.trim()) {
      pluginError.value = t("environment.base64Error");
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = t("environment.base64InvalidData");
      return;
    }

    formRef.value.setFieldValue("url", data.url);
    formRef.value.setFieldValue("clientId", data.clientId);
    formRef.value.setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("environment.base64InvalidFormat");
  }
}
</script>

<style>
.sitespeed-urls-container {
  margin-top: 0.5rem;
}

.sitespeed-url-row {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  align-items: center;
}

.sitespeed-url-row input {
  flex: 1;
}

.ui-button--icon-only {
  padding: 0.5rem;
  min-width: auto;
}
</style>
