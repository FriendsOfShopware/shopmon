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
  border-radius: 1rem;
  box-shadow:
    inset 0 0 0 1px var(--panel-border-color),
    var(--surface-shadow);
  margin-bottom: 2rem;
}

.panel-title {
  font-size: 1.05rem;
  line-height: 1.5rem;
  font-weight: 600;
  padding-bottom: 0.75rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--panel-border-color);
  color: var(--item-title-color);

  .icon {
    margin-right: 0.375rem;
  }
}

.panel-header {
  display: flex;
  gap: 1rem;
  padding-bottom: 0.9rem;
  margin-bottom: 1rem;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 1px solid var(--panel-border-color);

  h3 {
    font-size: 1.05rem;
    line-height: 1.5rem;
    font-weight: 600;
    color: var(--item-title-color);
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
  overflow: hidden;

  @media (min-width: 768px) {
    overflow-y: hidden;
  }
}

.panel-table > .panel-title {
  padding: 1.25rem 1.25rem 0.9rem;
  margin-bottom: 0;
}

.panel-table > .panel-header {
  padding: 1.25rem 1.25rem 1rem;
  margin-bottom: 0;
}
</style>
