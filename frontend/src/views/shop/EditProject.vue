<template>
  <header-container :title="project ? `Edit ${project.name}` : 'Edit Project'">
    <div class="header-actions">
      <router-link :to="{ name: 'account.project.list' }" type="button" class="btn">
        <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
        Back to Projects
      </router-link>

      <router-link
        v-if="project"
        :to="{ name: 'account.shops.new', query: { projectId: project.id } }"
        type="button"
        class="btn btn-secondary"
      >
        <icon-fa6-solid:plus class="icon" aria-hidden="true" />
        Add Shop
      </router-link>
    </div>
  </header-container>

  <main-container>
    <Panel v-if="isPageLoading" class="state-panel">
      <icon-line-md:loading-twotone-loop class="icon" />
      Loading project...
    </Panel>

    <Panel v-else-if="!project" class="state-panel">
      <p>Project not found.</p>
    </Panel>

    <template v-else>
      <Panel>
        <vee-form
          :key="`project-${project.id}`"
          v-slot="{ errors, isSubmitting }"
          :validation-schema="projectSchema"
          :initial-values="projectFormInitialValues"
          @submit="onSubmitProject"
        >
          <form-group title="Project Information">
            <InputField name="name" label="Name" :error="errors.name" />

            <TextareaField
              name="description"
              label="Description"
              placeholder="Optional project description..."
              :error="errors.description"
            />

            <InputField
              name="gitUrl"
              label="Git Repository URL"
              type="url"
              placeholder="https://github.com/org/repo"
              :error="errors.gitUrl"
            />
          </form-group>

          <div class="form-submit">
            <button
              :disabled="isSubmitting || isSavingProject"
              type="submit"
              class="btn btn-primary"
            >
              <icon-fa6-solid:floppy-disk
                v-if="!(isSubmitting || isSavingProject)"
                class="icon"
                aria-hidden="true"
              />
              <icon-line-md:loading-twotone-loop v-else class="icon" />
              Save
            </button>
          </div>
        </vee-form>
      </Panel>

      <Panel id="api-keys" title="API Keys">
        <template #action>
          <button type="button" class="btn btn-primary" @click="openAddKeyModal">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Create API Key
          </button>
        </template>

        <Alert type="info">
          <p><strong>API Key Information</strong></p>
          <p>
            API keys allow external applications to interact with your project. Each key can have
            specific scopes to limit what actions it can perform.
          </p>
        </Alert>

        <div v-if="isApiKeysLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          Loading API keys...
        </div>

        <div v-else-if="apiKeys.length === 0" class="api-keys-empty">
          <icon-fa6-solid:key class="icon icon-large" aria-hidden="true" />
          <p>No API keys created yet.</p>
          <p class="text-muted">
            Create an API key to allow external applications to access your project.
          </p>
        </div>

        <div v-else class="api-keys-list">
          <div v-for="apiKey in apiKeys" :key="apiKey.id" class="api-key-item">
            <div class="api-key-info">
              <h4>{{ apiKey.name }}</h4>
              <p class="text-muted">Created {{ formatDate(apiKey.createdAt) }}</p>
              <p v-if="apiKey.lastUsedAt" class="text-muted">
                Last used {{ formatDate(apiKey.lastUsedAt) }}
              </p>
              <div class="api-key-scopes">
                <span v-for="scope in apiKey.scopes" :key="scope" class="badge badge-primary">
                  {{ getScopeLabel(scope) }}
                </span>
              </div>
            </div>
            <div class="api-key-actions">
              <button type="button" class="btn btn-danger" @click="confirmDeleteKey(apiKey)">
                <icon-fa6-solid:trash class="icon" aria-hidden="true" />
                Delete
              </button>
            </div>
          </div>
        </div>
      </Panel>

      <Panel
        v-if="isPackagesConfigured"
        id="packages-tokens"
        title="Packages Tokens"
        description="Sync store packages through a Global CDN (~80ms vs ~6s). Tokens are synced automatically every hour."
      >
        <template #action>
          <button type="button" class="btn btn-primary" @click="showAddPackagesTokenModal = true">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            Add Token
          </button>
        </template>

        <div v-if="isPackagesTokensLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          Loading packages tokens...
        </div>

        <div v-else-if="packagesTokens.length === 0" class="api-keys-empty">
          <icon-fa6-solid:cube class="icon icon-large" aria-hidden="true" />
          <p>No packages tokens added yet.</p>
          <p class="text-muted">Add a Shopware store token to sync packages through the mirror.</p>
        </div>

        <template v-else>
          <div class="api-keys-list">
            <div v-for="pt in packagesTokens" :key="pt.id" class="api-key-item">
              <div class="api-key-info">
                <h4>Token #{{ pt.id }}</h4>
                <p v-if="pt.lastSyncedAt" class="text-muted">
                  Last synced {{ timeAgo(new Date(pt.lastSyncedAt).getTime() / 1000) }}
                </p>
                <p v-else class="text-muted">Not synced yet</p>
              </div>
              <div class="api-key-actions">
                <button
                  type="button"
                  class="btn btn-secondary"
                  :disabled="isSyncingPackagesToken === pt.id"
                  @click="syncPackagesToken(pt)"
                >
                  <icon-fa6-solid:arrows-rotate
                    v-if="isSyncingPackagesToken !== pt.id"
                    class="icon"
                    aria-hidden="true"
                  />
                  <icon-line-md:loading-twotone-loop v-else class="icon" />
                  Sync
                </button>
                <button
                  type="button"
                  class="btn btn-danger"
                  @click="confirmDeletePackagesToken(pt)"
                >
                  <icon-fa6-solid:trash class="icon" aria-hidden="true" />
                  Delete
                </button>
              </div>
            </div>
          </div>

          <div v-if="packagesComposerUrl" class="composer-setup">
            <h4>Composer Setup</h4>
            <Alert type="warning">
              <p>
                Make sure to remove <code>packages.shopware.com</code> from your
                <code>composer.json</code> repositories before adding the mirror, otherwise Composer
                will still use the original source.
              </p>
            </Alert>
            <p class="text-muted">
              Add the mirror as a repository in your <code>composer.json</code>:
            </p>
            <div class="code-block">
              <pre><code>{
    "repositories": [
        {
            "type": "composer",
            "url": "{{ packagesComposerUrl }}"
        }
    ]
}</code></pre>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                @click="copyComposerRepository"
              >
                <icon-fa6-solid:copy class="icon" aria-hidden="true" />
                Copy
              </button>
            </div>
            <p class="text-muted">Authenticate using your Shopware store token:</p>
            <div class="code-block">
              <pre><code>composer config --auth bearer.{{ packagesComposerHost }} &lt;your-token&gt;</code></pre>
            </div>
          </div>
        </template>
      </Panel>

      <Panel title="Danger Zone">
        <p>Once you delete your project, you will lose all data associated with it.</p>

        <p v-if="!canDeleteProject" class="delete-project-warning">
          This project still has {{ shopsInProjectCount }} shop(s). Move or delete them first.
        </p>

        <button
          type="button"
          class="btn btn-danger"
          :disabled="!canDeleteProject || isDeletingProject"
          @click="showDeleteProjectModal = true"
        >
          <icon-fa6-solid:trash class="icon" aria-hidden="true" />
          Delete project
        </button>
      </Panel>
    </template>

    <!-- Add API Key Modal -->
    <modal :show="showAddKeyModal" close-x-mark @close="closeAddKeyModal">
      <template #title> Create API Key </template>

      <template #content>
        <vee-form
          id="apiKeyForm"
          v-slot="{ errors }"
          :validation-schema="apiKeySchema"
          :initial-values="{ name: '', scopes: [] }"
          class="api-key-form"
          @submit="onSubmitApiKey"
        >
          <div class="form-group">
            <InputField name="name" label="Name" placeholder="My API Key" :error="errors.name" />
            <p class="field-help">A descriptive name to identify this API key</p>
          </div>

          <div class="form-group">
            <label>Scopes</label>
            <p class="field-help">Select the permissions this API key should have</p>
            <div class="scopes-list">
              <label v-for="scope in availableScopes" :key="scope.value" class="scope-checkbox">
                <field type="checkbox" name="scopes" :value="scope.value" />
                <div class="scope-info">
                  <span class="scope-label">{{ scope.label }}</span>
                  <span class="scope-description">{{ scope.description }}</span>
                </div>
              </label>
            </div>
            <div class="field-error-message">{{ errors.scopes }}</div>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <button type="button" class="btn" @click="closeAddKeyModal">Cancel</button>
        <button
          type="submit"
          class="btn btn-primary"
          form="apiKeyForm"
          :disabled="isCreatingApiKey"
        >
          <icon-fa6-solid:key v-if="!isCreatingApiKey" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          Create API Key
        </button>
      </template>
    </modal>

    <!-- Show Token Modal -->
    <modal :show="showTokenModal" @close="closeTokenModal">
      <template #title> API Key Created </template>

      <template #content>
        <Alert type="warning">
          <p><strong>Copy your API key now!</strong></p>
          <p>
            This is the only time you will see this key. Make sure to copy it and store it securely.
          </p>
        </Alert>

        <div class="token-display">
          <code class="token-value">{{ newToken }}</code>
          <button type="button" class="btn btn-secondary" @click="copyToken">
            <icon-fa6-solid:copy class="icon" aria-hidden="true" />
            Copy
          </button>
        </div>
      </template>

      <template #footer>
        <button type="button" class="btn btn-primary" @click="closeTokenModal">Done</button>
      </template>
    </modal>

    <!-- Delete API Key Modal -->
    <delete-confirmation-modal
      :show="showDeleteApiKeyModal"
      title="Delete API Key?"
      :entity-name="`the API key '${deletingApiKey?.name}'`"
      custom-consequence="Applications using this key will no longer be able to access your project."
      :reversed-buttons="true"
      :is-loading="isDeletingApiKey"
      confirm-button-text="Delete Key"
      @close="showDeleteApiKeyModal = false"
      @confirm="deleteApiKey"
    />

    <!-- Delete Project Modal -->
    <delete-confirmation-modal
      :show="showDeleteProjectModal"
      title="Delete Project"
      :entity-name="project?.name || 'this project'"
      @close="showDeleteProjectModal = false"
      @confirm="deleteProject"
    />

    <!-- Add Packages Token Modal -->
    <modal
      :show="showAddPackagesTokenModal"
      close-x-mark
      @close="showAddPackagesTokenModal = false"
    >
      <template #title> Add Packages Token </template>

      <template #content>
        <vee-form
          id="packagesTokenForm"
          v-slot="{ errors }"
          :validation-schema="packagesTokenSchema"
          :initial-values="{ token: '' }"
          class="api-key-form"
          @submit="onSubmitPackagesToken"
        >
          <div class="form-group">
            <InputField
              name="token"
              label="Shopware Store Token"
              placeholder="Enter your Shopware store token"
              :error="errors.token"
            />
            <p class="field-help">
              The token from your Shopware account used to access store packages
            </p>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <button type="button" class="btn" @click="showAddPackagesTokenModal = false">Cancel</button>
        <button
          type="submit"
          class="btn btn-primary"
          form="packagesTokenForm"
          :disabled="isCreatingPackagesToken"
        >
          <icon-fa6-solid:plus v-if="!isCreatingPackagesToken" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          Add Token
        </button>
      </template>
    </modal>

    <!-- Delete Packages Token Modal -->
    <delete-confirmation-modal
      :show="showDeletePackagesTokenModal"
      title="Delete Packages Token?"
      :entity-name="`Token #${deletingPackagesToken?.id}`"
      custom-consequence="Packages synced with this token will no longer be available through the mirror."
      :reversed-buttons="true"
      :is-loading="isDeletingPackagesToken"
      confirm-button-text="Delete Token"
      @close="showDeletePackagesTokenModal = false"
      @confirm="deletePackagesToken"
    />
  </main-container>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { fetchAccountShops } from "@/composables/useAccountShops";
import { formatDate, timeAgo } from "@/helpers/formatter";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Field, Form as VeeForm } from "vee-validate";
import { computed, nextTick, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

type Project = components["schemas"]["AccountProject"];
type ApiKey = components["schemas"]["ApiKey"];
type AvailableScope = components["schemas"]["ApiKeyScope"];
type PackagesToken = components["schemas"]["PackagesToken"];

const route = useRoute();
const router = useRouter();
const alert = useAlert();

const projectId = Number(route.params.projectId);

const project = ref<Project | null>(null);
const shopsInProjectCount = ref(0);
const apiKeys = ref<ApiKey[]>([]);
const availableScopes = ref<AvailableScope[]>([]);

const isPageLoading = ref(true);
const isSavingProject = ref(false);
const isApiKeysLoading = ref(false);
const isCreatingApiKey = ref(false);
const isDeletingApiKey = ref(false);
const isDeletingProject = ref(false);

const showAddKeyModal = ref(false);
const showTokenModal = ref(false);
const showDeleteApiKeyModal = ref(false);
const showDeleteProjectModal = ref(false);

const deletingApiKey = ref<ApiKey | null>(null);
const newToken = ref("");

// Packages tokens
const isPackagesConfigured = ref(false);
const packagesComposerUrl = ref<string | null>(null);
const packagesTokens = ref<PackagesToken[]>([]);
const isPackagesTokensLoading = ref(false);
const isCreatingPackagesToken = ref(false);
const isDeletingPackagesToken = ref(false);
const isSyncingPackagesToken = ref<number | null>(null);
const showAddPackagesTokenModal = ref(false);
const showDeletePackagesTokenModal = ref(false);
const deletingPackagesToken = ref<PackagesToken | null>(null);

const canDeleteProject = computed(() => shopsInProjectCount.value === 0);

const packagesComposerHost = computed(() => {
  if (!packagesComposerUrl.value) return "";
  try {
    return new URL(packagesComposerUrl.value).host;
  } catch {
    return packagesComposerUrl.value;
  }
});

const projectFormInitialValues = computed(() => ({
  name: project.value?.name ?? "",
  description: project.value?.description ?? "",
  gitUrl: project.value?.gitUrl ?? "",
}));

const projectSchema = Yup.object().shape({
  name: Yup.string().required("Project name is required"),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url("Must be a valid URL").optional(),
});

const apiKeySchema = Yup.object().shape({
  name: Yup.string()
    .required("Name is required")
    .min(1, "Name is required")
    .max(100, "Name must be at most 100 characters"),
  scopes: Yup.array()
    .of(Yup.string())
    .min(1, "At least one scope is required")
    .required("At least one scope is required"),
});

const packagesTokenSchema = Yup.object().shape({
  token: Yup.string().required("Token is required").min(1, "Token is required"),
});

async function loadProjectSummary() {
  const [projectsRes, shopsData] = await Promise.all([
    api.GET("/account/projects"),
    fetchAccountShops(),
  ]);

  const projectsData = projectsRes.data ?? [];

  project.value = projectsData.find((currentProject) => currentProject.id === projectId) ?? null;
  shopsInProjectCount.value = shopsData.filter((shop) => shop.projectId === projectId).length;
}

async function loadApiKeys() {
  if (!project.value) return;

  isApiKeysLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/projects/{projectId}/api-keys", {
      params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
    });
    apiKeys.value = data ?? [];
  } catch (error) {
    alert.error(`Failed to load API keys${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isApiKeysLoading.value = false;
  }
}

async function loadAvailableScopes() {
  try {
    const { data } = await api.GET("/api-key-scopes");
    availableScopes.value = data ?? [];
  } catch (error) {
    alert.error(`Failed to load scopes${error instanceof Error ? `: ${error.message}` : ""}`);
  }
}

async function scrollToHashSection() {
  if (route.hash !== "#api-keys") return;
  await nextTick();
  document.getElementById("api-keys")?.scrollIntoView({ behavior: "smooth" });
}

async function loadPageData() {
  isPageLoading.value = true;

  try {
    await loadProjectSummary();

    if (!project.value) {
      alert.error("Project not found");
      return;
    }

    await Promise.all([loadApiKeys(), loadAvailableScopes(), checkPackagesConfigured()]);
    await loadPackagesTokens();
    await scrollToHashSection();
  } catch (error) {
    alert.error(`Failed to load project${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isPageLoading.value = false;
  }
}

function getScopeLabel(scope: string): string {
  const found = availableScopes.value.find((availableScope) => availableScope.value === scope);
  return found?.label ?? scope;
}

async function onSubmitProject(values: Record<string, unknown>) {
  if (!project.value) return;

  const typedValues = values as Yup.InferType<typeof projectSchema>;

  isSavingProject.value = true;
  try {
    await api.PATCH("/organizations/{orgId}/projects/{projectId}", {
      params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
      body: {
        name: typedValues.name,
        description: typedValues.description ?? "",
        gitUrl: typedValues.gitUrl || undefined,
      },
    });

    await loadProjectSummary();
    alert.success("Project updated successfully");
  } catch (error) {
    alert.error(`Failed to update project${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isSavingProject.value = false;
  }
}

function openAddKeyModal() {
  showAddKeyModal.value = true;
}

function closeAddKeyModal() {
  showAddKeyModal.value = false;
}

function closeTokenModal() {
  showTokenModal.value = false;
  newToken.value = "";
}

async function onSubmitApiKey(values: Record<string, unknown>) {
  if (!project.value) return;

  const typedValues = values as Yup.InferType<typeof apiKeySchema>;

  isCreatingApiKey.value = true;
  try {
    const { data: result } = await api.POST(
      "/organizations/{orgId}/projects/{projectId}/api-keys",
      {
        params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
        body: {
          name: typedValues.name,
          scopes: typedValues.scopes as string[],
        },
      },
    );

    newToken.value = result?.token ?? "";
    closeAddKeyModal();
    showTokenModal.value = true;

    await loadApiKeys();
    alert.success("API key created successfully");
  } catch (error) {
    alert.error(`Failed to create API key${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isCreatingApiKey.value = false;
  }
}

function copyToken() {
  if (!navigator.clipboard) {
    alert.error("Clipboard is not available");
    return;
  }

  navigator.clipboard.writeText(newToken.value);
  alert.success("API key copied to clipboard");
}

function copyComposerRepository() {
  if (!navigator.clipboard || !packagesComposerUrl.value) return;

  const json = JSON.stringify(
    { repositories: [{ type: "composer", url: packagesComposerUrl.value }] },
    null,
    4,
  );
  navigator.clipboard.writeText(json);
  alert.success("Composer repository config copied to clipboard");
}

function confirmDeleteKey(apiKey: ApiKey) {
  deletingApiKey.value = apiKey;
  showDeleteApiKeyModal.value = true;
}

async function deleteApiKey() {
  if (!deletingApiKey.value || !project.value) return;

  isDeletingApiKey.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/projects/{projectId}/api-keys/{keyId}", {
      params: {
        path: {
          orgId: project.value.organizationId,
          projectId: project.value.id,
          keyId: deletingApiKey.value.id,
        },
      },
    });

    showDeleteApiKeyModal.value = false;
    deletingApiKey.value = null;
    await loadApiKeys();
    alert.success("API key deleted successfully");
  } catch (error) {
    alert.error(`Failed to delete API key${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isDeletingApiKey.value = false;
  }
}

async function checkPackagesConfigured() {
  try {
    const { data: config } = await api.GET("/packages-token/configuration");
    isPackagesConfigured.value = config?.configured ?? false;
    packagesComposerUrl.value = config?.composerUrl ?? null;
  } catch {
    isPackagesConfigured.value = false;
  }
}

async function loadPackagesTokens() {
  if (!project.value || !isPackagesConfigured.value) return;

  isPackagesTokensLoading.value = true;
  try {
    const { data } = await api.GET("/organizations/{orgId}/projects/{projectId}/packages-tokens", {
      params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
    });
    packagesTokens.value = data ?? [];
  } catch (error) {
    alert.error(
      `Failed to load packages tokens${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isPackagesTokensLoading.value = false;
  }
}

async function onSubmitPackagesToken(values: Record<string, unknown>) {
  if (!project.value) return;

  const typedValues = values as Yup.InferType<typeof packagesTokenSchema>;

  isCreatingPackagesToken.value = true;
  try {
    await api.POST("/organizations/{orgId}/projects/{projectId}/packages-tokens", {
      params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
      body: { token: typedValues.token },
    });

    showAddPackagesTokenModal.value = false;
    await loadPackagesTokens();
    alert.success("Packages token added successfully");
  } catch (error) {
    alert.error(
      `Failed to add packages token${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isCreatingPackagesToken.value = false;
  }
}

function confirmDeletePackagesToken(pt: PackagesToken) {
  deletingPackagesToken.value = pt;
  showDeletePackagesTokenModal.value = true;
}

async function deletePackagesToken() {
  if (!deletingPackagesToken.value || !project.value) return;

  isDeletingPackagesToken.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/projects/{projectId}/packages-tokens/{tokenId}", {
      params: {
        path: {
          orgId: project.value.organizationId,
          projectId: project.value.id,
          tokenId: deletingPackagesToken.value.id,
        },
      },
    });

    showDeletePackagesTokenModal.value = false;
    deletingPackagesToken.value = null;
    await loadPackagesTokens();
    alert.success("Packages token deleted successfully");
  } catch (error) {
    alert.error(
      `Failed to delete packages token${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isDeletingPackagesToken.value = false;
  }
}

async function syncPackagesToken(pt: PackagesToken) {
  if (!project.value) return;

  isSyncingPackagesToken.value = pt.id;
  try {
    await api.POST("/organizations/{orgId}/projects/{projectId}/packages-tokens/{tokenId}/sync", {
      params: {
        path: { orgId: project.value.organizationId, projectId: project.value.id, tokenId: pt.id },
      },
    });

    await loadPackagesTokens();
    alert.success("Token sync triggered successfully");
  } catch (error) {
    alert.error(
      `Failed to sync packages token${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isSyncingPackagesToken.value = null;
  }
}

async function deleteProject() {
  if (!project.value) return;

  isDeletingProject.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/projects/{projectId}", {
      params: { path: { orgId: project.value.organizationId, projectId: project.value.id } },
    });

    alert.success("Project deleted successfully");
    router.push({ name: "account.project.list" });
  } catch (error) {
    alert.error(`Failed to delete project${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isDeletingProject.value = false;
    showDeleteProjectModal.value = false;
  }
}

onMounted(() => {
  if (Number.isNaN(projectId)) {
    isPageLoading.value = false;
    alert.error("Invalid project");
    return;
  }

  loadPageData();
});
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 0.75rem;
}

.state-panel {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.delete-project-warning {
  color: var(--color-danger);
}

.api-keys-loading {
  padding: 3rem;
  text-align: center;
  color: var(--text-color-muted);
}

.api-keys-empty {
  padding: 4rem 2rem;
  text-align: center;

  .icon-large {
    font-size: 3rem;
    color: var(--text-color-muted);
    margin-bottom: 1rem;
  }

  p {
    margin: 0;

    &.text-muted {
      margin-top: 0.5rem;
      color: var(--text-color-muted);
      font-size: 0.875rem;
    }
  }
}

.api-keys-list {
  padding: 1rem 0 0;
}

.api-key-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border: 1px solid var(--panel-border-color);
  border-radius: 0.5rem;
  margin-bottom: 1rem;

  &:last-child {
    margin-bottom: 0;
  }
}

.api-key-info {
  h4 {
    margin: 0 0 0.25rem 0;
    font-size: 1.125rem;
    font-weight: 500;
  }

  p {
    margin: 0 0 0.25rem 0;
    font-size: 0.875rem;
  }
}

.api-key-scopes {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.api-key-actions {
  display: flex;
  gap: 0.5rem;
}

.api-key-form {
  .form-group {
    margin-bottom: 1.5rem;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

.field-help {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--text-color-muted);
}

.scopes-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.scope-checkbox {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--panel-border-color);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background-color 0.2s;

  &:hover {
    background-color: var(--panel-background-color);
  }

  input[type="checkbox"] {
    margin-top: 0.25rem;
  }
}

.scope-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.scope-label {
  font-weight: 500;
}

.scope-description {
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.token-display {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
  padding: 1rem;
  background-color: var(--panel-background-color);
  border-radius: 0.5rem;
}

.token-value {
  flex: 1;
  padding: 0.75rem;
  background-color: var(--input-background-color);
  border: 1px solid var(--input-border-color);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.875rem;
  word-break: break-all;
}

.composer-setup {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--panel-border-color);

  h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    font-weight: 500;
  }

  .text-muted {
    font-size: 0.875rem;
    color: var(--text-color-muted);
    margin: 0.5rem 0;

    code {
      padding: 0.125rem 0.375rem;
      background-color: var(--input-background-color);
      border-radius: 0.25rem;
      font-size: 0.8125rem;
    }
  }
}

.code-block {
  position: relative;
  margin: 0.5rem 0 1rem;
  background-color: var(--input-background-color);
  border: 1px solid var(--input-border-color);
  border-radius: 0.5rem;
  overflow: hidden;

  pre {
    margin: 0;
    padding: 1rem;
    overflow-x: auto;

    code {
      font-family: monospace;
      font-size: 0.8125rem;
      line-height: 1.5;
    }
  }

  .btn-sm {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }
}
</style>
