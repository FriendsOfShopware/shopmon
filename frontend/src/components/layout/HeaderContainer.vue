<template>
  <header>
    <div :class="['header-title', { 'mobile-hide': titleMobileHide }]">
      <Breadcrumbs v-if="breadcrumb?.length" class="header-breadcrumb" :items="breadcrumb" />
      <h1>{{ title }}</h1>
    </div>

    <div class="header-actions">
      <slot />
    </div>
  </header>
</template>

<script lang="ts" setup>
import Breadcrumbs from "@/components/layout/Breadcrumbs.vue";
import type { BreadcrumbItem } from "@/components/layout/breadcrumbs";

defineProps<{
  title: string;
  breadcrumb?: BreadcrumbItem[];
  titleMobileHide?: boolean;
}>();
</script>

<style>
header {
  margin: 0 auto 1rem;
  display: flex;
  justify-content: space-between;
  width: 100%;
  grid-area: header;
}

.header-title {
  h1 {
    font-size: 1.875rem;
    line-height: 2.25rem;
    font-weight: 700;
    color: var(--text-color);
  }

  &.mobile-hide {
    display: none;

    + .header-actions {
      margin-left: auto;
    }

    @media all and (min-width: 768px) {
      display: block;
    }
  }
}

.header-breadcrumb {
  margin-bottom: -0.125rem;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;

  .ui-button {
    background: var(--primary-color);
    border-color: transparent;
    color: #ffffff;
    transition: background 0.4s;

    &:hover {
      background: color-mix(in srgb, var(--primary-color) 70%, white);
    }
  }
}
</style>
