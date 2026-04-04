<template>
  <header class="flex items-start justify-between mb-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ $t('organization.title') }}</h1>
    </div>
    <div class="flex gap-2 items-start">
      <Button as-child>
        <RouterLink :to="{ name: 'account.organizations.new' }">
          <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
          {{ $t("organization.addOrganization") }}
        </RouterLink>
      </Button>
    </div>
  </header>

  <div>
    <Card v-if="!organizations || organizations.length === 0">
      <CardContent class="py-16 flex flex-col items-center gap-6 text-center">
        <svg
          class="size-12 text-muted-foreground"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            vector-effect="non-scaling-stroke"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
          />
        </svg>
        <h2 class="text-2xl font-semibold">{{ $t('organization.noOrganization') }}</h2>
        <p class="text-muted-foreground max-w-lg">{{ $t("organization.getStarted") }}</p>
        <Button as-child>
          <RouterLink :to="{ name: 'account.organizations.new' }">
            <icon-fa6-solid:plus class="size-4" aria-hidden="true" />
            {{ $t("organization.addOrganization") }}
          </RouterLink>
        </Button>
      </CardContent>
    </Card>

    <Card v-else class="p-0 overflow-hidden">
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
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
