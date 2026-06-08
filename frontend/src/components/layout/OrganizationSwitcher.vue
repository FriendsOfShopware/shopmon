<template>
  <Select :model-value="activeOrganizationId ?? undefined" @update:model-value="switchOrganization">
    <SelectTrigger class="h-8 gap-1.5 text-sm font-medium">
      <SelectValue :placeholder="currentOrgName" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="org in organizations" :key="org.id" :value="org.id">
        {{ org.name }}
      </SelectItem>
      <SelectSeparator />
      <button
        class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        @mousedown.prevent="$router.push({ name: 'account.organizations.new' })"
      >
        <icon-fa6-solid:plus class="size-3" />
        New organization
      </button>
    </SelectContent>
  </Select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useSession, setActiveOrganization } from "@/composables/useSession";
import {
  resetAccountEnvironments,
  fetchAccountEnvironments,
} from "@/composables/useAccountEnvironments";
import { api } from "@/helpers/api";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { AcceptableValue } from "reka-ui";

const { activeOrganizationId } = useSession();

interface Organization {
  id: string;
  name: string;
  logo?: string | null;
}

const organizations = ref<Organization[]>([]);

onMounted(async () => {
  const { data } = await api.GET("/auth/list-organizations");
  if (data) {
    organizations.value = data as Organization[];
  }
});

const currentOrgName = computed(() => {
  if (!activeOrganizationId.value) {
    return "...";
  }
  const org = organizations.value.find((o) => o.id === activeOrganizationId.value);
  return org?.name ?? "...";
});

async function switchOrganization(orgId: AcceptableValue) {
  if (typeof orgId !== "string") return;
  if (orgId === activeOrganizationId.value) return;
  await setActiveOrganization(orgId);
  resetAccountEnvironments();
  fetchAccountEnvironments();
}
</script>
