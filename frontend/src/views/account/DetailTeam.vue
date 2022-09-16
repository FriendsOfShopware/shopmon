<script setup lang="ts">
import {storeToRefs} from 'pinia';

import {useAlertStore} from '@/stores/alert.store';
import {useAuthStore} from '@/stores/auth.store';
import {useTeamStore} from '@/stores/team.store';
import {useRoute, useRouter} from 'vue-router';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import {Field, Form} from 'vee-validate';

import {ref} from 'vue';
import * as Yup from "yup";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const teamStore = useTeamStore();
const alertStore = useAlertStore();
const { user } = storeToRefs(authStore);

const teamId = parseInt(route.params.teamId as string, 10);
const team = user.value?.teams.find(team => team.id == teamId);

const showAddMemberModal = ref(false);
const isSubmitting = ref(false);

teamStore.loadMembers(teamId);

const schemaMembers = Yup.object().shape({
  email: Yup.string().email('Email address is not valid').required('Email address is required'),
});

async function onAddMember(values: any) {
  isSubmitting.value = true;
  if ( team ) {
    try {
      await teamStore.addMember(team.id, values);
      showAddMemberModal.value = false;
      await router.push({
        name: 'account.teams.detail',
        params: {
          teamId: team.id
        }
      })
    } catch (error: any) {
      alertStore.error(error);
    }
  }
  isSubmitting.value = false;
}

async function onRemoveMember(userId: number) {
  if ( team ) {
    try {
      await teamStore.removeMember(team.id, userId);
      await router.push({
        name: 'account.teams.detail',
        params: {
          teamId: team.id
        }
      })
    } catch (error: any) {
      alertStore.error(error);
    }
  }
}

</script>

<template>
  <Header v-if="team" :title="team.name">
    <div class="flex gap-2">
      <router-link
        v-if="team.is_owner"
        :to="{ name: 'account.teams.edit', params: { teamId } }"
        type="button" class="group btn btn-primary flex items-center">
        <icon-fa6-solid:pencil class="-ml-1 mr-2 opacity-25 group-hover:opacity-50"
          aria-hidden="true" />
        Edit Team
      </router-link>
    </div>
  </Header>

  <MainContainer>
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8">
        <h3 class="text-lg leading-6 font-medium">
          Team Information
        </h3>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:px-6 lg:px-8">
        <dl class="grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium">Team Name</dt>
            <dd class="mt-1 text-sm text-gray-500">
              {{ team.name }}
            </dd>
          </div>
          
          <div class="sm:col-span-1">
            <dt class="font-medium">Member Count</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ team.memberCount }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="font-medium">Shop Count</dt>
            <dd class="mt-1 text-sm text-gray-500">{{ team.shopCount }}</dd>
          </div>
        </dl>
      </div>
    </div>
    <div class="mb-12 bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="py-5 px-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <h3 class="text-lg leading-6 font-medium">
          Members
        </h3>
        <button class="btn btn-primary" @click="showAddMemberModal = true">
            <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                <icon-fa6-solid:plus class="h-5 w-5" aria-hidden="true" />
            </span>
            Add
        </button>
      </div>
      <div class="border-t border-gray-200">
      <DataTable
          :labels="{email: {name: 'Email'}, username: {name: 'Username'}}"
          :data="teamStore.members">
          <template #cell(username)="{ item, value }">
              <div class="flex justify-between">
                <span>{{ value }}</span>
                <button type="button" class="btn btn-danger w-full sm:ml-3 sm:w-auto" @click="onRemoveMember(item.id)">
                  <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                      <icon-fa6-solid:trash class="h-5 w-5" aria-hidden="true" />
                  </span>
                  Unassign
                </button>
              </div>
          </template>
      </DataTable>
      </div>
    </div>


      <Modal :show="showAddMemberModal" :closeXMark="true" @close="showAddMemberModal = false">
        <template #title>Add member</template>
        <template #content>
          <Form
              v-slot="{ errors, isSubmitting }"
              :validation-schema="schemaMembers"
              @submit="onAddMember"
              id="addMemberForm"
              class="pt-6"
          >
            <label for="email" class="block text-sm font-medium text-gray-700 mb-1"> Email </label>
            <Field type="text" name="email" id="email" autocomplete="email" class="field"
                   v-bind:class="{ 'is-invalid': errors.emailAddress }" />
            <div class="text-red-700">
              {{ errors.email }}
            </div>
          </Form>
        </template>
        <template #footer>
            <button type="reset" class="btn w-full sm:ml-3 sm:w-auto" @click="showAddMemberModal = false" form="addMemberForm">
              <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                  <icon-fa6-solid:xmark class="h-5 w-5" aria-hidden="true" />
              </span>
              Cancel
            </button>
            <button :disabled="isSubmitting" type="submit" class="btn btn-primary w-full sm:ml-3 sm:w-auto" form="addMemberForm">
              <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
                  <icon-fa6-solid:plus class="h-5 w-5" aria-hidden="true"
                                       v-if="!isSubmitting" />
                  <icon-line-md:loading-twotone-loop class="w-5 h-5" v-else />
              </span>
              Add
            </button>
        </template>
      </Modal>


  </MainContainer>
</template>
