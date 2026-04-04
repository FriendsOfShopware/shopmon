<template>
  <header v-if="organization" class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">
        {{ $t('organization.editTitle', { name: organization.name }) }}
      </h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button variant="outline" as-child>
        <RouterLink
          :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
        >
          {{ $t("common.cancel") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div v-if="organization" class="space-y-6">
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('organization.orgInfo') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="space-y-4">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ $t('common.name') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" autocomplete="name" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <div class="flex justify-end pt-6">
            <Button type="submit" :disabled="isSubmitting">
              <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="size-4" aria-hidden="true" />
              <icon-line-md:loading-twotone-loop v-else class="size-4" />
              {{ $t("common.save") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <Card v-if="canDeleteOrganization">
      <CardHeader>
        <CardTitle>{{ $t('organization.deleteOrgTitle', { name: organization.name }) }}</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-sm text-muted-foreground mb-4">{{ $t("organization.deleteOrgWarning") }}</p>

        <Button
          type="button"
          variant="destructive"
          @click="showOrganizationDeletionModal = true"
        >
          <icon-fa6-solid:trash class="size-4" />
          {{ $t("organization.deleteOrganization") }}
        </Button>
      </CardContent>
    </Card>

    <delete-confirmation-modal
      :show="showOrganizationDeletionModal"
      :title="$t('organization.deleteOrganization')"
      :entity-name="organization?.name || 'this organization'"
      @close="showOrganizationDeletionModal = false"
      @confirm="deleteOrganization"
    />
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();
const route = useRoute();

interface OrganizationData {
  id: string;
  name: string;
}

const organization = ref<OrganizationData | null>(null);
const canDeleteOrganization = ref<boolean>(false);

const validationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.orgNameRequired")),
  }),
);

const { handleSubmit, isSubmitting, setValues } = useForm({
  validationSchema,
});

watch(organization, (org) => {
  if (org) {
    setValues({ name: org.name });
  }
});

async function loadOrganization() {
  try {
    const { data } = await api.GET("/auth/get-full-organization", {
      params: { query: { organizationId: route.params.organizationId as string } },
    });

    if (!data) return;

    organization.value = data as unknown as OrganizationData;

    // Check delete permission
    try {
      const { data: permData } = await api.POST("/auth/has-permission", {
        body: {
          organizationId: (data as unknown as OrganizationData).id,
        },
      });
      canDeleteOrganization.value = permData?.success ?? false;
    } catch {
      // silently ignore
    }
  } catch {
    // silently ignore
  }
}

loadOrganization();

const showOrganizationDeletionModal = ref(false);

const onSubmit = handleSubmit(async (values) => {
  if (organization.value) {
    try {
      const { error: respError } = await api.PATCH("/auth/organizations/{organizationId}", {
        params: { path: { organizationId: organization.value.id } },
        body: {
          name: values.name,
        },
      });

      if (respError) {
        error((respError as { message?: string }).message ?? "Failed to update organization");
        return;
      }

      await router.push({
        name: "account.organizations.detail",
        params: {
          organizationId: organization.value.id,
        },
      });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
});

async function deleteOrganization() {
  if (organization.value) {
    try {
      const { error: respError } = await api.DELETE("/auth/organizations/{organizationId}", {
        params: { path: { organizationId: organization.value.id } },
      });

      if (respError) {
        error((respError as { message?: string }).message ?? t("organization.failedDeleteOrg"));
        return;
      }

      await router.push({ name: "account.organizations.list" });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
}
</script>
