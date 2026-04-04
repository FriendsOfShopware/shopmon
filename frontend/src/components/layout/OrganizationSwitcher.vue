<template>
  <Listbox :model-value="activeOrganizationId" @update:model-value="switchOrganization">
    <div class="org-switcher">
      <ListboxButton class="org-switcher-button">
        <span class="org-switcher-label">{{ currentOrgName }}</span>
        <icon-fa6-solid:chevron-down class="org-switcher-chevron" />
      </ListboxButton>

      <transition
        enter-active-class="transition ease-out duration-100"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-75"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95"
      >
        <ListboxOptions class="org-switcher-options">
          <ListboxOption
            v-for="org in organizations"
            :key="org.id"
            :value="org.id"
            class="org-switcher-option"
            v-slot="{ selected }"
          >
            <span :class="{ 'font-bold': selected }">{{ org.name }}</span>
            <icon-fa6-solid:check v-if="selected" class="org-switcher-check" />
          </ListboxOption>
        </ListboxOptions>
      </transition>
    </div>
  </Listbox>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { useSession, setActiveOrganization } from "@/composables/useSession";
import {
  resetAccountEnvironments,
  fetchAccountEnvironments,
} from "@/composables/useAccountEnvironments";
import { api } from "@/helpers/api";

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

async function switchOrganization(orgId: string) {
  if (orgId === activeOrganizationId.value) return;
  await setActiveOrganization(orgId);
  resetAccountEnvironments();
  fetchAccountEnvironments();
}
</script>

<style>
.org-switcher {
  position: relative;
  margin-left: 1rem;
}

.org-switcher-button {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-color);
  background-color: var(--item-background);
  border: 1px solid var(--panel-border-color);
  cursor: pointer;
  transition: all 0.15s ease-in-out;

  &:hover {
    color: var(--text-color);
    border-color: var(--section-title-border-color);
    background-color: var(--item-hover-background);
  }
}

.org-switcher-label {
  max-width: 12rem;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.org-switcher-chevron {
  width: 0.5rem;
  flex-shrink: 0;
  color: var(--text-color-muted);
}

.org-switcher-options {
  position: absolute;
  z-index: 20;
  min-width: 12rem;
  margin-top: 0.375rem;
  max-height: 15rem;
  overflow-y: auto;
  border-radius: 0.375rem;
  background-color: var(--panel-background);
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.1),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.org-switcher-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  cursor: pointer;
  color: var(--text-color);

  &:hover {
    background-color: var(--primary-color);
    color: #ffffff;
  }
}

.org-switcher-check {
  width: 0.75rem;
  flex-shrink: 0;
}

.font-bold {
  font-weight: 600;
}
</style>
