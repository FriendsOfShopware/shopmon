<template>
  <header-container :title="$t('organization.title')">
    <UiButton :to="{ name: 'account.organizations.new' }" type="button" variant="primary">
      <icon-fa6-solid:plus class="icon" aria-hidden="true" />
      {{ $t("organization.addOrganization") }}
    </UiButton>
  </header-container>

  <main-container>
    <Panel v-if="!organizations || organizations.length === 0">
      <element-empty
        :title="$t('organization.noOrganization')"
        :route="{ name: 'account.organizations.new' }"
        :button="$t('organization.addOrganization')"
      >
        {{ $t("organization.getStarted") }}
      </element-empty>
    </Panel>

    <Panel v-else variant="table">
      <data-table
        :columns="[{ key: 'name', name: $t('common.name'), sortable: true }]"
        :data="organizations"
        class="bg-white dark:bg-neutral-800"
      >
        <template #cell-name="{ row }">
          <router-link
            :to="{ name: 'account.organizations.detail', params: { organizationId: row.id } }"
          >
            {{ row.name }}
          </router-link>
        </template>
      </data-table>
    </Panel>
  </main-container>
</template>

<script setup lang="ts">
import ElementEmpty from "@/components/layout/ElementEmpty.vue";
import { api } from "@/helpers/api";
import { onMounted, ref } from "vue";

interface Organization {
  id: string;
  name: string;
}

const organizations = ref<Organization[]>([]);

async function loadOrganizations() {
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) {
      organizations.value = data;
    }
  } catch {
    // silently ignore
  }
}

onMounted(() => {
  loadOrganizations();
});
</script>
