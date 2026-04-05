<script setup lang="ts">
import { Input } from "@/components/ui/input";

defineProps<{
  filters: { label: string; value: string }[];
  searchPlaceholder?: string;
}>();

const filter = defineModel<string>("filter", { required: true });
const search = defineModel<string>("search", { required: true });
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3 max-sm:flex-col max-sm:w-full">
    <div class="flex gap-1 rounded-lg border bg-muted/50 p-1">
      <button
        v-for="f in filters"
        :key="f.value"
        :class="[
          'rounded-md px-3 py-1 text-sm font-medium transition-colors',
          filter === f.value
            ? 'bg-background text-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground',
        ]"
        @click="filter = f.value"
      >
        {{ f.label }}
      </button>
    </div>

    <div class="relative">
      <icon-fa6-solid:magnifying-glass
        class="pointer-events-none absolute left-2.5 top-1/2 size-3 -translate-y-1/2 text-muted-foreground"
      />
      <Input
        v-model="search"
        type="search"
        :placeholder="searchPlaceholder ?? 'Search...'"
        class="h-8 w-full sm:w-56 pl-8 text-sm"
      />
    </div>
  </div>
</template>
