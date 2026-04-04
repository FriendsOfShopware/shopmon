<template>
  <div class="form-group">
    <h3 class="form-group-title">{{ title }}</h3>
    <p v-if="subTitle" class="form-group-subtitle">{{ subTitle }}</p>

    <Banner v-if="$slots.info" variant="default">
      <slot name="info" />
    </Banner>

    <div class="content">
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
defineProps<{ title: string; subTitle?: string }>();
</script>

<style scoped>
.form-group-title {
  font-size: 1.125rem;
  line-height: 1.75rem;
  font-weight: 500;
  padding-bottom: 0.25rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--panel-border-color);
}

.form-group + .form-group {
  margin-top: 2rem;
}

.content {
  margin: 1.5rem 0 0;
  background-color: var(--panel-background);

  .form-group-table & {
    padding: 0;
  }

  > :not(:first-child) {
    margin-top: 1rem;
  }

  label {
    display: block;
    margin-bottom: 0.25rem;
    font-weight: 500;
    cursor: pointer;
  }
}

.dark .content {
  box-shadow: none;
}
</style>
