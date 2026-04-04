<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("organization.title") }}</h1>
      <Button size="sm" as-child>
        <RouterLink :to="{ name: 'account.organizations.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("organization.addOrganization") }}
        </RouterLink>
      </Button>
    </div>

    <!-- Empty state -->
    <div
      v-if="loaded && organizations.length === 0"
      class="flex flex-col items-center gap-4 rounded-xl border border-dashed py-20 text-center"
    >
      <div class="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
        <icon-fa6-solid:building class="size-6 text-primary" />
      </div>
      <h2 class="text-xl font-semibold">{{ $t("organization.noOrganization") }}</h2>
      <p class="max-w-md text-muted-foreground">{{ $t("organization.getStarted") }}</p>
      <Button as-child>
        <RouterLink :to="{ name: 'account.organizations.new' }">
          <icon-fa6-solid:plus class="mr-1.5 size-3" />
          {{ $t("organization.addOrganization") }}
        </RouterLink>
      </Button>
    </div>

    <!-- Organization list -->
    <div v-else-if="loaded" class="space-y-2">
      <RouterLink
        v-for="org in organizations"
        :key="org.id"
        :to="{ name: 'account.organizations.detail' }"
        @click="setActiveOrganization(org.id)"
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

onMounted(() => {
  loadOrganizations();
});
</script>
