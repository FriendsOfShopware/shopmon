<template>
  <header-container :title="$t('shop.newShop')" />
  <main-container>
    <Panel>
      <vee-form
        ref="formRef"
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="shops"
        @submit="onSubmit"
      >
        <form-group :title="$t('shop.shopInfo')">
          <div>
            <label for="Name">{{ $t("common.name") }}</label>

            <field
              id="name"
              type="text"
              name="name"
              class="field"
              :class="{ 'has-error': errors.name }"
            />

            <div class="field-error-message">
              {{ errors.name }}
            </div>
          </div>

          <div>
            <label for="projectId">{{ $t("shop.project") }}</label>

            <field id="projectId" name="projectId">
              <select v-model="selectedProjectId" class="field" required>
                <option v-for="project in projects" :key="project.id" :value="project.id">
                  {{ project.nameCombined }}
                </option>
              </select>
            </field>

            <div class="field-error-message">
              {{ errors.projectId }}
            </div>
          </div>

          <div>
            <label for="shopUrl">{{ $t("common.url") }}</label>

            <field
              id="shopUrl"
              type="text"
              name="shopUrl"
              autocomplete="url"
              class="field"
              :class="{ 'has-error': errors.shop_url }"
            />

            <div class="field-error-message">
              {{ errors.shopUrl }}
            </div>
          </div>
        </form-group>

        <form-group :title="$t('shop.bypassAuthHeader')">
          <template #info>
            {{ $t("shop.bypassAuthHeaderDesc") }}
          </template>

          <div class="shop-token-display">
            <code>{{ shopToken }}</code>

            <button type="button" class="btn btn-sm btn-icon" @click="copyToken">
              <icon-fa6-solid:copy />
            </button>
          </div>
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

          <div>
            <label for="clientId">{{ $t("shop.clientId") }}</label>

            <field
              id="clientId"
              type="text"
              name="clientId"
              class="field"
              :class="{ 'has-error': errors.clientId }"
            />

            <div class="field-error-message">
              {{ errors.clientId }}
            </div>
          </div>

          <div>
            <label for="clientSecret">{{ $t("shop.clientSecret") }}</label>

            <field
              id="clientSecret"
              type="text"
              name="clientSecret"
              class="field"
              :class="{ 'has-error': errors.clientSecret }"
            />

            <div class="field-error-message">
              {{ errors.clientSecret }}
            </div>
          </div>
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

import { type RouterInput, type RouterOutput, trpcClient } from "@/helpers/trpc";
import { Field, Form as VeeForm } from "vee-validate";
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
  success(t("shop.tokenCopied"));
}

const projects = ref<RouterOutput["account"]["currentUserProjects"]>([]);
const selectedProjectId = ref<number>(route.query.projectId ? Number(route.query.projectId) : 0);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

const formRef = ref();

trpcClient.account.currentUserProjects.query().then((data) => {
  projects.value = data;
  if (!selectedProjectId.value && data.length > 0) {
    selectedProjectId.value = data[0].id;
  }
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
  name: Yup.string().required(t("validation.shopNameRequired")),
  shopUrl: Yup.string()
    .required(t("validation.shopUrlRequired"))
    .test("is-url-valid", t("validation.shopUrlInvalid"), (value) => isValidUrl(value)),
  projectId: Yup.number().required(t("validation.projectRequired")),
  clientId: Yup.string().required(t("validation.clientIdRequired")),
  clientSecret: Yup.string().required(t("validation.clientSecretRequired")),
});

const shops = {
  projectId: selectedProjectId.value,
};

const onSubmit = async (values: Record<string, unknown>) => {
  try {
    const typedValues = values as Yup.InferType<typeof schema>;
    const input: RouterInput["organization"]["shop"]["create"] = {
      name: typedValues.name,
      shopUrl: typedValues.shopUrl.replace(/\/+$/, ""),
      clientId: typedValues.clientId,
      clientSecret: typedValues.clientSecret,
      projectId: selectedProjectId.value,
      shopToken,
    };
    await trpcClient.organization.shop.create.mutate(input);

    router.push({ name: "account.project.list" });
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
      pluginError.value = t("shop.base64Error");
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = t("shop.base64InvalidData");
      return;
    }

    formRef.value.setFieldValue("shopUrl", data.url);
    formRef.value.setFieldValue("clientId", data.clientId);
    formRef.value.setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = t("shop.base64InvalidFormat");
  }
}
</script>

<style scoped>
.shop-token-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;

  code {
    font-size: 0.8rem;
    word-break: break-all;
  }
}
</style>
