<template>
  <header-container :title="$t('organization.newOrganization')" />
  <main-container v-if="session?.user">
    <Panel>
      <vee-form
        v-slot="{ errors, isSubmitting }"
        :validation-schema="schema"
        @submit="onCreateOrganization"
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
          <UiButton :disabled="isSubmitting" type="submit" variant="primary">
            <icon-fa6-solid:floppy-disk v-if="!isSubmitting" class="icon" aria-hidden="true" />
            <icon-line-md:loading-twotone-loop v-else class="icon" />
            {{ $t("common.save") }}
          </UiButton>
        </div>
      </vee-form>
    </Panel>
  </main-container>
</template>

<script setup lang="ts">
import { useAlert } from "@/composables/useAlert";
import { useSession, fetchSession } from "@/composables/useSession";
import { api } from "@/helpers/api";

import { Form as VeeForm } from "vee-validate";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import * as Yup from "yup";

const { t } = useI18n();
const { session } = useSession();

const { error } = useAlert();
const router = useRouter();

const schema = Yup.object().shape({
  name: Yup.string().required(t("validation.orgNameRequired")),
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
    await fetchSession();
    await router.push({ name: "account.dashboard" });
  } catch (e) {
    error(e instanceof Error ? e.message : String(e));
  }
}
</script>
