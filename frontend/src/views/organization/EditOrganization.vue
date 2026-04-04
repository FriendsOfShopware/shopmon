<template>
  <header-container
    v-if="organization"
    :title="$t('organization.editTitle', { name: organization.name })"
  >
    <UiButton
      :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
      type="button"
    >
      {{ $t("common.cancel") }}
    </UiButton>
  </header-container>

  <main-container v-if="organization">
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="organization"
        @submit="onSaveOrganization"
      >
        <form-group :title="$t('organization.orgInfo')">
          <InputField
            name="name"
            :label="$t('common.name')"
            autocomplete="name"
            :error="errors.name"
          />
        </form-group>

        <div class="form-submit">
          <UiButton type="submit" variant="primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </UiButton>
        </div>
      </vee-form>
    </Panel>

    <Panel
      v-if="canDeleteOrganization"
      :title="$t('organization.deleteOrgTitle', { name: organization.name })"
    >
      <p>{{ $t("organization.deleteOrgWarning") }}</p>

      <UiButton
        type="button"
        variant="destructive"
        @click="showOrganizationDeletionModal = true"
      >
        <icon-fa6-solid:trash class="icon" />
        {{ $t("organization.deleteOrganization") }}
      </UiButton>
    </Panel>

    <delete-confirmation-modal
      :show="showOrganizationDeletionModal"
      :title="$t('organization.deleteOrganization')"
      :entity-name="organization?.name || 'this organization'"
      @close="showOrganizationDeletionModal = false"
      @confirm="deleteOrganization"
    />
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { api } from "@/helpers/api";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";

import { Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

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

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.orgNameRequired")),
});

async function onSaveOrganization(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  if (organization.value) {
    try {
      const { error: respError } = await api.PATCH("/auth/organizations/{organizationId}", {
        params: { path: { organizationId: organization.value.id } },
        body: {
          name: typedValues.name,
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
}

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
