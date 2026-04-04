<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ $t('sso.title') }}</h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button variant="outline" as-child>
        <RouterLink
          :to="{
            name: 'account.organizations.detail',
            params: { organizationId: route.params.organizationId },
          }"
        >
          <icon-fa6-solid:arrow-left class="size-4" aria-hidden="true" />
          {{ $t("sso.backToOrg") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div class="space-y-6">
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('sso.providers') }}</CardTitle>
        <CardAction>
          <Button
            v-if="canManageOrganization"
            type="button"
            @click="openAddProviderModal"
          >
            <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
            {{ $t("sso.addProvider") }}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Alert class="mb-4">
          <AlertTitle>{{ $t("sso.infoTitle") }}</AlertTitle>
          <AlertDescription>
            {{ $t("sso.infoDesc") }}
          </AlertDescription>
        </Alert>

        <div v-if="isLoading" class="py-12 text-center text-muted-foreground">
          <icon-line-md:loading-twotone-loop class="size-6 mx-auto mb-2" />
          {{ $t("sso.loadingProviders") }}
        </div>

        <div v-else-if="ssoProviders.length === 0" class="py-16 text-center">
          <icon-fa6-solid:key class="size-12 text-muted-foreground mx-auto mb-4" aria-hidden="true" />
          <p>{{ $t("sso.noProviders") }}</p>
          <p class="mt-2 text-muted-foreground text-sm">
            {{ $t("sso.noProvidersHint") }}
          </p>
        </div>

        <div v-else class="space-y-4 pt-4">
          <div
            v-for="provider in ssoProviders"
            :key="provider.id"
            class="flex justify-between items-center p-4 border rounded-lg"
          >
            <div>
              <h4 class="text-lg font-medium mb-1">{{ provider.domain }}</h4>
              <p class="text-sm text-muted-foreground mb-2">{{ provider.issuer }}</p>
              <div class="flex gap-2 mt-2">
                <Badge>{{ provider.oidcConfig ? "OIDC" : "SAML" }}</Badge>
              </div>
            </div>

            <div class="flex gap-2">
              <Button
                v-if="canManageOrganization"
                type="button"
                variant="outline"
                @click="openEditProviderModal(provider)"
              >
                <icon-fa6-solid:pencil class="size-4" aria-hidden="true" />
                {{ $t("common.edit") }}
              </Button>

              <Button
                v-if="canManageOrganization"
                type="button"
                variant="destructive"
                @click="confirmDeleteProvider(provider)"
              >
                <icon-fa6-solid:trash class="size-4" aria-hidden="true" />
                {{ $t("common.delete") }}
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Add Provider Modal -->
    <modal :show="showAddProviderModal" close-x-mark @close="closeProviderModal">
      <template #title>
        {{ isEditMode ? $t("sso.editProvider") : $t("sso.addProviderTitle") }}
      </template>

      <template #content>
        <form id="providerForm" class="space-y-6" @submit="onSubmitProvider">
          <div class="space-y-4">
            <FormField v-slot="{ componentField }" name="domain">
              <FormItem>
                <FormLabel>{{ $t('organization.domain') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" placeholder="example.com" />
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">{{ $t("sso.domainHelp") }}</p>
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="issuer">
              <FormItem>
                <FormLabel>{{ $t('sso.issuerUrl') }}</FormLabel>
                <FormControl>
                  <div class="flex gap-2">
                    <Input
                      v-bind="componentField"
                      type="url"
                      placeholder="https://auth.example.com"
                      class="flex-1"
                      @update:model-value="issuerUrl = $event as string"
                    />
                    <Button
                      v-if="!isEditMode"
                      type="button"
                      variant="outline"
                      :disabled="isDiscovering || !issuerUrl"
                      @click="discoverOpenIdConfig"
                    >
                      <icon-fa6-solid:magnifying-glass
                        v-if="!isDiscovering"
                        class="size-4"
                        aria-hidden="true"
                      />
                      <icon-line-md:loading-twotone-loop v-else class="size-4" />
                      {{ $t("sso.discover") }}
                    </Button>
                  </div>
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">
                  {{ isEditMode ? $t("sso.issuerHelp") : $t("sso.issuerDiscoverHelp") }}
                </p>
              </FormItem>
            </FormField>
          </div>

          <div class="pt-6 border-t space-y-4">
            <h4 class="text-base font-medium">{{ $t("sso.oidcConfig") }}</h4>

            <div>
              <label class="text-sm font-medium">{{ $t('sso.callbackUrl') }}</label>
              <Input
                :model-value="callbackUrl"
                readonly
                class="mt-2"
              />
              <p class="text-xs text-muted-foreground mt-1">{{ $t("sso.callbackHelp") }}</p>
            </div>

            <FormField v-slot="{ componentField }" name="clientId">
              <FormItem>
                <FormLabel>{{ $t('sso.clientId') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="clientSecret">
              <FormItem>
                <FormLabel>{{ $t('sso.clientSecret') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="password"
                    :placeholder="isEditMode ? t('sso.clientSecretKeep') : ''"
                  />
                </FormControl>
                <FormMessage />
                <p v-if="isEditMode" class="text-xs text-muted-foreground">
                  {{ $t("sso.clientSecretKeep") }}
                </p>
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="authorizationEndpoint">
              <FormItem>
                <FormLabel>{{ $t('sso.authEndpoint') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://auth.example.com/oauth2/authorize"
                  />
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">{{ $t("sso.authEndpointHelp") }}</p>
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="tokenEndpoint">
              <FormItem>
                <FormLabel>{{ $t('sso.tokenEndpoint') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://auth.example.com/oauth2/token"
                  />
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">{{ $t("sso.tokenEndpointHelp") }}</p>
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="jwksEndpoint">
              <FormItem>
                <FormLabel>{{ $t('sso.jwksEndpoint') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="url"
                    placeholder="https://auth.example.com/.well-known/jwks.json"
                  />
                </FormControl>
                <FormMessage />
                <p class="text-xs text-muted-foreground">{{ $t("sso.jwksEndpointHelp") }}</p>
              </FormItem>
            </FormField>
          </div>
        </form>
      </template>

      <template #footer>
        <Button variant="outline" type="button" @click="closeProviderModal">
          {{ $t("common.cancel") }}
        </Button>
        <Button
          type="submit"
          form="providerForm"
          :disabled="isSubmitting"
        >
          <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="size-4" aria-hidden="true" />
          <icon-line-md:loading-twotone-loop v-else class="size-4" />
          {{ isEditMode ? $t("sso.updateProvider") : $t("sso.addProvider") }}
        </Button>
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
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { usePermissions } from "@/composables/usePermissions";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { api } from "@/helpers/api";
import type { components } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

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

const providerValidationSchema = toTypedSchema(
  z.object({
    domain: z
      .string()
      .min(1, t("validation.domainRequired"))
      .regex(
        /^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?\.[a-zA-Z]{2,}$/,
        t("validation.domainInvalid"),
      ),
    issuer: z.string().url(t("validation.urlRequired")).min(1, t("validation.issuerUrlRequired")),
    clientId: z.string().min(1, t("validation.clientIdRequired")),
    clientSecret: z.string().optional().refine(
      (value) => {
        if (!isEditMode.value && !value) {
          return false;
        }
        return true;
      },
      { message: t("validation.clientSecretRequired") },
    ),
    authorizationEndpoint: z
      .string()
      .url(t("validation.urlRequired"))
      .min(1, t("validation.authEndpointRequired")),
    tokenEndpoint: z
      .string()
      .url(t("validation.urlRequired"))
      .min(1, t("validation.tokenEndpointRequired")),
    jwksEndpoint: z
      .string()
      .url(t("validation.urlRequired"))
      .min(1, t("validation.jwksEndpointRequired")),
  }),
);

const { handleSubmit: handleProviderSubmit, setValues: setProviderValues, resetForm: resetProviderForm } = useForm({
  validationSchema: providerValidationSchema,
  initialValues: {
    domain: "",
    issuer: "",
    clientId: "",
    clientSecret: "",
    authorizationEndpoint: "",
    tokenEndpoint: "",
    jwksEndpoint: "",
  },
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
  resetProviderForm({
    values: {
      domain: "",
      issuer: "",
      clientId: "",
      clientSecret: "",
      authorizationEndpoint: "",
      tokenEndpoint: "",
      jwksEndpoint: "",
    },
  });
  showAddProviderModal.value = true;
}

function openEditProviderModal(provider: SSOProvider) {
  editingProvider.value = provider;
  isEditMode.value = true;
  issuerUrl.value = provider.issuer;
  const oidcConfig = provider.oidcConfig ?? {};
  setProviderValues({
    domain: provider.domain,
    issuer: provider.issuer,
    clientId: oidcConfig.clientId ?? "",
    clientSecret: "",
    authorizationEndpoint: oidcConfig.authorizationEndpoint ?? "",
    tokenEndpoint: oidcConfig.tokenEndpoint ?? "",
    jwksEndpoint: oidcConfig.jwksEndpoint ?? "",
  });
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

    setProviderValues({
      issuer: config.issuer,
      authorizationEndpoint: config.authorizationEndpoint,
      tokenEndpoint: config.tokenEndpoint,
      jwksEndpoint: config.jwksEndpoint,
    });

    alert.success(t("sso.discoveredSuccess"));
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : t("sso.failedDiscover");
    alert.error(errorMessage);
  } finally {
    isDiscovering.value = false;
  }
}

const onSubmitProvider = handleProviderSubmit(async (values) => {
  isSubmitting.value = true;

  try {
    if (isEditMode.value && editingProvider.value) {
      // Update existing provider
      await api.PUT("/organizations/{orgId}/sso-providers/{providerId}", {
        params: {
          path: { orgId: organization.value!.id, providerId: editingProvider.value.providerId },
        },
        body: {
          domain: values.domain,
          issuer: values.issuer,
          clientId: values.clientId,
          clientSecret: values.clientSecret ?? undefined,
          authorizationEndpoint: values.authorizationEndpoint,
          tokenEndpoint: values.tokenEndpoint,
          jwksEndpoint: values.jwksEndpoint,
        },
      });

      alert.success(t("sso.providerUpdated"));
    } else {
      // Create new provider
      const { error: ssoError } = await api.POST("/auth/sso/register", {
        body: {
          domain: values.domain,
          issuer: values.issuer,
          organizationId: organization.value!.id,
          clientId: values.clientId,
          clientSecret: values.clientSecret ?? "",
          authorizationEndpoint: values.authorizationEndpoint,
          tokenEndpoint: values.tokenEndpoint,
          jwksEndpoint: values.jwksEndpoint,
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
});

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
