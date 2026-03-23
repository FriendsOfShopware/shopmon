<template>
  <header-container v-if="organization" :title="'Edit ' + organization.name">
    <router-link
      :to="{ name: 'account.organizations.detail', params: { organizationId: organization.id } }"
      type="button"
      class="btn"
    >
      Cancel
    </router-link>
  </header-container>

  <main-container v-if="organization">
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="organization"
        @submit="onSaveOrganization"
      >
        <form-group title="Organization Information">
          <div>
            <label for="Name">Name</label>
            <field
              id="name"
              type="text"
              name="name"
              autocomplete="name"
              class="field"
              :class="{ 'has-error': errors.name }"
            />
            <div class="field-error-message">
              {{ errors.name }}
            </div>
          </div>
        </form-group>

        <div class="form-submit">
          <button type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            Save
          </button>
        </div>
      </vee-form>
    </Panel>

    <Panel v-if="canDeleteOrganization" :title="'Deleting organization ' + organization.name">
      <p>Once you delete your organization, you will lose all data associated with it.</p>

      <button type="button" class="btn btn-danger" @click="showOrganizationDeletionModal = true">
        <icon-fa6-solid:trash class="icon" />
        Delete organization
      </button>
    </Panel>

    <delete-confirmation-modal
      :show="showOrganizationDeletionModal"
      title="Delete organization"
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

import { Field, Form as VeeForm } from "vee-validate";
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

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
  name: Yup.string().required("Name of organization is required"),
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
        error((respError as { message?: string }).message ?? "Failed to delete organization");
        return;
      }

      await router.push({ name: "account.organizations.list" });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
}
</script>
