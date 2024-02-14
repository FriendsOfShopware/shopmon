<script setup lang="ts">
import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import FormGroup from '@/components/layout/FormGroup.vue';

import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import { useOrganizationStore } from '@/stores/organization.store';

import { Field, Form } from 'vee-validate';
import { useRouter } from 'vue-router';
import * as Yup from 'yup';

const authStore = useAuthStore();
const organizationStore = useOrganizationStore();
const alertStore = useAlertStore();
const router = useRouter();

const schema = Yup.object().shape({
  name: Yup.string().required('Name for organization is required')
});

const owner = {
  ownerId: authStore.user?.id
}

async function onCreateOrganization(values: Yup.InferType<typeof schema>) {
  try {
    await organizationStore.createOrganization(values.name);
    await router.push({ name: 'account.organizations.list' });
  } catch (e: any) {
    alertStore.error(e);
  }
}
</script>

<template>
  <Header title="New Organization" />
  <MainContainer v-if="authStore.user">
    <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" :initial-values="owner" @submit="onCreateOrganization">
      <FormGroup title="Organization Information" subTitle="">
        <div class="sm:col-span-6">
          <label for="Name" class="block text-sm font-medium mb-1"> Name </label>
          <Field type="text" name="name" id="name" autocomplete="name" class="field"
            v-bind:class="{ 'is-invalid': errors.name }" />
          <div class="text-red-700">
            {{ errors.name }}
          </div>
        </div>
      </FormGroup>

      <div class="flex justify-end group">
        <button :disabled="isSubmitting" type="submit" class="btn btn-primary">
          <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
            <icon-fa6-solid:floppy-disk class="h-5 w-5" aria-hidden="true" v-if="!isSubmitting" />
            <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
          </span>
          Save
        </button>
      </div>
    </Form>
  </MainContainer>
</template>
