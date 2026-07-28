<template>
  <div v-if="organization" class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">
        {{ $t("organization.editTitle", { name: organization.name }) }}
      </h1>
      <Button variant="outline" size="sm" as-child>
        <RouterLink
          :to="{
            name: 'account.organizations.detail',
          }"
        >
          {{ $t("common.cancel") }}
        </RouterLink>
      </Button>
    </div>

    <!-- Organization info -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <icon-fa6-solid:building class="size-4 text-muted-foreground" />
          {{ $t("organization.orgInfo") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit" class="space-y-4">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem>
              <FormLabel>{{ $t("common.name") }}</FormLabel>
              <FormControl>
                <Input v-bind="componentField" autocomplete="name" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <div class="flex justify-end">
            <Button type="submit" :disabled="isSubmitting">
              <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="mr-1.5 size-3.5" />
              <icon-line-md:loading-twotone-loop v-else class="mr-1.5 size-3.5" />
              {{ $t("common.save") }}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Danger zone -->
    <Card v-if="canDeleteOrganization" class="border-destructive/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-base text-destructive">
          <icon-fa6-solid:triangle-exclamation class="size-4" />
          {{ $t("organization.deleteOrgTitle", { name: organization.name }) }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("organization.deleteOrgWarning") }}</p>
        <Button variant="destructive" @click="showOrganizationDeletionModal = true">
          <icon-fa6-solid:trash class="mr-1.5 size-3.5" />
          {{ $t("organization.deleteOrganization") }}
        </Button>
      </CardContent>
    </Card>

    <DeleteConfirmationModal
      :show="showOrganizationDeletionModal"
      :title="$t('organization.deleteOrganization')"
      :entity-name="organization?.name || $t('organization.thisOrganization')"
      @close="showOrganizationDeletionModal = false"
      @confirm="deleteOrganization"
    />
  </div>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession } from "@/composables/useSession";
import {
  deleteOrganization as apiDeleteOrganization,
  getFullOrganization,
  hasPermission,
  updateOrganization,
} from "@/api/generated";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRouter } from "vue-router";

const { t } = useI18n();
const { activeOrganizationId } = useSession();
const { error } = useAlert();
const router = useRouter();

interface OrganizationData {
  id: string;
  name: string;
}

const organization = ref<OrganizationData | null>(null);
const canDeleteOrganization = ref(false);

const validationSchema = toTypedSchema(
  z.object({
    name: z.string().min(1, t("validation.orgNameRequired")),
  }),
);

const { handleSubmit, isSubmitting, setValues } = useForm({ validationSchema });

watch(organization, (org) => {
  if (org) setValues({ name: org.name });
});

async function loadOrganization() {
  try {
    const { data } = await getFullOrganization({
      query: { organizationId: activeOrganizationId.value! },
    });
    if (!data) return;
    organization.value = data as unknown as OrganizationData;

    try {
      const { data: permData } = await hasPermission({
        body: { organizationId: (data as unknown as OrganizationData).id },
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
  if (!organization.value) return;
  try {
    const { error: respError } = await updateOrganization({
      path: { organizationId: organization.value.id },
      body: { name: values.name },
    });
    if (respError) {
      error((respError as { message?: string }).message ?? "Failed to update organization");
      return;
    }
    await router.push({
      name: "account.organizations.detail",
    });
  } catch (err) {
    error(err instanceof Error ? err.message : String(err));
  }
});

async function deleteOrganization() {
  if (!organization.value) return;
  try {
    const { error: respError } = await apiDeleteOrganization({
      path: { organizationId: organization.value.id },
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
</script>
