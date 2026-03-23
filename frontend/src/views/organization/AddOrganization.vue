<template>
  <header-container title="New Organization" />
  <main-container v-if="session?.user">
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        @submit="onCreateOrganization"
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
import { useSession } from "@/composables/useSession";
import { api } from "@/helpers/api";

import { Field, Form as VeeForm } from "vee-validate";
import { useRouter } from "vue-router";
import * as Yup from "yup";

const { session } = useSession();

const { error } = useAlert();
const router = useRouter();

const schema = Yup.object().shape({
  name: Yup.string().required("Name for organization is required"),
});

async function onCreateOrganization(values: Record<string, unknown>) {
  const typedValues = values as Yup.InferType<typeof schema>;
  try {
    const { error: respError } = await api.POST("/auth/organizations", {
      body: {
        name: typedValues.name,
      },
    });

    if (respError) {
      error((respError as { message?: string }).message ?? "Failed to create organization");
      return;
    }
    await router.push({ name: "account.organizations.list" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
}
</script>
