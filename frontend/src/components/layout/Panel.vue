<template>
  <div class="panel" :class="{ 'panel-table': variant === 'table' }">
    <template v-if="$slots.title">
      <h2 v-if="!$slots.action" class="panel-title">
        <slot name="title" />
      </h2>

      <div v-else class="panel-header">
        <div>
          <h3><slot name="title" /></h3>
          <p v-if="description" class="panel-description">{{ description }}</p>
        </div>
        <slot name="action" />
      </div>
    </template>

    <template v-else-if="title">
      <div v-if="$slots.action" class="panel-header">
        <div>
          <h3>{{ title }}</h3>
          <p v-if="description" class="panel-description">{{ description }}</p>
        </div>
        <slot name="action" />
      </div>

      <h2 v-else class="panel-title">{{ title }}</h2>
    </template>

    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: "default" | "table";
    title?: string;
    description?: string;
  }>(),
  {
    variant: "default",
    title: undefined,
    description: undefined,
  },
);
</script>

<style>
.panel {
  background-color: var(--panel-background);
  padding: 1.25rem;
  border-radius: 0.375rem;
  box-shadow:
    0 1px 3px 0 rgba(0, 0, 0, 0.1),
    0 1px 2px 0 rgba(0, 0, 0, 0.06);
  margin-bottom: 2rem;
}

.dark .panel {
  box-shadow: none;
}

.panel-title {
  font-size: 1.125rem;
  line-height: 1.75rem;
  font-weight: 500;
  padding-bottom: 0.25rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--panel-border-color);

  .icon {
    margin-right: 0.25rem;
  }
}

.panel-header {
  display: flex;
  padding-bottom: 0.75rem;
  margin-bottom: 1rem;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 1px solid var(--panel-border-color);

  h3 {
    font-size: 1.125rem;
    line-height: 1.75rem;
    font-weight: 500;
  }
}

.panel-description {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  color: var(--text-color-muted);
}

.panel-title:not(:first-child),
.panel-header:not(:first-child) {
  margin-top: 1.5rem;
}

.panel-table {
  padding: 0;
  overflow-y: auto;
  overflow-x: hidden;

  @media (min-width: 768px) {
    overflow-y: hidden;
  }
}
</style>
