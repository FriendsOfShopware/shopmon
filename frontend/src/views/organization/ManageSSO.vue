<template>
  <header-container title="SSO Configuration">
    <router-link
      :to="{
        name: 'account.organizations.detail',
        params: { organizationId: route.params.organizationId },
      }"
      type="button"
      class="btn"
    >
      <icon-fa6-solid:arrow-left class="icon" aria-hidden="true" />
      Back to Organization
    </router-link>
  </header-container>

  <main-container>
    <Panel title="SSO Providers">
      <template #action>
        <button
          v-if="canManageOrganization"
          type="button"
          class="btn btn-primary"
          @click="openAddProviderModal"
        >
          <icon-fa6-solid:plus class="icon" aria-hidden="true" />
          Add Provider
        </button>
      </template>

      <Alert type="info">
        <p><strong>Single Sign-On Information</strong></p>
        <p>
          Users who sign in through SSO will automatically become members of this organization.
          Configure your identity provider to allow users from your domain to access Shopmon.
        </p>
      </Alert>

      <div v-if="isLoading" class="sso-loading">
        <icon-line-md:loading-twotone-loop class="icon" />
        Loading providers...
      </div>

      <div v-else-if="ssoProviders.length === 0" class="sso-empty">
        <icon-fa6-solid:key class="icon icon-large" aria-hidden="true" />
        <p>No SSO providers configured yet.</p>
        <p class="text-muted">
          Add an SSO provider to enable single sign-on for your organization.
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
              Edit
            </button>

            <button
              v-if="canManageOrganization"
              type="button"
              class="btn btn-danger"
              @click="confirmDeleteProvider(provider)"
            >
              <icon-fa6-solid:trash class="icon" aria-hidden="true" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </Panel>

    <!-- Add Provider Modal -->
    <modal :show="showAddProviderModal" close-x-mark @close="closeProviderModal">
      <template #title>
        {{ isEditMode ? "Edit SSO Provider" : "Add SSO Provider" }}
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
              label="Domain"
              placeholder="example.com"
              :error="errors.domain"
            />
            <p class="field-help">
              Users with email addresses from this domain will use this SSO provider
            </p>
          </div>

          <div class="form-group">
            <InputField
              name="issuer"
              label="Issuer URL"
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
                  Discover
                </button>
              </template>
            </InputField>
            <p class="field-help">
              {{
                isEditMode
                  ? "The issuer URL from your identity provider"
                  : 'Enter the issuer URL and click "Discover" to auto-fill endpoints'
              }}
            </p>
          </div>

          <div class="oidc-fields">
            <h4>OIDC Configuration</h4>

            <div class="form-group">
              <BaseInput
                name="callbackUrl"
                label="Callback URL"
                :model-value="callbackUrl"
                readonly
              />
              <p class="field-help">
                Use this URL as the redirect/callback URL in your identity provider configuration
              </p>
            </div>

            <div class="form-group">
              <InputField name="clientId" label="Client ID" :error="errors.clientId" />
            </div>

            <div class="form-group">
              <InputField
                name="clientSecret"
                label="Client Secret"
                type="password"
                :placeholder="isEditMode ? 'Leave empty to keep existing secret' : ''"
                :error="errors.clientSecret"
              />
              <p v-if="isEditMode" class="field-help">
                Leave empty to keep the existing client secret
              </p>
            </div>

            <div class="form-group">
              <InputField
                name="authorizationEndpoint"
                label="Authorization Endpoint"
                type="url"
                placeholder="https://auth.example.com/oauth2/authorize"
                :error="errors.authorizationEndpoint"
              />
              <p class="field-help">OAuth2 authorization endpoint URL</p>
            </div>

            <div class="form-group">
              <InputField
                name="tokenEndpoint"
                label="Token Endpoint"
                type="url"
                placeholder="https://auth.example.com/oauth2/token"
                :error="errors.tokenEndpoint"
              />
              <p class="field-help">OAuth2 token endpoint URL</p>
            </div>

            <div class="form-group">
              <InputField
                name="jwksEndpoint"
                label="JWKS Endpoint"
                type="url"
                placeholder="https://auth.example.com/.well-known/jwks.json"
                :error="errors.jwksEndpoint"
              />
              <p class="field-help">JSON Web Key Set (JWKS) endpoint URL</p>
            </div>
          </div>
        </vee-form>
      </template>

      <template #footer>
        <button type="button" class="btn" @click="closeProviderModal">Cancel</button>
        <button type="submit" class="btn btn-primary" form="providerForm" :disabled="isSubmitting">
          <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="icon" />
          {{ isEditMode ? "Update Provider" : "Add Provider" }}
        </button>
      </template>
    </modal>

    <!-- Delete Confirmation Modal -->
    <delete-confirmation-modal
      :show="showDeleteModal"
      title="Delete SSO Provider?"
      :entity-name="`the SSO provider for ${deletingProvider?.domain}`"
      custom-consequence="Users from this domain will no longer be able to use SSO to sign in."
      :reversed-buttons="true"
      :is-loading="isDeleting"
      confirm-button-text="Delete Provider"
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
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import * as Yup from "yup";

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
    .required("Domain is required")
    .matches(
      /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$/,
      "Please enter a valid domain",
    ),
  issuer: Yup.string().url("Must be a valid URL").required("Issuer URL is required"),
  clientId: Yup.string().required("Client ID is required"),
  clientSecret: Yup.string().test("required-when-adding", "Client Secret is required", (value) => {
    if (!isEditMode.value && !value) {
      return false;
    }
    return true;
  }),
  authorizationEndpoint: Yup.string()
    .url("Must be a valid URL")
    .required("Authorization Endpoint is required"),
  tokenEndpoint: Yup.string().url("Must be a valid URL").required("Token Endpoint is required"),
  jwksEndpoint: Yup.string().url("Must be a valid URL").required("JWKS Endpoint is required"),
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
    alert.error("Please enter an issuer URL first");
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

    alert.success("OpenID configuration discovered and form fields updated");
  } catch (error: unknown) {
    const errorMessage =
      error instanceof Error ? error.message : "Failed to discover OpenID configuration";
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

      alert.success("SSO provider updated successfully");
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

      alert.success("SSO provider added successfully");
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
    alert.success("SSO provider deleted successfully");
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
