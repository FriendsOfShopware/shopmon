<template>
  <header-container v-if="shop" :title="$t('shop.editShop', { name: shop.name })">
    <router-link
      :to="{
        name: 'account.shops.detail',
        params: {
          organizationId: route.params.organizationId as string,
          shopId: shop.id,
        },
      }"
      type="button"
      class="btn"
    >
      {{ $t("common.cancel") }}
    </router-link>
  </header-container>

  <main-container v-if="shop">
    <Panel>
      <vee-form
        ref="formRef"
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="shop"
        @submit="onSubmit"
      >
        <form-group :title="$t('shop.shopInfo')">
          <InputField name="name" :label="$t('common.name')" autocomplete="name" :error="errors.name" />

          <BaseSelect
            v-model="selectedProjectId"
            name="projectId"
            :label="$t('shop.project')"
            :error="errors.projectId"
          >
            <option v-for="project in projects" :key="project.id" :value="project.id">
              {{ project.organizationName + " / " + project.name }}
            </option>
          </BaseSelect>

          <InputField name="shopUrl" :label="$t('common.url')" autocomplete="url" :error="errors.shopUrl" />
        </form-group>

        <form-group :title="$t('shop.integration')">
          <template #info>
            <i18n-t keypath="shop.integrationDesc" tag="span">
              <template #pluginLink>
                <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank">{{
                  $t("shop.pluginName")
                }}</a>
              </template>
            </i18n-t>
            <a
              href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions"
            >
              {{ $t("shop.permissions") }}
            </a>
          </template>

          <button type="button" class="btn btn-secondary" @click="openPluginModal">
            {{ $t("shop.connectPlugin") }}
          </button>

          <InputField name="clientId" :label="$t('shop.clientId')" :error="errors.clientId" />
          <InputField name="clientSecret" :label="$t('shop.clientSecret')" :error="errors.clientSecret" />
        </form-group>

        <div class="form-submit">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </button>
        </div>
      </vee-form>
    </Panel>

    <Panel v-if="shop" :title="$t('shop.sitespeed')">
      <p>
        {{ $t("shop.sitespeedDesc") }}
      </p>
      <p>
        {{ $t("shop.sitespeedActivateDesc") }}
      </p>
      <p>
        {{ $t("shop.sitespeedScheduleDesc") }}
      </p>

      <form @submit.prevent="onSitespeedSubmit">
        <div class="mb-1">
          <label for="sitespeedEnabled">{{ $t("shop.sitespeedEnabled") }}</label>
          <input id="sitespeedEnabled" v-model="sitespeedEnabled" type="checkbox" class="field" />
        </div>

        <div v-if="sitespeedEnabled">
          <label for="sitespeedUrls">{{ $t("shop.sitespeedUrls") }}</label>

          <div class="sitespeed-urls-container">
            <div v-for="(url, index) in sitespeedUrls" :key="index" class="sitespeed-url-row">
              <BaseInput
                :model-value="sitespeedUrls[index]"
                type="url"
                placeholder="https://example.com"
                @update:model-value="sitespeedUrls[index] = $event"
              >
                <template #append>
                  <button type="button" class="btn btn-icon" @click="removeSitespeedUrl(index)">
                    <icon-fa6-solid:xmark />
                  </button>
                </template>
              </BaseInput>
            </div>

            <button
              v-if="sitespeedUrls.length < 5"
              type="button"
              class="btn btn-secondary"
              @click="addSitespeedUrl"
            >
              <icon-fa6-solid:plus class="icon" />
              {{ $t("shop.newUrl") }}
            </button>
          </div>
        </div>

        <div class="form-submit">
          <button
            :disabled="isSitespeedSubmitting || !isSitespeedFormValid"
            type="submit"
            class="btn btn-primary"
          >
            <icon-fa6-solid:floppy-disk
              v-if="!isSitespeedSubmitting"
              class="icon"
              aria-hidden="true"
            />

            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("shop.saveSitespeedSettings") }}
          </button>
        </div>
      </form>
    </Panel>

    <Panel :title="$t('shop.deleteShopTitle', { name: shop.name })">
      <p>{{ $t("shop.deleteShopWarning") }}</p>

      <button type="button" class="btn btn-danger" @click="showShopDeletionModal = true">
        <icon-fa6-solid:trash class="icon icon-delete" />
        {{ $t("shop.deleteShop") }}
      </button>
    </Panel>

    <delete-confirmation-modal
      :show="showShopDeletionModal"
      :title="$t('shop.deleteShop')"
      :entity-name="shop?.name || 'this shop'"
      @close="showShopDeletionModal = false"
      @confirm="deleteShop"
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
const shop = ref<components["schemas"]["ShopDetail"] | null>(null);
const isLoading = ref(false);
const projects = ref<components["schemas"]["AccountProject"][]>([]);
const selectedProjectId = ref<number>(0);

const shopId = Number.parseInt(route.params.shopId as string, 10);

api.GET("/account/projects").then(({ data }) => {
  if (data) projects.value = data;
});

async function loadShop() {
  isLoading.value = true;
  const { data } = await api.GET("/shops/{shopId}", {
    params: { path: { shopId } },
  });
  shop.value = data ?? null;

  // Load projects for the organization
  if (shop.value?.organizationId) {
    const { data: projectsData } = await api.GET("/account/projects");
    if (projectsData) projects.value = projectsData;
    selectedProjectId.value = shop.value.projectId ?? 0;
  }

  // Initialize sitespeed settings
  if (shop.value) {
    sitespeedEnabled.value = shop.value.sitespeedEnabled ?? false;
    sitespeedUrls.value = shop.value.sitespeedUrls ? [...shop.value.sitespeedUrls] : [];
  }

  isLoading.value = false;
}

loadShop();

const showShopDeletionModal = ref(false);
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
  name: Yup.string().required(t("validation.shopNameRequired")),
  url: Yup.string().required(t("validation.shopUrlRequired")).url(),
  projectId: Yup.number().required(t("validation.projectRequired")),
  clientId: Yup.string().when("url", {
    is: (url: string) => url !== shop.value?.url,
    // eslint-disable-next-line unicorn/no-thenable -- Valid Yup schema method
    then: () => Yup.string().required(t("validation.urlChangeClientId")),
  }),
  clientSecret: Yup.string().when("url", {
    is: (url: string) => url !== shop.value?.url,
    // eslint-disable-next-line unicorn/no-thenable -- Valid Yup schema method
    then: () => Yup.string().required(t("validation.urlChangeClientSecret")),
  }),
});

async function onSubmit(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  if (shop.value) {
    try {
      if (typedValues.url) {
        typedValues.url = typedValues.url.replace(/\/+$/, "");
      }
      await api.PATCH("/shops/{shopId}", {
        params: { path: { shopId: shop.value.id } },
        body: {
          name: typedValues.name,
          shopUrl: typedValues.url,
          clientId: typedValues.clientId,
          clientSecret: typedValues.clientSecret,
          projectId: selectedProjectId.value,
        },
      });

      router.push({
        name: "account.shops.detail",
        params: {
          organizationId: route.params.organizationId as string,
          shopId: shop.value.id,
        },
      });
    } catch (e) {
      error(e instanceof Error ? e.message : String(e));
    }
  }
}

async function deleteShop() {
  if (shop.value) {
    try {
      await api.DELETE("/shops/{shopId}", {
        params: { path: { shopId: shop.value.id } },
      });

      router.push({ name: "account.project.list" });
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
  if (shop.value) {
    try {
      isSitespeedSubmitting.value = true;

      // Validate that if enabled, at least one URL is provided
      if (sitespeedEnabled.value && sitespeedUrls.value.length === 0) {
        error(t("shop.sitespeedUrlRequired"));
        return;
      }

      await api.PUT("/shops/{shopId}/sitespeed-settings", {
        params: { path: { shopId: shop.value.id } },
        body: {
          enabled: sitespeedEnabled.value,
          urls: sitespeedUrls.value,
        },
      });

      // Reload shop data to refresh the UI
      await loadShop();
      success(t("shop.sitespeedSaved"));
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
      pluginError.value = t("shop.base64Error");
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = t("shop.base64InvalidData");
      return;
    }

    formRef.value.setFieldValue("url", data.url);
    formRef.value.setFieldValue("clientId", data.clientId);
    formRef.value.setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("shop.base64InvalidFormat");
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

.btn-icon {
  padding: 0.5rem;
  min-width: auto;
}
</style>
