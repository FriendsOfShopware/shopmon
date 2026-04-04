<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("sso.title") }}</h1>
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" as-child>
          <RouterLink
            :to="{
              name: 'account.organizations.detail',
            }"
          >
            <icon-fa6-solid:arrow-left class="mr-1.5 size-3" />
            {{ $t("sso.backToOrg") }}
          </RouterLink>
        </Button>
        <Button v-if="canManageOrganization" size="sm" @click="openAddProviderModal">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("sso.addProvider") }}
        </Button>
      </div>
    </div>

    <Alert>
      <AlertDescription>{{ $t("sso.infoDesc") }}</AlertDescription>
    </Alert>

    <!-- Loading -->
    <div
      v-if="isLoading"
      class="flex items-center justify-center gap-2 py-12 text-muted-foreground"
    >
      <icon-line-md:loading-twotone-loop class="size-5" />
      {{ $t("sso.loadingProviders") }}
    </div>

    <!-- Empty state -->
    <div
      v-else-if="ssoProviders.length === 0"
      class="flex flex-col items-center gap-4 rounded-xl border border-dashed py-20 text-center"
    >
      <div class="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
        <icon-fa6-solid:shield-halved class="size-6 text-primary" />
      </div>
      <h2 class="text-xl font-semibold">{{ $t("sso.noProviders") }}</h2>
      <p class="max-w-md text-sm text-muted-foreground">{{ $t("sso.noProvidersHint") }}</p>
    </div>

    <!-- Provider list -->
    <div v-else class="space-y-2">
      <div
        v-for="provider in ssoProviders"
        :key="provider.id"
        class="flex items-center gap-4 rounded-xl border bg-card px-4 py-3"
      >
        <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
          <icon-fa6-solid:shield-halved class="size-4 text-primary" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="font-medium">{{ provider.domain }}</div>
          <div class="mt-0.5 flex items-center gap-2 text-xs text-muted-foreground">
            <span class="truncate">{{ provider.issuer }}</span>
            <Badge variant="secondary" class="text-[10px]">{{
              provider.oidcConfig ? "OIDC" : "SAML"
            }}</Badge>
          </div>
        </div>
        <div v-if="canManageOrganization" class="flex shrink-0 items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            class="size-7"
            :title="$t('common.edit')"
            @click="openEditProviderModal(provider)"
          >
            <icon-fa6-solid:pencil class="size-3" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="size-7 text-muted-foreground hover:text-destructive"
            :title="$t('common.delete')"
            @click="confirmDeleteProvider(provider)"
          >
            <icon-fa6-solid:trash class="size-3" />
          </Button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Provider Dialog -->
    <Dialog :open="showAddProviderModal" @update:open="(v: boolean) => !v && closeProviderModal()">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>{{
            isEditMode ? $t("sso.editProvider") : $t("sso.addProviderTitle")
          }}</DialogTitle>
        </DialogHeader>

        <form id="providerForm" class="space-y-4" @submit="onSubmitProvider">
          <FormField v-slot="{ componentField }" name="domain">
            <FormItem>
              <FormLabel>{{ $t("organization.domain") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" placeholder="example.com" />
              </FormControl>
              <FormMessage />
              <p class="text-xs text-muted-foreground">{{ $t("sso.domainHelp") }}</p>
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="issuer">
            <FormItem>
              <FormLabel>{{ $t("sso.issuerUrl") }}</FormLabel>
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
                    size="sm"
                    :disabled="isDiscovering || !issuerUrl"
                    @click="discoverOpenIdConfig"
                  >
                    <icon-fa6-solid:magnifying-glass v-if="!isDiscovering" class="mr-1 size-3" />
                    <icon-line-md:loading-twotone-loop v-else class="mr-1 size-3" />
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

          <Separator />

          <h4 class="text-sm font-semibold">{{ $t("sso.oidcConfig") }}</h4>

          <div>
            <label class="text-sm font-medium">{{ $t("sso.callbackUrl") }}</label>
            <Input :model-value="callbackUrl" readonly class="mt-1.5" />
            <p class="mt-1 text-xs text-muted-foreground">{{ $t("sso.callbackHelp") }}</p>
          </div>

          <FormField v-slot="{ componentField }" name="clientId">
            <FormItem>
              <FormLabel>{{ $t("sso.clientId") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="clientSecret">
            <FormItem>
              <FormLabel>{{ $t("sso.clientSecret") }}</FormLabel>
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
              <FormLabel>{{ $t("sso.authEndpoint") }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  type="url"
                  placeholder="https://auth.example.com/oauth2/authorize"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="tokenEndpoint">
            <FormItem>
              <FormLabel>{{ $t("sso.tokenEndpoint") }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  type="url"
                  placeholder="https://auth.example.com/oauth2/token"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="jwksEndpoint">
            <FormItem>
              <FormLabel>{{ $t("sso.jwksEndpoint") }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  type="url"
                  placeholder="https://auth.example.com/.well-known/jwks.json"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>

        <DialogFooter>
          <Button variant="outline" @click="closeProviderModal">{{ $t("common.cancel") }}</Button>
          <Button type="submit" form="providerForm" :disabled="isSubmitting">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="mr-1.5 size-3.5" />
            <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
            {{ isEditMode ? $t("sso.updateProvider") : $t("sso.addProvider") }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation -->
    <DeleteConfirmationModal
      :show="showDeleteModal"
      :title="$t('sso.deleteProviderTitle')"
      :entity-name="`the SSO provider for ${deletingProvider?.domain}`"
      :custom-consequence="$t('sso.deleteProviderWarning')"
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
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { computed, onMounted, ref } from "vue";
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

interface OrganizationData {
  id: string;
  name: string;
}
const organization = ref<OrganizationData | null>(null);

const canManageOrganization = usePermissions(
  computed(() => ({
    organizationId: organization.value?.id,
    permissions: { organization: ["update"] },
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
    clientSecret: z
      .string()
      .optional()
      .refine((v) => isEditMode.value || !!v, { message: t("validation.clientSecretRequired") }),
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

const {
  handleSubmit: handleProviderSubmit,
  setValues: setProviderValues,
  resetForm: resetProviderForm,
} = useForm({
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
    const { data } = await api.GET("/auth/get-full-organization");
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
    const { data: providers } = await api.GET("/organization/sso-providers");
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
  const oidc = provider.oidcConfig ?? {};
  setProviderValues({
    domain: provider.domain,
    issuer: provider.issuer,
    clientId: oidc.clientId ?? "",
    clientSecret: "",
    authorizationEndpoint: oidc.authorizationEndpoint ?? "",
    tokenEndpoint: oidc.tokenEndpoint ?? "",
    jwksEndpoint: oidc.jwksEndpoint ?? "",
  });
  showAddProviderModal.value = true;
}

function closeProviderModal() {
  showAddProviderModal.value = false;
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
    alert.error(error instanceof Error ? error.message : t("sso.failedDiscover"));
  } finally {
    isDiscovering.value = false;
  }
}

const onSubmitProvider = handleProviderSubmit(async (values) => {
  isSubmitting.value = true;
  try {
    if (isEditMode.value && editingProvider.value) {
      await api.PUT("/organization/sso-providers/{providerId}", {
        params: {
          path: { providerId: editingProvider.value.providerId },
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
    alert.error(error instanceof Error ? error.message : "Failed to save SSO provider");
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
    await api.DELETE("/organization/sso-providers/{providerId}", {
      params: {
        path: { providerId: deletingProvider.value.providerId },
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
