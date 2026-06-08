<template>
  <Breadcrumb>
    <BreadcrumbList>
      <!-- Mobile: show ellipsis + last 2 items -->
      <template v-if="normalizedItems.length > 2">
        <BreadcrumbItem class="sm:hidden">
          <BreadcrumbEllipsis />
        </BreadcrumbItem>
        <BreadcrumbSeparator class="sm:hidden" />
        <BreadcrumbItem class="sm:hidden">
          <BreadcrumbLink v-if="normalizedItems[normalizedItems.length - 2].to" as-child>
            <RouterLink :to="normalizedItems[normalizedItems.length - 2].to!">
              {{ normalizedItems[normalizedItems.length - 2].label }}
            </RouterLink>
          </BreadcrumbLink>
          <BreadcrumbPage v-else>{{
            normalizedItems[normalizedItems.length - 2].label
          }}</BreadcrumbPage>
        </BreadcrumbItem>
        <BreadcrumbSeparator class="sm:hidden" />
        <BreadcrumbItem class="sm:hidden">
          <BreadcrumbPage>{{ normalizedItems[normalizedItems.length - 1].label }}</BreadcrumbPage>
        </BreadcrumbItem>
      </template>

      <!-- Desktop: show all items -->
      <template v-for="(item, index) in normalizedItems" :key="index">
        <BreadcrumbItem :class="normalizedItems.length > 2 ? 'hidden sm:inline-flex' : ''">
          <BreadcrumbLink v-if="index < normalizedItems.length - 1 && item.to" as-child>
            <RouterLink :to="item.to">
              <component :is="item.icon" v-if="item.icon" class="mr-1 inline size-4" />
              {{ item.label }}
            </RouterLink>
          </BreadcrumbLink>
          <BreadcrumbLink
            v-else-if="index < normalizedItems.length - 1 && item.href"
            :href="item.href"
            target="_blank"
          >
            <component :is="item.icon" v-if="item.icon" class="mr-1 inline size-4" />
            {{ item.label }}
          </BreadcrumbLink>
          <BreadcrumbPage v-else>
            <component :is="item.icon" v-if="item.icon" class="mr-1 inline size-4" />
            {{ item.label }}
          </BreadcrumbPage>
        </BreadcrumbItem>
        <BreadcrumbSeparator
          v-if="index < normalizedItems.length - 1"
          :class="normalizedItems.length > 2 ? 'hidden sm:inline-flex' : ''"
        />
      </template>
    </BreadcrumbList>
  </Breadcrumb>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { BreadcrumbItem as BreadcrumbItemType } from "@/components/layout/breadcrumbs";
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

const props = defineProps<{
  items: BreadcrumbItemType[];
  size?: "sm" | "base";
}>();

const normalizedItems = computed(() => {
  return props.items.filter((item) => item.label.trim().length > 0);
});
</script>
