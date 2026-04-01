<template>
  <header-container
    v-if="organization?.data"
    :title="$t('organization.editTitle', { name: organization.data.name })"
  >
    <router-link
      :to="{ name: 'account.organizations.detail', params: { slug: organization.data.slug } }"
      type="button"
      class="btn"
    >
      {{ $t("common.cancel") }}
    </router-link>
  </header-container>

  <main-container v-if="organization?.data">
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="organization.data"
        @submit="onSaveOrganization"
      >
        <form-group :title="$t('organization.orgInfo')">
          <div>
            <label for="Name">{{ $t("common.name") }}</label>
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

          <div>
            <label for="slug">{{ $t("common.slug") }}</label>
            <field
              id="slug"
              type="text"
              name="slug"
              autocomplete="slug"
              class="field"
              :class="{ 'has-error': errors.slug }"
            />
            <div class="field-error-message">
              {{ errors.slug }}
            </div>
          </div>
        </form-group>

        <div class="form-submit">
          <button type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </button>
        </div>
      </vee-form>
    </Panel>

    <Panel
      v-if="canDeleteOrganization"
      :title="$t('organization.deleteOrgTitle', { name: organization.data.name })"
    >
      <p>{{ $t("organization.deleteOrgWarning") }}</p>

      <button type="button" class="btn btn-danger" @click="showOrganizationDeletionModal = true">
        <icon-fa6-solid:trash class="icon" />
        {{ $t("organization.deleteOrganization") }}
      </button>
    </Panel>

    <delete-confirmation-modal
      :show="showOrganizationDeletionModal"
      :title="$t('organization.deleteOrganization')"
      :entity-name="organization?.data?.name || 'this organization'"
      @close="showOrganizationDeletionModal = false"
      @confirm="deleteOrganization"
    />
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";
import DeleteConfirmationModal from "@/components/modal/DeleteConfirmationModal.vue";

import { Field, Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();
const route = useRoute();

const organization = ref<Awaited<ReturnType<typeof authClient.organization.getFullOrganization>>>();
const canDeleteOrganization = ref<boolean>(false);

authClient.organization
  .getFullOrganization({
    query: { organizationSlug: route.params.slug as string },
  })
  .then((org) => {
    organization.value = org;
    authClient.organization
      .hasPermission({
        organizationId: org.data?.id,
        permissions: {
          organization: ["delete"],
        },
      })
      .then((hasPermission) => {
        canDeleteOrganization.value = hasPermission.data?.success ?? false;
      });
  });

const showOrganizationDeletionModal = ref(false);

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.orgNameRequired")),
  slug: Yup.string()
    .required(t("validation.orgSlugRequired"))
    .matches(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, t("validation.slugFormat")),
});

async function onSaveOrganization(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  if (organization.value) {
    try {
      await authClient.organization.update({
        organizationId: organization.value.data.id,
        data: {
          name: typedValues.name,
          slug: typedValues.slug,
        },
      });

      await router.push({
        name: "account.organizations.detail",
        params: {
          slug: typedValues.slug,
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
      const resp = await authClient.organization.delete({
        organizationId: organization.value.data.id,
      });

      if (resp.error) {
        error(resp.error.message ?? t("organization.failedDeleteOrg"));
        return;
      }

      await router.push({ name: "account.organizations.list" });
    } catch (err) {
      error(err instanceof Error ? err.message : String(err));
    }
  }
}
</script>
