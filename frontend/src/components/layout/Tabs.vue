<template>
  <Panel class="tab-container">
    <tab-group>
      <tab-list class="tabs-list">
        <tab v-for="label in $props.labels" :key="label.key" v-slot="{ selected }" as="template">
          <button class="tab" :class="{ 'tab-active': selected }" type="button">
            <component :is="label.icon" v-if="label.icon" class="icon" />
            {{ label.title }}
            <span v-if="label.count !== undefined" class="pill">
              {{ label.count }}
            </span>
          </button>
        </tab>
      </tab-list>

      <tab-panels class="tab-panels">
        <tab-panel v-for="label in $props.labels" :key="label.key" class="tab-panel">
          <slot :name="`panel-${label.key}`" :label="label">
            <div class="tab-panel-default">
              <strong>{{ label.title }} Tab Panel</strong>. Use
              <pre class="tab-panel-code">
&lt;template #panel-{{ label.key }}="{ label }"&gt;...&lt;/template&gt;</pre
              >
              to fill with content
            </div>
          </slot>
        </tab-panel>
      </tab-panels>
    </tab-group>
  </Panel>
</template>

<script setup lang="ts" generic="T extends string">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from "@headlessui/vue";
import type { FunctionalComponent } from "vue";

defineProps<{
  labels: Array<{
    key: string;
    title: string;
    count?: number;
    icon?: FunctionalComponent;
  }>;
}>();
</script>

<style scoped>
.tab-container {
  width: 100%;
  margin-bottom: 4rem;
  padding: 0;
}

.tabs-list {
  padding: 0.5rem;
  gap: 0.375rem;
  display: grid;
  overflow-x: auto;
  overflow-y: hidden;
  background: var(--recessed-background);
  box-shadow: inset 0 -1px 0 var(--panel-border-color);

  @media (min-width: 640px) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  @media (min-width: 1024px) {
    grid-auto-columns: min-content;
    grid-template-columns: none;
    grid-auto-flow: column;
  }
}

.tab {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  padding: 0.5rem 0.875rem;
  font-size: 0.95rem;
  font-weight: 500;
  position: relative;
  color: var(--text-color-muted);
  border-radius: 0.75rem;
  transition:
    background-color 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
  white-space: nowrap;

  &:hover {
    color: var(--text-color);
    background: var(--button-ghost-hover-background);
  }

  .icon {
    margin-right: 0.125rem;
  }

  .pill {
    margin-left: 0.125rem;
  }

  &:focus-visible {
    box-shadow: inset 0 0 0 1px var(--field-focus-ring-color);
  }
}

.tab-active {
  color: var(--text-color);
  background: var(--panel-background);
  box-shadow:
    inset 0 0 0 1px var(--panel-border-color),
    var(--surface-shadow);
}

.tab-panels {
  margin-top: 0;
}

.tab-panel {
  overflow-y: auto;
}

.tab-panel-default {
  padding: 1.5rem;
}

.tab-panel-code {
  background-color: var(--panel-border-color);
  display: inline-block;
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
}
</style>
