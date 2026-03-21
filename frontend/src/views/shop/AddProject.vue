<template>
  <header-container title="New Project" />
  <main-container>
    <Panel v-if="!isLoadingOrgs">
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
              v-for="organization in organizations"
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
import { api } from "@/helpers/api";
import { Form as VeeForm } from "vee-validate";
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as Yup from "yup";

const { error } = useAlert();
const router = useRouter();
const route = useRoute();

interface Organization {
  id: string;
  name: string;
}

const organizations = ref<Organization[]>([]);
const isLoadingOrgs = ref(true);

async function loadOrganizations() {
  isLoadingOrgs.value = true;
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) {
      organizations.value = data;
    }
  } catch {
    // silently ignore
  } finally {
    isLoadingOrgs.value = false;
  }
}

const schema = Yup.object().shape({
  name: Yup.string().required("Project name is required"),
  description: Yup.string().optional(),
  gitUrl: Yup.string().url("Must be a valid URL").optional(),
  organizationId: Yup.string().required("Organization is required"),
});

const initialValues = computed(() => {
  const firstOrgId = organizations.value[0]?.id ?? "";
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
    const { error: apiError } = await api.POST("/organizations/{orgId}/projects", {
      params: { path: { orgId: typedValues.organizationId } },
      body: {
        name: typedValues.name,
        description: typedValues.description ?? undefined,
        gitUrl: typedValues.gitUrl || undefined,
      },
    });

    if (apiError) {
      error(apiError.message);
      return;
    }

    if (route.query.redirect) {
      router.push(route.query.redirect as string);
    } else {
      router.push({ name: "account.project.list" });
    }
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
};

onMounted(() => {
  loadOrganizations();
});
</script>
