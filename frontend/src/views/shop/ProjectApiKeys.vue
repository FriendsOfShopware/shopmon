<template>
  <header-container title="Project API Keys">
    <router-link :to="{ name: 'account.project.list' }" type="button" class="btn">
      <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
      Back to Projects
    </router-link>
  </header-container>

  <main-container>
    <div class="panel">
      <Alert type="info">
        <p><strong>API Key Information</strong></p>
        <p>
          API keys allow external applications to interact with your project. Each key can have
          specific scopes to limit what actions it can perform.
        </p>
      </Alert>

      <div class="panel-header">
        <h3>{{ project?.name ?? "Loading..." }}</h3>
        <button type="button" class="btn btn-primary" @click="openAddKeyModal">
          <icon-fa6-solid:plus class="icon" aria-hidden="true" />
          Create API Key
        </button>
      </div>

      <div v-if="isLoading" class="api-keys-loading">
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
    </div>

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
            <label for="name">Name</label>
            <field
              id="name"
              type="text"
              name="name"
              placeholder="My API Key"
              class="field"
              :class="{ 'has-error': errors.name }"
            />
            <div class="field-error-message">{{ errors.name }}</div>
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
        <button type="submit" class="btn btn-primary" form="apiKeyForm" :disabled="isSubmitting">
          <icon-fa6-solid:key v-if="!isSubmitting" class="icon" aria-hidden="true" />
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

    <!-- Delete Confirmation Modal -->
    <delete-confirmation-modal
      :show="showDeleteModal"
      title="Delete API Key?"
      :entity-name="`the API key '${deletingKey?.name}'`"
      custom-consequence="Applications using this key will no longer be able to access your project."
      :reversed-buttons="true"
      :is-loading="isDeleting"
      confirm-button-text="Delete Key"
      @close="showDeleteModal = false"
      @confirm="deleteApiKey"
    />
  </main-container>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { formatDate } from "@/helpers/formatter";
import { type RouterOutput, trpcClient } from "@/helpers/trpc";
import { Field, Form as VeeForm } from "vee-validate";
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import * as Yup from "yup";

type ApiKey = RouterOutput["organization"]["apiKey"]["list"][number];
type AvailableScope = RouterOutput["organization"]["apiKey"]["scopes"][number];

const route = useRoute();
const alert = useAlert();

const projectId = Number(route.params.projectId);
const isLoading = ref(true);
const isSubmitting = ref(false);
const isDeleting = ref(false);

const project = ref<RouterOutput["account"]["currentUserProjects"][number] | null>(null);
const apiKeys = ref<ApiKey[]>([]);
const availableScopes = ref<AvailableScope[]>([]);

const showAddKeyModal = ref(false);
const showTokenModal = ref(false);
const showDeleteModal = ref(false);
const newToken = ref("");
const deletingKey = ref<ApiKey | null>(null);

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

function getScopeLabel(scope: string): string {
  const found = availableScopes.value.find((s) => s.value === scope);
  return found?.label ?? scope;
}

async function loadProject() {
  try {
    const projects = await trpcClient.account.currentUserProjects.query();
    project.value = projects.find((p) => p.id === projectId) ?? null;

    if (!project.value) {
      alert.error("Project not found");
      return;
    }

    await loadApiKeys();
  } catch (error) {
    alert.error(`Failed to load project${error instanceof Error ? `: ${error.message}` : ""}`);
  }
}

async function loadApiKeys() {
  if (!project.value) return;

  isLoading.value = true;
  try {
    apiKeys.value = await trpcClient.organization.apiKey.list.query({
      orgId: project.value.organizationId,
      projectId: project.value.id,
    });
  } catch (error) {
    alert.error(`Failed to load API keys${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isLoading.value = false;
  }
}

async function loadAvailableScopes() {
  try {
    availableScopes.value = await trpcClient.organization.apiKey.scopes.query();
  } catch (error) {
    alert.error(`Failed to load scopes${error instanceof Error ? `: ${error.message}` : ""}`);
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

  isSubmitting.value = true;
  try {
    const result = await trpcClient.organization.apiKey.create.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      name: typedValues.name,
      scopes: typedValues.scopes as "deployments"[],
    });

    newToken.value = result.token;
    closeAddKeyModal();
    showTokenModal.value = true;
    await loadApiKeys();

    alert.success("API key created successfully");
  } catch (error) {
    alert.error(`Failed to create API key${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isSubmitting.value = false;
  }
}

function copyToken() {
  navigator.clipboard.writeText(newToken.value);
  alert.success("API key copied to clipboard");
}

function confirmDeleteKey(apiKey: ApiKey) {
  deletingKey.value = apiKey;
  showDeleteModal.value = true;
}

async function deleteApiKey() {
  if (!deletingKey.value || !project.value) return;

  isDeleting.value = true;
  try {
    await trpcClient.organization.apiKey.delete.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      apiKeyId: deletingKey.value.id,
    });

    alert.success("API key deleted successfully");
    showDeleteModal.value = false;
    await loadApiKeys();
  } catch (error) {
    alert.error(`Failed to delete API key${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isDeleting.value = false;
    deletingKey.value = null;
  }
}

onMounted(() => {
  loadProject();
  loadAvailableScopes();
});
</script>

<style scoped>
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
  padding: 1rem;
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
</style>
