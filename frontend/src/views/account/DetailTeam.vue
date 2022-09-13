<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useAuthStore } from '@/stores/auth.store';
import { useTeamStore } from '@/stores/team.store';
import { useRoute } from 'vue-router';

import Header from '@/components/layout/Header.vue';
import MainContainer from '@/components/layout/MainContainer.vue';
import { ref } from 'vue';

const route = useRoute();
const authStore = useAuthStore();
const teamStore = useTeamStore();
const { user } = storeToRefs(authStore);

const teamId = parseInt(route.params.teamId as string, 10);
const team = user.value?.teams.find(team => team.id == teamId);

const showAddMemberModal = ref(false);

teamStore.loadMembers(teamId);
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
      </DataTable>
      </div>
    </div>

    <Modal :show="showAddMemberModal" :closeXMark="true" @close="showAddMemberModal = false">
      <template #title>Add member</template>
      <template #content>
        Member Form here
      </template>
      <template #footer>
        <button type="button" class="btn btn-primary w-full sm:ml-3 sm:w-auto" @click="addMember">
          <span class="-ml-1 mr-2 flex items-center opacity-25 group-hover:opacity-50 ">
              <icon-fa6-solid:plus class="h-5 w-5" aria-hidden="true" />
          </span>
          Add
        </button>
      </template>
    </Modal>
  </MainContainer>
</template>
