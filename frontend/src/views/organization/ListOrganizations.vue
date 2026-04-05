<template>
  <div class="space-y-6">
    <PageHeader :title="$t('organization.title')">
      <Button size="sm" as-child>
        <RouterLink :to="{ name: 'account.organizations.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("organization.addOrganization") }}
        </RouterLink>
      </Button>
    </PageHeader>

    <EmptyState
      v-if="loaded && organizations.length === 0"
      :icon="IconBuilding"
      :title="$t('organization.noOrganization')"
      :description="$t('organization.getStarted')"
    >
      <Button as-child>
        <RouterLink :to="{ name: 'account.organizations.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("organization.addOrganization") }}
        </RouterLink>
      </Button>
    </EmptyState>

    <!-- Organization list -->
    <div v-else-if="loaded" class="space-y-2">
      <RouterLink
        v-for="org in organizations"
        :key="org.id"
        :to="{ name: 'account.organizations.detail' }"
        @click="switchToOrg(org.id)"
        class="group flex items-center gap-3 rounded-xl border bg-card p-4 transition-all duration-200 hover:border-primary/30 hover:shadow-sm"
      >
        <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
          <icon-fa6-solid:building class="size-4 text-primary" />
        </div>
        <div class="min-w-0 flex-1">
          <span class="font-medium group-hover:text-primary transition-colors">{{ org.name }}</span>
        </div>
        <icon-fa6-solid:chevron-right
          class="size-3 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
        />
      </RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from "@/components/ui/button";
import PageHeader from "@/components/PageHeader.vue";
import EmptyState from "@/components/EmptyState.vue";
import IconBuilding from "~icons/fa6-solid/building";
import { api } from "@/helpers/api";
import { setActiveOrganization } from "@/composables/useSession";
import { onMounted, ref } from "vue";

interface Organization {
  id: string;
  name: string;
}

const organizations = ref<Organization[]>([]);
const loaded = ref(false);

async function loadOrganizations() {
  try {
    const { data } = await api.GET("/auth/list-organizations");
    if (data) {
      organizations.value = data;
    }
  } catch {
    // silently ignore
  } finally {
    loaded.value = true;
  }
}

async function switchToOrg(orgId: string) {
  await setActiveOrganization(orgId);
}

onMounted(() => {
  loadOrganizations();
});
</script>
