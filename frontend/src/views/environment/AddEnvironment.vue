<template>
  <header-container :title="$t('environment.newEnvironment')" />
  <main-container>
    <Panel v-if="shops.length === 0 && shopsLoaded">
      <div class="empty-shop-state">
        <div class="empty-shop-icon">
          <icon-fa6-solid:folder-tree aria-hidden="true" />
        </div>

        <h3 class="empty-shop-title">Create a shop first</h3>

        <p class="empty-shop-description">
          Shops group related environments together. For example:
        </p>

        <div class="empty-shop-example">
          <div class="example-shop">
            <icon-fa6-solid:folder class="example-icon" aria-hidden="true" />
            <span class="example-shop-name">Toy Shop X</span>
          </div>
          <div class="example-environments">
            <div class="example-environment">
              <icon-fa6-solid:shop class="example-icon" aria-hidden="true" />
              Production
            </div>
            <div class="example-environment">
              <icon-fa6-solid:shop class="example-icon" aria-hidden="true" />
              Staging
            </div>
          </div>
        </div>

        <div class="empty-shop-cta">
          <UiButton
            :to="{ name: 'account.shops.new', query: { redirect: $route.fullPath } }"
            variant="primary"
          >
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Create Shop
          </UiButton>
        </div>
      </div>
    </Panel>

    <Panel v-else>
      <vee-form
        ref="formRef"
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="initialValues"
        @submit="onSubmit"
      >
        <form-group :title="$t('environment.environmentInfo')">
          <InputField name="name" :label="$t('common.name')" :error="errors.name" />

          <BaseSelect
            v-model="selectedShopId"
            name="shopId"
            :label="$t('environment.shop')"
            required
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

        <form-group :title="$t('environment.bypassAuthHeader')">
          <template #info>
            {{ $t("environment.bypassAuthHeaderDesc") }}
          </template>

          <div class="environment-token-display">
            <code>{{ shopToken }}</code>

            <UiButton type="button" size="sm" icon-only @click="copyToken">
              <icon-fa6-solid:copy />
            </UiButton>
          </div>
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
import { ref } from "vue";
import { useI18n } from "vue-i18n";

import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { error, success } = useAlert();
const router = useRouter();
const route = useRoute();

function generateShopToken(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32));
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

const shopToken = generateShopToken();

async function copyToken() {
  await navigator.clipboard.writeText(shopToken);
  success(t("environment.tokenCopied"));
}

const shops = ref<components["schemas"]["AccountShop"][]>([]);
const shopsLoaded = ref(false);
const selectedShopId = ref<number>(route.query.shopId ? Number(route.query.shopId) : 0);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

const formRef = ref();

api.GET("/account/shops").then(({ data }) => {
  if (data) {
    shops.value = data;
    if (!selectedShopId.value && data.length > 0) {
      selectedShopId.value = data[0].id;
    }
  }
  shopsLoaded.value = true;
});

const isValidUrl = (url: string) => {
  try {
    const parsedUrl = new URL(url);
    return !!parsedUrl;
  } catch (_e) {
    return false;
  }
};

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.environmentNameRequired")),
  shopUrl: Yup.string()
    .required(t("validation.environmentUrlRequired"))
    .test("is-url-valid", t("validation.environmentUrlInvalid"), (value) => isValidUrl(value)),
  shopId: Yup.number().required(t("validation.shopRequired")),
  clientId: Yup.string().required(t("validation.clientIdRequired")),
  clientSecret: Yup.string().required(t("validation.clientSecretRequired")),
});

const initialValues = {
  shopId: selectedShopId.value,
};

const onSubmit = async (values: Record<string, unknown>) => {
  try {
    const typedValues = values as Yup.InferType<typeof schema>;
    const { error: apiError } = await api.POST("/environments", {
      body: {
        name: typedValues.name,
        shopUrl: typedValues.shopUrl.replace(/\/+$/, ""),
        clientId: typedValues.clientId,
        clientSecret: typedValues.clientSecret,
        shopId: selectedShopId.value,
        shopToken,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

    router.push({ name: "account.shop.list" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
};

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

    formRef.value.setFieldValue("shopUrl", data.url);
    formRef.value.setFieldValue("clientId", data.clientId);
    formRef.value.setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("environment.base64InvalidFormat");
  }
}
</script>

<style scoped>
.environment-token-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;

  code {
    font-size: 0.8rem;
    word-break: break-all;
  }
}

.empty-shop-state {
  text-align: center;
  padding: 2rem 1rem 1rem;
}

.empty-shop-icon {
  font-size: 2.5rem;
  color: var(--primary-color);
  opacity: 0.7;
  margin-bottom: 1rem;
}

.empty-shop-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--item-title-color);
  margin-bottom: 0.5rem;
}

.empty-shop-description {
  color: var(--text-color-muted);
  margin-bottom: 1.5rem;
}

.empty-shop-example {
  display: inline-flex;
  flex-direction: column;
  gap: 0.5rem;
  background-color: var(--item-background);
  border: 1px solid var(--panel-border-color);
  border-radius: 0.5rem;
  padding: 1rem 1.5rem;
  margin-bottom: 2rem;
  text-align: left;
}

.example-shop {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  color: var(--item-title-color);
}

.example-environments {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-left: 1.5rem;
  padding-left: 0.75rem;
  border-left: 2px solid var(--panel-border-color);
}

.example-environment {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.example-icon {
  font-size: 0.75rem;
  color: var(--primary-color);
  opacity: 0.8;
}

.empty-shop-cta {
  margin-top: 2rem;
}
</style>
