<template>
  <header-container title="New Shop" />
  <main-container>
    <Panel>
      <vee-form
        ref="formRef"
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="shops"
        @submit="onSubmit"
      >
        <form-group title="Shop information">
          <InputField name="name" label="Name" :error="errors.name" />

          <BaseSelect
            v-model="selectedProjectId"
            name="projectId"
            label="Project"
            required
            :error="errors.projectId"
          >
            <option v-for="project in projects" :key="project.id" :value="project.id">
              {{ project.nameCombined }}
            </option>
          </BaseSelect>

          <InputField name="shopUrl" label="URL" autocomplete="url" :error="errors.shopUrl" />
        </form-group>

        <form-group title="Bypass Authentication Header">
          <template #info>
            If your website is protected by authentication, please configure the header
            <code>shopmon-shop-token</code> with the value below to be excluded, so Shopmon can
            function normally.
          </template>

          <div class="shop-token-display">
            <code>{{ shopToken }}</code>

            <button type="button" class="btn btn-sm btn-icon" @click="copyToken">
              <icon-fa6-solid:copy />
            </button>
          </div>
        </form-group>

        <form-group title="Integration">
          <template #info>
            The easiest way to get started is to install the
            <a href="https://github.com/FriendsOfShopware/FroshShopmon" target="_blank"
              >Shopmon Plugin</a
            >
            or create an integration in your Shopware Administration with the following
            <a
              href="https://github.com/FriendsOfShopware/FroshShopmon?tab=readme-ov-file#permissions"
            >
              permissions
            </a>
          </template>

          <button type="button" class="btn btn-secondary" @click="openPluginModal">
            Connect using Shopmon Plugin
          </button>

          <InputField name="clientId" label="Client-ID" :error="errors.clientId" />
          <InputField name="clientSecret" label="Client-Secret" :error="errors.clientSecret" />
        </form-group>

        <div class="form-submit">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            Save
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

import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

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
  success("Token copied to clipboard");
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
  name: Yup.string().required("Shop name is required"),
  shopUrl: Yup.string()
    .required("Shop URL is required")
    .test("is-url-valid", "Shop URL is not valid", (value) => isValidUrl(value)),
  projectId: Yup.number().required("Project is required"),
  clientId: Yup.string().required("Client ID is required"),
  clientSecret: Yup.string().required("Client Secret is required"),
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
      pluginError.value = "Please enter a base64 string";
      return;
    }

    const decodedString = window.atob(pluginBase64.value.trim());
    const data = JSON.parse(decodedString);

    if (!data.url || !data.clientId || !data.clientSecret) {
      pluginError.value = "Invalid data: missing required fields (url, clientId, clientSecret)";
      return;
    }

    formRef.value.setFieldValue("shopUrl", data.url);
    formRef.value.setFieldValue("clientId", data.clientId);
    formRef.value.setFieldValue("clientSecret", data.clientSecret);

    closePluginModal();
  } catch (_e) {
    pluginError.value = "Invalid base64 string or JSON format";
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
