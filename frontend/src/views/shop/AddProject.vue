<template>
  <header-container title="New Project" />
  <main-container>
    <Panel v-if="!organizations.isPending">
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        :initial-values="initialValues"
        @submit="onSubmit"
      >
        <form-group title="Project information">
          <InputField name="name" label="Name" :error="errors.name" />

          <TextareaField
            name="description"
            label="Description"
            placeholder="Optional project description..."
            :error="errors.description"
          />

          <InputField
            name="gitUrl"
            label="Git Repository URL"
            type="url"
            placeholder="https://github.com/org/repo"
            :error="errors.gitUrl"
          />

          <SelectField name="organizationId" label="Organization" :error="errors.organizationId">
            <option value="">Select an organization</option>
            <option
              v-for="organization in organizations.data"
              :key="organization.id"
              :value="organization.id"
            >
              {{ organization.name }}
            </option>
          </SelectField>
        </form-group>

        <div class="form-submit">
          <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            Save
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
import { Form as VeeForm } from "vee-validate";
import { computed } from "vue";
import { useRouter } from "vue-router";
import * as Yup from "yup";

const { error } = useAlert();
const router = useRouter();

const organizations = authClient.useListOrganizations();

const schema = Yup.object().shape({
  name: Yup.string().required("Project name is required"),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url("Must be a valid URL").optional(),
  organizationId: Yup.string().required("Organization is required"),
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
