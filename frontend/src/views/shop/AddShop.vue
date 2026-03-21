<template>
  <header-container title="New Shop" />
  <main-container>
    <Panel v-if="projects.length === 0 && projectsLoaded">
      <div class="empty-project-state">
        <div class="empty-project-icon">
          <icon-fa6-solid:folder-tree aria-hidden="true" />
        </div>

        <h3 class="empty-project-title">Create a project first</h3>

        <p class="empty-project-description">
          Projects group related shop environments together. For example:
        </p>

        <div class="empty-project-example">
          <div class="example-project">
            <icon-fa6-solid:folder class="example-icon" aria-hidden="true" />
            <span class="example-project-name">Toy Shop X</span>
          </div>
          <div class="example-shops">
            <div class="example-shop">
              <icon-fa6-solid:shop class="example-icon" aria-hidden="true" />
              Production
            </div>
            <div class="example-shop">
              <icon-fa6-solid:shop class="example-icon" aria-hidden="true" />
              Staging
            </div>
          </div>
        </div>

        <div class="empty-project-cta">
          <router-link :to="{ name: 'account.projects.new', query: { redirect: $route.fullPath } }" class="btn btn-primary">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Create Project
          </router-link>
        </div>
      </div>
    </Panel>

    <Panel v-else>
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
              {{ project.organizationName + " / " + project.name }}
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

import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Form as VeeForm } from "vee-validate";
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

const projects = ref<components["schemas"]["AccountProject"][]>([]);
const projectsLoaded = ref(false);
const selectedProjectId = ref<number>(route.query.projectId ? Number(route.query.projectId) : 0);

const showPluginModal = ref(false);
const pluginBase64 = ref("");
const pluginError = ref("");

const formRef = ref();

api.GET("/account/projects").then(({ data }) => {
  if (data) {
    projects.value = data;
    if (!selectedProjectId.value && data.length > 0) {
      selectedProjectId.value = data[0].id;
    }
  }
  projectsLoaded.value = true;
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
    const { error: apiError } = await api.POST("/shops", {
      body: {
        name: typedValues.name,
        shopUrl: typedValues.shopUrl.replace(/\/+$/, ""),
        clientId: typedValues.clientId,
        clientSecret: typedValues.clientSecret,
        projectId: selectedProjectId.value,
        shopToken,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

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

.empty-project-state {
  text-align: center;
  padding: 2rem 1rem 1rem;
}

.empty-project-icon {
  font-size: 2.5rem;
  color: var(--primary-color);
  opacity: 0.7;
  margin-bottom: 1rem;
}

.empty-project-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--item-title-color);
  margin-bottom: 0.5rem;
}

.empty-project-description {
  color: var(--text-color-muted);
  margin-bottom: 1.5rem;
}

.empty-project-example {
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

.example-project {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  color: var(--item-title-color);
}

.example-shops {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-left: 1.5rem;
  padding-left: 0.75rem;
  border-left: 2px solid var(--panel-border-color);
}

.example-shop {
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

.empty-project-cta {
  margin-top: 2rem;
}
</style>
