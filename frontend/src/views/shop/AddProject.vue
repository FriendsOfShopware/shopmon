<template>
  <header-container :title="$t('project.newProject')" />
  <main-container>
    <Panel v-if="!organizations.isPending">
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="initialValues"
        @submit="onSubmit"
      >
        <form-group :title="$t('project.projectInfo')">
          <div>
            <label for="name">{{ $t("common.name") }}</label>

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
            <label for="description">{{ $t("common.description") }}</label>

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
            <label for="gitUrl">{{ $t("project.gitRepoUrl") }}</label>

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

          <div>
            <label for="organizationId">{{ $t("settings.organization") }}</label>

            <field id="organizationId" v-slot="{ field }" name="organizationId">
              <select v-bind="field" class="field" :class="{ 'has-error': errors.organizationId }">
                <option value="">{{ $t("project.selectOrganization") }}</option>
                <option
                  v-for="organization in organizations.data"
                  :key="organization.id"
                  :value="organization.id"
                >
                  {{ organization.name }}
                </option>
              </select>
            </field>

            <div class="field-error-message">
              {{ errors.organizationId }}
            </div>
          </div>
        </form-group>

        <div class="form-submit">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </button>
        </div>
      </vee-form>
    </Panel>
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { authClient } from "@/helpers/auth-client";
import { type RouterInput, trpcClient } from "@/helpers/trpc";
import { Field, Form as VeeForm } from "vee-validate";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { error } = useAlert();
const router = useRouter();

const organizations = authClient.useListOrganizations();

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.projectNameRequired")),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url(t("validation.urlInvalid")).optional(),
  organizationId: Yup.string().required(t("validation.orgRequired")),
});

const initialValues = computed(() => {
  const firstOrgId = organizations.value.data?.[0]?.id ?? "";
  return {
    name: "",
    description: "",
    gitUrl: "",
    organizationId: firstOrgId,
  };
});

const onSubmit = async (values: Record<string, unknown>) => {
  try {
    const typedValues = values as Yup.InferType<typeof schema>;
    const input: RouterInput["organization"]["project"]["create"] = {
      orgId: typedValues.organizationId,
      name: typedValues.name,
      description: typedValues.description ?? undefined,
      gitUrl: typedValues.gitUrl || undefined,
    };
    await trpcClient.organization.project.create.mutate(input);

    router.push({ name: "account.project.list" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
};
</script>
