<template>
  <header-container :title="$t('sso.title')">
    <router-link
      :to="{
        name: 'account.organizations.detail',
        params: { organizationId: route.params.organizationId },
      }"
      type="button"
      class="btn"
    >
      <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
      {{ $t("sso.backToOrg") }}
    </router-link>
  </header-container>

  <main-container>
    <Panel :title="$t('sso.providers')">
      <template #action>
        <button
          v-if="canManageOrganization"
          type="button"
          class="btn btn-primary"
          @click="openAddProviderModal"
        >
          <icon-fa6-solid:plus class="icon" aria-hidden="true" />
          {{ $t("sso.addProvider") }}
        </button>
      </template>

      <Alert type="info">
        <p>
          <strong>{{ $t("sso.infoTitle") }}</strong>
        </p>
        <p>
          {{ $t("sso.infoDesc") }}
        </p>
      </Alert>

      <div v-if="isLoading" class="sso-loading">
        <icon-line-md:loading-twotone-loop class="icon" />
        {{ $t("sso.loadingProviders") }}
      </div>

      <div v-else-if="ssoProviders.length === 0" class="sso-empty">
        <icon-fa6-solid:key class="icon icon-large" aria-hidden="true" />
        <p>{{ $t("sso.noProviders") }}</p>
        <p class="text-muted">
          {{ $t("sso.noProvidersHint") }}
        </p>
      </div>

      <div v-else class="sso-providers">
        <div v-for="provider in ssoProviders" :key="provider.id" class="sso-provider">
          <div class="sso-provider-info">
            <h4>{{ provider.domain }}</h4>

            <p class="text-muted">{{ provider.issuer }}</p>

            <div class="sso-provider-type">
              <span class="badge badge-primary">{{ provider.oidcConfig ? "OIDC" : "SAML" }}</span>
            </div>
          </div>

          <div class="sso-provider-actions">
            <button
              v-if="canManageOrganization"
              type="button"
              class="btn btn-secondary"
              @click="openEditProviderModal(provider)"
            >
              <icon-fa6-solid:pencil class="icon" aria-hidden="true" />
              {{ $t("common.edit") }}
            </button>

            <button
              v-if="canManageOrganization"
              type="button"
              class="btn btn-danger"
              @click="confirmDeleteProvider(provider)"
            >
              <icon-fa6-solid:trash class="icon" aria-hidden="true" />
              {{ $t("common.delete") }}
            </button>
          </div>
        </div>
      </div>
    </Panel>

    <!-- Add Provider Modal -->
    <modal :show="showAddProviderModal" close-x-mark @close="closeProviderModal">
      <template #title>
        {{ isEditMode ? $t("sso.editProvider") : $t("sso.addProviderTitle") }}
      </template>

      <template #content>
        <vee-form
          id="providerForm"
          v-slot="{ errors }"
          :validation-schema="providerSchema"
          :initial-values="currentFormValues"
          class="sso-form"
          @submit="onSubmitProvider"
        >
          <div class="form-group">
            <InputField
              name="domain"
              :label="$t('organization.domain')"
              placeholder="example.com"
              :error="errors.domain"
            />
            <p class="field-help">
              {{ $t("sso.domainHelp") }}
            </p>
          </div>

          <div class="form-group">
            <InputField
              name="issuer"
              :label="$t('sso.issuerUrl')"
              type="url"
              placeholder="https://auth.example.com"
              :error="errors.issuer"
              @update:model-value="issuerUrl = $event"
            >
              <template v-if="!isEditMode" #append>
                <button
                  type="button"
                  class="btn btn-secondary"
                  :disabled="isDiscovering || !issuerUrl"
                  @click="discoverOpenIdConfig"
                >
                  <icon-fa6-solid:magnifying-glass
                    v-if="!isDiscovering"
                    class="icon"
                    aria-hidden="true"
                  />
                  <icon-line-md:loading-twotone-loop v-else class="icon" />
                  {{ $t("sso.discover") }}
                </button>
              </template>
            </InputField>
            <p class="field-help">
              {{ isEditMode ? $t("sso.issuerHelp") : $t("sso.issuerDiscoverHelp") }}
            </p>
          </div>

          <div class="oidc-fields">
            <h4>{{ $t("sso.oidcConfig") }}</h4>

            <div class="form-group">
              <BaseInput
                name="callbackUrl"
                :label="$t('sso.callbackUrl')"
                :model-value="callbackUrl"
                readonly
              />
              <p class="field-help">
                {{ $t("sso.callbackHelp") }}
              </p>
            </div>

            <div class="form-group">
              <InputField name="clientId" :label="$t('sso.clientId')" :error="errors.clientId" />
            </div>

            <div class="form-group">
              <InputField
                name="clientSecret"
                :label="$t('sso.clientSecret')"
                type="password"
                :placeholder="isEditMode ? t('sso.clientSecretKeep') : ''"
                :error="errors.clientSecret"
              />
              <p v-if="isEditMode" class="field-help">
                {{ $t("sso.clientSecretKeep") }}
              </p>
            </div>

            <div class="form-group">
              <InputField
                name="authorizationEndpoint"
                :label="$t('sso.authEndpoint')"
                type="url"
                placeholder="https://auth.example.com/oauth2/authorize"
                :error="errors.authorizationEndpoint"
              />
              <p class="field-help">{{ $t("sso.authEndpointHelp") }}</p>
            </div>

            <div class="form-group">
              <InputField
                name="tokenEndpoint"
                :label="$t('sso.tokenEndpoint')"
                type="url"
                placeholder="https://auth.example.com/oauth2/token"
                :error="errors.tokenEndpoint"
              />
              <p class="field-help">{{ $t("sso.tokenEndpointHelp") }}</p>
            </div>

            <div class="form-group">
              <InputField
                name="jwksEndpoint"
                :label="$t('sso.jwksEndpoint')"
                type="url"
                placeholder="https://auth.example.com/.well-known/jwks.json"
                :error="errors.jwksEndpoint"
              />
              <p class="field-help">{{ $t("sso.jwksEndpointHelp") }}</p>
            </div>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <button type="button" class="btn" @click="closeProviderModal">
          {{ $t("common.cancel") }}
        </button>
        <button type="submit" class="btn btn-primary" form="providerForm" :disabled="isSubmitting">
          <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ isEditMode ? $t("sso.updateProvider") : $t("sso.addProvider") }}
        </button>
      </template>
    </modal>

    <!-- Delete Confirmation Modal -->
    <delete-confirmation-modal
      :show="showDeleteModal"
      :title="$t('sso.deleteProviderTitle')"
      :entity-name="`the SSO provider for ${deletingProvider?.domain}`"
      :custom-consequence="$t('sso.deleteProviderWarning')"
      :reversed-buttons="true"
      :is-loading="isDeleting"
      :confirm-button-text="$t('sso.deleteProviderConfirm')"
      @close="showDeleteModal = false"
      @confirm="deleteProvider"
    />
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { usePermissions } from "@/composables/usePermissions";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const route = useRoute();
const alert = useAlert();

const isLoading = ref(true);
const isSubmitting = ref(false);
const isDeleting = ref(false);
const isDiscovering = ref(false);
const issuerUrl = ref("");
const currentProviderId = ref("");
const editingProvider = ref<SSOProvider | null>(null);
const isEditMode = ref(false);
interface SSOProvider {
  id: string;
  domain: string;
  issuer: string;
  providerId: string;
  oidcConfig?: {
    clientId?: string;
    authorizationEndpoint?: string;
    tokenEndpoint?: string;
    jwksEndpoint?: string;
  } | null;
  samlConfig: string | null;
  userId: string | null;
  organizationId: string | null;
}

const ssoProviders = ref<SSOProvider[]>([]);
const showAddProviderModal = ref(false);
const showDeleteModal = ref(false);
const deletingProvider = ref<SSOProvider | null>(null);
const providerType = ref("oidc");
interface OrganizationData {
  id: string;
  name: string;
}
const organization = ref<OrganizationData | null>(null);

const canManageOrganization = usePermissions(
  computed(() => ({
    organizationId: organization.value?.id,
    permissions: {
      organization: ["update"],
    },
  })),
);

const providerSchema = Yup.object().shape({
  domain: Yup.string()
    .required(t("validation.domainRequired"))
    .matches(
      /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$/,
      t("validation.domainInvalid"),
    ),
  issuer: Yup.string().url(t("validation.urlRequired")).required(t("validation.issuerUrlRequired")),
  clientId: Yup.string().required(t("validation.clientIdRequired")),
  clientSecret: Yup.string().test(
    "required-when-adding",
    t("validation.clientSecretRequired"),
    (value) => {
      if (!isEditMode.value && !value) {
        return false;
      }
      return true;
    },
  ),
  authorizationEndpoint: Yup.string()
    .url(t("validation.urlRequired"))
    .required(t("validation.authEndpointRequired")),
  tokenEndpoint: Yup.string()
    .url(t("validation.urlRequired"))
    .required(t("validation.tokenEndpointRequired")),
  jwksEndpoint: Yup.string()
    .url(t("validation.urlRequired"))
    .required(t("validation.jwksEndpointRequired")),
});

const providerFormValues = computed(() => {
  return {
    domain: "",
    issuer: "",
    clientId: "",
    clientSecret: "",
    authorizationEndpoint: "",
    tokenEndpoint: "",
    jwksEndpoint: "",
  };
});

const currentFormValues = computed(() => {
  if (isEditMode.value && editingProvider.value) {
    const oidcConfig = editingProvider.value.oidcConfig ?? {};

    return {
      domain: editingProvider.value.domain,
      issuer: editingProvider.value.issuer,
      clientId: oidcConfig.clientId ?? "",
      clientSecret: "", // Always empty in edit mode since we don't expose the secret
      authorizationEndpoint: oidcConfig.authorizationEndpoint ?? "",
      tokenEndpoint: oidcConfig.tokenEndpoint ?? "",
      jwksEndpoint: oidcConfig.jwksEndpoint ?? "",
    };
  }
  return providerFormValues.value;
});

const callbackUrl = computed(() => {
  const providerId =
    isEditMode.value && editingProvider.value
      ? editingProvider.value.providerId
      : currentProviderId.value;
  return `${window.location.origin}/auth/sso/callback/${providerId}`;
});

async function loadOrganization() {
  try {
    const { data } = await api.GET("/auth/get-full-organization", {
      params: { query: { organizationId: route.params.organizationId as string } },
    });

    if (!data) {
      alert.error("Failed to load organization");
      return;
    }

    organization.value = data as unknown as OrganizationData;
    await loadProviders();
  } catch (error) {
    alert.error(`Failed to load organization${error instanceof Error ? `: ${error.message}` : ""}`);
  }
}

async function loadProviders() {
  if (!organization.value) return;

  isLoading.value = true;
  try {
    const { data: providers } = await api.GET("/organizations/{orgId}/sso-providers", {
      params: { path: { orgId: organization.value.id } },
    });
    ssoProviders.value = (providers ?? []) as unknown as SSOProvider[];
  } catch (error) {
    alert.error(
      `Failed to load SSO providers${error instanceof Error ? `: ${error.message}` : ""}`,
    );
  } finally {
    isLoading.value = false;
  }
}

function openAddProviderModal() {
  currentProviderId.value = crypto.randomUUID();
  isEditMode.value = false;
  editingProvider.value = null;
  showAddProviderModal.value = true;
}

function openEditProviderModal(provider: SSOProvider) {
  editingProvider.value = provider;
  isEditMode.value = true;
  issuerUrl.value = provider.issuer;
  showAddProviderModal.value = true;
}

function closeProviderModal() {
  showAddProviderModal.value = false;
  providerType.value = "oidc";
  issuerUrl.value = "";
  currentProviderId.value = "";
  editingProvider.value = null;
  isEditMode.value = false;
}

async function discoverOpenIdConfig() {
  if (!issuerUrl.value) {
    alert.error(t("sso.issuerUrlFirst"));
    return;
  }

  isDiscovering.value = true;
  try {
    const { data: config } = await api.GET("/sso/discover", {
      params: { query: { issuer: issuerUrl.value } },
    });

    if (!config) {
      alert.error("Failed to discover OpenID configuration");
      return;
    }

    // Update form fields with discovered values
    const form = document.getElementById("providerForm") as HTMLFormElement;
    if (form) {
      // Get form elements and update their values
      const issuerField = form.querySelector('[name="issuer"]') as HTMLInputElement;
      const authEndpointField = form.querySelector(
        '[name="authorizationEndpoint"]',
      ) as HTMLInputElement;
      const tokenEndpointField = form.querySelector('[name="tokenEndpoint"]') as HTMLInputElement;
      const jwksEndpointField = form.querySelector('[name="jwksEndpoint"]') as HTMLInputElement;

      if (issuerField) issuerField.value = config.issuer;
      if (authEndpointField) authEndpointField.value = config.authorizationEndpoint;
      if (tokenEndpointField) tokenEndpointField.value = config.tokenEndpoint;
      if (jwksEndpointField) jwksEndpointField.value = config.jwksEndpoint;

      // Trigger input events to update vee-validate
      for (const field of [issuerField, authEndpointField, tokenEndpointField, jwksEndpointField]) {
        if (field) {
          field.dispatchEvent(new Event("input", { bubbles: true }));
        }
      }
    }

    alert.success(t("sso.discoveredSuccess"));
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : t("sso.failedDiscover");
    alert.error(errorMessage);
  } finally {
    isDiscovering.value = false;
  }
}

async function onSubmitProvider(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof providerSchema>;

  isSubmitting.value = true;

  try {
    if (isEditMode.value && editingProvider.value) {
      // Update existing provider
      await api.PUT("/organizations/{orgId}/sso-providers/{providerId}", {
        params: {
          path: { orgId: organization.value!.id, providerId: editingProvider.value.providerId },
        },
        body: {
          domain: typedValues.domain,
          issuer: typedValues.issuer,
          clientId: typedValues.clientId,
          clientSecret: typedValues.clientSecret ?? undefined,
          authorizationEndpoint: typedValues.authorizationEndpoint,
          tokenEndpoint: typedValues.tokenEndpoint,
          jwksEndpoint: typedValues.jwksEndpoint,
        },
      });

      alert.success(t("sso.providerUpdated"));
    } else {
      // Create new provider
      const { error: ssoError } = await api.POST("/auth/sso/register", {
        body: {
          domain: typedValues.domain,
          issuer: typedValues.issuer,
          organizationId: organization.value!.id,
          clientId: typedValues.clientId,
          clientSecret: typedValues.clientSecret ?? "",
          authorizationEndpoint: typedValues.authorizationEndpoint,
          tokenEndpoint: typedValues.tokenEndpoint,
          jwksEndpoint: typedValues.jwksEndpoint,
        },
      });

      if (ssoError) {
        alert.error(
          (ssoError as { message?: string }).message ?? "Failed to register SSO provider",
        );
        isSubmitting.value = false;
        return;
      }

      alert.success(t("sso.providerAdded"));
    }

    closeProviderModal();
    await loadProviders();
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : "Failed to save SSO provider";
    alert.error(errorMessage);
  } finally {
    isSubmitting.value = false;
  }
}

function confirmDeleteProvider(provider: SSOProvider) {
  deletingProvider.value = provider;
  showDeleteModal.value = true;
}

async function deleteProvider() {
  if (!deletingProvider.value) return;

  isDeleting.value = true;
  try {
    await api.DELETE("/organizations/{orgId}/sso-providers/{providerId}", {
      params: {
        path: { orgId: organization.value!.id, providerId: deletingProvider.value.providerId },
      },
    });
    alert.success(t("sso.providerDeleted"));
    showDeleteModal.value = false;
    await loadProviders();
  } catch (error: unknown) {
    alert.error(error instanceof Error ? error.message : "Failed to delete SSO provider");
  } finally {
    isDeleting.value = false;
    deletingProvider.value = null;
  }
}

onMounted(() => {
  loadOrganization();
});
</script>

<style scoped>
.sso-loading {
  padding: 3rem;
  text-align: center;
  color: var(--text-color-muted);
}

.sso-empty {
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

.sso-providers {
  padding: 1rem;
}

.sso-provider {
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

.sso-provider-info {
  h4 {
    margin: 0 0 0.25rem 0;
    font-size: 1.125rem;
    font-weight: 500;
  }

  p {
    margin: 0 0 0.5rem 0;
    font-size: 0.875rem;
  }
}

.sso-provider-actions {
  display: flex;
  gap: 0.5rem;
}

.oidc-fields {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--panel-border-color);

  h4 {
    margin: 0 0 1rem 0;
    font-size: 1rem;
    font-weight: 500;
  }
}

.field-help {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--text-color-muted);
}

.sso-form {
  .form-group {
    margin-bottom: 1.5rem;

    &:last-child {
      margin-bottom: 0;
    }
  }
}
</style>
