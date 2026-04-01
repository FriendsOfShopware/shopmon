<template>
  <header-container :title="project ? $t('project.editProject', { name: project.name }) : $t('nav.editProject')">
    <div class="header-actions">
      <router-link :to="{ name: 'account.project.list' }" type="button" class="btn">
        <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
        {{ $t('project.backToProjects') }}
      </router-link>

      <router-link
        v-if="project"
        :to="{ name: 'account.shops.new', query: { projectId: project.id } }"
        type="button"
        class="btn btn-secondary"
      >
        <icon-fa6-solid:plus class="icon" aria-hidden="true" />
        {{ $t('shop.addShop') }}
      </router-link>
    </div>
  </header-container>

  <main-container>
    <Panel v-if="isPageLoading" class="state-panel">
      <icon-line-md:loading-twotone-loop class="icon" />
      {{ $t('project.loadingProject') }}
    </Panel>

    <Panel v-else-if="!project" class="state-panel">
      <p>{{ $t('project.projectNotFound') }}</p>
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
          <form-group :title="$t('project.projectInfo')">
            <div>
              <label for="name">{{ $t('common.name') }}</label>

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
              <label for="description">{{ $t('common.description') }}</label>

              <field id="description" v-slot="{ field }" name="description">
                <textarea
                  v-bind="field"
                  id="description"
                  class="field"
                  rows="4"
                  :placeholder="$t('project.optionalDescription')"
                  :class="{ 'has-error': errors.description }"
                />
              </field>

              <div class="field-error-message">
                {{ errors.description }}
              </div>
            </div>

            <div>
              <label for="gitUrl">{{ $t('project.gitRepoUrl') }}</label>

              <field
                id="gitUrl"
                type="url"
                name="gitUrl"
                class="field"
                placeholder="https://github.com/org/repo"
                :class="{ 'has-error': errors.gitUrl }"
              />

              <div class="field-error-message">
                {{ errors.gitUrl }}
              </div>
            </div>
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
              {{ $t('common.save') }}
            </button>
          </div>
        </vee-form>
      </Panel>

      <Panel id="api-keys" :title="$t('project.apiKeys')">
        <template #action>
          <button type="button" class="btn btn-primary" @click="openAddKeyModal">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            {{ $t('project.createApiKey') }}
          </button>
        </template>

        <Alert type="info">
          <p><strong>{{ $t('project.apiKeys') }}</strong></p>
          <p>
            {{ $t('project.apiKeyInfo') }}
          </p>
        </Alert>

        <div v-if="isApiKeysLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          {{ $t('project.loadingApiKeys') }}
        </div>

        <div v-else-if="apiKeys.length === 0" class="api-keys-empty">
          <icon-fa6-solid:key class="icon icon-large" aria-hidden="true" />
          <p>{{ $t('project.noApiKeys') }}</p>
          <p class="text-muted">
            {{ $t('project.noApiKeysHint') }}
          </p>
        </div>

        <div v-else class="api-keys-list">
          <div v-for="apiKey in apiKeys" :key="apiKey.id" class="api-key-item">
            <div class="api-key-info">
              <h4>{{ apiKey.name }}</h4>
              <p class="text-muted">{{ $t('project.createdDate', { date: formatDate(apiKey.createdAt) }) }}</p>
              <p v-if="apiKey.lastUsedAt" class="text-muted">
                {{ $t('project.lastUsedDate', { date: formatDate(apiKey.lastUsedAt) }) }}
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
                {{ $t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </Panel>

      <Panel
        v-if="isPackagesConfigured"
        id="packages-tokens"
        :title="$t('packages.title')"
        :description="$t('packages.description')"
      >
        <template #action>
          <button type="button" class="btn btn-primary" @click="showAddPackagesTokenModal = true">
            <icon-fa6-solid:plus class="icon" aria-hidden="true" />
            {{ $t('packages.addToken') }}
          </button>
        </template>

        <div v-if="isPackagesTokensLoading" class="api-keys-loading">
          <icon-line-md:loading-twotone-loop class="icon" />
          {{ $t('packages.loading') }}
        </div>

        <div v-else-if="packagesTokens.length === 0" class="api-keys-empty">
          <icon-fa6-solid:cube class="icon icon-large" aria-hidden="true" />
          <p>{{ $t('packages.noTokens') }}</p>
          <p class="text-muted">{{ $t('packages.noTokensHint') }}</p>
        </div>

        <template v-else>
          <div class="api-keys-list">
            <div v-for="pt in packagesTokens" :key="pt.id" class="api-key-item">
              <div class="api-key-info">
                <h4>Token #{{ pt.id }}</h4>
                <p v-if="pt.lastSyncedAt" class="text-muted">
                  {{ $t('packages.lastSynced', { time: timeAgo(pt.lastSyncedAt) }) }}
                </p>
                <p v-else class="text-muted">{{ $t('packages.notSyncedYet') }}</p>
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
                  {{ $t('packages.sync') }}
                </button>
                <button
                  type="button"
                  class="btn btn-danger"
                  @click="confirmDeletePackagesToken(pt)"
                >
                  <icon-fa6-solid:trash class="icon" aria-hidden="true" />
                  {{ $t('common.delete') }}
                </button>
              </div>
            </div>
          </div>

          <div v-if="packagesComposerUrl" class="composer-setup">
            <h4>{{ $t('packages.composerSetup') }}</h4>
            <Alert type="warning">
              <p>
                {{ $t('packages.composerWarning') }}
              </p>
            </Alert>
            <p class="text-muted">
              {{ $t('packages.composerRepoHint') }}
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
                {{ $t('common.copy') }}
              </button>
            </div>
            <p class="text-muted">{{ $t('packages.composerAuthHint') }}</p>
            <div class="code-block">
              <pre><code>composer config --auth bearer.{{ packagesComposerHost }} &lt;your-token&gt;</code></pre>
            </div>
          </div>
        </template>
      </Panel>

      <Panel :title="$t('project.dangerZone')">
        <p>{{ $t('project.deleteProjectWarning') }}</p>

        <p v-if="!canDeleteProject" class="delete-project-warning">
          {{ $t('project.deleteProjectShopsWarning', { count: shopsInProjectCount }) }}
        </p>

        <button
          type="button"
          class="btn btn-danger"
          :disabled="!canDeleteProject || isDeletingProject"
          @click="showDeleteProjectModal = true"
        >
          <icon-fa6-solid:trash class="icon" aria-hidden="true" />
          {{ $t('project.deleteProject') }}
        </button>
      </Panel>
    </template>

    <!-- Add API Key Modal -->
    <modal :show="showAddKeyModal" close-x-mark @close="closeAddKeyModal">
      <template #title> {{ $t('project.createApiKeyTitle') }} </template>

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
            <label for="apiKeyName">{{ $t('common.name') }}</label>
            <field
              id="apiKeyName"
              type="text"
              name="name"
              :placeholder="$t('project.apiKeyPlaceholder')"
              class="field"
              :class="{ 'has-error': errors.name }"
            />
            <div class="field-error-message">{{ errors.name }}</div>
            <p class="field-help">{{ $t('packages.apiKeyHelp') }}</p>
          </div>

          <div class="form-group">
            <label>{{ $t('project.scopes') }}</label>
            <p class="field-help">{{ $t('project.scopesHelp') }}</p>
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
        <button type="button" class="btn" @click="closeAddKeyModal">{{ $t('common.cancel') }}</button>
        <button
          type="submit"
          class="btn btn-primary"
          form="apiKeyForm"
          :disabled="isCreatingApiKey"
        >
          <icon-fa6-solid:key v-if="!isCreatingApiKey" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t('project.createApiKey') }}
        </button>
      </template>
    </modal>

    <!-- Show Token Modal -->
    <modal :show="showTokenModal" @close="closeTokenModal">
      <template #title> {{ $t('project.apiKeyCreatedTitle') }} </template>

      <template #content>
        <Alert type="warning">
          <p><strong>{{ $t('project.apiKeyCopyWarning') }}</strong></p>
          <p>
            {{ $t('project.apiKeyCopyDesc') }}
          </p>
        </Alert>

        <div class="token-display">
          <code class="token-value">{{ newToken }}</code>
          <button type="button" class="btn btn-secondary" @click="copyToken">
            <icon-fa6-solid:copy class="icon" aria-hidden="true" />
            {{ $t('common.copy') }}
          </button>
        </div>
      </template>

      <template #footer>
        <button type="button" class="btn btn-primary" @click="closeTokenModal">{{ $t('common.done') }}</button>
      </template>
    </modal>

    <!-- Delete API Key Modal -->
    <delete-confirmation-modal
      :show="showDeleteApiKeyModal"
      :title="$t('project.deleteApiKeyTitle')"
      :entity-name="`the API key '${deletingApiKey?.name}'`"
      :custom-consequence="$t('project.deleteApiKeyWarning')"
      :reversed-buttons="true"
      :is-loading="isDeletingApiKey"
      :confirm-button-text="$t('project.deleteApiKeyConfirm')"
      @close="showDeleteApiKeyModal = false"
      @confirm="deleteApiKey"
    />

    <!-- Delete Project Modal -->
    <delete-confirmation-modal
      :show="showDeleteProjectModal"
      :title="$t('project.deleteProjectTitle')"
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
      <template #title> {{ $t('packages.addTokenTitle') }} </template>

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
            <label for="packagesToken">{{ $t('packages.tokenLabel') }}</label>
            <field
              id="packagesToken"
              type="text"
              name="token"
              :placeholder="$t('packages.tokenPlaceholder')"
              class="field"
              :class="{ 'has-error': errors.token }"
            />
            <div class="field-error-message">{{ errors.token }}</div>
            <p class="field-help">
              {{ $t('packages.tokenHelp') }}
            </p>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <button type="button" class="btn" @click="showAddPackagesTokenModal = false">{{ $t('common.cancel') }}</button>
        <button
          type="submit"
          class="btn btn-primary"
          form="packagesTokenForm"
          :disabled="isCreatingPackagesToken"
        >
          <icon-fa6-solid:plus v-if="!isCreatingPackagesToken" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ $t('packages.addToken') }}
        </button>
      </template>
    </modal>

    <!-- Delete Packages Token Modal -->
    <delete-confirmation-modal
      :show="showDeletePackagesTokenModal"
      :title="$t('packages.deleteTokenTitle')"
      :entity-name="`Token #${deletingPackagesToken?.id}`"
      :custom-consequence="$t('packages.deleteTokenWarning')"
      :reversed-buttons="true"
      :is-loading="isDeletingPackagesToken"
      :confirm-button-text="$t('packages.deleteTokenConfirm')"
      @close="showDeletePackagesTokenModal = false"
      @confirm="deletePackagesToken"
    />
  </main-container>
</template>

<script setup lang="ts">
import Modal from "@/components/layout/Modal.vue";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useAlert } from "@/composables/useAlert";
import { formatDate, timeAgo } from "@/helpers/formatter";
import { type RouterOutput, trpcClient } from "@/helpers/trpc";
import { Field, Form as VeeForm } from "vee-validate";
import { computed, nextTick, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

type Project = RouterOutput["account"]["currentUserProjects"][number];
type ApiKey = RouterOutput["organization"]["apiKey"]["list"][number];
type AvailableScope = RouterOutput["organization"]["apiKey"]["scopes"][number];
type PackagesToken = RouterOutput["organization"]["packagesToken"]["list"][number];

const { t } = useI18n();
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
  name: Yup.string().required(t('validation.projectNameRequired')),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url(t('validation.urlInvalid')).optional(),
});

const apiKeySchema = Yup.object().shape({
  name: Yup.string()
    .required(t('validation.nameRequired'))
    .min(1, t('validation.nameRequired'))
    .max(100, t('validation.nameMaxLength')),
  scopes: Yup.array()
    .of(Yup.string())
    .min(1, t('validation.required', { field: 'Scope' }))
    .required(t('validation.required', { field: 'Scope' })),
});

const packagesTokenSchema = Yup.object().shape({
  token: Yup.string().required(t('validation.tokenRequired')).min(1, t('validation.tokenRequired')),
});

async function loadProjectSummary() {
  const [projectsData, shopsData] = await Promise.all([
    trpcClient.account.currentUserProjects.query(),
    trpcClient.account.currentUserShops.query(),
  ]);

  project.value = projectsData.find((currentProject) => currentProject.id === projectId) ?? null;
  shopsInProjectCount.value = shopsData.filter((shop) => shop.projectId === projectId).length;
}

async function loadApiKeys() {
  if (!project.value) return;

  isApiKeysLoading.value = true;
  try {
    apiKeys.value = await trpcClient.organization.apiKey.list.query({
      orgId: project.value.organizationId,
      projectId: project.value.id,
    });
  } catch (error) {
    alert.error(`${t('project.failedLoadApiKeys')}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isApiKeysLoading.value = false;
  }
}

async function loadAvailableScopes() {
  try {
    availableScopes.value = await trpcClient.organization.apiKey.scopes.query();
  } catch (error) {
    alert.error(`${t('project.failedLoadScopes')}${error instanceof Error ? `: ${error.message}` : ""}`);
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
      alert.error(t('project.projectNotFound'));
      return;
    }

    await Promise.all([loadApiKeys(), loadAvailableScopes(), checkPackagesConfigured()]);
    await loadPackagesTokens();
    await scrollToHashSection();
  } catch (error) {
    alert.error(`${t('project.failedLoadProject')}${error instanceof Error ? `: ${error.message}` : ""}`);
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
    await trpcClient.organization.project.update.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      name: typedValues.name,
      description: typedValues.description ?? "",
      gitUrl: typedValues.gitUrl || null,
    });

    await loadProjectSummary();
    alert.success(t('project.projectUpdated'));
  } catch (error) {
    alert.error(`${t('project.failedUpdateProject')}${error instanceof Error ? `: ${error.message}` : ""}`);
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
    alert.success(t('project.apiKeyCreated'));
  } catch (error) {
    alert.error(`${t('project.failedCreateApiKey')}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isCreatingApiKey.value = false;
  }
}

function copyToken() {
  if (!navigator.clipboard) {
    alert.error(t('project.clipboardUnavailable'));
    return;
  }

  navigator.clipboard.writeText(newToken.value);
  alert.success(t('project.apiKeyCopied'));
}

function copyComposerRepository() {
  if (!navigator.clipboard || !packagesComposerUrl.value) return;

  const json = JSON.stringify(
    { repositories: [{ type: "composer", url: packagesComposerUrl.value }] },
    null,
    4,
  );
  navigator.clipboard.writeText(json);
  alert.success(t('packages.composerCopied'));
}

function confirmDeleteKey(apiKey: ApiKey) {
  deletingApiKey.value = apiKey;
  showDeleteApiKeyModal.value = true;
}

async function deleteApiKey() {
  if (!deletingApiKey.value || !project.value) return;

  isDeletingApiKey.value = true;
  try {
    await trpcClient.organization.apiKey.delete.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      apiKeyId: deletingApiKey.value.id,
    });

    showDeleteApiKeyModal.value = false;
    deletingApiKey.value = null;
    await loadApiKeys();
    alert.success(t('project.apiKeyDeleted'));
  } catch (error) {
    alert.error(`${t('project.failedDeleteApiKey')}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isDeletingApiKey.value = false;
  }
}

async function checkPackagesConfigured() {
  try {
    const config = await trpcClient.organization.packagesToken.configuration.query();
    isPackagesConfigured.value = config.configured;
    packagesComposerUrl.value = config.composerUrl;
  } catch {
    isPackagesConfigured.value = false;
  }
}

async function loadPackagesTokens() {
  if (!project.value || !isPackagesConfigured.value) return;

  isPackagesTokensLoading.value = true;
  try {
    packagesTokens.value = await trpcClient.organization.packagesToken.list.query({
      orgId: project.value.organizationId,
      projectId: project.value.id,
    });
  } catch (error) {
    alert.error(
      `${t('packages.failedLoad')}${error instanceof Error ? `: ${error.message}` : ""}`,
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
    await trpcClient.organization.packagesToken.create.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      token: typedValues.token,
    });

    showAddPackagesTokenModal.value = false;
    await loadPackagesTokens();
    alert.success(t('packages.tokenAdded'));
  } catch (error) {
    alert.error(
      `${t('packages.failedAdd')}${error instanceof Error ? `: ${error.message}` : ""}`,
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
    await trpcClient.organization.packagesToken.delete.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      tokenId: deletingPackagesToken.value.id,
    });

    showDeletePackagesTokenModal.value = false;
    deletingPackagesToken.value = null;
    await loadPackagesTokens();
    alert.success(t('packages.tokenDeleted'));
  } catch (error) {
    alert.error(
      `${t('packages.failedDelete')}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isDeletingPackagesToken.value = false;
  }
}

async function syncPackagesToken(pt: PackagesToken) {
  if (!project.value) return;

  isSyncingPackagesToken.value = pt.id;
  try {
    await trpcClient.organization.packagesToken.sync.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
      tokenId: pt.id,
    });

    await loadPackagesTokens();
    alert.success(t('packages.syncTriggered'));
  } catch (error) {
    alert.error(
      `${t('packages.failedSync')}${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isSyncingPackagesToken.value = null;
  }
}

async function deleteProject() {
  if (!project.value) return;

  isDeletingProject.value = true;
  try {
    await trpcClient.organization.project.delete.mutate({
      orgId: project.value.organizationId,
      projectId: project.value.id,
    });

    alert.success(t('project.projectDeleted'));
    router.push({ name: "account.project.list" });
  } catch (error) {
    alert.error(`${t('project.failedDeleteProject')}${error instanceof Error ? `: ${error.message}` : ""}`);
  } finally {
    isDeletingProject.value = false;
    showDeleteProjectModal.value = false;
  }
}

onMounted(() => {
  if (Number.isNaN(projectId)) {
    isPageLoading.value = false;
    alert.error(t('project.invalidProject'));
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
