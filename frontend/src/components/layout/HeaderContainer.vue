<template>
  <header class="mb-6 flex items-start justify-between">
    <div :class="{ 'hidden md:block': titleMobileHide }">
      <Breadcrumbs v-if="breadcrumb?.length" class="-mb-0.5" :items="breadcrumb" />
      <h1 class="text-3xl font-bold tracking-tight">{{ title }}</h1>
    </div>

    <div class="flex items-start gap-2" :class="{ 'ml-auto': titleMobileHide }">
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
